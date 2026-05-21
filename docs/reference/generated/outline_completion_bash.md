---
title: "outline completion bash"
description: "CLI reference for outline completion bash"
llmsDescription: "Auto-generated CLI reference for outline completion bash. Contains usage, flags, and examples."
generated: true
---

## outline completion bash

Generate the autocompletion script for bash

### Synopsis

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(outline completion bash)

To load completions for every new session, execute once:

#### Linux:

	outline completion bash > /etc/bash_completion.d/outline

#### macOS:

	outline completion bash > $(brew --prefix)/etc/bash_completion.d/outline

You will need to start a new shell for this setup to take effect.


```
outline completion bash
```

### Options

```
  -h, --help              help for bash
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

