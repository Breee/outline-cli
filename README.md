# outline-cli

A modern Go Cobra CLI named `outline` for pushing markdown files to an [Outline](https://github.com/outline/outline) wiki.

## Features

- Push one markdown file or full directory trees to Outline
- Authentication support:
  - API token (`--api-token` / `OUTLINE_API_TOKEN`)
  - Basic auth (`--username` and `--password`)
  - OIDC login via device flow (`outline auth oidc-login`)
- Docker Compose based e2e test workflow
- Automated release workflow + Renovate config based on `Breee/kubeswitch`
- Release artifacts for Linux (`amd64`, `arm64`, `armv7`), macOS (`amd64`, `arm64`), and Windows (`amd64`, `arm64`)
- Shell autocompletion for bash, zsh, and fish

## Install

Download binaries from the [latest release](https://github.com/Breee/outline-cli/releases/latest).

Linux/macOS quick install:

```bash
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) ARCH=amd64 ;;
  aarch64|arm64) ARCH=arm64 ;;
  armv7l) ARCH=armv7 ;;
esac
curl -sL "https://github.com/Breee/outline-cli/releases/latest/download/outline_${OS}_${ARCH}.tar.gz" | sudo tar xz -C /usr/local/bin outline
```

## Usage

```bash
go run . --server-url https://outline.example.com --api-token "$OUTLINE_API_TOKEN" push --collection-id <collection-id> --path ./docs
```

### OIDC login

```bash
go run . auth oidc-login --issuer https://issuer.example.com --client-id outline-cli
```

This stores the OIDC access token at `~/.config/outline-cli/config.json`.

### Autocompletion

```bash
# Bash
source <(outline completion bash)

# Zsh
source <(outline completion zsh)

# Fish
outline completion fish | source
```

Release assets also include a bundled `outline_completions.tar.gz`.

## Development

```bash
make test
make e2e
make completions
```
