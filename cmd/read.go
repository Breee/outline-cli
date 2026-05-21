package cmd

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/Breee/outline-cli/internal/tui"
	"github.com/spf13/cobra"
)

var readCollection string

func init() {
	readCmd := &cobra.Command{
		Use:   "read [query]",
		Short: "Interactive TUI for browsing and reading wiki documents",
		Long: `Launch an interactive terminal UI for browsing collections and reading
documents from Outline. Optionally pass a search query to jump directly
to search results.`,
		Example: `  # Open collection browser
  outline read

  # Jump to search results
  outline read "deploy guide"

  # Browse a specific collection
  outline read --collection ops`,
		RunE: runRead,
	}

	readCmd.Flags().StringVar(&readCollection, "collection", "", "Browse a specific collection")
	rootCmd.AddCommand(readCmd)
}

func runRead(cmd *cobra.Command, args []string) error {
	client, err := newOutlineClient()
	if err != nil {
		return err
	}

	query := ""
	if len(args) > 0 {
		query = args[0]
	}

	m := tui.New(client, serverURL, query)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		return err
	}
	return nil
}
