package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Breee/outline-cli/internal/outline"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	getOutput     string
	getCollection string
)

func init() {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get resources from Outline (kubectl-style)",
		Long:  "Retrieve and display resources from Outline in various output formats.",
	}

	collectionsCmd := &cobra.Command{
		Use:     "collections",
		Aliases: []string{"col", "cols"},
		Short:   "List all collections",
		Example: `  # List collections (default table)
  outline get collections

  # JSON output
  outline get collections -o json

  # YAML output
  outline get collections -o yaml`,
		RunE: runGetCollections,
	}

	documentsCmd := &cobra.Command{
		Use:     "documents",
		Aliases: []string{"docs", "doc"},
		Short:   "List or get documents",
		Example: `  # List documents in a collection
  outline get documents --collection test

  # Get a specific document by title (markdown output)
  outline get documents "Deployment Guide" -o md

  # Get by title, JSON output
  outline get documents "API Reference" -o json

  # YAML metadata
  outline get documents "FAQ" -o yaml`,
		RunE: runGetDocuments,
	}

	collectionsCmd.Flags().StringVarP(&getOutput, "output", "o", "table", "Output format: table, json, yaml")
	documentsCmd.Flags().StringVarP(&getOutput, "output", "o", "table", "Output format: table, json, yaml, md, raw")
	documentsCmd.Flags().StringVar(&getCollection, "collection", "", "Filter by collection (name, slug, or UUID)")

	getCmd.AddCommand(collectionsCmd)
	getCmd.AddCommand(documentsCmd)
	rootCmd.AddCommand(getCmd)
}

func runGetCollections(cmd *cobra.Command, args []string) error {
	client, err := newOutlineClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collections, err := client.ListCollections(ctx)
	if err != nil {
		return err
	}

	switch getOutput {
	case "json":
		return printJSON(cmd, collections)
	case "yaml", "yml":
		return printYAML(cmd, collections)
	default:
		return printCollectionsTable(cmd, collections)
	}
}

func runGetDocuments(cmd *cobra.Command, args []string) error {
	client, err := newOutlineClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// If a positional arg is given, fetch a specific document by title/id.
	if len(args) > 0 {
		return getSpecificDocument(cmd, client, ctx, strings.Join(args, " "))
	}

	// List documents, optionally filtered by collection.
	var collectionID string
	if getCollection != "" {
		collectionID, err = client.ResolveCollectionID(ctx, getCollection)
		if err != nil {
			return err
		}
	}

	docs, err := client.ListAllDocuments(ctx, collectionID)
	if err != nil {
		return err
	}

	// Build collection name map for display.
	collections, _ := client.ListCollections(ctx)
	colMap := make(map[string]string, len(collections))
	for _, col := range collections {
		colMap[col.ID] = col.Name
	}

	switch getOutput {
	case "json":
		return printJSON(cmd, docs)
	case "yaml", "yml":
		return printYAML(cmd, docs)
	default:
		return printDocumentsTable(cmd, docs, colMap)
	}
}

func getSpecificDocument(cmd *cobra.Command, client *outline.Client, ctx context.Context, query string) error {
	// Try search by title.
	results, err := client.Search(ctx, query, "", 5)
	if err != nil {
		return fmt.Errorf("searching: %w", err)
	}
	if len(results) == 0 {
		return fmt.Errorf("no document found matching %q", query)
	}

	// Find best match (exact title match preferred).
	var docID string
	for _, r := range results {
		if strings.EqualFold(r.Document.Title, query) {
			docID = r.Document.ID
			break
		}
	}
	if docID == "" {
		docID = results[0].Document.ID
	}

	doc, err := client.GetDocument(ctx, docID)
	if err != nil {
		return fmt.Errorf("fetching document: %w", err)
	}

	switch getOutput {
	case "json":
		return printJSON(cmd, doc)
	case "yaml", "yml":
		return printDocYAML(cmd, doc)
	case "md", "markdown", "raw":
		cmd.Println(doc.Text)
		return nil
	default:
		// Table-like single doc view.
		cmd.Printf("TITLE:      %s\n", doc.Title)
		cmd.Printf("ID:         %s\n", doc.ID)
		cmd.Printf("COLLECTION: %s\n", doc.CollectionID)
		if doc.ParentDocumentID != "" {
			cmd.Printf("PARENT:     %s\n", doc.ParentDocumentID)
		}
		if doc.UpdatedAt != "" {
			cmd.Printf("UPDATED:    %s\n", doc.UpdatedAt)
		}
		if doc.URL != "" {
			cmd.Printf("URL:        %s\n", doc.URL)
		}
		cmd.Println()
		cmd.Println(doc.Text)
		return nil
	}
}

// --- Output formatters ---

func printJSON(cmd *cobra.Command, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	cmd.Println(string(data))
	return nil
}

func printYAML(cmd *cobra.Command, v any) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	cmd.Print(string(data))
	return nil
}

func printDocYAML(cmd *cobra.Command, doc outline.Document) error {
	out := map[string]any{
		"id":           doc.ID,
		"title":        doc.Title,
		"collectionId": doc.CollectionID,
		"updatedAt":    doc.UpdatedAt,
		"url":          doc.URL,
		"text":         doc.Text,
	}
	if doc.ParentDocumentID != "" {
		out["parentDocumentId"] = doc.ParentDocumentID
	}
	return printYAML(cmd, out)
}

func printCollectionsTable(cmd *cobra.Command, collections []outline.Collection) error {
	cmd.Printf("%-36s  %-30s  %s\n", "ID", "NAME", "URL")
	for _, c := range collections {
		fullURL := strings.TrimRight(serverURL, "/") + c.URL
		cmd.Printf("%-36s  %-30s  %s\n", c.ID, c.Name, fullURL)
	}
	return nil
}

func printDocumentsTable(cmd *cobra.Command, docs []outline.Document, colMap map[string]string) error {
	cmd.Printf("%-36s  %-20s  %-40s  %s\n", "ID", "COLLECTION", "TITLE", "UPDATED")
	for _, d := range docs {
		updated := d.UpdatedAt
		if len(updated) > 10 {
			updated = updated[:10]
		}
		title := d.Title
		if len(title) > 40 {
			title = title[:37] + "..."
		}
		col := colMap[d.CollectionID]
		if len(col) > 20 {
			col = col[:17] + "..."
		}
		cmd.Printf("%-36s  %-20s  %-40s  %s\n", d.ID, col, title, updated)
	}
	return nil
}
