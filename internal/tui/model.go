package tui

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
	"github.com/Breee/outline-cli/internal/outline"
)

// View represents the current TUI view state.
type View int

const (
	ViewBrowser View = iota
	ViewReader
	ViewSearch
)

// Item represents a navigable item in the browser.
type Item struct {
	ID       string
	Title    string
	IsParent bool // has children
	Depth    int
	URL      string
}

// Model is the root TUI model.
type Model struct {
	client  *outline.Client
	ctx     context.Context
	keys    KeyMap
	width   int
	height  int
	view    View
	err     error
	baseURL string

	// Browser state
	collections []outline.Collection
	items       []Item
	cursor      int
	loading     bool
	spinner     spinner.Model

	// Reader state
	viewport    viewport.Model
	docTitle    string
	docURL      string
	docContent  string
	breadcrumbs string

	// Search state
	searchInput  string
	searchTyping bool
	results      []outline.SearchResult
	searchCursor int
	searchTickID int

	// Initial query (if launched with argument)
	initialQuery string
}

// New creates a new TUI model.
func New(client *outline.Client, baseURL string, initialQuery string) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot

	return Model{
		client:       client,
		ctx:          context.Background(),
		keys:         DefaultKeyMap,
		spinner:      s,
		loading:      true,
		baseURL:      baseURL,
		initialQuery: initialQuery,
	}
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.spinner.Tick}
	if m.initialQuery != "" {
		cmds = append(cmds, m.doSearch(m.initialQuery))
	} else {
		cmds = append(cmds, m.fetchCollections())
	}
	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport = viewport.New(viewport.WithWidth(m.width), viewport.WithHeight(m.height-3))
		if m.docContent != "" {
			m.viewport.SetContent(m.docContent)
		}
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)

	case collectionsMsg:
		m.loading = false
		m.collections = msg.collections
		m.items = collectionsToItems(msg.collections)
		if msg.err != nil {
			m.err = msg.err
		}
		return m, nil

	case documentsMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.items = docsToItems(msg.docs, msg.collectionName)
		m.cursor = 0
		return m, nil

	case documentContentMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.view = ViewReader
		m.docTitle = msg.title
		m.docURL = msg.url
		rendered, err := renderMarkdown(msg.content, m.width)
		if err != nil {
			m.docContent = msg.content
		} else {
			m.docContent = rendered
		}
		m.viewport = viewport.New(viewport.WithWidth(m.width), viewport.WithHeight(m.height-3))
		m.viewport.SetContent(m.docContent)
		return m, nil

	case searchResultsMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.view = ViewSearch
		m.results = msg.results
		m.searchCursor = 0
		// Keep searchTyping=true so user can continue refining query
		return m, nil

	case searchTickMsg:
		if msg.id == m.searchTickID && m.searchInput != "" {
			return m, m.doSearch(m.searchInput)
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	// Pass viewport updates when in reader view.
	if m.view == ViewReader {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) View() tea.View {
	if m.err != nil {
		v := tea.NewView(
			errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n\nPress q to quit.",
		)
		v.AltScreen = true
		return v
	}

	if m.loading {
		v := tea.NewView(
			fmt.Sprintf("\n  %s Loading...\n", m.spinner.View()),
		)
		v.AltScreen = true
		return v
	}

	switch m.view {
	case ViewReader:
		return m.readerView()
	case ViewSearch:
		return m.searchView()
	default:
		return m.browserView()
	}
}

// --- Key handling ---

func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// If typing in search, handle differently.
	if m.searchTyping {
		return m.handleSearchInput(msg)
	}

	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Back):
		return m.goBack()

	case key.Matches(msg, m.keys.Search):
		m.searchTyping = true
		return m, nil

	case key.Matches(msg, m.keys.Up):
		if m.view == ViewReader {
			// If at top of document and came from search, go back to search
			if m.viewport.ScrollPercent() == 0 && len(m.results) > 0 {
				m.view = ViewSearch
				m.docContent = ""
				return m, nil
			}
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
		if m.view == ViewSearch && m.searchCursor == 0 {
			// At first result, re-enter search typing
			m.searchTyping = true
			return m, nil
		}
		return m.moveCursor(-1), nil

	case key.Matches(msg, m.keys.Down):
		if m.view == ViewReader {
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
		return m.moveCursor(1), nil

	case key.Matches(msg, m.keys.Enter):
		return m.selectItem()

	case key.Matches(msg, m.keys.OpenURL):
		return m.openInBrowser()

	case key.Matches(msg, m.keys.PageDown):
		if m.view == ViewReader {
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
		return m.moveCursor(10), nil

	case key.Matches(msg, m.keys.PageUp):
		if m.view == ViewReader {
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
		return m.moveCursor(-10), nil

	case key.Matches(msg, m.keys.HalfPageDn):
		if m.view == ViewReader {
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
		return m, nil

	case key.Matches(msg, m.keys.HalfPageUp):
		if m.view == ViewReader {
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
		return m, nil

	case key.Matches(msg, m.keys.Top):
		if m.view == ViewReader {
			m.viewport.GotoTop()
			return m, nil
		}
		m.cursor = 0
		return m, nil

	case key.Matches(msg, m.keys.Bottom):
		if m.view == ViewReader {
			m.viewport.GotoBottom()
			return m, nil
		}
		max := m.maxCursor()
		if max >= 0 {
			m.cursor = max
		}
		return m, nil
	}

	return m, nil
}

func (m Model) goBack() (tea.Model, tea.Cmd) {
	switch m.view {
	case ViewReader:
		m.view = ViewBrowser
		m.docContent = ""
		if len(m.results) > 0 {
			m.view = ViewSearch
		}
		return m, nil
	case ViewSearch:
		m.view = ViewBrowser
		m.results = nil
		return m, nil
	default:
		// If viewing docs in a collection, go back to collections.
		if len(m.collections) > 0 && len(m.items) > 0 && m.items[0].Depth > 0 {
			m.items = collectionsToItems(m.collections)
			m.cursor = 0
			return m, nil
		}
		return m, tea.Quit
	}
}

func (m Model) moveCursor(delta int) Model {
	max := m.maxCursor()
	if m.view == ViewSearch {
		m.searchCursor += delta
		if m.searchCursor < 0 {
			m.searchCursor = 0
		}
		if m.searchCursor > max {
			m.searchCursor = max
		}
	} else {
		m.cursor += delta
		if m.cursor < 0 {
			m.cursor = 0
		}
		if m.cursor > max {
			m.cursor = max
		}
	}
	return m
}

func (m Model) maxCursor() int {
	switch m.view {
	case ViewSearch:
		return len(m.results) - 1
	default:
		return len(m.items) - 1
	}
}

func (m Model) selectItem() (tea.Model, tea.Cmd) {
	switch m.view {
	case ViewSearch:
		if m.searchCursor >= 0 && m.searchCursor < len(m.results) {
			doc := m.results[m.searchCursor].Document
			m.loading = true
			return m, m.fetchDocument(doc.ID)
		}
	default:
		if m.cursor >= 0 && m.cursor < len(m.items) {
			item := m.items[m.cursor]
			if item.IsParent {
				// It's a collection — load its documents.
				m.loading = true
				return m, m.fetchDocuments(item.ID, item.Title)
			}
			// It's a document — load content.
			m.loading = true
			return m, m.fetchDocument(item.ID)
		}
	}
	return m, nil
}

func (m Model) openInBrowser() (tea.Model, tea.Cmd) {
	// TODO: open URL in browser
	return m, nil
}

// --- Search input ---

// searchTickMsg is sent after the debounce delay to trigger a search.
type searchTickMsg struct{ id int }

func (m Model) handleSearchInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.searchTyping = false
		if len(m.results) == 0 {
			m.view = ViewBrowser
		}
		return m, nil
	case "enter":
		if m.searchInput != "" {
			m.searchTyping = false
			m.loading = true
			return m, m.doSearch(m.searchInput)
		}
		m.searchTyping = false
		return m, nil
	case "backspace":
		if len(m.searchInput) > 0 {
			m.searchInput = m.searchInput[:len(m.searchInput)-1]
		}
		return m, m.debounceSearch()
	case "down", "j":
		// Exit typing mode and move to next result
		if len(m.results) > 0 {
			m.searchTyping = false
			if m.searchCursor < len(m.results)-1 {
				m.searchCursor++
			}
			return m, nil
		}
		return m, nil
	default:
		if len(msg.String()) == 1 || msg.String() == " " {
			m.searchInput += msg.String()
		}
		return m, m.debounceSearch()
	}
}

func (m *Model) debounceSearch() tea.Cmd {
	m.searchTickID++
	id := m.searchTickID
	return tea.Tick(300*time.Millisecond, func(time.Time) tea.Msg {
		return searchTickMsg{id: id}
	})
}

// --- Views ---

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	headerStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Bold(true).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	highlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("228")).
			Bold(true)
)

func (m Model) browserView() tea.View {
	s := ""
	if m.searchTyping {
		s += headerStyle.Render("Search: "+m.searchInput+"█") + "\n\n"
	} else {
		s += headerStyle.Render("Outline Wiki") + "\n\n"
	}

	visibleItems := m.height - 5
	if visibleItems < 1 {
		visibleItems = 10
	}

	start := 0
	if m.cursor >= visibleItems {
		start = m.cursor - visibleItems + 1
	}

	for i := start; i < len(m.items) && i < start+visibleItems; i++ {
		item := m.items[i]
		prefix := "  "
		if i == m.cursor {
			prefix = "> "
		}

		icon := "📄 "
		if item.IsParent {
			icon = "📁 "
		}

		line := prefix + icon + item.Title
		if i == m.cursor {
			s += selectedStyle.Render(line) + "\n"
		} else {
			s += normalStyle.Render(line) + "\n"
		}
	}

	s += "\n" + statusStyle.Render("/ search • enter select • q quit")
	v := tea.NewView(s)
	v.AltScreen = true
	return v
}

func (m Model) readerView() tea.View {
	header := headerStyle.Render(m.docTitle)
	footer := statusStyle.Render(fmt.Sprintf(
		"↑↓ scroll • esc back • q quit • %d%%",
		int(m.viewport.ScrollPercent()*100),
	))
	v := tea.NewView(header + "\n" + m.viewport.View() + "\n" + footer)
	v.AltScreen = true
	return v
}

func (m Model) searchView() tea.View {
	s := ""
	if m.searchTyping {
		s += headerStyle.Render("Search: "+m.searchInput+"█") + "\n\n"
	} else {
		s += headerStyle.Render(fmt.Sprintf("Search results (%d)", len(m.results))) + "\n\n"
	}

	// Each result takes ~3 lines (title + snippet + blank)
	linesPerResult := 3
	visibleItems := (m.height - 5) / linesPerResult
	if visibleItems < 1 {
		visibleItems = 5
	}

	cursor := m.searchCursor
	start := 0
	if cursor >= visibleItems {
		start = cursor - visibleItems + 1
	}

	maxSnippetWidth := m.width - 6
	if maxSnippetWidth < 40 {
		maxSnippetWidth = 40
	}

	for i := start; i < len(m.results) && i < start+visibleItems; i++ {
		r := m.results[i]
		prefix := "  "
		if i == cursor {
			prefix = "> "
		}

		// Title line with collection info
		titleLine := prefix + r.Document.Title
		if r.Document.CollectionID != "" {
			collName := m.collectionName(r.Document.CollectionID)
			if collName != "" {
				titleLine += dimStyle.Render("  [" + collName + "]")
			}
		}

		if i == cursor {
			s += selectedStyle.Render(prefix+r.Document.Title) + dimStyle.Render(titleLine[len(prefix)+len(r.Document.Title):]) + "\n"
		} else {
			s += normalStyle.Render(prefix+r.Document.Title) + dimStyle.Render(titleLine[len(prefix)+len(r.Document.Title):]) + "\n"
		}

		// Snippet line
		snippet := formatSnippet(r.Context, m.searchInput, maxSnippetWidth)
		if snippet != "" {
			s += "    " + snippet + "\n"
		}
	}

	if len(m.results) == 0 && !m.searchTyping {
		s += dimStyle.Render("  No results found.") + "\n"
	}

	s += "\n" + statusStyle.Render("enter open • esc back • / search • q quit")
	v := tea.NewView(s)
	v.AltScreen = true
	return v
}

// collectionName resolves a collection ID to its name from cached collections.
func (m Model) collectionName(id string) string {
	for _, c := range m.collections {
		if c.ID == id {
			return c.Name
		}
	}
	return ""
}

// formatSnippet cleans up the search context, strips HTML, and highlights matches.
func formatSnippet(context, query string, maxWidth int) string {
	if context == "" {
		return ""
	}

	// Strip HTML tags (Outline returns <b>match</b>)
	htmlTagRe := regexp.MustCompile(`<[^>]*>`)
	clean := htmlTagRe.ReplaceAllString(context, "")

	// Collapse whitespace
	clean = strings.Join(strings.Fields(clean), " ")

	// Truncate
	if len(clean) > maxWidth {
		clean = clean[:maxWidth] + "…"
	}

	// Highlight query terms
	if query != "" {
		terms := strings.Fields(query)
		for _, term := range terms {
			re, err := regexp.Compile(`(?i)` + regexp.QuoteMeta(term))
			if err != nil {
				continue
			}
			clean = re.ReplaceAllStringFunc(clean, func(match string) string {
				return highlightStyle.Render(match)
			})
		}
	}

	return dimStyle.Render(clean)
}

// --- Commands / Messages ---

type collectionsMsg struct {
	collections []outline.Collection
	err         error
}

type documentsMsg struct {
	docs           []outline.Document
	collectionName string
	err            error
}

type documentContentMsg struct {
	title   string
	content string
	url     string
	err     error
}

type searchResultsMsg struct {
	results []outline.SearchResult
	err     error
}

func (m Model) fetchCollections() tea.Cmd {
	return func() tea.Msg {
		cols, err := m.client.ListCollections(m.ctx)
		return collectionsMsg{collections: cols, err: err}
	}
}

func (m Model) fetchDocuments(collectionID, collectionName string) tea.Cmd {
	return func() tea.Msg {
		docs, err := m.client.ListDocuments(m.ctx, collectionID)
		return documentsMsg{docs: docs, collectionName: collectionName, err: err}
	}
}

func (m Model) fetchDocument(id string) tea.Cmd {
	return func() tea.Msg {
		doc, err := m.client.GetDocument(m.ctx, id)
		if err != nil {
			return documentContentMsg{err: err}
		}
		return documentContentMsg{
			title:   doc.Title,
			content: doc.Text,
			url:     doc.URL,
		}
	}
}

func (m Model) doSearch(query string) tea.Cmd {
	return func() tea.Msg {
		results, err := m.client.Search(m.ctx, query, "", 25)
		return searchResultsMsg{results: results, err: err}
	}
}

// --- Helpers ---

func collectionsToItems(cols []outline.Collection) []Item {
	items := make([]Item, 0, len(cols))
	for _, c := range cols {
		items = append(items, Item{
			ID:       c.ID,
			Title:    c.Name,
			IsParent: true,
			Depth:    0,
			URL:      c.URL,
		})
	}
	return items
}

func docsToItems(docs []outline.Document, collectionName string) []Item {
	items := make([]Item, 0, len(docs))
	for _, d := range docs {
		items = append(items, Item{
			ID:    d.ID,
			Title: d.Title,
			Depth: 1,
			URL:   d.URL,
		})
	}
	return items
}

func renderMarkdown(content string, width int) (string, error) {
	if width <= 0 {
		width = 80
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStylePath("dark"),
		glamour.WithWordWrap(width-4),
		glamour.WithEmoji(),
	)
	if err != nil {
		return "", err
	}
	out, err := r.Render(content)
	if err != nil {
		return "", err
	}
	return out, nil
}
