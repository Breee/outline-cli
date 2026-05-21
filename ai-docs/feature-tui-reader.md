<!-- Collection: test -->
<!-- Title: Feature: Terminal UI Reader (BubbleTea TUI) -->

# Feature: Terminal UI Reader (BubbleTea TUI)

## Why This Matters

Documentation is useless if people don't read it. The biggest barrier to reading docs is context switching — leaving your terminal to open a browser, navigate to the wiki, find the right page. A TUI reader eliminates this friction entirely.

The best CLI tools (lazygit, k9s, htop) succeed because they meet you where you already are.

## Commands

### `outline read [query]`

Interactive TUI for browsing and reading wiki documents.

```bash
outline read                     # open TUI with collection browser
outline read "deploy guide"      # jump to search results for "deploy guide"
outline read --collection ops    # browse a specific collection
```

### `outline cat <doc>`

Non-interactive — prints a single document to stdout (pipe-friendly):

```bash
outline cat "Deployment Guide" | less
outline cat --id abc123 | grep -i "rollback"
```

## TUI Design

Built with [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) + [glamour](https://github.com/charmbracelet/glamour) for markdown rendering.

### Views

1. **Collection Browser** — tree view of all collections and docs (vim-style j/k navigation)
2. **Document Reader** — rendered markdown with syntax-highlighted code blocks, scrollable
3. **Search** — fuzzy search across all documents, live results as you type
4. **Breadcrumbs** — show current location: `Collection > Parent > Document`

### Key Bindings

| Key | Action |
|-----|--------|
| `/` | Open search |
| `Enter` | Open document / expand collection |
| `q`, `Esc` | Back / quit |
| `j/k` | Navigate up/down |
| `h/l` | Collapse/expand tree nodes |
| `y` | Copy document URL to clipboard |
| `e` | Open in `$EDITOR` (pulls to temp file, pushes on save) |
| `o` | Open in browser |
| `?` | Help |

### Rendering

- Use `glamour` with terminal-aware dark/light theme detection
- Syntax highlighting in code blocks via `chroma`
- Tables rendered as proper aligned columns
- Images show `[image: alt text]` placeholder with URL
- Links are numbered `[1]` with footnote list at bottom (like `lynx`)

## Implementation Plan

1. Add `github.com/charmbracelet/bubbletea`, `bubbles`, `lipgloss`, `glamour` dependencies
2. Create `internal/tui/` package:
   - `model.go` — root model with view state machine
   - `browser.go` — collection/document tree browser
   - `reader.go` — document viewer with scroll
   - `search.go` — search input + results list
3. Create `cmd/read.go` and `cmd/cat.go` commands
4. API calls needed: `collections.list`, `documents.search`, `documents.info`, `documents.list` (by collection)

## Caching

- Cache collection tree and document metadata locally in `~/.outline-cli/cache/`
- TTL-based: refresh if older than 5 minutes (configurable)
- `outline cache clear` to force refresh
- Document content fetched on-demand, cached for session

## Offline Mode

If the server is unreachable, show cached content with a `[OFFLINE - cached at <time>]` banner. This makes the CLI useful on planes, in restricted networks, etc.

## Testing

- Unit test model state transitions (view switches, key handling)
- Mock API client for browser/search integration tests
- Golden file tests for glamour rendering output
