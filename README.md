# outline-cli

A CLI for managing [Outline](https://github.com/outline/outline) wiki documents — push, pull, search, and browse from your terminal.

**[Documentation](https://breee.github.io/outline-cli/)** · **[Install](https://breee.github.io/outline-cli/docs/install/)** · **[Commands Reference](https://breee.github.io/outline-cli/docs/commands/outline/)**

## Features

- **Push** markdown files or directory trees to Outline (upsert by title)
- **Pull** documents back to local markdown
- **Search** from the command line with filtering and multiple output formats
- **Interactive TUI** for browsing and reading documents with rendered markdown
- **Get** documents rendered in the terminal (`outline get documents <title>`)
- Metadata headers (`<!-- Collection: ... -->`, `<!-- Title: ... -->`, `<!-- Parent: ... -->`) for fine-grained control
- Directory structure maps to document hierarchy (`index.md` = parent doc)
- Automatic image upload and URL rewriting
- Auth: API token, OIDC browser login (PKCE), basic auth — with OS keyring storage
- Shell completion for bash, zsh, fish with tab-complete for document titles
- Cross-platform: Linux, macOS, Windows (`amd64`, `arm64`, `armv7`)

## Quick Start

```bash
# Install (Linux/macOS amd64 — see docs for other architectures)
curl -sL "https://github.com/Breee/outline-cli/releases/latest/download/outline_$(uname -s | tr '[:upper:]' '[:lower:]')_amd64.tar.gz" | sudo tar xz -C /usr/local/bin outline

# Configure
outline config set server_url https://outline.example.com
outline config set api_token sk-your-token

# Push docs
outline push --path ./docs --collection-id "Engineering"

# Pull docs
outline pull --collection "Engineering" --output ./local-docs/

# Search
outline search "deployment guide"

# Browse interactively
outline tui
```

## CI/CD

No config file or keyring needed — just set environment variables:

```yaml
env:
  OUTLINE_SERVER_URL: https://outline.example.com
  OUTLINE_API_TOKEN: ${{ secrets.OUTLINE_API_TOKEN }}

steps:
  - run: outline push --path ./docs --collection-id "Docs" --yes
```

## Development

```bash
make all        # vet, test, build
make dev-up     # start local Outline + Dex OIDC (Docker)
make dev-down   # tear down
make e2e        # run e2e tests
make docs       # regenerate reference docs
```

See the [development guide](https://breee.github.io/outline-cli/docs/) for details.

## License

[MIT](LICENSE)
