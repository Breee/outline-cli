package config

import (
	"os"
	"path/filepath"
	"testing"
)

// Tests use token_storage=file to avoid OS keyring dependency in CI.

func tempConfigPath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "config.yaml")
}

func TestSaveAndLoadAPIToken_FileStorage(t *testing.T) {
	t.Parallel()
	path := tempConfigPath(t)
	cfg := &File{TokenStorage: TokenStorageFile}

	if err := SaveAPIToken(path, cfg, "sk-test-123"); err != nil {
		t.Fatalf("SaveAPIToken: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	got := LoadAPIToken(loaded)
	if got != "sk-test-123" {
		t.Fatalf("LoadAPIToken = %q, want %q", got, "sk-test-123")
	}
}

func TestSaveAndLoadPassword_FileStorage(t *testing.T) {
	t.Parallel()
	path := tempConfigPath(t)
	cfg := &File{TokenStorage: TokenStorageFile}

	if err := SavePassword(path, cfg, "s3cret!"); err != nil {
		t.Fatalf("SavePassword: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	got := LoadPassword(loaded)
	if got != "s3cret!" {
		t.Fatalf("LoadPassword = %q, want %q", got, "s3cret!")
	}
}

func TestSaveAndLoadToken_FileStorage(t *testing.T) {
	t.Parallel()
	path := tempConfigPath(t)
	cfg := &File{TokenStorage: TokenStorageFile}

	if err := SaveToken(path, cfg, "oidc-tok-abc"); err != nil {
		t.Fatalf("SaveToken: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	got := LoadToken(loaded)
	if got != "oidc-tok-abc" {
		t.Fatalf("LoadToken = %q, want %q", got, "oidc-tok-abc")
	}
}

func TestSaveSecret_UnknownStorage(t *testing.T) {
	t.Parallel()
	path := tempConfigPath(t)
	cfg := &File{TokenStorage: "invalid"}

	err := SaveSecret(path, cfg, "test", "val", func(f *File, v string) {})
	if err == nil {
		t.Fatal("expected error for unknown token_storage")
	}
}

func TestLoadSecret_UnknownStorage_FallsBackToFile(t *testing.T) {
	t.Parallel()
	cfg := File{TokenStorage: "bogus", APIToken: "from-file"}
	got := LoadSecret(cfg, KeyringUserAPIToken, cfg.APIToken)
	if got != "from-file" {
		t.Fatalf("LoadSecret = %q, want %q", got, "from-file")
	}
}

func TestDeleteSecret_FileStorage_NoError(t *testing.T) {
	t.Parallel()
	cfg := File{TokenStorage: TokenStorageFile}
	if err := DeleteSecret(cfg, KeyringUserAPIToken); err != nil {
		t.Fatalf("DeleteSecret: %v", err)
	}
}

func TestConfigFilePermissions(t *testing.T) {
	t.Parallel()
	path := tempConfigPath(t)
	cfg := &File{TokenStorage: TokenStorageFile}

	if err := SaveAPIToken(path, cfg, "secret"); err != nil {
		t.Fatalf("SaveAPIToken: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Fatalf("config file permissions = %o, want 0600", perm)
	}
}

func TestMultipleSecrets_Independent(t *testing.T) {
	t.Parallel()
	path := tempConfigPath(t)
	cfg := &File{TokenStorage: TokenStorageFile}

	if err := SaveAPIToken(path, cfg, "api-123"); err != nil {
		t.Fatalf("SaveAPIToken: %v", err)
	}
	if err := SavePassword(path, cfg, "pw-456"); err != nil {
		t.Fatalf("SavePassword: %v", err)
	}
	if err := SaveToken(path, cfg, "oidc-789"); err != nil {
		t.Fatalf("SaveToken: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got := LoadAPIToken(loaded); got != "api-123" {
		t.Fatalf("LoadAPIToken = %q, want %q", got, "api-123")
	}
	if got := LoadPassword(loaded); got != "pw-456" {
		t.Fatalf("LoadPassword = %q, want %q", got, "pw-456")
	}
	if got := LoadToken(loaded); got != "oidc-789" {
		t.Fatalf("LoadToken = %q, want %q", got, "oidc-789")
	}
}
