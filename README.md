# outline-cli

A modern Go Cobra CLI named `outline` for pushing markdown files to an [Outline](https://github.com/outline/outline) wiki.

**[Documentation](https://breee.github.io/outline-cli/)**

## Features

- Push one markdown file or full directory trees to Outline
- Upsert logic: updates existing documents by title, creates new ones if not found
- Per-file metadata headers (`<!-- Collection: ... -->`, `<!-- Title: ... -->`, `<!-- Icon: ... -->`, `<!-- Parent: ... -->`)
- Automatic parent/child nesting via directory structure (`index.md`/`README.md` become parent docs)
- Automatic image upload: local image references are uploaded as attachments and URLs rewritten
- Collection resolution by name, slug, urlId, or UUID
- Auto-create collections with `--create-collection`
- Authentication support:
  - API token (`--api-token` / `OUTLINE_API_TOKEN`)
  - Basic auth (`--username` / `--password`)
  - OIDC browser login with PKCE + Dynamic Client Registration (`outline auth oidc-login`)
- Auto-reauthentication when token expires (configurable via `auth_method`)
- Secure credential storage in OS keyring (API token, password, OIDC token) with file fallback
- YAML config file at `~/.outline-cli/config.yaml`
- Release artifacts for Linux (`amd64`, `arm64`, `armv7`), macOS (`amd64`, `arm64`), and Windows (`amd64`, `arm64`)
- Shell autocompletion for bash, zsh, and fish
- Interactive TUI (`outline tui`) for browsing and reading wiki documents with rendered markdown
- `outline get documents <title>` prints glamour-rendered markdown (use `-o raw` for source)

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

### Push

```bash
outline --server-url https://outline.example.com --api-token "$OUTLINE_API_TOKEN" \
  push --collection-id <collection-id> --path ./docs
```

| Flag | Default | Description |
|------|---------|-------------|
| `-p, --path` | `.` | Markdown file or directory to push |
| `--collection-id` | — | Target collection (name, slug, or UUID) |
| `--publish` | `true` | Publish created documents |
| `--create-collection` | `false` | Create collection if it doesn't exist |

### Metadata Headers

Add HTML comments at the top of markdown files to control behavior:

```markdown
<!-- Collection: My Wiki -->
<!-- Title: Getting Started -->
<!-- Icon: 🚀 -->
<!-- Parent: User Guide -->

# Getting Started
...
```

Title resolution order: `<!-- Title -->` → first `# H1` → filename (without extension).

### OIDC Login

```bash
outline auth oidc-login --port 10800
```

Opens a browser for OAuth2 authorization (PKCE + Dynamic Client Registration). The access token is stored in the OS keyring.

### Auth Check

```bash
outline auth check
```

Prints the authenticated user name, email, and team.

### Config Management

```bash
outline config set server_url https://outline.example.com
outline config set auth_method oidc
outline config set api_token sk-abc123   # stored in OS keyring
outline config set password s3cret       # stored in OS keyring
outline config get server_url
outline config list                      # secrets are masked
outline config path
```

Valid keys: `server_url`, `auth_method`, `token_storage`, `oidc_port`, `api_token`, `password`

Secrets (`api_token`, `password`) are stored in the OS keyring by default. Set `token_storage=file` to fall back to plaintext (for headless/CI environments).

### Global Flags

| Flag | Env Var | Description |
|------|---------|-------------|
| `--config` | — | Config file path (default `~/.outline-cli/config.yaml`) |
| `--server-url` | `OUTLINE_SERVER_URL` | Outline server URL |
| `--api-token` | `OUTLINE_API_TOKEN` | API token |
| `--username` | `OUTLINE_USERNAME` | Basic auth username |
| `--password` | `OUTLINE_PASSWORD` | Basic auth password |
| `--oidc-access-token` | `OUTLINE_OIDC_ACCESS_TOKEN` | OIDC access token |

Credential resolution order (highest priority first):
1. CLI flags
2. Environment variables
3. OS keyring
4. Config file

### CI/CD Usage

For CI/CD pipelines, use environment variables — no config file or keyring needed:

```yaml
env:
  OUTLINE_SERVER_URL: https://outline.example.com
  OUTLINE_API_TOKEN: ${{ secrets.OUTLINE_API_TOKEN }}
```

```bash
outline push --collection-id my-docs --path ./docs
```

### Interactive TUI

```bash
# Browse collections and documents
outline tui

# Jump to search results
outline tui "deploy guide"

# Browse a specific collection
outline tui --collection ops
```

Key bindings: `/` search, `j/k` navigate, `enter` open, `esc` back, `q` quit.
Documents are rendered with syntax-highlighted markdown. Search is live with debounce.

### Get Document (Rendered)

```bash
# Rendered markdown output (default)
outline get documents "Deployment Guide"

# Raw markdown (pipe-friendly)
outline get documents "Deployment Guide" -o raw | less

# JSON metadata
outline get documents "Deployment Guide" -o json
```

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

### Dev environment (real Outline + Dex OIDC)

Requires: Docker.

```bash
make dev-up    # start traefik, postgres, redis, dex, outline
make dev-logs  # follow logs
make dev-down  # tear down
```

| Service | URL |
|---------|-----|
| Outline | http://outline.localhost |
| Dex     | http://dex.localhost |
| Traefik | http://traefik.localhost |

Login with `admin@example.com` / `password` via the Dex OIDC flow.

Then test the CLI:

```bash
go run . --server-url http://outline.localhost --api-token <token> push --collection-id <id> --path ./docs
```

### Tests

```bash
make test
make e2e
make completions
```

### Build

```bash
make build   # produces ./outline binary with version info
make clean   # remove binary
```

## Examples

See the [examples/](examples/) directory for usage patterns:

- `minimal.md` — minimal metadata
- `different-collection.md` — per-file collection override
- `title-from-h1.md` / `title-from-filename.md` — title resolution demos
- `rich-content.md` — tables, code blocks, full API reference
- `markdown-reference.md` — complete metadata header reference
- `guide/` — hierarchical directory push with nested parents
