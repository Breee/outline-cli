package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type File struct {
	OIDCAccessToken string `json:"oidc_access_token"`
}

func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".outline-cli.json"
	}
	return filepath.Join(home, ".config", "outline-cli", "config.json")
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
	if err := json.Unmarshal(data, &cfg); err != nil {
		return File{}, err
	}
	if cfg.OIDCAccessToken == "" {
		return File{}, errors.New("oidc access token is empty")
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
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
