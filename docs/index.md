---
title: "outline-cli"
description: "The best wiki CLI: push, pull, search, and read Outline wiki documents from your terminal."
llmsDescription: "outline-cli is a Go CLI tool for managing Outline wiki documents. Core commands: `outline push --path ./docs/ --collection-id <id>` to publish markdown files, `outline auth oidc-login` for browser-based OIDC authentication, `outline config set <key> <value>` for configuration. Supports metadata headers (<!-- Collection: ... -->, <!-- Title: ... -->, <!-- Icon: ... -->, <!-- Parent: ... -->), automatic image upload, directory-based parent/child nesting, collection resolution by name/slug/UUID, and OS keyring credential storage. Auth priority: CLI flags > env vars (OUTLINE_API_TOKEN, OUTLINE_HOST) > keyring > config file."
---

# outline-cli

The best wiki CLI: push, pull, search, and read [Outline](https://github.com/outline/outline) wiki documents from your terminal.

## Why outline-cli?

- **Stay in your terminal** — no context switching to a browser
- **Docs-as-code** — version control your wiki content in git
- **CI/CD native** — publish docs automatically on every merge
- **Secure by default** — credentials stored in OS keyring (OWASP-aligned)

## Quick Start

```bash
# Install
curl -sL "https://github.com/Breee/outline-cli/releases/latest/download/outline_linux_amd64.tar.gz" | sudo tar xz -C /usr/local/bin outline

# Authenticate
export OUTLINE_HOST=https://outline.example.com
export OUTLINE_API_TOKEN=sk-your-token

# Push docs
outline push --collection-id "Engineering" --path ./docs/
```

## Features

| Feature | Description |
|---------|-------------|
| **Push** | Push markdown files or directory trees to Outline |
| **Upsert** | Updates existing docs by title, creates if not found |
| **Metadata** | Per-file collection, title, icon, parent via HTML comments |
| **Images** | Auto-upload local images as attachments |
| **OIDC** | Browser login with PKCE + Dynamic Client Registration |
| **Keyring** | All credentials in OS keyring (API token, password, OIDC) |
| **Config** | YAML config with `config set/get/list` commands |
| **CI/CD** | Works with env vars only — no config file needed |

## AI Integration

This documentation is optimized for AI consumption:

- [`llms.txt`](/llms.txt) — structured index for LLM discovery
- [`llms-full.txt`](/llms-full.txt) — complete documentation as single file
- Every page has `llmsDescription` frontmatter for precise AI context

## Next Steps

- [Installation](guide/install.md) — download and set up
- [Quick Start](guide/quickstart.md) — first push in 2 minutes
- [CI/CD Integration](guide/cicd.md) — automate doc publishing
