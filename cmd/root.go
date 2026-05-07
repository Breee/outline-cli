package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Breee/outline-cli/internal/config"
	"github.com/Breee/outline-cli/internal/outline"
	"github.com/spf13/cobra"
)

var (
	cfgFile         string
	serverURL       string
	apiToken        string
	basicUser       string
	basicPassword   string
	oidcAccessToken string
)

var rootCmd = &cobra.Command{
	Use:   "outline",
	Short: "Push markdown documents to Outline",
	Long:  "A modern CLI for pushing markdown files to an Outline wiki.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(loadAuthFromConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", config.DefaultPath(), "config file")
	rootCmd.PersistentFlags().StringVar(&serverURL, "server-url", os.Getenv("OUTLINE_SERVER_URL"), "Outline server URL")
	rootCmd.PersistentFlags().StringVar(&apiToken, "api-token", os.Getenv("OUTLINE_API_TOKEN"), "Outline API token")
	rootCmd.PersistentFlags().StringVar(&basicUser, "username", os.Getenv("OUTLINE_USERNAME"), "Basic auth username")
	rootCmd.PersistentFlags().StringVar(&basicPassword, "password", os.Getenv("OUTLINE_PASSWORD"), "Basic auth password")
	rootCmd.PersistentFlags().StringVar(&oidcAccessToken, "oidc-access-token", os.Getenv("OUTLINE_OIDC_ACCESS_TOKEN"), "OIDC access token")
}

func loadAuthFromConfig() {
	if oidcAccessToken != "" {
		return
	}
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return
	}
	oidcAccessToken = cfg.OIDCAccessToken
}

func newOutlineClient() (*outline.Client, error) {
	auth, err := outline.NewAuthConfig(apiToken, basicUser, basicPassword, oidcAccessToken)
	if err != nil {
		return nil, err
	}
	return outline.NewClient(serverURL, auth, http.DefaultClient)
}
