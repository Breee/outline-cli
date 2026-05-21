package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/Breee/outline-cli/internal/config"
	"github.com/Breee/outline-cli/internal/oidc"
	"github.com/spf13/cobra"
)

var (
	oidcIssuer   string
	oidcClientID string
	oidcScopes   []string
	oidcPort     int
)

func init() {
	authCmd := &cobra.Command{Use: "auth", Short: "Authentication helpers"}
	oidcCmd := &cobra.Command{
		Use:   "oidc-login",
		Short: "Login via browser-based OIDC flow through Outline",
		Long: `Initiates Outline's OIDC login flow in your browser.
A local HTTP server captures the callback and obtains an Outline session token.
Requires --server-url to be set (the Outline instance URL).`,
		Example: `  # Login to your Outline instance
  outline auth oidc-login --server-url https://wiki.example.com

  # Use a custom callback port
  outline auth oidc-login --server-url https://wiki.example.com --port 9090`,
		RunE: runOIDCLogin,
	}
	oidcCmd.Flags().IntVar(&oidcPort, "port", 10800, "Local port for OIDC callback server")

	checkCmd := &cobra.Command{
		Use:     "check",
		Short:   "Verify that the stored credentials are valid",
		Example: "  outline auth check --server-url https://wiki.example.com",
		RunE:    runAuthCheck,
	}

	authCmd.AddCommand(oidcCmd, checkCmd)
	rootCmd.AddCommand(authCmd)
}

func runOIDCLogin(cmd *cobra.Command, _ []string) error {
	if serverURL == "" {
		return fmt.Errorf("--server-url is required for OIDC login")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	cmd.Printf("Starting OIDC login via %s ...\n", serverURL)

	session, err := oidc.StartBrowserLogin(ctx, serverURL, oidcPort)
	if err != nil {
		return err
	}

	// Open browser automatically, fall back to printing the URL.
	if openErr := openBrowser(session.BrowserURL); openErr != nil {
		cmd.Printf("Open this URL in your browser:\n  %s\n", session.BrowserURL)
	} else {
		cmd.Println("Browser opened. Waiting for authentication...")
	}

	result, err := session.Wait(ctx)
	if err != nil {
		return err
	}

	cfg, _ := config.Load(cfgFile)
	if serverURL != "" {
		cfg.ServerURL = serverURL
	}
	if err := config.SaveToken(cfgFile, &cfg, result.AccessToken); err != nil {
		return err
	}

	storage := cfg.TokenStorage
	if storage == "" {
		storage = config.TokenStorageKeyring
	}
	if storage == config.TokenStorageKeyring {
		cmd.Printf("Token saved to OS keyring (service: outline-cli)\n")
	} else {
		cmd.Printf("Token saved to %s\n", cfgFile)
	}
	return nil
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}
}

func runAuthCheck(cmd *cobra.Command, _ []string) error {
	client, err := newOutlineClientRaw()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	info, err := client.GetAuthInfo(ctx)
	if err != nil {
		return fmt.Errorf("not authenticated: %w", err)
	}
	cmd.Printf("Authenticated as %s (%s) on team %s\n", info.User.Name, info.User.Email, info.Team.Name)
	return nil
}
