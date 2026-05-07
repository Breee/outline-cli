package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	pushPath       string
	pushCollection string
	pushPublish    bool
)

func init() {
	pushCmd := &cobra.Command{
		Use:   "push",
		Short: "Push markdown files to Outline",
		RunE:  runPush,
	}

	pushCmd.Flags().StringVarP(&pushPath, "path", "p", ".", "Path to markdown file or directory")
	pushCmd.Flags().StringVar(&pushCollection, "collection-id", "", "Outline collection ID")
	pushCmd.Flags().BoolVar(&pushPublish, "publish", true, "Publish created documents")
	_ = pushCmd.MarkFlagRequired("collection-id")
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

	ctx := context.Background()
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		title := documentTitle(file, string(content))
		doc, err := client.CreateDocument(ctx, pushCollection, title, string(content), pushPublish)
		if err != nil {
			return fmt.Errorf("push %s: %w", file, err)
		}
		cmd.Printf("pushed %s -> %s\n", file, doc.ID)
	}

	return nil
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
