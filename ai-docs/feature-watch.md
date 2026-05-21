<!-- Collection: test -->
<!-- Title: Feature: Watch Mode (Live Push) -->

# Feature: Watch Mode (Live Push)

## Why This Matters

Writing docs should feel as fluid as writing code with hot reload. Save a file → it's live on the wiki in seconds. No manual `push` commands, no forgetting to update the wiki. The docs are always current.

## Command

### `outline watch`

```bash
# Watch a directory and auto-push on changes
outline watch --path ./docs/ --collection "Engineering"

# Watch with debounce (wait for burst of saves to settle)
outline watch --path ./docs/ --collection "Engineering" --debounce 2s

# Watch specific file
outline watch --path ./docs/deploy.md --collection "Engineering"
```

### Output

```
[watch] Watching ./docs/ for changes...
[watch] docs/deploy.md changed → pushing... done (updated "Deployment Guide")
[watch] docs/new-page.md created → pushing... done (created "New Page")
[watch] docs/old-page.md deleted → skipping (use --delete to remove from wiki)
```

## Behavior

- Uses `fsnotify` for filesystem watching
- Debounces rapid saves (default 1s) to avoid hammering the API
- Only pushes changed files (compares content hash)
- Respects metadata headers (same as `push`)
- Shows live status in terminal (file, action, result)
- Ctrl+C to stop gracefully

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--path` | `.` | Path to watch |
| `--collection-id` | — | Target collection |
| `--debounce` | `1s` | Wait time after last change before pushing |
| `--delete` | `false` | Delete remote docs when local file deleted |
| `--ignore` | — | Glob patterns to ignore (e.g. `*.tmp`) |

## Implementation Plan

1. Add `github.com/fsnotify/fsnotify` dependency
2. Create `cmd/watch.go`:
   - Setup fsnotify watcher on path (recursive)
   - Debounce timer per file
   - On trigger: run same logic as `push` for that single file
   - Track file hashes to skip no-op saves (editor write without change)
3. Reuse existing push infrastructure (`documentTitle`, `parseMetadata`, upsert logic)

## Edge Cases

- Editor creates temp files (`.swp`, `~`, `.tmp`) → ignore by default
- File renamed → treat as delete + create (if `--delete` enabled)
- Permission errors → log warning, continue watching
- Network offline → queue changes, retry when connection restored

## Testing

- Unit test debounce logic
- Unit test file change detection (create, modify, delete, rename)
- Unit test ignore patterns
- Integration test: start watch, write file, verify push occurred
