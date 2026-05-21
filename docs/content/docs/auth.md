---
title: "Authentication"
weight: 30
description: "Configure authentication for outline-cli: API token, OIDC, or basic auth."
llmsDescription: "outline-cli authentication methods: 1) API token: set OUTLINE_API_TOKEN env var or `--api-token` flag or `outline config set api_token <token>` (stored in OS keyring). 2) OIDC browser login: run `outline auth oidc-login --port 10800`, uses PKCE + Dynamic Client Registration, token stored in keyring. 3) Basic auth: set OUTLINE_USERNAME and OUTLINE_PASSWORD env vars or `--username`/`--password` flags. Credential resolution priority: CLI flags > env vars > OS keyring > config file. Auto-reauthentication triggers OIDC flow when token expires (configurable via auth_method). Verify with `outline auth check`."
---


outline-cli supports three authentication methods. Choose based on your use case.

## API Token (Recommended for CI/CD)

The simplest method. Get a token from **Settings → API** in Outline.

{{< tabs >}}

{{< tab name="Environment Variable" >}}
```bash
export OUTLINE_API_TOKEN=sk-your-token
export OUTLINE_HOST=https://outline.example.com
outline push --path ./docs/ --collection-id "Engineering"
```
{{< /tab >}}

{{< tab name="Config (stored in keyring)" >}}
```bash
outline config set server_url https://outline.example.com
outline config set api_token sk-your-token
# Token is now in your OS keyring — no env var needed
outline push --path ./docs/ --collection-id "Engineering"
```
{{< /tab >}}

{{< tab name="Flag" >}}
```bash
outline --api-token sk-your-token --server-url https://outline.example.com \
  push --path ./docs/ --collection-id "Engineering"
```
{{< /tab >}}

{{< /tabs >}}

## OIDC Browser Login (Recommended for developers)

Interactive browser login using OAuth2 with PKCE. No token management needed — the CLI handles everything.

```bash
outline config set server_url https://outline.example.com
outline auth oidc-login
```

This will:

1. Discover your Outline instance's OAuth2 configuration
2. Register a public OAuth2 client (Dynamic Client Registration)
3. Open your browser for authentication
4. Store the access token in your OS keyring

After login, all commands work without additional flags.

## Basic Auth

```bash
export OUTLINE_USERNAME=user@example.com
export OUTLINE_PASSWORD=secret
export OUTLINE_HOST=https://outline.example.com
outline push --path ./docs/ --collection-id "Docs"
```

## Credential Resolution Order

When multiple credentials are available, this priority applies (highest wins):

1. CLI flags (`--api-token`, `--username`/`--password`, `--oidc-access-token`)
2. Environment variables (`OUTLINE_API_TOKEN`, `OUTLINE_HOST`)
3. OS keyring (set via `config set` or `auth oidc-login`)
4. Config file (`~/.outline-cli/config.yaml`)

## Auto-Reauthentication

If your token expires during a `push`, the CLI automatically re-authenticates using your configured method:

```bash
outline config set auth_method oidc    # default — triggers browser login
outline config set auth_method api-token  # requires valid token in env/keyring
```

## Verify Credentials

```bash
outline auth check
```

Output:
```
✓ Authenticated as jane@example.com (Engineering team)
```

## Security

- All persisted credentials are stored in the **OS keyring** (libsecret on Linux, Keychain on macOS, Credential Manager on Windows)
- Config file is written with `0600` permissions
- If keyring is unavailable (headless/container), falls back to config file with a warning
- Secrets are never logged; `config list` masks them as `***`
