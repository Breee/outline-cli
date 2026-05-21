package config

// Option describes a single configuration option.
type Option struct {
	Key         string // config file key, e.g. "server_url"; empty for env-only options
	EnvVar      string // env var name, e.g. "OUTLINE_SERVER_URL"; empty if no env binding
	Secret      bool   // stored in OS keyring
	Description string // human-readable, used in docs and `config list`
}

// Registry is the single source of truth for all config options.
var Registry = []Option{
	{Key: "server_url", EnvVar: "OUTLINE_SERVER_URL", Description: "Outline server base URL"},
	{Key: "auth_method", EnvVar: "OUTLINE_AUTH_METHOD", Description: "Auth method: api_token, oidc, basic"},
	{Key: "token_storage", EnvVar: "OUTLINE_TOKEN_STORAGE", Description: "Secret storage backend: keyring, file"},
	{Key: "oidc_port", EnvVar: "OUTLINE_OIDC_PORT", Description: "Local port for OIDC callback"},
	{Key: "api_token", EnvVar: "OUTLINE_API_TOKEN", Secret: true, Description: "API bearer token"},
	{Key: "password", EnvVar: "OUTLINE_PASSWORD", Secret: true, Description: "Basic auth password"},
	{Key: "", EnvVar: "OUTLINE_USERNAME", Description: "Basic auth username"},
	{Key: "oidc_access_token", EnvVar: "OUTLINE_OIDC_ACCESS_TOKEN", Secret: true, Description: "OIDC access token (set by auth oidc-login)"},
}
