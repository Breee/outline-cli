---
title: "CI/CD Integration"
weight: 48
description: "Automatically publish documentation on every git push using GitHub Actions or GitLab CI."
llmsDescription: "outline-cli CI/CD integration: use env vars OUTLINE_SERVER_URL and OUTLINE_API_TOKEN (no config file or keyring needed). GitHub Actions can install the release tarball or run in `ghcr.io/breee/outline-cli:latest`. GitLab CI can use the published container image directly and run `outline push`. Example GitHub workflow triggers on push to main with path filter on docs/**. The CLI exits 0 on success, non-zero on failure for CI gate usage."
---


Publish documentation automatically on every git push. No config file or keyring needed — environment variables are all that's required.

## GitHub Actions

### Using the Action (Recommended)

```yaml
name: Publish Docs

on:
  push:
    branches: [main]
    paths: ['docs/**']

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6

      - name: Install outline-cli
        run: |
          curl -sL "https://github.com/Breee/outline-cli/releases/latest/download/outline_linux_amd64.tar.gz" | sudo tar xz -C /usr/local/bin outline

      - name: Push docs to Outline
        env:
          OUTLINE_SERVER_URL: ${{ vars.OUTLINE_URL }}
          OUTLINE_API_TOKEN: ${{ secrets.OUTLINE_API_TOKEN }}
        run: |
          outline push \
            --collection-id "Engineering" \
            --path ./docs/ \
            --create-collection
```

### Setup

1. Go to your Outline instance → **Settings → API** → create a token
2. In your GitHub repo → **Settings → Secrets and variables → Actions**:
   - Add secret: `OUTLINE_API_TOKEN` = your token
   - Add variable: `OUTLINE_URL` = `https://outline.example.com`

## GitLab CI

```yaml
publish-docs:
  image: ghcr.io/breee/outline-cli:latest
  stage: deploy
  only:
    changes:
      - docs/**
    refs:
      - main
  script:
    - outline push --collection-id "$OUTLINE_COLLECTION" --path ./docs/ --create-collection
  variables:
    OUTLINE_SERVER_URL: https://outline.example.com
    # OUTLINE_API_TOKEN set in CI/CD settings
```

## Generic CI

For any CI system that can run shell commands:

```bash
#!/bin/bash
set -e

# Option 1: install the release binary
curl -sL "https://github.com/Breee/outline-cli/releases/latest/download/outline_linux_amd64.tar.gz" | tar xz -C /usr/local/bin

# Push (env vars must be set: OUTLINE_SERVER_URL, OUTLINE_API_TOKEN)
outline push --collection-id "${OUTLINE_COLLECTION}" --path ./docs/ --create-collection
```

Or use the published container image directly:

```bash
docker run --rm \
  -e OUTLINE_SERVER_URL \
  -e OUTLINE_API_TOKEN \
  -v "$PWD:/workspace" \
  -w /workspace \
  ghcr.io/breee/outline-cli:latest \
  push --collection-id "${OUTLINE_COLLECTION}" --path ./docs/ --create-collection
```

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `OUTLINE_SERVER_URL` | Yes | Outline server URL |
| `OUTLINE_API_TOKEN` | Yes | API token |
| `OUTLINE_COLLECTION` | No | Default collection (can use `--collection-id` flag instead) |

## Conditional Publishing

Only publish when docs actually change:

```yaml
on:
  push:
    branches: [main]
    paths:
      - 'docs/**'
      - '*.md'
```

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | All documents pushed successfully |
| `1` | Error (auth failure, network error, invalid config) |

Use this for CI gates — a failed push breaks the build.
