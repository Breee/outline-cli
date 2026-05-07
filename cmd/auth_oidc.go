package cmd

import (
	"context"
	"time"

	"github.com/Breee/outline-cli/internal/config"
	"github.com/Breee/outline-cli/internal/oidc"
	"github.com/spf13/cobra"
)

var (
	oidcIssuer   string
	oidcClientID string
	oidcScopes   []string
)

func init() {
	authCmd := &cobra.Command{Use: "auth", Short: "Authentication helpers"}
	oidcCmd := &cobra.Command{
		Use:   "oidc-login",
		Short: "Login using OIDC device authorization flow",
		RunE:  runOIDCLogin,
	}
	oidcCmd.Flags().StringVar(&oidcIssuer, "issuer", "", "OIDC issuer URL")
	oidcCmd.Flags().StringVar(&oidcClientID, "client-id", "", "OIDC client ID")
	oidcCmd.Flags().StringSliceVar(&oidcScopes, "scopes", []string{"openid", "profile", "email"}, "OIDC scopes")
	_ = oidcCmd.MarkFlagRequired("issuer")
	_ = oidcCmd.MarkFlagRequired("client-id")

	authCmd.AddCommand(oidcCmd)
	rootCmd.AddCommand(authCmd)
}

func runOIDCLogin(cmd *cobra.Command, _ []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	token, device, err := oidc.DeviceLogin(ctx, oidcIssuer, oidcClientID, oidcScopes)
	if err != nil {
		return err
	}

	if device.VerificationURIComplete != "" {
		cmd.Printf("Open this URL in your browser: %s\n", device.VerificationURIComplete)
	} else {
		cmd.Printf("Open %s and enter code: %s\n", device.VerificationURI, device.UserCode)
	}

	cfg := config.File{OIDCAccessToken: token.AccessToken}
	if err := config.Save(cfgFile, cfg); err != nil {
		return err
	}
	cmd.Printf("OIDC token saved to %s\n", cfgFile)
	return nil
}
