---
title: "Directory Structure"
description: "How directory layout maps to document hierarchy in Outline."
llmsDescription: "outline-cli maps directory structure to Outline document hierarchy. Rules: index.md or README.md in a directory becomes the parent document for all sibling .md files in that directory. Subdirectories create nested levels. Files without a directory parent become top-level documents in the collection. Sort order: shallower depth first, index/README first, then alphabetical. Example: docs/guide/index.md becomes parent, docs/guide/install.md and docs/guide/setup.md become children of that parent."
---

# Directory Structure

When pushing a directory, the folder layout automatically creates document hierarchy in Outline.

## Rules

1. `index.md` or `README.md` in a directory → becomes the **parent document**
2. Other `.md` files in the same directory → become **children** of that parent
3. Subdirectories → create **nested levels**
4. Files at the root of the push path → become **top-level** collection documents

## Example

Given this directory:

```
docs/
├── index.md              → "Documentation" (top-level parent)
├── quickstart.md         → "Quick Start" (child of Documentation)
├── guide/
│   ├── index.md          → "Guide" (child of Documentation, parent of below)
│   ├── install.md        → "Installation" (child of Guide)
│   └── auth.md           → "Authentication" (child of Guide)
└── reference/
    ├── index.md          → "Reference" (child of Documentation, parent of below)
    └── commands.md       → "Commands" (child of Reference)
```

Results in this Outline structure:

```
Collection: Engineering
├── Documentation
│   ├── Quick Start
│   ├── Guide
│   │   ├── Installation
│   │   └── Authentication
│   └── Reference
│       └── Commands
```

## Sort Order

Documents are pushed in this order (ensuring parents exist before children):

1. Shallower depth first
2. `index.md` / `README.md` before other files at the same level
3. Alphabetical within the same priority

## Overriding with Metadata

You can override the automatic parent with a `<!-- Parent: ... -->` header:

```markdown
<!-- Parent: Reference -->
<!-- Title: API Reference -->

# API Reference
```

This places the document under "Reference" regardless of its file location.

## Tips

- Use `index.md` for section landing pages (they become group parents)
- Keep directory names meaningful — they don't directly map to titles, but help organize your source
- The title always comes from the file content, not the directory name
