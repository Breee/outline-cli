---
title: "outline config set"
description: "CLI reference for outline config set"
llmsDescription: "Auto-generated CLI reference for outline config set. Contains usage, flags, and examples."
generated: true
---

## outline config set

Set a config key to a value

### Synopsis

Set a config key to a value.

Available keys:
  server_url
  auth_method
  token_storage
  oidc_port
  api_token
  password

```
outline config set <key> <value> [flags]
```

### Options

```
  -h, --help   help for set
```

### Options inherited from parent commands

```
      --api-token string           Outline API token
      --config string              config file (default "/home/bree/.outline-cli/config.yaml")
      --oidc-access-token string   OIDC access token
      --password string            Basic auth password
      --server-url string          Outline server URL
      --username string            Basic auth username
```

### SEE ALSO

* [outline config](outline_config.md)	 - Manage CLI configuration

