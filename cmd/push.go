package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Breee/outline-cli/internal/outline"
	"github.com/spf13/cobra"
)

var (
	pushPath             string
	pushCollection       string
	pushPublish          bool
	pushCreateCollection bool
)

func init() {
	pushCmd := &cobra.Command{
		Use:   "push",
		Short: "Push markdown files to Outline",
		RunE:  runPush,
	}

	pushCmd.Flags().StringVarP(&pushPath, "path", "p", ".", "Path to markdown file or directory")
	pushCmd.Flags().StringVar(&pushCollection, "collection-id", "", "Default Outline collection (name, slug, or UUID)")
	pushCmd.Flags().BoolVar(&pushPublish, "publish", true, "Publish created documents")
	pushCmd.Flags().BoolVar(&pushCreateCollection, "create-collection", false, "Create collection if it does not exist")
	rootCmd.AddCommand(pushCmd)
}

func runPush(cmd *cobra.Command, _ []string) error {
	client, err := newOutlineClient()
	if err != nil {
		return err
	}

	files, err := markdownFiles(pushPath)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no markdown files found under %s", pushPath)
	}

	// Sort: by depth first, index/README files before others at same depth, then alpha.
	sort.Slice(files, func(i, j int) bool {
		di := strings.Count(files[i], string(filepath.Separator))
		dj := strings.Count(files[j], string(filepath.Separator))
		if di != dj {
			return di < dj
		}
		ii := isIndexFile(files[i])
		ij := isIndexFile(files[j])
		if ii != ij {
			return ii
		}
		return files[i] < files[j]
	})

	ctx := context.Background()

	// Cache resolved collection IDs to avoid repeated API calls.
	collectionCache := map[string]string{}
	// Track pushed docs: title -> ID and dir -> ID (for directory-based parenting).
	docIDByTitle := map[string]string{}
	docIDByDir := map[string]string{} // key: collectionID + ":" + relDir

	// Determine the push root for relative path computation.
	pushRoot, _ := filepath.Abs(pushPath)
	if info, _ := os.Stat(pushRoot); info != nil && !info.IsDir() {
		pushRoot = filepath.Dir(pushRoot)
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		meta := parseMetadata(string(content))
		body := stripMetadata(string(content))

		// Determine collection: per-file header > CLI flag.
		collection := meta.Collection
		if collection == "" {
			collection = pushCollection
		}
		if collection == "" {
			return fmt.Errorf("push %s: no collection specified (use --collection-id flag or <!-- Collection: ... --> header)", file)
		}

		collectionID, ok := collectionCache[collection]
		if !ok {
			collectionID, err = client.ResolveCollectionID(ctx, collection)
			if err != nil {
				if !pushCreateCollection {
					return fmt.Errorf("push %s: %w", file, err)
				}
				// Create the collection.
				collectionID, err = client.CreateCollection(ctx, collection)
				if err != nil {
					return fmt.Errorf("push %s: creating collection: %w", file, err)
				}
				cmd.Printf("created collection %q\n", collection)
			}
			collectionCache[collection] = collectionID
		}

		// Determine title: per-file header > first H1 > filename.
		title := meta.Title
		if title == "" {
			title = documentTitle(file, body)
		}

		// Determine parent document ID.
		// Priority: <!-- Parent: title --> header > directory structure.
		parentDocID, err := resolveParentDocID(ctx, client, meta.Parent, file, pushRoot, collectionID, docIDByTitle, docIDByDir)
		if err != nil {
			return fmt.Errorf("push %s: resolving parent: %w", file, err)
		}

		// Upload local images and rewrite references.
		body, err = uploadImages(ctx, client, file, body)
		if err != nil {
			return fmt.Errorf("push %s: uploading images: %w", file, err)
		}

		// Search for existing document with matching title in the collection.
		existing, err := client.SearchDocuments(ctx, collectionID, title)
		if err != nil {
			return fmt.Errorf("push %s: searching documents: %w", file, err)
		}

		var doc outline.Document
		var found bool
		for _, d := range existing {
			if d.Title == title {
				doc = d
				found = true
				break
			}
		}

		opts := outline.DocumentOptions{
			Icon:             meta.Icon,
			ParentDocumentID: parentDocID,
		}

		if found {
			doc, err = client.UpdateDocument(ctx, doc.ID, title, body, opts)
			if err != nil {
				return fmt.Errorf("push %s: %w", file, err)
			}
			cmd.Printf("updated %s -> %s\n", file, doc.ID)
		} else {
			doc, err = client.CreateDocument(ctx, collectionID, title, body, pushPublish, opts)
			if err != nil {
				return fmt.Errorf("push %s: %w", file, err)
			}
			cmd.Printf("created %s -> %s\n", file, doc.ID)
		}

		// Record this doc for later parent resolution.
		docIDByTitle[collectionID+":"+title] = doc.ID

		// If this is an index/README file, register it as the directory's doc
		// so sibling and sub-directory files can find it as their parent.
		absFile, _ := filepath.Abs(file)
		if isIndexFile(absFile) {
			rel, _ := filepath.Rel(pushRoot, filepath.Dir(absFile))
			if rel == "." {
				rel = ""
			}
			docIDByDir[collectionID+":"+rel] = doc.ID
		}
	}

	return nil
}

// resolveParentDocID determines the parent document ID for a file.
// Priority: explicit <!-- Parent: title --> header > directory structure.
//
// Directory logic:
//   - Non-index files become children of the index.md/README.md in the same directory.
//   - Index files in subdirectories become children of the index.md in the parent directory.
//   - If no local index was pushed, searches Outline by the explicit parent title.
func resolveParentDocID(ctx context.Context, client *outline.Client, explicitParent, file, pushRoot, collectionID string, docIDByTitle, docIDByDir map[string]string) (string, error) {
	if explicitParent != "" {
		// Check local cache first.
		if id, ok := docIDByTitle[collectionID+":"+explicitParent]; ok {
			return id, nil
		}
		// Search Outline for existing doc with this title.
		docs, err := client.SearchDocuments(ctx, collectionID, explicitParent)
		if err != nil {
			return "", fmt.Errorf("searching for parent %q: %w", explicitParent, err)
		}
		for _, d := range docs {
			if d.Title == explicitParent {
				docIDByTitle[collectionID+":"+explicitParent] = d.ID
				return d.ID, nil
			}
		}
		return "", fmt.Errorf("parent document %q not found in collection", explicitParent)
	}

	// Directory-based resolution.
	absFile, _ := filepath.Abs(file)
	rel, err := filepath.Rel(pushRoot, absFile)
	if err != nil {
		return "", nil
	}
	dir := filepath.Dir(rel)
	if dir == "." {
		dir = ""
	}

	if isIndexFile(absFile) {
		// Index file: parent is the index doc of the parent directory.
		if dir == "" {
			return "", nil // root-level index, no parent
		}
		parentDir := filepath.Dir(dir)
		if parentDir == "." {
			parentDir = ""
		}
		if id, ok := docIDByDir[collectionID+":"+parentDir]; ok {
			return id, nil
		}
	} else {
		// Regular file: parent is the index doc in the same directory.
		if id, ok := docIDByDir[collectionID+":"+dir]; ok {
			return id, nil
		}
	}

	return "", nil
}

// isIndexFile returns true if the file is an index or README file.
func isIndexFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	return base == "index.md" || base == "readme.md"
}

// docMetadata holds metadata parsed from HTML comment headers in markdown files.
type docMetadata struct {
	Collection string
	Title      string
	Parent     string
	Icon       string
}

var metaPattern = regexp.MustCompile(`(?m)^<!--\s*(Collection|Title|Parent|Icon):\s*(.+?)\s*-->`)

// parseMetadata extracts metadata from HTML comment headers at the top of the file.
// Only lines before the first non-comment, non-blank line are considered.
func parseMetadata(content string) docMetadata {
	var meta docMetadata
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		match := metaPattern.FindStringSubmatch(line)
		if match == nil {
			break // stop at first non-metadata line
		}
		key, value := match[1], match[2]
		switch key {
		case "Collection":
			meta.Collection = value
		case "Title":
			meta.Title = value
		case "Parent":
			meta.Parent = value
		case "Icon":
			meta.Icon = value
		}
	}
	return meta
}

// stripMetadata removes metadata comment lines from the top of the content.
func stripMetadata(content string) string {
	lines := strings.Split(content, "\n")
	start := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if metaPattern.MatchString(line) {
			start = i + 1
			continue
		}
		break
	}
	// Also skip any leading blank lines after metadata.
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	return strings.Join(lines[start:], "\n")
}

func markdownFiles(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		if strings.HasSuffix(strings.ToLower(path), ".md") {
			return []string{path}, nil
		}
		return nil, nil
	}

	var files []string
	err = filepath.WalkDir(path, func(p string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			files = append(files, p)
		}
		return nil
	})
	return files, err
}

func documentTitle(path, content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	name := filepath.Base(path)
	return strings.TrimSuffix(name, filepath.Ext(name))
}

// imagePattern matches markdown image references: ![alt](path)
var imagePattern = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)

// uploadImages finds local image references in markdown, uploads them,
// and rewrites the references to use the Outline attachment URL.
func uploadImages(ctx context.Context, client *outline.Client, mdFile, body string) (string, error) {
	baseDir := filepath.Dir(mdFile)

	return replaceAllStringSubmatchFunc(imagePattern, body, func(match []string) (string, error) {
		imgPath := match[2]

		// Skip URLs (http/https) and already-absolute paths.
		if strings.HasPrefix(imgPath, "http://") || strings.HasPrefix(imgPath, "https://") || strings.HasPrefix(imgPath, "/api/") {
			return match[0], nil
		}

		// Resolve relative to the markdown file.
		absPath := imgPath
		if !filepath.IsAbs(imgPath) {
			absPath = filepath.Join(baseDir, imgPath)
		}

		// Check file exists.
		if _, err := os.Stat(absPath); err != nil {
			// Not a local file, leave reference as-is.
			return match[0], nil
		}

		// Upload and get the outline URL.
		attachmentURL, err := client.UploadAttachment(ctx, absPath)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("![%s](%s)", match[1], attachmentURL), nil
	})
}

// replaceAllStringSubmatchFunc is like ReplaceAllStringFunc but provides submatches.
func replaceAllStringSubmatchFunc(re *regexp.Regexp, str string, repl func([]string) (string, error)) (string, error) {
	var result strings.Builder
	lastIndex := 0
	for _, match := range re.FindAllStringSubmatchIndex(str, -1) {
		groups := make([]string, len(match)/2)
		for i := range groups {
			if match[i*2] >= 0 {
				groups[i] = str[match[i*2]:match[i*2+1]]
			}
		}
		replacement, err := repl(groups)
		if err != nil {
			return "", err
		}
		result.WriteString(str[lastIndex:match[0]])
		result.WriteString(replacement)
		lastIndex = match[1]
	}
	result.WriteString(str[lastIndex:])
	return result.String(), nil
}
