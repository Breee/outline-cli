package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"charm.land/glamour/v2"
	"github.com/spf13/cobra"
)

var (
	catDocID string
	catRaw   bool
)

func init() {
	catCmd := &cobra.Command{
		Use:   "cat <title>",
		Short: "Print a document to stdout",
		Long: `Retrieve a document from Outline and print it to stdout.
Renders markdown by default; use --raw for the original markdown source.`,
		Example: `  # Print rendered document
  outline cat "Deployment Guide"

  # Print raw markdown (pipe-friendly)
  outline cat "Deployment Guide" --raw | less

  # Fetch by document ID
  outline cat --id abc123`,
		RunE: runCat,
	}

	catCmd.Flags().StringVar(&catDocID, "id", "", "Document ID (instead of title)")
	catCmd.Flags().BoolVar(&catRaw, "raw", false, "Print raw markdown without rendering")
	rootCmd.AddCommand(catCmd)
}

func runCat(cmd *cobra.Command, args []string) error {
	if catDocID == "" && len(args) == 0 {
		return fmt.Errorf("provide a document title or --id")
	}

	client, err := newOutlineClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var docText, docTitle string

	if catDocID != "" {
		doc, err := client.GetDocument(ctx, catDocID)
		if err != nil {
			return fmt.Errorf("fetching document: %w", err)
		}
		docText = doc.Text
		docTitle = doc.Title
	} else {
		title := strings.Join(args, " ")
		results, err := client.Search(ctx, title, "", 5)
		if err != nil {
			return fmt.Errorf("searching: %w", err)
		}
		if len(results) == 0 {
			return fmt.Errorf("no document found matching %q", title)
		}
		// Use the first exact-title match, or fall back to first result.
		var bestDoc *struct {
			id    string
			title string
		}
		for _, r := range results {
			if strings.EqualFold(r.Document.Title, title) {
				bestDoc = &struct {
					id    string
					title string
				}{r.Document.ID, r.Document.Title}
				break
			}
		}
		if bestDoc == nil {
			bestDoc = &struct {
				id    string
				title string
			}{results[0].Document.ID, results[0].Document.Title}
		}
		doc, err := client.GetDocument(ctx, bestDoc.id)
		if err != nil {
			return fmt.Errorf("fetching document: %w", err)
		}
		docText = doc.Text
		docTitle = doc.Title
	}

	if catRaw {
		fmt.Println(docText)
		return nil
	}

	// Render markdown.
	r, err := glamour.NewTermRenderer(
		glamour.WithWordWrap(80),
	)
	if err != nil {
		// Fall back to raw output.
		fmt.Println(docText)
		return nil
	}

	// Prepend title as H1 if not already present.
	content := docText
	if !strings.HasPrefix(strings.TrimSpace(content), "# ") {
		content = "# " + docTitle + "\n\n" + content
	}

	rendered, err := r.Render(content)
	if err != nil {
		fmt.Println(docText)
		return nil
	}
	fmt.Print(rendered)
	return nil
}
