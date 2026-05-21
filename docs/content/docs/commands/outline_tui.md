---
title: "outline tui"
description: "CLI reference for outline tui"
llmsDescription: "Auto-generated CLI reference for outline tui. Contains usage, flags, and examples."
generated: true
---

## outline tui

Interactive TUI for browsing and reading wiki documents

### Synopsis

Launch an interactive terminal UI for browsing collections and reading
documents from Outline. Optionally pass a search query to jump directly
to search results.

```
outline tui [query] [flags]
```

### Examples

```
  # Open collection browser
  outline tui

  # Jump to search results
  outline tui "deploy guide"

  # Browse a specific collection
  outline tui --collection ops
```

### Options

```
      --collection string   Browse a specific collection
  -h, --help                help for tui
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

