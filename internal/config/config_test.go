package config

import (
	"testing"
)

func TestValidKeys_ContainsNewKeys(t *testing.T) {
	t.Parallel()
	required := []string{"api_token", "password"}
	for _, key := range required {
		found := false
		for _, k := range ValidKeys {
			if k == key {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ValidKeys missing %q", key)
		}
	}
}

func TestSecretKeys(t *testing.T) {
	t.Parallel()
	if !SecretKeys["api_token"] {
		t.Error("SecretKeys missing api_token")
	}
	if !SecretKeys["password"] {
		t.Error("SecretKeys missing password")
	}
	if SecretKeys["server_url"] {
		t.Error("server_url should not be a secret key")
	}
}

func TestSetAndGet_APIToken(t *testing.T) {
	t.Parallel()
	cfg := &File{TokenStorage: TokenStorageFile}
	if err := cfg.Set("api_token", "my-token"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if cfg.APIToken != "my-token" {
		t.Fatalf("APIToken = %q, want %q", cfg.APIToken, "my-token")
	}
}

func TestSetAndGet_Password(t *testing.T) {
	t.Parallel()
	cfg := &File{TokenStorage: TokenStorageFile}
	if err := cfg.Set("password", "secret123"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if cfg.Password != "secret123" {
		t.Fatalf("Password = %q, want %q", cfg.Password, "secret123")
	}
}

func TestGet_APIToken_ReadsFromFileField(t *testing.T) {
	t.Parallel()
	cfg := &File{TokenStorage: TokenStorageFile, APIToken: "stored-token"}
	val, err := cfg.Get("api_token")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "stored-token" {
		t.Fatalf("Get(api_token) = %q, want %q", val, "stored-token")
	}
}

func TestGet_Password_ReadsFromFileField(t *testing.T) {
	t.Parallel()
	cfg := &File{TokenStorage: TokenStorageFile, Password: "stored-pw"}
	val, err := cfg.Get("password")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "stored-pw" {
		t.Fatalf("Get(password) = %q, want %q", val, "stored-pw")
	}
}

func TestGet_UnknownKey(t *testing.T) {
	t.Parallel()
	cfg := &File{}
	_, err := cfg.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown key")
	}
}

func TestSet_UnknownKey(t *testing.T) {
	t.Parallel()
	cfg := &File{}
	err := cfg.Set("nonexistent", "value")
	if err == nil {
		t.Fatal("expected error for unknown key")
	}
}
