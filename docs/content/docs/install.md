---
title: "Installation"
weight: 10
description: "Install outline-cli on Linux, macOS, Windows, or from the published container image."
llmsDescription: "Install outline-cli: download binary from GitHub releases. Linux/macOS: curl tar.gz to /usr/local/bin. Windows: download zip. Container images are available at `ghcr.io/breee/outline-cli` with both `latest` and versioned release tags. Also available via `go install github.com/Breee/outline-cli@latest`. Requires no runtime dependencies. Shell completions available for bash, zsh, fish via `outline completion <shell>`."
---


## Binary Download (Recommended)

Download the latest release for your platform from [GitHub Releases](https://github.com/Breee/outline-cli/releases/latest).

### Linux / macOS

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

### Windows

Download `outline_windows_amd64.zip` from the [releases page](https://github.com/Breee/outline-cli/releases/latest) and add `outline.exe` to your PATH.

## Go Install

```bash
go install github.com/Breee/outline-cli@latest
```

## Container Image

Use the published image from GitHub Container Registry:

```bash
docker run --rm ghcr.io/breee/outline-cli:latest version
```

For reproducible environments, replace `latest` with a release tag such as `vX.Y.Z`.

## Build from Source

```bash
git clone https://github.com/Breee/outline-cli.git
cd outline-cli
make build
# Binary at ./outline
```

## Shell Completions

Source directly in your current session:

```bash
# Bash
source <(outline completion bash)

# Zsh
source <(outline completion zsh)

# Fish
outline completion fish | source
```

Or install persistently:

```bash
# Bash
outline completion bash > /etc/bash_completion.d/outline

# Zsh
outline completion zsh > "${fpath[1]}/_outline"

# Fish
outline completion fish > ~/.config/fish/completions/outline.fish
```

## Verify Installation

```bash
outline --help
```
