# Feature: Push Confirmation & Diff

## Summary

When pushing to a collection that already contains documents, show the user what will change and ask for confirmation before proceeding. Skip confirmation with `--yes` / `-y`.

## Motivation

Currently `outline push` silently upserts documents. If a collection already has content, the user has no visibility into what will be created vs updated vs left unchanged. This can lead to accidental overwrites.

## Behavior

### Default (interactive)

```bash
outline push --collection-id "Engineering" --path ./docs/
```

Output:
```
Collection "Engineering" contains 12 existing documents.

Changes:
  update  "Getting Started"        (content changed)
  update  "API Reference"          (content changed)
  create  "Migration Guide"        (new document)
  unchanged  "FAQ"                 (no changes)
  unchanged  "Architecture"        (no changes)

3 documents will be modified, 1 created, 2 unchanged.
Proceed? [y/N]
```

### Force mode (non-interactive / CI)

```bash
outline push --collection-id "Engineering" --path ./docs/ -y
outline push --collection-id "Engineering" --path ./docs/ --yes
```

Skips confirmation, pushes immediately. Required for CI/CD pipelines.

### Verbose diff

```bash
outline push --collection-id "Engineering" --path ./docs/ --diff
```

Shows inline content diff for each changed document before prompting.

## Implementation

### New flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--yes` | `-y` | `false` | Skip confirmation prompt |
| `--diff` | | `false` | Show content diff for changed documents |

### Logic

1. Resolve collection (existing behavior)
2. Fetch all documents in the target collection via Outline API
3. For each local file:
   - Match by title to existing document
   - If matched: compare content (strip metadata headers) → classify as `update` or `unchanged`
   - If not matched: classify as `create`
4. If any `update` or `create` exists and `--yes` is not set:
   - Print summary table
   - If `--diff`: print unified diff for each `update`
   - Prompt user for confirmation
   - Abort on `n` / non-`y` input
5. Proceed with push (existing upsert logic)

### API calls needed

- `documents.list` with `collectionId` — get existing docs in collection
- `documents.info` — get full content for diff comparison (if `--diff`)

### Edge cases

- Empty collection → no confirmation needed, push directly
- `--create-collection` with new collection → no confirmation needed
- Non-interactive terminal (no TTY) without `--yes` → error with message to use `--yes`
- Large collections (100+ docs) → paginate API calls, summarize instead of listing all unchanged

## Testing

| Scenario | Package | Type |
|----------|---------|------|
| Empty collection skips confirmation | `cmd` | Unit |
| New collection skips confirmation | `cmd` | Unit |
| `--yes` flag skips confirmation | `cmd` | Unit |
| Non-TTY without `--yes` errors | `cmd` | Unit |
| Diff output matches expected format | `cmd` | Unit |
| Update/create/unchanged classification | `cmd` | Unit |
| E2E: push with confirmation | `test/e2e` | E2E |
