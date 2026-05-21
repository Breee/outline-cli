---
title: "Metadata Headers"
weight: 47
description: "Control document title, collection, icon, and parent using HTML comment metadata."
llmsDescription: "outline-cli metadata headers are HTML comments at the top of markdown files. Format: `<!-- Key: Value -->`. Supported keys: Collection (target collection name/slug/UUID, overrides --collection-id flag), Title (explicit document title), Icon (emoji or icon name for Outline UI), Parent (parent document title for nesting). Title resolution order: <!-- Title --> header > first # H1 heading > filename without extension. Example: `<!-- Collection: Engineering -->\n<!-- Title: Deploy Guide -->\n<!-- Icon: 🚀 -->\n<!-- Parent: Operations -->`"
---


Control per-file behavior using HTML comments at the top of your markdown files.

## Format

```markdown
<!-- Collection: Engineering -->
<!-- Title: Deployment Guide -->
<!-- Icon: 🚀 -->
<!-- Parent: Operations Runbooks -->

# Deployment Guide

Your content here...
```

## Supported Headers

| Header | Description | Example |
|--------|-------------|---------|
| `Collection` | Target collection (overrides `--collection-id`) | `<!-- Collection: Engineering -->` |
| `Title` | Explicit document title | `<!-- Title: Getting Started -->` |
| `Icon` | Document icon in Outline UI | `<!-- Icon: 🚀 -->` |
| `Parent` | Parent document for nesting | `<!-- Parent: User Guide -->` |

## Title Resolution

The document title is determined in this order:

1. `<!-- Title: ... -->` header (if present)
2. First `# H1` heading in the file
3. Filename without extension (e.g., `deploy-guide.md` → `deploy-guide`)

## Collection Override

Each file can target a different collection, useful for mixed-content directories:

```markdown
<!-- Collection: Operations -->
# Runbook: Database Failover
...
```

```markdown
<!-- Collection: Engineering -->
# Architecture Decision: Event Sourcing
...
```

## Parent Documents

Explicit parent assignment for nesting:

```markdown
<!-- Parent: User Guide -->
<!-- Title: Installation -->

# Installation
...
```

This creates `Installation` as a child of `User Guide` in the Outline tree.

{{< callout type="info" >}}
The parent document must already exist (or be pushed in the same batch). See [Directory Structure]({{< relref "directories" >}}) for automatic parent resolution.
{{< /callout >}}

## Icons

Use emoji or Outline-supported icon names:

```markdown
<!-- Icon: 🚀 -->
<!-- Icon: 📝 -->
<!-- Icon: ⚙️ -->
```
