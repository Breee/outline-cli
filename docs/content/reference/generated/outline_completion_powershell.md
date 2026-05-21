---
title: "outline completion powershell"
description: "CLI reference for outline completion powershell"
llmsDescription: "Auto-generated CLI reference for outline completion powershell. Contains usage, flags, and examples."
generated: true
---

## outline completion powershell

Generate the autocompletion script for powershell

### Synopsis

Generate the autocompletion script for powershell.

To load completions in your current shell session:

	outline completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.


```
outline completion powershell [flags]
```

### Options

```
  -h, --help              help for powershell
      --no-descriptions   disable completion descriptions
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

* [outline completion](outline_completion.md)	 - Generate the autocompletion script for the specified shell

