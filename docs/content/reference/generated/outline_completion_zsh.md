---
title: "outline completion zsh"
description: "CLI reference for outline completion zsh"
llmsDescription: "Auto-generated CLI reference for outline completion zsh. Contains usage, flags, and examples."
generated: true
---

## outline completion zsh

Generate the autocompletion script for zsh

### Synopsis

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(outline completion zsh)

To load completions for every new session, execute once:

#### Linux:

	outline completion zsh > "${fpath[1]}/_outline"

#### macOS:

	outline completion zsh > $(brew --prefix)/share/zsh/site-functions/_outline

You will need to start a new shell for this setup to take effect.


```
outline completion zsh [flags]
```

### Options

```
  -h, --help              help for zsh
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

