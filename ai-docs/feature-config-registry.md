# Feature: Declarative Config Registry

## Problem

`docs.go` is too complex. Functions like `discoverEnvVars` (parse source for quoted strings) and `buildConfigKeys` (iterate exported slices) are fragile heuristics. Adding a new config key or env var requires knowing the internals of the discovery code.

## Solution

A single declarative config registry (struct slice) that is the **source of truth** for:
- Config keys (`outline config set <key> <value>`)
- Env var bindings (`OUTLINE_*`)
- Whether it's a secret (stored in keyring)
- Human-readable description (used in docs generation)

## Design

```go
// internal/config/registry.go

type Option struct {
    Key         string // config file key, e.g. "server_url"
    EnvVar      string // env var name, e.g. "OUTLINE_SERVER_URL"
    Secret      bool   // stored in OS keyring
    Description string // used in generated docs and `outline config list`
}

var Registry = []Option{
    {Key: "server_url", EnvVar: "OUTLINE_SERVER_URL", Description: "Outline server base URL"},
    {Key: "auth_method", EnvVar: "OUTLINE_AUTH_METHOD", Description: "Auth method: api_token, oidc, basic"},
    {Key: "token_storage", EnvVar: "", Description: "Secret storage backend: keyring, file"},
    {Key: "oidc_port", EnvVar: "", Description: "Local port for OIDC callback"},
    {Key: "api_token", EnvVar: "OUTLINE_API_TOKEN", Secret: true, Description: "API bearer token"},
    {Key: "password", EnvVar: "OUTLINE_PASSWORD", Secret: true, Description: "Basic auth password"},
}
```

## What This Replaces

| Current code | After |
|---|---|
| `config.ValidKeys` (string slice) | `config.Registry` |
| `config.SecretKeys` (map) | `opt.Secret` field |
| `discoverEnvVars()` in docs.go (parses root.go source) | `config.Registry` iteration |
| `buildConfigKeys()` in docs.go | `config.Registry` iteration |
| Manual `viper.BindEnv()` calls in root.go | Loop over `Registry` |

## Impact on docs.go

`buildInstructionContent()` becomes:

```go
data := instructionData{
    Targets:  readMakefileTargets(),
    Packages: discoverPackages(),
    Options:  config.Registry,
}
```

Delete: `discoverEnvVars()`, `extractQuotedStrings()`, `buildConfigKeys()`, the separate `EnvVars`/`ConfigKeys` template sections.

The template merges env vars and config keys into one table:

```
## Configuration

| Key | Env Var | Secret | Description |
|-----|---------|--------|-------------|
{{- range .Options}}
| `{{.Key}}` | {{if .EnvVar}}`{{.EnvVar}}`{{else}}—{{end}} | {{if .Secret}}yes{{else}}no{{end}} | {{.Description}} |
{{- end}}
```

## Benefits

- One place to add a config option (registry) — docs, CLI, env binding all update automatically
- No source-parsing heuristics
- `outline config list` can print descriptions
- Template is simpler and more readable
