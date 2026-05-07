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

## Usage

```bash
go run . --server-url https://outline.example.com --api-token "$OUTLINE_API_TOKEN" push --collection-id <collection-id> --path ./docs
```

### OIDC login

```bash
go run . auth oidc-login --issuer https://issuer.example.com --client-id outline-cli
```

This stores the OIDC access token at `~/.config/outline-cli/config.json`.

## Development

```bash
make test
make e2e
```
