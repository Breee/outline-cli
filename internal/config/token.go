package config

import (
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	keyringService = "outline-cli"

	// Keyring user keys for each credential type.
	KeyringUserOIDC     = "oidc_access_token"
	KeyringUserAPIToken = "api_token"
	KeyringUserPassword = "basic_password"

	// TokenStorageKeyring stores secrets in the OS keyring.
	TokenStorageKeyring = "keyring"
	// TokenStorageFile stores secrets in the config file (plaintext).
	TokenStorageFile = "file"
)

// SaveSecret saves a secret using the configured storage method.
func SaveSecret(cfgPath string, cfg *File, keyringUser string, value string, setFileField func(*File, string)) error {
	storage := cfg.TokenStorage
	if storage == "" {
		storage = TokenStorageKeyring
	}

	switch storage {
	case TokenStorageKeyring:
		if err := keyring.Set(keyringService, keyringUser, value); err != nil {
			// Fall back to file if keyring unavailable (e.g. headless server).
			setFileField(cfg, value)
			cfg.TokenStorage = TokenStorageFile
			return Save(cfgPath, *cfg)
		}
		// Clear any value previously stored in the file.
		setFileField(cfg, "")
		return Save(cfgPath, *cfg)
	case TokenStorageFile:
		setFileField(cfg, value)
		return Save(cfgPath, *cfg)
	default:
		return fmt.Errorf("unknown token_storage %q (use keyring or file)", storage)
	}
}

// LoadSecret retrieves a secret from the configured storage.
func LoadSecret(cfg File, keyringUser string, fileFieldValue string) string {
	storage := cfg.TokenStorage
	if storage == "" {
		storage = TokenStorageKeyring
	}

	switch storage {
	case TokenStorageKeyring:
		secret, err := keyring.Get(keyringService, keyringUser)
		if err == nil && secret != "" {
			return secret
		}
		// Fall back to file if keyring read fails.
		return fileFieldValue
	case TokenStorageFile:
		return fileFieldValue
	default:
		return fileFieldValue
	}
}

// DeleteSecret removes a stored secret.
func DeleteSecret(cfg File, keyringUser string) error {
	storage := cfg.TokenStorage
	if storage == "" {
		storage = TokenStorageKeyring
	}

	if storage == TokenStorageKeyring {
		_ = keyring.Delete(keyringService, keyringUser)
	}
	return nil
}

// --- OIDC Token (backward-compatible wrappers) ---

// SaveToken saves the OIDC token using the configured storage method.
func SaveToken(cfgPath string, cfg *File, token string) error {
	return SaveSecret(cfgPath, cfg, KeyringUserOIDC, token, func(f *File, v string) {
		f.OIDCAccessToken = v
	})
}

// LoadToken retrieves the OIDC token from the configured storage.
func LoadToken(cfg File) string {
	return LoadSecret(cfg, KeyringUserOIDC, cfg.OIDCAccessToken)
}

// DeleteToken removes the stored OIDC token.
func DeleteToken(cfg File) error {
	return DeleteSecret(cfg, KeyringUserOIDC)
}

// --- API Token ---

// SaveAPIToken saves the API token using the configured storage method.
func SaveAPIToken(cfgPath string, cfg *File, token string) error {
	return SaveSecret(cfgPath, cfg, KeyringUserAPIToken, token, func(f *File, v string) {
		f.APIToken = v
	})
}

// LoadAPIToken retrieves the API token from the configured storage.
func LoadAPIToken(cfg File) string {
	return LoadSecret(cfg, KeyringUserAPIToken, cfg.APIToken)
}

// DeleteAPIToken removes the stored API token.
func DeleteAPIToken(cfg File) error {
	return DeleteSecret(cfg, KeyringUserAPIToken)
}

// --- Basic Auth Password ---

// SavePassword saves the basic auth password using the configured storage method.
func SavePassword(cfgPath string, cfg *File, password string) error {
	return SaveSecret(cfgPath, cfg, KeyringUserPassword, password, func(f *File, v string) {
		f.Password = v
	})
}

// LoadPassword retrieves the basic auth password from the configured storage.
func LoadPassword(cfg File) string {
	return LoadSecret(cfg, KeyringUserPassword, cfg.Password)
}

// DeletePassword removes the stored password.
func DeletePassword(cfg File) error {
	return DeleteSecret(cfg, KeyringUserPassword)
}
