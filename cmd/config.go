package cmd

import (
	"fmt"
	"strings"

	"github.com/Breee/outline-cli/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
		Long:  "Get, set, or list configuration values stored in " + config.DefaultPath(),
	}

	configCmd.AddCommand(
		&cobra.Command{
			Use:   "get <key>",
			Short: "Print the value of a config key",
			Args:  cobra.ExactArgs(1),
			RunE:  runConfigGet,
		},
		&cobra.Command{
			Use:   "set <key> <value>",
			Short: "Set a config key to a value",
			Long: fmt.Sprintf("Set a config key to a value.\n\nAvailable keys:\n  %s",
				strings.Join(config.ValidKeys, "\n  ")),
			Args: cobra.ExactArgs(2),
			RunE: runConfigSet,
		},
		&cobra.Command{
			Use:   "list",
			Short: "List all config values",
			Args:  cobra.NoArgs,
			RunE:  runConfigList,
		},
		&cobra.Command{
			Use:   "path",
			Short: "Print the config file path",
			Args:  cobra.NoArgs,
			Run:   func(cmd *cobra.Command, _ []string) { cmd.Println(cfgFile) },
		},
	)

	rootCmd.AddCommand(configCmd)
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	cfg, _ := config.Load(cfgFile)
	val, err := cfg.Get(args[0])
	if err != nil {
		return err
	}
	cmd.Println(val)
	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	cfg, _ := config.Load(cfgFile)
	key, value := args[0], args[1]

	if err := cfg.Set(key, value); err != nil {
		return err
	}

	// Route secret keys through keyring storage.
	switch key {
	case "api_token":
		if err := config.SaveAPIToken(cfgFile, &cfg, value); err != nil {
			return err
		}
	case "password":
		if err := config.SavePassword(cfgFile, &cfg, value); err != nil {
			return err
		}
	default:
		if err := config.Save(cfgFile, cfg); err != nil {
			return err
		}
	}

	if config.SecretKeys[key] {
		cmd.Printf("%s = ***\n", key)
	} else {
		cmd.Printf("%s = %s\n", key, value)
	}
	return nil
}

func runConfigList(cmd *cobra.Command, _ []string) error {
	cfg, _ := config.Load(cfgFile)
	for _, key := range config.ValidKeys {
		val, _ := cfg.Get(key)
		if val != "" {
			if config.SecretKeys[key] {
				cmd.Printf("%s = ***\n", key)
			} else {
				cmd.Printf("%s = %s\n", key, val)
			}
		}
	}
	return nil
}
