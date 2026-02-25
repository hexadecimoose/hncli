package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hexadecimoose/hncli/internal/api"
	"github.com/hexadecimoose/hncli/internal/util"
)

// ItemLoaded is sent when a single item (and its comment tree) is ready.
type ItemLoaded struct {
	Story    *api.Item
	Comments []*api.Item
	Err      error
}

// BackMsg is sent when the user wants to go back to the list.
type BackMsg struct{}

// flatComment is a comment flattened with its indent level.
type flatComment struct {
	item   *api.Item
	depth  int
	hidden bool // collapsed
}

// CommentsModel is a bubbletea model for a threaded comment view.
type CommentsModel struct {
	story   *api.Item
	flat    []flatComment
	lines   []string // all content lines, pre-rendered (excluding fixed header/footer)
	scroll  int      // first visible line index into m.lines
	height  int
	width   int
	loading bool
	err     error
}

// NewCommentsModel returns a loading comments model.
func NewCommentsModel() CommentsModel {
	return CommentsModel{loading: true, height: 24, width: 80}
}

func (m CommentsModel) Init() tea.Cmd { return nil }

// buildLines re-renders all scrollable content into m.lines.
func (m *CommentsModel) buildLines() {
	var lines []string
	add := func(s string) {
		for _, l := range strings.Split(s, "\n") {
			lines = append(lines, l)
		}
	}

	if m.story != nil {
		// Story meta.
		meta := fmt.Sprintf("  %s  ▲ %d  %s comments  by %s  %s",
			URLStyle.Render(m.story.URL),
			m.story.Score,
			commentsStr(m.story.Descendants),
			CommentAuthorStyle.Render(m.story.By),
			MetaStyle.Render(m.story.Age()),
		)
		add(meta)
		if m.story.Text != "" {
			add("")
			add(wrapText(util.StripHTML(m.story.Text), m.width-4, "  "))
		}
		add("")
	}

	for _, fc := range m.flat {
		if fc.item == nil || fc.item.Deleted || fc.item.Dead {
			continue
		}
		indent := strings.Repeat("  ", fc.depth)
		renderedBar := IndentStyle.Render("│ ")
		displayPrefix := indent + renderedBar + "  "
		// Use a plain-text prefix for width measurement — ANSI escapes in
		// renderedBar would inflate len() and cause premature line wraps.
		plainPrefixLen := len(indent) + len("│   ") // "│ " + "  " = 4 visible chars

		// Comment header line — always the first line of a comment.
		author := CommentAuthorStyle.Render(fc.item.By)
		age := CommentTimeStyle.Render(fc.item.Age())
		add(indent + renderedBar + author + "  " + age)

		// Comment body: wrap to plain lines then prepend the rendered prefix.
		text := util.StripHTML(fc.item.Text)
		wrapWidth := m.width - plainPrefixLen
		for _, paragraph := range strings.Split(text, "\n") {
			for _, wline := range wrapToLines(paragraph, wrapWidth) {
				lines = append(lines, displayPrefix+wline)
			}
		}
		add("") // blank separator between comments
	}

	m.lines = lines
}

func (m CommentsModel) Update(msg tea.Msg) (CommentsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case ItemLoaded:
		m.loading = false
		m.err = msg.Err
		m.story = msg.Story
		m.flat = flattenComments(msg.Comments, 0)
		m.scroll = 0
		m.buildLines()

	case tea.WindowSizeMsg:
		m.height = msg.Height - 2 // 1 fixed header + 1 fixed footer
		m.width = msg.Width
		m.buildLines()

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.scroll > 0 {
				m.scroll--
			}
		case "down", "j":
			if m.scroll < len(m.lines)-m.height {
				m.scroll++
			}
		case "g":
			m.scroll = 0
		case "G":
			m.scroll = max(0, len(m.lines)-m.height)
		case "o":
			if m.story != nil && m.story.URL != "" {
				util.OpenBrowser(m.story.URL) //nolint:errcheck
			}
		case "c":
			if m.story != nil {
				util.OpenBrowser(fmt.Sprintf("https://news.ycombinator.com/item?id=%d", m.story.ID)) //nolint:errcheck
			}
		case "q", "backspace", "esc", "left", "h":
			return m, func() tea.Msg { return BackMsg{} }
		}
	}
	return m, nil
}

func (m CommentsModel) View() string {
	var b strings.Builder

	// Fixed header — always line 1.
	title := "Loading…"
	if m.story != nil {
		title = m.story.Title
	}
	b.WriteString(HeaderStyle.Width(m.width).Render("  " + title))
	b.WriteString("\n")

	if m.loading {
		b.WriteString("\n" + StatusStyle.Render("  Loading comments…"))
		return b.String()
	}
	if m.err != nil {
		b.WriteString(fmt.Sprintf("\n  Error: %v", m.err))
		return b.String()
	}

	// Scrollable body: exactly m.height lines.
	end := m.scroll + m.height
	if end > len(m.lines) {
		end = len(m.lines)
	}
	for _, line := range m.lines[m.scroll:end] {
		b.WriteString(line + "\n")
	}

	// Fixed footer — always last line.
	pct := 100
	if len(m.lines) > m.height {
		pct = m.scroll * 100 / (len(m.lines) - m.height)
		if pct > 100 {
			pct = 100
		}
	}
	b.WriteString(HelpStyle.Render(fmt.Sprintf(
		"  ↑/↓ scroll · o: open url · c: open hn · r: refresh · ←/esc: back · q: quit  [%d%%]", pct,
	)))
	return b.String()
}

// flattenComments converts a tree of comments into a flat list with depth info.
func flattenComments(comments []*api.Item, depth int) []flatComment {
	var out []flatComment
	for _, c := range comments {
		if c == nil {
			continue
		}
		out = append(out, flatComment{item: c, depth: depth})
		// Note: child comments are not yet fetched; they require separate API calls.
		// For simplicity we don't recurse beyond first level in this version.
	}
	return out
}

// wrapToLines word-wraps text at width and returns the resulting lines (no prefix).
func wrapToLines(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	var lines []string
	line := ""
	for _, w := range words {
		if line == "" {
			line = w
		} else if len(line)+1+len(w) > width {
			lines = append(lines, line)
			line = w
		} else {
			line += " " + w
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

// wrapText hard-wraps text at width, prefixing every line with indent.
func wrapText(text string, width int, indent string) string {
	if width <= 0 {
		return indent + text
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}
	var lines []string
	line := indent
	for _, w := range words {
		if len(line)+len(w)+1 > width && line != indent {
			lines = append(lines, line)
			line = indent + w
		} else {
			if line == indent {
				line += w
			} else {
				line += " " + w
			}
		}
	}
	if line != indent {
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// indentColor maps depth to a lipgloss colour for variety (used when rendering nested threads).
func indentColor(depth int) lipgloss.Color {
	palette := []lipgloss.Color{
		lipgloss.Color("#FF6600"),
		lipgloss.Color("#E8C547"),
		lipgloss.Color("#72C472"),
		lipgloss.Color("#5BC8DB"),
		lipgloss.Color("#A78BFA"),
	}
	return palette[depth%len(palette)]
}
