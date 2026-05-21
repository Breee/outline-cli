<!-- Collection: test -->
<!-- Title: Feature: Enhanced Search with Excerpts -->

# Feature: Enhanced Search with Excerpts

## Problem

Current search shows only document titles. Users can't tell which result is relevant without opening each one. The TUI search view needs to show context snippets and feel instant.

## Requirements

- Show excerpt/snippet below each search result (the `context` field from Outline's API)
- Highlight matching terms in the excerpt
- Search must feel fast — debounce input, show spinner during API call
- Results update as you type (live search with ~300ms debounce)
- Compact multi-line result format: title + snippet + collection breadcrumb

## TUI Search View Design

```
┌─────────────────────────────────────────────────┐
│ Search: deploy rollba█                          │
├─────────────────────────────────────────────────┤
│ > Deployment Guide              [infrastructure]│
│   ...perform a **rollback** by running the      │
│   deploy script with --rollback flag...         │
│                                                 │
│   Rollback Procedures           [infrastructure]│
│   ...automated **rollback** is triggered when   │
│   health checks fail after **deploy**...        │
│                                                 │
│   CI/CD Pipeline                    [platform]  │
│   ...the **deploy** stage pushes to production  │
│   and waits for canary...                       │
├─────────────────────────────────────────────────┤
│ 3 results • enter open • esc back • / clear     │
└─────────────────────────────────────────────────┘
```

## Implementation

### Changes to `internal/tui/model.go`

1. **Live search with debounce**: On each keystroke in search mode, reset a 300ms timer. When it fires, send the API request.
2. **Show snippets**: Render `SearchResult.Context` below each title, truncated to 2 lines.
3. **Highlight matches**: Bold/color the query terms in the snippet using lipgloss inline styling.
4. **Collection breadcrumb**: Show collection name right-aligned on the title line.

### Data already available

The Outline API `documents.search` already returns a `context` field with the matching snippet (includes `<b>` tags around matches). We just need to:
- Strip HTML tags from the context
- Re-apply highlighting using lipgloss styles
- Truncate to fit 2 lines

### Debounce approach

Use a `tea.Tick` command with a short delay. Each keystroke cancels the previous tick by incrementing a counter. Only the tick matching the current counter triggers the search.

```go
type searchTickMsg struct{ id int }

func (m Model) handleSearchInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
    // ... update m.searchInput ...
    m.searchTickID++
    id := m.searchTickID
    return m, tea.Tick(300*time.Millisecond, func(time.Time) tea.Msg {
        return searchTickMsg{id: id}
    })
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    case searchTickMsg:
        if msg.id == m.searchTickID {
            return m, m.doSearch(m.searchInput)
        }
}
```

### CLI `outline search` improvements

- `--format default` already shows snippets — keep as-is
- Strip `<b>` tags from context in default/oneline formats (already done)
- Consider adding `--highlight` flag for terminal bold on matches

## Testing

| Scenario | Package | Type |
|----------|---------|------|
| Debounce fires only on last keystroke | `tui` | Unit |
| HTML tags stripped from context | `tui` | Unit |
| Highlight positions calculated correctly | `tui` | Unit |
| Empty query shows no results | `tui` | Unit |
| Search results navigable with j/k | `tui` | Unit |
