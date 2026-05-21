---
title: "outline auth oidc-login"
description: "CLI reference for outline auth oidc-login"
llmsDescription: "Auto-generated CLI reference for outline auth oidc-login. Contains usage, flags, and examples."
generated: true
---

## outline auth oidc-login

Login via browser-based OIDC flow through Outline

### Synopsis

Initiates Outline's OIDC login flow in your browser.
A local HTTP server captures the callback and obtains an Outline session token.
Requires --server-url to be set (the Outline instance URL).

```
outline auth oidc-login [flags]
```

### Options

```
  -h, --help       help for oidc-login
      --port int   Local port for OIDC callback server (default 10800)
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

* [outline auth](outline_auth.md)	 - Authentication helpers

