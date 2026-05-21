package cmd

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/Breee/outline-cli/internal/outline"
	"github.com/spf13/cobra"
)

var (
	searchCollection string
	searchLimit      int
	searchFormat     string
)

func init() {
	searchCmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search documents in Outline",
		Example: `  # Basic search
  outline search "kubernetes rollback"

  # Filter by collection
  outline search "deploy" --collection infrastructure

  # JSON output for scripting
  outline search "API key" --format json

  # Compact output
  outline search "auth" --format oneline

  # Limit results
  outline search "setup" --limit 5`,
		Args: cobra.MinimumNArgs(1),
		RunE: runSearch,
	}

	searchCmd.Flags().StringVar(&searchCollection, "collection", "", "Filter by collection (name, slug, or UUID)")
	searchCmd.Flags().IntVar(&searchLimit, "limit", 25, "Maximum number of results")
	searchCmd.Flags().StringVar(&searchFormat, "format", "default", "Output format: default, json, oneline")
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	client, err := newOutlineClient()
	if err != nil {
		return err
	}

	query := strings.Join(args, " ")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var collectionID string
	if searchCollection != "" {
		collectionID, err = client.ResolveCollectionID(ctx, searchCollection)
		if err != nil {
			return err
		}
	}

	results, err := client.Search(ctx, query, collectionID, searchLimit)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		cmd.PrintErrln("No results found.")
		return nil
	}

	switch searchFormat {
	case "json":
		return printSearchJSON(cmd, results)
	case "oneline":
		return printSearchOneline(cmd, results)
	default:
		return printSearchDefault(cmd, results)
	}
}

func printSearchDefault(cmd *cobra.Command, results []outline.SearchResult) error {
	for i, r := range results {
		doc := r.Document
		snippet := strings.TrimSpace(r.Context)
		if snippet == "" {
			// Fall back to first 120 chars of document text.
			snippet = doc.Text
			if len(snippet) > 120 {
				snippet = snippet[:120] + "..."
			}
		}
		// Clean up snippet whitespace.
		snippet = strings.ReplaceAll(snippet, "\n", " ")
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}

		cmd.Printf("%2d. %s\n", i+1, doc.Title)
		if snippet != "" {
			cmd.Printf("    %s\n", snippet)
		}
		if doc.URL != "" {
			cmd.Printf("    URL: %s\n", doc.URL)
		}
		cmd.Println()
	}
	return nil
}

func printSearchJSON(cmd *cobra.Command, results []outline.SearchResult) error {
	type jsonResult struct {
		ID         string `json:"id"`
		Title      string `json:"title"`
		Collection string `json:"collectionId"`
		Snippet    string `json:"snippet"`
		URL        string `json:"url"`
		UpdatedAt  string `json:"updatedAt"`
	}

	out := make([]jsonResult, 0, len(results))
	for _, r := range results {
		out = append(out, jsonResult{
			ID:         r.Document.ID,
			Title:      r.Document.Title,
			Collection: r.Document.CollectionID,
			Snippet:    r.Context,
			URL:        r.Document.URL,
			UpdatedAt:  r.Document.UpdatedAt,
		})
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	cmd.Println(string(data))
	return nil
}

func printSearchOneline(cmd *cobra.Command, results []outline.SearchResult) error {
	for _, r := range results {
		doc := r.Document
		url := doc.URL
		if url == "" {
			url = doc.ID
		}
		cmd.Printf("%s\t%s\n", doc.Title, url)
	}
	return nil
}
