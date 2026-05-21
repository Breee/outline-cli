<!-- Collection: test -->
<!-- Title: Feature: Terminal UI Reader (BubbleTea TUI) -->

# Feature: Terminal UI Reader (BubbleTea TUI)

## Why This Matters

Documentation is useless if people don't read it. The biggest barrier to reading docs is context switching — leaving your terminal to open a browser, navigate to the wiki, find the right page. A TUI reader eliminates this friction entirely.

The best CLI tools (lazygit, k9s, htop) succeed because they meet you where you already are.

## Commands

### `outline tui [query]`

Interactive TUI for browsing and reading wiki documents.

```bash
outline tui                     # open TUI with collection browser
outline tui "deploy guide"      # jump to search results for "deploy guide"
outline tui --collection ops    # browse a specific collection
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

## Implementation

### Done

1. Added `charm.land/bubbletea/v2`, `bubbles/v2`, `lipgloss/v2`, `glamour/v2` dependencies
2. Created `internal/tui/` package:
   - `keys.go` — key bindings (KeyMap)
   - `model.go` — root model with view state machine, browser, reader, search views
3. Created `cmd/read.go` (`outline tui`) and `cmd/cat.go` (`outline cat`) commands
4. API integration: `collections.list`, `documents.search`, `documents.info`, `documents.list`
5. Enhanced search: live debounced search (300ms), context excerpts, highlighted matches
6. Alt screen mode, viewport scrolling, seamless search→navigate→read→back flow

### Planned

- Preview pane (split view for wide terminals ≥120 cols, toggle with `p`)
- Document caching (in-memory for session, optional disk cache with TTL)
- Offline mode with cached content banner
- `outline search --interactive` flag to launch TUI search directly

## Architecture

```
cmd/read.go          — cobra command, creates tui.Model, runs tea.Program
cmd/cat.go           — non-interactive, fetches doc + glamour render to stdout
internal/tui/
  keys.go            — KeyMap struct with all bindings
  model.go           — Model struct, Init/Update/View, all views + messages
```

### State Machine

```
ViewBrowser ──enter──▶ ViewReader
     │                      │
     │◀──────esc────────────┘
     │
     │──/──▶ ViewSearch ──enter──▶ ViewReader
     │◀──esc──┘       │◀──up(top)──┘
```

### Key Patterns

- `searchTyping` bool gates whether keystrokes go to search input vs navigation
- Debounce via `tea.Tick` + incrementing `searchTickID` (only latest tick fires search)
- Viewport (bubbles) handles scroll in reader view; keys forwarded via Update passthrough

## Testing

- Unit test model state transitions (view switches, key handling)
- Mock API client for browser/search integration tests
- Golden file tests for glamour rendering output
