package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Breee/outline-cli/internal/config"
	"github.com/Breee/outline-cli/internal/oidc"
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
	cobra.OnInitialize(loadFromConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", config.DefaultPath(), "config file")
	rootCmd.PersistentFlags().StringVar(&serverURL, "server-url", envOrDefault("OUTLINE_SERVER_URL", "OUTLINE_HOST"), "Outline server URL")
	rootCmd.PersistentFlags().StringVar(&apiToken, "api-token", os.Getenv("OUTLINE_API_TOKEN"), "Outline API token")
	rootCmd.PersistentFlags().StringVar(&basicUser, "username", os.Getenv("OUTLINE_USERNAME"), "Basic auth username")
	rootCmd.PersistentFlags().StringVar(&basicPassword, "password", os.Getenv("OUTLINE_PASSWORD"), "Basic auth password")
	rootCmd.PersistentFlags().StringVar(&oidcAccessToken, "oidc-access-token", os.Getenv("OUTLINE_OIDC_ACCESS_TOKEN"), "OIDC access token")
}

// envOrDefault returns the value of the first non-empty environment variable.
func envOrDefault(keys ...string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}

// loadFromConfig populates flags from the config file when not set via CLI/env.
func loadFromConfig() {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return
	}
	if serverURL == "" && cfg.ServerURL != "" {
		serverURL = cfg.ServerURL
	}
	if apiToken == "" {
		apiToken = config.LoadAPIToken(cfg)
	}
	if basicPassword == "" {
		basicPassword = config.LoadPassword(cfg)
	}
	if oidcAccessToken == "" {
		oidcAccessToken = config.LoadToken(cfg)
	}
}

// isAuthenticated checks whether the current credentials are valid.
func isAuthenticated() bool {
	if apiToken == "" && oidcAccessToken == "" && basicUser == "" {
		return false
	}
	client, err := newOutlineClientRaw()
	if err != nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = client.GetAuthInfo(ctx)
	return err == nil
}

// autoAuth performs automatic authentication using the configured method.
// Returns nil if auth succeeds or no auto-auth is configured.
func autoAuth(cmd *cobra.Command) error {
	cfg, _ := config.Load(cfgFile)
	method := os.Getenv("OUTLINE_AUTH_METHOD")
	if method == "" {
		method = cfg.AuthMethod
	}
	if method == "" {
		method = "oidc" // default
	}

	switch method {
	case "oidc":
		if serverURL == "" {
			return fmt.Errorf("auto-auth requires --server-url or server_url in config")
		}
		port := 10800
		if cfg.OIDCPort > 0 {
			port = cfg.OIDCPort
		}
		cmd.PrintErrln("Token expired or missing. Re-authenticating via OIDC...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		session, err := oidc.StartBrowserLogin(ctx, serverURL, port)
		if err != nil {
			return fmt.Errorf("auto-auth: %w", err)
		}
		if openErr := openBrowser(session.BrowserURL); openErr != nil {
			cmd.PrintErrf("Open this URL in your browser:\n  %s\n", session.BrowserURL)
		} else {
			cmd.PrintErrln("Browser opened. Waiting for authentication...")
		}
		result, err := session.Wait(ctx)
		if err != nil {
			return fmt.Errorf("auto-auth: %w", err)
		}
		oidcAccessToken = result.AccessToken
		if serverURL != "" {
			cfg.ServerURL = serverURL
		}
		if err := config.SaveToken(cfgFile, &cfg, result.AccessToken); err != nil {
			return fmt.Errorf("auto-auth: saving token: %w", err)
		}
		cmd.PrintErrln("Re-authenticated successfully.")
		return nil
	case "api-token":
		return fmt.Errorf("api-token auth requires --api-token flag or OUTLINE_API_TOKEN env var")
	case "basic":
		return fmt.Errorf("basic auth requires --username and --password flags")
	default:
		return fmt.Errorf("unknown auth_method %q (use oidc, api-token, or basic)", method)
	}
}

// newOutlineClientRaw creates a client without auto-auth.
func newOutlineClientRaw() (*outline.Client, error) {
	auth, err := outline.NewAuthConfig(apiToken, basicUser, basicPassword, oidcAccessToken)
	if err != nil {
		return nil, err
	}
	return outline.NewClient(serverURL, auth, http.DefaultClient)
}

// newOutlineClient creates a client, auto-authenticating if needed.
func newOutlineClient() (*outline.Client, error) {
	if !isAuthenticated() {
		if err := autoAuth(rootCmd); err != nil {
			return nil, err
		}
	}
	return newOutlineClientRaw()
}
