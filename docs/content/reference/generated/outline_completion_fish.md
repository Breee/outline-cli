---
title: "outline completion fish"
description: "CLI reference for outline completion fish"
llmsDescription: "Auto-generated CLI reference for outline completion fish. Contains usage, flags, and examples."
generated: true
---

## outline completion fish

Generate the autocompletion script for fish

### Synopsis

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	outline completion fish | source

To load completions for every new session, execute once:

	outline completion fish > ~/.config/fish/completions/outline.fish

You will need to start a new shell for this setup to take effect.


```
outline completion fish [flags]
```

### Options

```
  -h, --help              help for fish
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

