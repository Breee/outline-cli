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

	// Preview pane state
	previewEnabled bool
	previewContent string
	previewTitle   string
	previewVP      viewport.Model
	previewTickID  int
	docCache       map[string]string // id -> rendered content

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
		docCache:     make(map[string]string),
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
		if m.previewEnabled {
			pw := m.previewWidth()
			m.previewVP = viewport.New(viewport.WithWidth(pw), viewport.WithHeight(m.height-4))
			if m.previewContent != "" {
				m.previewVP.SetContent(m.previewContent)
			}
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

	case previewTickMsg:
		if msg.id == m.previewTickID {
			return m, m.fetchPreview(msg.docID)
		}
		return m, nil

	case previewContentMsg:
		if msg.err != nil {
			m.previewContent = "Error loading preview"
		} else {
			m.docCache[msg.id] = msg.rendered
			m.previewContent = msg.rendered
			m.previewTitle = msg.title
		}
		pw := m.previewWidth()
		m.previewVP = viewport.New(viewport.WithWidth(pw), viewport.WithHeight(m.height-4))
		m.previewVP.SetContent(m.previewContent)
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
		m = m.moveCursor(-1)
		var cmd tea.Cmd
		m, cmd = m.triggerPreview()
		return m, cmd

	case key.Matches(msg, m.keys.Down):
		if m.view == ViewReader {
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
		m = m.moveCursor(1)
		var cmd tea.Cmd
		m, cmd = m.triggerPreview()
		return m, cmd

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
		m = m.moveCursor(10)
		var cmd tea.Cmd
		m, cmd = m.triggerPreview()
		return m, cmd

	case key.Matches(msg, m.keys.PageUp):
		if m.view == ViewReader {
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
		m = m.moveCursor(-10)
		var cmd tea.Cmd
		m, cmd = m.triggerPreview()
		return m, cmd

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
		var cmd tea.Cmd
		m, cmd = m.triggerPreview()
		return m, cmd

	case key.Matches(msg, m.keys.Bottom):
		if m.view == ViewReader {
			m.viewport.GotoBottom()
			return m, nil
		}
		max := m.maxCursor()
		if max >= 0 {
			m.cursor = max
		}
		var cmd tea.Cmd
		m, cmd = m.triggerPreview()
		return m, cmd

	case key.Matches(msg, m.keys.TogglePreview):
		if m.view == ViewBrowser || m.view == ViewSearch {
			m.previewEnabled = !m.previewEnabled
			if m.previewEnabled {
				var cmd tea.Cmd
				m, cmd = m.triggerPreview()
				return m, cmd
			}
			m.previewContent = ""
			m.previewTitle = ""
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

// previewTickMsg is sent after debounce to trigger preview fetch.
type previewTickMsg struct {
	id    int
	docID string
}

// previewContentMsg returns rendered preview content.
type previewContentMsg struct {
	id       string
	title    string
	rendered string
	err      error
}

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
		m.searchTickID++
		id := m.searchTickID
		return m, tea.Tick(300*time.Millisecond, func(time.Time) tea.Msg {
			return searchTickMsg{id: id}
		})
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
		m.searchTickID++
		id := m.searchTickID
		return m, tea.Tick(300*time.Millisecond, func(time.Time) tea.Msg {
			return searchTickMsg{id: id}
		})
	}
}

// --- Preview helpers ---

// previewWidth returns the width of the preview pane (roughly half).
func (m Model) previewWidth() int {
	pw := m.width / 2
	if pw < 30 {
		pw = 30
	}
	return pw
}

// listWidth returns the width available for the list when preview is active.
func (m Model) listWidth() int {
	if !m.previewEnabled {
		return m.width
	}
	return m.width - m.previewWidth() - 1 // 1 for separator
}

// triggerPreview schedules a debounced preview fetch for the current item.
func (m Model) triggerPreview() (Model, tea.Cmd) {
	if !m.previewEnabled {
		return m, nil
	}
	docID := m.currentDocID()
	if docID == "" {
		m.previewContent = ""
		m.previewTitle = ""
		return m, nil
	}
	// If cached, load immediately.
	if content, ok := m.docCache[docID]; ok {
		m.previewContent = content
		pw := m.previewWidth()
		m.previewVP = viewport.New(viewport.WithWidth(pw), viewport.WithHeight(m.height-4))
		m.previewVP.SetContent(content)
		return m, nil
	}
	m.previewTickID++
	id := m.previewTickID
	return m, tea.Tick(150*time.Millisecond, func(time.Time) tea.Msg {
		return previewTickMsg{id: id, docID: docID}
	})
}

// currentDocID returns the document ID at the current cursor position.
func (m Model) currentDocID() string {
	switch m.view {
	case ViewSearch:
		if m.searchCursor >= 0 && m.searchCursor < len(m.results) {
			return m.results[m.searchCursor].Document.ID
		}
	default:
		if m.cursor >= 0 && m.cursor < len(m.items) {
			item := m.items[m.cursor]
			if !item.IsParent {
				return item.ID
			}
		}
	}
	return ""
}

// fetchPreview fetches and renders a document for preview.
func (m Model) fetchPreview(docID string) tea.Cmd {
	return func() tea.Msg {
		doc, err := m.client.GetDocument(m.ctx, docID)
		if err != nil {
			return previewContentMsg{id: docID, err: err}
		}
		pw := m.previewWidth()
		rendered, err := renderMarkdown(doc.Text, pw-2)
		if err != nil {
			rendered = doc.Text
		}
		return previewContentMsg{id: docID, title: doc.Title, rendered: rendered}
	}
}

// renderPreviewPane renders the preview panel as a string.
func (m Model) renderPreviewPane() string {
	pw := m.previewWidth()
	h := m.height - 2

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Width(pw - 4).
		Height(h - 2)

	if m.previewContent == "" {
		placeholder := dimStyle.Render("No preview available")
		return style.Render(placeholder)
	}

	title := headerStyle.Render(m.previewTitle)
	return style.Render(title + "\n" + m.previewVP.View())
}

// renderSplitView combines a left list string with the preview pane.
func (m Model) renderSplitView(left string) string {
	listW := m.listWidth()

	leftStyle := lipgloss.NewStyle().Width(listW)
	rightContent := m.renderPreviewPane()

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftStyle.Render(left),
		rightContent,
	)
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

	listW := m.listWidth()
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
		if len(line) > listW-1 {
			line = line[:listW-4] + "..."
		}
		if i == m.cursor {
			s += selectedStyle.Render(line) + "\n"
		} else {
			s += normalStyle.Render(line) + "\n"
		}
	}

	helpStr := "/ search • enter select • p preview • q quit"
	s += "\n" + statusStyle.Render(helpStr)

	if m.previewEnabled {
		v := tea.NewView(m.renderSplitView(s))
		v.AltScreen = true
		return v
	}

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

	listW := m.listWidth()

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

	maxSnippetWidth := listW - 6
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

	s += "\n" + statusStyle.Render("enter open • esc back • / search • p preview • q quit")

	if m.previewEnabled {
		v := tea.NewView(m.renderSplitView(s))
		v.AltScreen = true
		return v
	}

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
