<!-- Title: Feature: Static Site Export -->

# Feature: Static Site Export (`outline site`)

## Concept

Eat our own dogfood: push markdown docs to Outline, then export the rendered collection as a static site for GitHub Pages.

```
docs/*.md → outline push → Outline wiki → outline site export → static HTML → GitHub Pages
```

## Why

- Single source of truth lives in Outline (what we already push to)
- Outline already renders markdown beautifully (ToC, syntax highlighting, embeds)
- No SSG dependency (no MkDocs, Hugo, Docusaurus — just our own CLI)
- Dogfooding validates the CLI on every docs deploy

## Proposed Commands

```bash
# Export a collection as static HTML
outline site export --collection-id "Docs" --output ./site/

# Export as raw markdown tree (for AI consumption)
outline site export --collection-id "Docs" --output ./site/ --format markdown

# Serve locally for preview
outline site serve --collection-id "Docs" --port 8080
```

## Implementation Sketch

1. `outline site export`:
   - List all documents in collection (recursive, preserving hierarchy)
   - For each document: fetch rendered HTML via Outline API
   - Write to output directory preserving parent/child structure as directories
   - Generate index.html with navigation from document tree
   - Copy/download attachments (images)
   - Generate `llms.txt` from document titles + first paragraphs

2. Outline API endpoints needed:
   - `POST /api/collections.documents` — list docs in collection
   - `POST /api/documents.info` — get document content (markdown or HTML)
   - `POST /api/attachments.redirect` — resolve attachment URLs

3. Static site output:
   ```
   site/
   ├── index.html           # collection landing page with nav
   ├── guide/
   │   ├── install.html
   │   ├── install.md        # raw markdown for AI
   │   └── push.html
   ├── llms.txt             # auto-generated from doc tree
   ├── llms-full.txt        # all docs concatenated
   └── assets/              # downloaded images
   ```

## CI/CD Pipeline (Dogfooding)

```yaml
name: Deploy Docs
on:
  push:
    branches: [main]
    paths: ['docs/**']

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install outline-cli
        run: go install github.com/Breee/outline-cli@latest

      - name: Push docs to Outline
        env:
          OUTLINE_HOST: ${{ vars.OUTLINE_URL }}
          OUTLINE_API_TOKEN: ${{ secrets.OUTLINE_API_TOKEN }}
        run: outline push --collection-id "Docs" --path ./docs/ --create-collection

      - name: Export as static site
        env:
          OUTLINE_HOST: ${{ vars.OUTLINE_URL }}
          OUTLINE_API_TOKEN: ${{ secrets.OUTLINE_API_TOKEN }}
        run: outline site export --collection-id "Docs" --output ./site/

      - name: Deploy to Pages
        uses: actions/upload-pages-artifact@v3
        with:
          path: ./site/
```

## Open Questions

- Use Outline's HTML rendering or re-render from markdown ourselves?
  - Outline HTML: consistent with wiki appearance, but adds coupling
  - Self-render: more control, but then we ARE an SSG
- Minimal CSS: ship a small CSS file or go unstyled (let GitHub render)?
- Should `outline site` be a separate binary/plugin to keep core CLI small?

## Relation to Current `make docs`

Current: `make docs` generates reference from cobra + llms.txt from frontmatter (no Outline involved).

Future: `make docs` could be `outline push` + `outline site export` — the CLI tests itself.

## Priority

Medium. Requires `outline pull` (fetching documents) which is prerequisite feature.
See: ai-docs/feature-pull-sync.md
