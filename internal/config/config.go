package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type File struct {
	ServerURL       string `yaml:"server_url,omitempty" mapstructure:"server_url"`
	AuthMethod      string `yaml:"auth_method,omitempty" mapstructure:"auth_method"`
	TokenStorage    string `yaml:"token_storage,omitempty" mapstructure:"token_storage"`
	OIDCAccessToken string `yaml:"oidc_access_token,omitempty" mapstructure:"oidc_access_token"`
	APIToken        string `yaml:"api_token,omitempty" mapstructure:"api_token"`
	Password        string `yaml:"password,omitempty" mapstructure:"password"`
	Username        string `yaml:"username,omitempty" mapstructure:"username"`
	OIDCPort        int    `yaml:"oidc_port,omitempty" mapstructure:"oidc_port"`
}

// ValidKeys lists all configurable keys (derived from Registry).
var ValidKeys = func() []string {
	var keys []string
	for _, opt := range Registry {
		if opt.Key != "" {
			keys = append(keys, opt.Key)
		}
	}
	return keys
}()

// SecretKeys are keys that are stored in the keyring (derived from Registry).
var SecretKeys = func() map[string]bool {
	m := make(map[string]bool)
	for _, opt := range Registry {
		if opt.Secret {
			m[opt.Key] = true
		}
	}
	return m
}()

// InitViper sets up Viper with env var bindings from the Registry.
// Call this once during CLI initialization.
func InitViper() {
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("")        // No prefix — env vars are explicit in Registry.
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	for _, opt := range Registry {
		if opt.EnvVar != "" {
			_ = viper.BindEnv(opt.Key, opt.EnvVar)
		}
	}
}

// SetConfigFile tells Viper which file to read.
func SetConfigFile(path string) {
	if path == "" {
		path = DefaultPath()
	}
	viper.SetConfigFile(path)
	_ = viper.ReadInConfig() // Ignore error if file doesn't exist yet.
}

// Get returns the value of a config key (Viper-resolved: flag > env > file).
func (f *File) Get(key string) (string, error) {
	switch key {
	case "server_url":
		return f.ServerURL, nil
	case "auth_method":
		return f.AuthMethod, nil
	case "token_storage":
		if f.TokenStorage == "" {
			return TokenStorageKeyring, nil
		}
		return f.TokenStorage, nil
	case "oidc_access_token":
		return f.OIDCAccessToken, nil
	case "api_token":
		return LoadAPIToken(*f), nil
	case "password":
		return LoadPassword(*f), nil
	case "username":
		return f.Username, nil
	case "oidc_port":
		if f.OIDCPort == 0 {
			return "", nil
		}
		return strconv.Itoa(f.OIDCPort), nil
	default:
		return "", fmt.Errorf("unknown config key %q (valid: %v)", key, ValidKeys)
	}
}

// Set sets the value of a config key.
func (f *File) Set(key, value string) error {
	switch key {
	case "server_url":
		f.ServerURL = value
	case "auth_method":
		switch value {
		case "oidc", "api-token", "basic":
		default:
			return fmt.Errorf("invalid auth_method %q (use oidc, api-token, or basic)", value)
		}
		f.AuthMethod = value
	case "oidc_port":
		port, err := strconv.Atoi(value)
		if err != nil || port < 1 || port > 65535 {
			return fmt.Errorf("invalid port %q", value)
		}
		f.OIDCPort = port
	case "token_storage":
		switch value {
		case TokenStorageKeyring, TokenStorageFile:
		default:
			return fmt.Errorf("invalid token_storage %q (use keyring or file)", value)
		}
		f.TokenStorage = value
	case "api_token":
		f.APIToken = value
	case "password":
		f.Password = value
	case "username":
		f.Username = value
	default:
		return fmt.Errorf("unknown config key %q (valid: %v)", key, ValidKeys)
	}
	return nil
}

func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".outline-cli.yaml"
	}
	return filepath.Join(home, ".outline-cli", "config.yaml")
}

func Load(path string) (File, error) {
	if path == "" {
		path = DefaultPath()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return File{}, err
	}
	var cfg File
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return File{}, err
	}
	return cfg, nil
}

func Save(path string, cfg File) error {
	if path == "" {
		path = DefaultPath()
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
