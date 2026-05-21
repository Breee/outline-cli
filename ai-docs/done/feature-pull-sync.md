<!-- Collection: test -->
<!-- Title: Feature: Pull & Bidirectional Sync -->

# Feature: Pull & Bidirectional Sync

## Why This Matters

Docs-as-code only works in one direction today (local → Outline). But the real power is bidirectional:
- Developers edit in their IDE/editor with git version control
- Non-developers edit in the Outline web UI
- Both stay in sync without conflicts

This is what makes a wiki CLI truly indispensable — it becomes the bridge between the "docs in git" and "docs in wiki" worlds.

## Commands

### `outline pull`

Pull documents from Outline to local markdown files.

```bash
# Pull entire collection to a directory
outline pull --collection "Engineering" --output ./docs/engineering/

# Pull a single document
outline pull --doc "Deployment Guide" --output ./docs/deploy.md

# Pull with metadata headers preserved
outline pull --collection "Ops" --output ./ops/ --with-metadata
```

### `outline sync`

Bidirectional sync: detect which side changed and reconcile.

```bash
# Sync a directory with a collection
outline sync --collection "Engineering" --path ./docs/

# Dry run — show what would change without doing it
outline sync --collection "Engineering" --path ./docs/ --dry-run

# Force local wins (useful in CI: git is source of truth)
outline sync --collection "Engineering" --path ./docs/ --strategy local-wins

# Force remote wins (useful for initial pull)
outline sync --collection "Engineering" --path ./docs/ --strategy remote-wins
```

### `outline diff`

Show differences between local files and remote documents.

```bash
outline diff --path ./docs/ --collection "Engineering"
```

Output:
```
 M docs/deploy.md          — local modified (remote unchanged)
 R docs/monitoring.md      — remote modified (local unchanged)
 C docs/auth.md            — conflict (both modified)
 + docs/new-feature.md     — local only (will create)
 - runbook.md              — remote only (not in local)
```

## Sync Strategy

### Tracking State

Store sync metadata in `.outline-sync.yaml` at the directory root:

```yaml
collection: Engineering
collection_id: abc-123
last_sync: 2026-05-20T10:30:00Z
documents:
  docs/deploy.md:
    id: doc-456
    remote_updated_at: 2026-05-19T08:00:00Z
    local_hash: sha256:abcdef...
  docs/monitoring.md:
    id: doc-789
    remote_updated_at: 2026-05-20T09:00:00Z
    local_hash: sha256:123456...
```

### Conflict Detection

1. Compare local file hash against stored `local_hash` → local changed?
2. Compare remote `updatedAt` against stored `remote_updated_at` → remote changed?
3. If both changed → conflict

### Conflict Resolution

| Strategy | Behavior |
|----------|----------|
| `local-wins` | Local file always overwrites remote (CI/CD mode) |
| `remote-wins` | Remote always overwrites local (initial setup) |
| `interactive` | Prompt user for each conflict (default) |
| `skip` | Skip conflicts, sync non-conflicting files |

## File Format on Pull

```markdown
<!-- outline-id: doc-456 -->
<!-- Collection: Engineering -->
<!-- Title: Deployment Guide -->
<!-- Icon: 🚀 -->

# Deployment Guide

Content here...
```

The `outline-id` comment enables reliable matching even if title changes.

## Implementation Plan

1. Add `documents.list` (by collection) and `documents.info` (full content) to API client
2. Create `internal/sync/` package:
   - `state.go` — load/save `.outline-sync.yaml`
   - `diff.go` — compute local vs remote differences
   - `pull.go` — download and write files
   - `push.go` — reuse existing push logic
   - `resolve.go` — conflict resolution strategies
3. Create `cmd/pull.go`, `cmd/sync.go`, `cmd/diff.go`

## Testing

- Unit test diff detection with various change combinations
- Unit test conflict resolution strategies
- Integration test: push → modify remote → pull → verify
- Golden file tests for `.outline-sync.yaml` format
- Test that `outline-id` survives round-trip (push + pull)
