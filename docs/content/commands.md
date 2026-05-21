---
title: "CLI Command Reference"
weight: 55
description: "Complete reference for all outline-cli commands and flags."
llmsDescription: "outline-cli commands: 1) `outline push` — push markdown to Outline. Flags: --path (default '.'), --collection-id (name/slug/UUID), --publish (default true), --create-collection (default false). 2) `outline auth oidc-login` — browser OIDC login. Flags: --port (default 10800). 3) `outline auth check` — verify credentials, prints user/email/team. 4) `outline config set <key> <value>` — set config value. 5) `outline config get <key>` — get config value. 6) `outline config list` — list all values (secrets masked). 7) `outline config path` — print config file path. 8) `outline completion <bash|zsh|fish>` — generate shell completions. Global flags: --config (path), --server-url / OUTLINE_SERVER_URL / OUTLINE_HOST, --api-token / OUTLINE_API_TOKEN, --username / OUTLINE_USERNAME, --password / OUTLINE_PASSWORD, --oidc-access-token / OUTLINE_OIDC_ACCESS_TOKEN."
---

# CLI Command Reference

## Global Flags

These flags apply to all commands:

| Flag | Env Var | Description |
|------|---------|-------------|
| `--config` | — | Config file path (default `~/.outline-cli/config.yaml`) |
| `--server-url` | `OUTLINE_SERVER_URL`, `OUTLINE_HOST` | Outline server URL |
| `--api-token` | `OUTLINE_API_TOKEN` | API token |
| `--username` | `OUTLINE_USERNAME` | Basic auth username |
| `--password` | `OUTLINE_PASSWORD` | Basic auth password |
| `--oidc-access-token` | `OUTLINE_OIDC_ACCESS_TOKEN` | OIDC access token |

---

## `outline push`

Push markdown files to an Outline collection.

```bash
outline push --path ./docs/ --collection-id "Engineering"
```

| Flag | Default | Description |
|------|---------|-------------|
| `-p, --path` | `.` | Markdown file or directory to push |
| `--collection-id` | — | Target collection (name, slug, or UUID) |
| `--publish` | `true` | Publish created documents |
| `--create-collection` | `false` | Create collection if it doesn't exist |

---

## `outline auth oidc-login`

Authenticate via browser-based OIDC flow.

```bash
outline auth oidc-login
```

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `10800` | Local port for OAuth callback server |

---

## `outline auth check`

Verify stored credentials are valid.

```bash
outline auth check
```

Prints: authenticated user name, email, and team. Exit code 1 if not authenticated.

---

## `outline config set`

Set a configuration value.

```bash
outline config set <key> <value>
```

Secret keys (`api_token`, `password`) are stored in the OS keyring.

---

## `outline config get`

Print a configuration value.

```bash
outline config get <key>
```

---

## `outline config list`

List all set configuration values. Secrets are masked as `***`.

```bash
outline config list
```

---

## `outline config path`

Print the config file path.

```bash
outline config path
```

---

## `outline completion`

Generate shell completion scripts.

```bash
outline completion bash
outline completion zsh
outline completion fish
```
