---
title: "outline pull"
description: "CLI reference for outline pull"
llmsDescription: "Auto-generated CLI reference for outline pull. Contains usage, flags, and examples."
generated: true
---

## outline pull

Pull documents from Outline to local markdown files

```
outline pull [flags]
```

### Examples

```
  # Pull entire collection to a directory
  outline pull --collection "Engineering" --output ./docs/

  # Pull a single document by title
  outline pull --doc "Deployment Guide" --output ./deploy.md

  # Pull with metadata headers
  outline pull --collection "Ops" --output ./ops/ --with-metadata
```

### Options

```
      --collection string   Collection to pull (name, slug, or UUID)
      --doc string          Pull a single document by title
  -h, --help                help for pull
  -o, --output string       Output directory or file path (default ".")
      --with-metadata       Include metadata comment headers in output
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

