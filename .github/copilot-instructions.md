# outline-cli

A Go CLI for pushing markdown documents to Outline wiki.

## Build & Test

```bash
make all          # vet, test, build (default)
make build        # build binary
make test         # run unit tests
make vet          # run go vet
make e2e          # run e2e tests
make completions  # generate shell completions
make docs         # generate all docs (reference, llms.txt, LLM instructions)
make dev-up       # start dev environment
make dev-down     # stop dev environment
make dev-logs     # tail dev logs
make clean        # remove binary
make help         # show this help
```

## Architecture
- `cmd/` — CLI commands
- `internal/config/` — configuration, registry, keyring
- `internal/oidc/` — OIDC browser/device auth flows
- `internal/outline/` — Outline API client

## Key Patterns

- Auth priority: CLI flags > env vars > OS keyring > config file
- Secrets (api_token, password, oidc token) stored in OS keyring via `github.com/zalando/go-keyring`
- Push uses upsert: searches by title, updates if found, creates if not
- Metadata headers in markdown: `<!-- Collection: X -->`, `<!-- Title: X -->`, `<!-- Parent: X -->`
- Directory structure maps to document hierarchy (index.md = parent)

## Configuration

| Key | Env Var | Secret | Description |
|-----|---------|--------|-------------|
| `server_url` | `OUTLINE_SERVER_URL` | no | Outline server base URL |
| `auth_method` | `OUTLINE_AUTH_METHOD` | no | Auth method: api_token, oidc, basic |
| `token_storage` | — | no | Secret storage backend: keyring, file |
| `oidc_port` | — | no | Local port for OIDC callback |
| `api_token` | `OUTLINE_API_TOKEN` | yes | API bearer token |
| `password` | `OUTLINE_PASSWORD` | yes | Basic auth password |

## Docs Generation

Reference docs are auto-generated from the cobra command tree. Never hand-edit files in `docs/reference/generated/`.

```bash
make docs  # runs: go run . docs generate
```

This produces:
- `docs/reference/generated/*.md` — CLI reference pages
- `llms.txt` — structured index for AI tools
- `llms-full.txt` — complete docs as single file
- `.github/copilot-instructions.md`, `CLAUDE.md`, `.cursor/rules` — AI tool instructions
