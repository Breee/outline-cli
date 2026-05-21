---
title: "Examples"
description: "Example markdown files demonstrating outline-cli features."
llmsDescription: "outline-cli examples directory contains sample markdown files: minimal.md (bare minimum push), title-from-h1.md (title from heading), title-from-filename.md (title from filename), rich-content.md (tables/code/lists), different-collection.md (Collection metadata header), no-headers.md (no metadata at all), markdown-reference.md (full markdown feature showcase), guide/ (directory with index.md parent and nested subdirectories demonstrating hierarchy push)."
---

# Examples

The [`examples/`](https://github.com/Breee/outline-cli/tree/main/examples) directory contains sample markdown files you can use to test outline-cli.

## Single File Examples

| File | Demonstrates |
|------|-------------|
| `minimal.md` | Bare minimum: just content, title from filename |
| `title-from-h1.md` | Title resolved from `# H1` heading |
| `title-from-filename.md` | No heading, title from filename |
| `rich-content.md` | Tables, code blocks, lists |
| `different-collection.md` | `<!-- Collection: ... -->` override |
| `no-headers.md` | No metadata at all |
| `markdown-reference.md` | Full markdown feature showcase |

## Directory Example

```
examples/guide/
├── index.md
├── faq.md
├── setup/
│   ├── index.md
│   └── install.md
└── usage/
    ├── index.md
    └── push.md
```

Push with:

```bash
outline push --path ./examples/guide/ --collection-id "Test" --create-collection
```

This creates the full nested document tree under a "Test" collection.

## Try It

```bash
# Single file
outline push --path ./examples/minimal.md --collection-id "Test" --create-collection

# Directory with hierarchy
outline push --path ./examples/guide/ --collection-id "Test Guide" --create-collection
```
