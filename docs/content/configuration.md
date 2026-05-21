---
title: "Configuration"
weight: 40
description: "Manage outline-cli configuration with config set/get/list commands."
llmsDescription: "outline-cli config stored at ~/.outline-cli/config.yaml. Commands: `outline config set <key> <value>`, `outline config get <key>`, `outline config list`, `outline config path`. Valid keys: server_url, auth_method (oidc|api-token|basic), token_storage (keyring|file), oidc_port (int), api_token (stored in keyring), password (stored in keyring). Secret keys (api_token, password) are stored in OS keyring by default, masked in `config list`. Set token_storage=file for headless environments."
---

# Configuration

outline-cli uses a YAML config file at `~/.outline-cli/config.yaml` with OS keyring for secrets.

## Commands

```bash
outline config set server_url https://outline.example.com
outline config get server_url
outline config list
outline config path
```

## Available Keys

| Key | Values | Description |
|-----|--------|-------------|
| `server_url` | URL | Outline server URL |
| `auth_method` | `oidc`, `api-token`, `basic` | Authentication method for auto-auth |
| `token_storage` | `keyring`, `file` | Where to store secrets |
| `oidc_port` | `1`–`65535` | Local port for OIDC callback |
| `api_token` | string | API token (stored in keyring) |
| `password` | string | Basic auth password (stored in keyring) |

## Secret Storage

Keys `api_token` and `password` are **stored in the OS keyring** by default:

```bash
outline config set api_token sk-abc123
# → stored in libsecret/Keychain/Credential Manager, NOT in config.yaml
```

For headless environments (CI, containers) where keyring is unavailable:

```bash
outline config set token_storage file
outline config set api_token sk-abc123
# → stored in config.yaml (0600 permissions)
```

## Config File Format

```yaml
server_url: https://outline.example.com
auth_method: oidc
token_storage: keyring
oidc_port: 10800
```

!!! warning
    Never commit your config file to git. Add `config.yaml` to `.gitignore`.
