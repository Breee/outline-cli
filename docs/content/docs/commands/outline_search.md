---
title: "outline search"
description: "CLI reference for outline search"
llmsDescription: "Auto-generated CLI reference for outline search. Contains usage, flags, and examples."
generated: true
---

## outline search

Search documents in Outline

```
outline search <query> [flags]
```

### Examples

```
  # Basic search
  outline search "kubernetes rollback"

  # Filter by collection
  outline search "deploy" --collection infrastructure

  # JSON output for scripting
  outline search "API key" --format json

  # Compact output
  outline search "auth" --format oneline

  # Limit results
  outline search "setup" --limit 5
```

### Options

```
      --collection string   Filter by collection (name, slug, or UUID)
      --format string       Output format: default, json, oneline (default "default")
  -h, --help                help for search
      --limit int           Maximum number of results (default 25)
```

### Options inherited from parent commands

```
      --api-token string           Outline API token
      --config string              config file (default "$HOME/.outline-cli/config.yaml")
      --oidc-access-token string   OIDC access token
      --password string            Basic auth password
      --server-url string          Outline server URL
      --username string            Basic auth username
```

### SEE ALSO

* [outline](outline.md)	 - Push markdown documents to Outline

