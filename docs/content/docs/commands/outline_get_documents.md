---
title: "outline get documents"
description: "CLI reference for outline get documents"
llmsDescription: "Auto-generated CLI reference for outline get documents. Contains usage, flags, and examples."
generated: true
---

## outline get documents

List or get documents

```
outline get documents [flags]
```

### Examples

```
  # List documents in a collection
  outline get documents --collection test

  # Get a specific document by title (markdown output)
  outline get documents "Deployment Guide" -o md

  # Get by title, JSON output
  outline get documents "API Reference" -o json

  # YAML metadata
  outline get documents "FAQ" -o yaml
```

### Options

```
      --collection string   Filter by collection (name, slug, or UUID)
  -h, --help                help for documents
  -o, --output string       Output format: table, json, yaml, md, raw (default "table")
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

* [outline get](outline_get.md)	 - Get resources from Outline (kubectl-style)

