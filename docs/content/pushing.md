---
title: "Pushing Documents"
weight: 45
description: "Push markdown files and directories to Outline collections."
llmsDescription: "outline push command: `outline push --path <file-or-dir> --collection-id <name|slug|uuid>`. Flags: --path (default '.'), --collection-id (required unless set in file metadata), --publish (default true), --create-collection (default false). Behavior: recursively walks directories for .md files, upserts by document title (searches existing, updates if found, creates if not). Sorts: shallower files first, index.md/README.md first, then alphabetical. Supports automatic image upload (local ![alt](path) rewritten to attachment URLs)."
---

# Pushing Documents

The `push` command uploads markdown files to your Outline wiki.

## Basic Usage

```bash
# Push a single file
outline push --path ./deploy.md --collection-id "Engineering"

# Push a directory tree
outline push --path ./docs/ --collection-id "Engineering"

# Auto-create collection if it doesn't exist
outline push --path ./docs/ --collection-id "New Collection" --create-collection
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-p, --path` | `.` | Path to markdown file or directory |
| `--collection-id` | — | Target collection (name, slug, or UUID) |
| `--publish` | `true` | Publish created documents |
| `--create-collection` | `false` | Create collection if it doesn't exist |

## Upsert Behavior

`push` is idempotent. It searches for an existing document by title:

- **Found** → updates the content
- **Not found** → creates a new document

This means you can run `push` repeatedly (e.g., in CI) without creating duplicates.

## Collection Resolution

The `--collection-id` flag accepts:

- Collection name: `"Engineering"`
- Collection slug: `engineering`
- URL ID: `engineering-abc123`
- UUID: `550e8400-e29b-41d4-a716-446655440000`

## Image Upload

Local image references are automatically uploaded:

```markdown
![architecture diagram](./images/arch.png)
```

Becomes an Outline attachment with the URL rewritten in the pushed document.

## What Gets Pushed

When pushing a directory:

- All `.md` files are included (recursive)
- Non-markdown files are ignored (except images referenced in markdown)
- Hidden files/directories (`.git`, `.DS_Store`) are skipped
- Sort order: shallower files first, then `index.md`/`README.md` first, then alphabetical
