<!-- Collection: test -->
<!-- Title: Development Guide -->

# Development Guide

## Prerequisites

- Go 1.25+
- Docker & Docker Compose
- Traefik running on the `proxy` network

## Dev Environment

```bash
make dev-up    # Start postgres, redis, dex, outline
make dev-down  # Stop all services
make dev-logs  # Tail logs
```

## Building

```bash
make build     # Build binary with version info
make test      # Run unit tests
make vet       # Run go vet
```

## Testing the Auth Flow

```bash
./outline --server-url http://outline.localhost auth oidc-login
```

Login with: `admin@example.com` / `password`

## Pushing Documents

```bash
# Push a single file
./outline --server-url http://outline.localhost push --path README.md --collection-id test

# Push a directory (each .md file becomes a document)
./outline --server-url http://outline.localhost push --path ai-docs/
```
