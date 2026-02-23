package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/moose/hncli/internal/api"
	"github.com/moose/hncli/internal/util"
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
	story    *api.Item
	flat     []flatComment
	scroll   int
	height   int
	width    int
	loading  bool
	err      error
}

// NewCommentsModel returns a loading comments model.
func NewCommentsModel() CommentsModel {
	return CommentsModel{loading: true, height: 24, width: 80}
}

func (m CommentsModel) Init() tea.Cmd { return nil }

func (m CommentsModel) Update(msg tea.Msg) (CommentsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case ItemLoaded:
		m.loading = false
		m.err = msg.Err
		m.story = msg.Story
		m.flat = flattenComments(msg.Comments, 0)
		m.scroll = 0

	case tea.WindowSizeMsg:
		m.height = msg.Height - 4
		m.width = msg.Width

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.scroll > 0 {
				m.scroll--
			}
		case "down", "j":
			m.scroll++
		case "g":
			m.scroll = 0
		case "G":
			m.scroll = len(m.flat)
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

	// Header.
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

	// Story meta line.
	if m.story != nil {
		meta := fmt.Sprintf("  %s  ▲ %d  %s comments  by %s  %s",
			URLStyle.Render(m.story.URL),
			m.story.Score,
			commentsStr(m.story.Descendants),
			CommentAuthorStyle.Render(m.story.By),
			MetaStyle.Render(m.story.Age()),
		)
		b.WriteString("\n" + meta + "\n")
		if m.story.Text != "" {
			b.WriteString("\n" + wrapText(util.StripHTML(m.story.Text), m.width-4, "  ") + "\n")
		}
		b.WriteString("\n")
	}

	// Render visible comments.
	end := m.scroll + m.height
	if end > len(m.flat) {
		end = len(m.flat)
	}
	start := m.scroll
	if start > len(m.flat) {
		start = len(m.flat)
	}

	for _, fc := range m.flat[start:end] {
		if fc.item == nil {
			continue
		}
		if fc.item.Deleted || fc.item.Dead {
			continue
		}
		indent := strings.Repeat("  ", fc.depth)
		bar := IndentStyle.Render("│ ")

		author := CommentAuthorStyle.Render(fc.item.By)
		age := CommentTimeStyle.Render(fc.item.Age())
		header := indent + bar + author + "  " + age
		b.WriteString(header + "\n")

		text := util.StripHTML(fc.item.Text)
		for _, line := range strings.Split(text, "\n") {
			wrapped := wrapText(line, m.width-4-fc.depth*2, indent+bar+"  ")
			b.WriteString(wrapped + "\n")
		}
		b.WriteString("\n")
	}

	// Scroll indicator.
	if len(m.flat) > 0 {
		pct := 0
		if len(m.flat) > m.height {
			pct = m.scroll * 100 / (len(m.flat) - m.height)
			if pct > 100 {
				pct = 100
			}
		} else {
			pct = 100
		}
		b.WriteString(StatusStyle.Render(fmt.Sprintf("  %d/%d comments  %d%%", start, len(m.flat), pct)))
		b.WriteString("\n")
	}

	b.WriteString(HelpStyle.Render("  ↑/↓ scroll · o: open url · c: open hn · ←/esc: back · q: quit"))
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

// wrapText hard-wraps text at width, prefixing continuation lines with indent.
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
