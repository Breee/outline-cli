---
title: "Config File Reference"
weight: 60
description: "YAML config file format for outline-cli."
generated: true
---

Config file location: `~/.outline-cli/config.yaml` (override with `--config <path>`)

## Format

```yaml
server_url: # Outline server base URL
auth_method: # Auth method: api_token, oidc, basic
token_storage: # Secret storage backend: keyring, file
oidc_port: # Local port for OIDC callback
username: # Basic auth username
```

## Fields

| Key | Secret | Env Var | Description |
|-----|--------|---------|-------------|
| `server_url` | no | `OUTLINE_SERVER_URL` | Outline server base URL |
| `auth_method` | no | `OUTLINE_AUTH_METHOD` | Auth method: api_token, oidc, basic |
| `token_storage` | no | `OUTLINE_TOKEN_STORAGE` | Secret storage backend: keyring, file |
| `oidc_port` | no | `OUTLINE_OIDC_PORT` | Local port for OIDC callback |
| `api_token` | yes | `OUTLINE_API_TOKEN` | API bearer token |
| `password` | yes | `OUTLINE_PASSWORD` | Basic auth password |
| `username` | no | `OUTLINE_USERNAME` | Basic auth username |
| `oidc_access_token` | yes | `OUTLINE_OIDC_ACCESS_TOKEN` | OIDC access token (set by auth oidc-login) |

## File Permissions

The config file is always written with `0600` permissions (owner read/write only).

## Secret Storage

When `token_storage=keyring` (default), secret keys are stored in the OS keyring under service `outline-cli`.
Set `token_storage=file` for headless environments (CI, containers) — secrets are then written to the YAML file in plaintext.
