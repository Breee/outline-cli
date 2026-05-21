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
	"github.com/spf13/viper"
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
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", config.DefaultPath(), "config file")
	rootCmd.PersistentFlags().StringVar(&serverURL, "server-url", "", "Outline server URL")
	rootCmd.PersistentFlags().StringVar(&apiToken, "api-token", "", "Outline API token")
	rootCmd.PersistentFlags().StringVar(&basicUser, "username", "", "Basic auth username")
	rootCmd.PersistentFlags().StringVar(&basicPassword, "password", "", "Basic auth password")
	rootCmd.PersistentFlags().StringVar(&oidcAccessToken, "oidc-access-token", "", "OIDC access token")

	// Bind flags to Viper keys.
	_ = viper.BindPFlag("server_url", rootCmd.PersistentFlags().Lookup("server-url"))
	_ = viper.BindPFlag("api_token", rootCmd.PersistentFlags().Lookup("api-token"))
	_ = viper.BindPFlag("username", rootCmd.PersistentFlags().Lookup("username"))
	_ = viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))
	_ = viper.BindPFlag("oidc_access_token", rootCmd.PersistentFlags().Lookup("oidc-access-token"))
}

// initConfig initializes Viper: env bindings from Registry + config file.
func initConfig() {
	config.InitViper()
	config.SetConfigFile(cfgFile)

	// Resolve values from Viper (flag > env > config file).
	if serverURL == "" {
		serverURL = viper.GetString("server_url")
	}
	if apiToken == "" {
		apiToken = viper.GetString("api_token")
	}
	if basicUser == "" {
		basicUser = viper.GetString("username")
	}
	if basicPassword == "" {
		basicPassword = viper.GetString("password")
	}
	if oidcAccessToken == "" {
		oidcAccessToken = viper.GetString("oidc_access_token")
	}

	// Secrets may live in the OS keyring — check there if still empty.
	if apiToken == "" || basicPassword == "" || oidcAccessToken == "" {
		cfg, err := config.Load(cfgFile)
		if err == nil {
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
func autoAuth(cmd *cobra.Command) error {
	method := viper.GetString("auth_method")
	if method == "" {
		method = "oidc"
	}

	switch method {
	case "oidc":
		if serverURL == "" {
			return fmt.Errorf("auto-auth requires --server-url or server_url in config")
		}
		port := viper.GetInt("oidc_port")
		if port == 0 {
			port = 10800
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
		cfg, _ := config.Load(cfgFile)
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
