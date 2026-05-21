<!-- Collection: test -->
<!-- Title: Feature: Search -->

# Feature: Search

## Why This Matters

The #1 reason people can't find docs: search sucks. Outline has full-text search but you have to open a browser. A CLI search that returns results instantly — with context snippets — makes docs feel as accessible as `grep`.

## Commands

### `outline search <query>`

```bash
# Basic search
outline search "kubernetes rollback"

# Filter by collection
outline search "deploy" --collection infrastructure

# Output formats
outline search "API key" --format json      # for scripting
outline search "auth" --format oneline      # compact: title + URL

# Limit results
outline search "setup" --limit 5
```

### Output (default)

```
 1. Deployment Guide (Infrastructure)
    ...to perform a **rollback**, run `kubectl rollout undo`...
    Updated: 3 days ago | URL: https://outline.example.com/doc/deploy-abc123

 2. Kubernetes Runbook (Operations)
    ...if **rollback** fails, check the pod events...
    Updated: 2 weeks ago | URL: https://outline.example.com/doc/k8s-runbook-def456
```

### `outline search --interactive`

Opens TUI search (same as pressing `/` in `outline tui`). Live results as you type, press Enter to read.

## Implementation

- Uses Outline API `documents.search` endpoint
- Highlights matching terms in snippets (bold/color in terminal)
- Respects `--server-url` and auth config (same as push)
- JSON output includes: id, title, collection, snippet, url, updatedAt

## Scripting Integration

```bash
# Open first result in browser
outline search "deploy guide" --format json | jq -r '.[0].url' | xargs open

# Find all docs mentioning a deprecated API
outline search "v1/legacy" --format oneline | wc -l

# Pipe doc content to fzf
outline search "setup" --format json | jq -r '.[].title' | fzf | xargs outline cat
```

## Testing

- Mock API responses for search endpoint
- Test output formatting (default, json, oneline)
- Test snippet highlighting
- Integration test with dev environment
