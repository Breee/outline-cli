package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Breee/outline-cli/internal/outline"
	"github.com/spf13/cobra"
)

var (
	pullCollection   string
	pullOutput       string
	pullDoc          string
	pullWithMetadata bool
)

func init() {
	pullCmd := &cobra.Command{
		Use:   "pull",
		Short: "Pull documents from Outline to local markdown files",
		Example: `  # Pull entire collection to a directory
  outline pull --collection "Engineering" --output ./docs/

  # Pull a single document by title
  outline pull --doc "Deployment Guide" --output ./deploy.md

  # Pull with metadata headers
  outline pull --collection "Ops" --output ./ops/ --with-metadata`,
		RunE: runPull,
	}

	pullCmd.Flags().StringVar(&pullCollection, "collection", "", "Collection to pull (name, slug, or UUID)")
	pullCmd.Flags().StringVarP(&pullOutput, "output", "o", ".", "Output directory or file path")
	pullCmd.Flags().StringVar(&pullDoc, "doc", "", "Pull a single document by title")
	pullCmd.Flags().BoolVar(&pullWithMetadata, "with-metadata", false, "Include metadata comment headers in output")
	rootCmd.AddCommand(pullCmd)
}

func runPull(cmd *cobra.Command, _ []string) error {
	if pullCollection == "" && pullDoc == "" {
		return fmt.Errorf("specify --collection or --doc")
	}

	client, err := newOutlineClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if pullDoc != "" {
		return pullSingleDoc(ctx, cmd, client)
	}
	return pullCollectionDocs(ctx, cmd, client)
}

func pullSingleDoc(ctx context.Context, cmd *cobra.Command, client *outline.Client) error {
	var collectionID string
	var err error
	if pullCollection != "" {
		collectionID, err = client.ResolveCollectionID(ctx, pullCollection)
		if err != nil {
			return err
		}
	}

	// Search for the document by title.
	results, err := client.Search(ctx, pullDoc, collectionID, 10)
	if err != nil {
		return err
	}

	var doc outline.Document
	var found bool
	for _, r := range results {
		if r.Document.Title == pullDoc {
			doc = r.Document
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("document %q not found", pullDoc)
	}

	// Get full document content.
	doc, err = client.GetDocument(ctx, doc.ID)
	if err != nil {
		return err
	}

	content := formatPulledDoc(doc, pullCollection)
	return writeDocToFile(cmd, pullOutput, doc.Title, content)
}

func pullCollectionDocs(ctx context.Context, cmd *cobra.Command, client *outline.Client) error {
	collectionID, err := client.ResolveCollectionID(ctx, pullCollection)
	if err != nil {
		return err
	}

	docs, err := client.ListDocuments(ctx, collectionID)
	if err != nil {
		return err
	}

	if len(docs) == 0 {
		cmd.PrintErrln("Collection is empty.")
		return nil
	}

	// Build parent→children map for directory structure.
	byID := make(map[string]outline.Document, len(docs))
	children := make(map[string][]outline.Document)
	var roots []outline.Document

	for _, doc := range docs {
		byID[doc.ID] = doc
		if doc.ParentDocumentID == "" {
			roots = append(roots, doc)
		} else {
			children[doc.ParentDocumentID] = append(children[doc.ParentDocumentID], doc)
		}
	}

	// Create output directory.
	if err := os.MkdirAll(pullOutput, 0o755); err != nil {
		return err
	}

	// Write documents recursively preserving hierarchy.
	var writeTree func(docs []outline.Document, dir string) error
	writeTree = func(docs []outline.Document, dir string) error {
		for _, doc := range docs {
			kids := children[doc.ID]
			content := formatPulledDoc(doc, pullCollection)

			if len(kids) > 0 {
				// This doc has children → create subdirectory with index.md.
				subdir := filepath.Join(dir, sanitizeFilename(doc.Title))
				if err := os.MkdirAll(subdir, 0o755); err != nil {
					return err
				}
				outPath := filepath.Join(subdir, "index.md")
				if err := os.WriteFile(outPath, []byte(content), 0o644); err != nil {
					return err
				}
				cmd.Printf("pulled %s\n", outPath)
				if err := writeTree(kids, subdir); err != nil {
					return err
				}
			} else {
				// Leaf document → write as file.
				outPath := filepath.Join(dir, sanitizeFilename(doc.Title)+".md")
				if err := os.WriteFile(outPath, []byte(content), 0o644); err != nil {
					return err
				}
				cmd.Printf("pulled %s\n", outPath)
			}
		}
		return nil
	}

	return writeTree(roots, pullOutput)
}

func formatPulledDoc(doc outline.Document, collection string) string {
	var sb strings.Builder
	if pullWithMetadata {
		if collection != "" {
			sb.WriteString(fmt.Sprintf("<!-- Collection: %s -->\n", collection))
		}
		sb.WriteString(fmt.Sprintf("<!-- Title: %s -->\n", doc.Title))
		sb.WriteString("\n")
	}
	sb.WriteString(doc.Text)
	if !strings.HasSuffix(doc.Text, "\n") {
		sb.WriteString("\n")
	}
	return sb.String()
}

func writeDocToFile(cmd *cobra.Command, output, title, content string) error {
	// If output looks like a directory (ends with / or exists as dir), use title as filename.
	info, err := os.Stat(output)
	if (err == nil && info.IsDir()) || strings.HasSuffix(output, "/") {
		if err := os.MkdirAll(output, 0o755); err != nil {
			return err
		}
		output = filepath.Join(output, sanitizeFilename(title)+".md")
	}

	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(output, []byte(content), 0o644); err != nil {
		return err
	}
	cmd.Printf("pulled %s\n", output)
	return nil
}

func sanitizeFilename(title string) string {
	// Replace path separators and problematic characters.
	r := strings.NewReplacer("/", "-", "\\", "-", ":", "-", "*", "", "?", "", "\"", "", "<", "", ">", "", "|", "")
	name := r.Replace(title)
	name = strings.TrimSpace(name)
	if name == "" {
		name = "untitled"
	}
	return name
}
