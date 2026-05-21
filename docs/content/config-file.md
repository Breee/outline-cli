---
title: "Config File Reference"
weight: 60
description: "YAML config file format for outline-cli."
llmsDescription: "outline-cli config file at ~/.outline-cli/config.yaml. YAML format with fields: server_url (string), auth_method (oidc|api-token|basic), token_storage (keyring|file), oidc_port (int 1-65535), oidc_access_token (string, set by auth flow), api_token (string, only if token_storage=file), password (string, only if token_storage=file). File permissions are 0600. Secret fields are only written to file when token_storage=file; otherwise stored in OS keyring under service 'outline-cli'."
---

# Config File Reference

Config file location: `~/.outline-cli/config.yaml`

## Format

```yaml
server_url: https://outline.example.com
auth_method: oidc
token_storage: keyring
oidc_port: 10800
```

## Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `server_url` | string | — | Outline server URL |
| `auth_method` | string | `oidc` | Auto-auth method (`oidc`, `api-token`, `basic`) |
| `token_storage` | string | `keyring` | Where to store secrets (`keyring`, `file`) |
| `oidc_port` | int | `10800` | Local port for OIDC callback |
| `oidc_access_token` | string | — | OIDC token (only if `token_storage=file`) |
| `api_token` | string | — | API token (only if `token_storage=file`) |
| `password` | string | — | Basic auth password (only if `token_storage=file`) |

## File Permissions

The config file is always written with `0600` permissions (owner read/write only).

## Keyring Storage

When `token_storage=keyring` (default), secrets are stored in the OS keyring:

| Keyring Service | Keyring User | Credential |
|-----------------|--------------|------------|
| `outline-cli` | `oidc_access_token` | OIDC token |
| `outline-cli` | `api_token` | API token |
| `outline-cli` | `basic_password` | Basic auth password |

The config file only stores non-secret values in this mode.

## File Storage (Headless)

For environments without a keyring (CI, containers, SSH sessions):

```bash
outline config set token_storage file
```

Secrets are then written to the YAML file in plaintext. Ensure the file is not committed to git.
