---
title: "outline get collections"
description: "CLI reference for outline get collections"
llmsDescription: "Auto-generated CLI reference for outline get collections. Contains usage, flags, and examples."
generated: true
---

## outline get collections

List all collections

```
outline get collections [flags]
```

### Examples

```
  # List collections (default table)
  outline get collections

  # JSON output
  outline get collections -o json

  # YAML output
  outline get collections -o yaml
```

### Options

```
  -h, --help            help for collections
  -o, --output string   Output format: table, json, yaml (default "table")
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

