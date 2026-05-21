package cmd

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// VersionInfo holds structured version metadata.
type VersionInfo struct {
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit"`
	BuildDate string `json:"buildDate"`
	GoVersion string `json:"goVersion"`
	Compiler  string `json:"compiler"`
	Platform  string `json:"platform"`
}

var (
	versionStr string
	commitStr  string
	dateStr    string
	versionOut string
)

// SetVersionInfo sets the version info from linker flags.
func SetVersionInfo(version, commit, date string) {
	versionStr = version
	commitStr = commit
	dateStr = date
}

func init() {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Example: `  outline version
  outline version -o json`,
		RunE: runVersion,
	}
	versionCmd.Flags().StringVarP(&versionOut, "output", "o", "", "Output format: json")
	rootCmd.AddCommand(versionCmd)
}

func runVersion(cmd *cobra.Command, _ []string) error {
	info := VersionInfo{
		Version:   versionStr,
		GitCommit: commitStr,
		BuildDate: dateStr,
		GoVersion: runtime.Version(),
		Compiler:  runtime.Compiler,
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}

	if versionOut == "json" {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(info)
	}

	fmt.Fprintln(cmd.OutOrStdout(), info.Version)
	return nil
}
