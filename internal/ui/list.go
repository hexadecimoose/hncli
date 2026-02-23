package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/moose/hncli/internal/api"
	"github.com/moose/hncli/internal/util"
)

// StoriesLoaded is sent when story items have been fetched.
type StoriesLoaded struct {
	Items []*api.Item
	Err   error
}

// OpenItem is sent when the user wants to open a story's comments.
type OpenItem struct{ ID int }

// OpenURL is sent when the user wants to open a URL.
type OpenURL struct{ URL string }

// ListModel is a bubbletea model for a scrollable list of stories.
type ListModel struct {
	title    string
	items    []*api.Item
	cursor   int
	offset   int
	height   int
	width    int
	loading  bool
	err      error
}

// NewListModel creates a list model with a given title. Items are populated later.
func NewListModel(title string) ListModel {
	return ListModel{title: title, loading: true, height: 24, width: 80}
}

func (m ListModel) Init() tea.Cmd { return nil }

func (m ListModel) Update(msg tea.Msg) (ListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case StoriesLoaded:
		m.loading = false
		m.err = msg.Err
		m.items = msg.Items
		m.cursor = 0
		m.offset = 0

	case tea.WindowSizeMsg:
		m.height = msg.Height - 4 // leave room for header + help
		m.width = msg.Width

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.offset {
					m.offset = m.cursor
				}
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
				if m.cursor >= m.offset+m.visibleLines() {
					m.offset++
				}
			}
		case "g":
			m.cursor = 0
			m.offset = 0
		case "G":
			m.cursor = len(m.items) - 1
			m.offset = max(0, m.cursor-m.visibleLines()+1)
		case "enter":
			if len(m.items) > 0 {
				return m, func() tea.Msg { return OpenItem{m.items[m.cursor].ID} }
			}
		case "o":
			if len(m.items) > 0 {
				u := m.items[m.cursor].URL
				if u == "" {
					u = fmt.Sprintf("https://news.ycombinator.com/item?id=%d", m.items[m.cursor].ID)
				}
				util.OpenBrowser(u) //nolint:errcheck
			}
		case "c":
			if len(m.items) > 0 {
				util.OpenBrowser(fmt.Sprintf("https://news.ycombinator.com/item?id=%d", m.items[m.cursor].ID)) //nolint:errcheck
			}
		}
	}
	return m, nil
}

func (m ListModel) visibleLines() int {
	// Each story takes 2 lines.
	return m.height / 2
}

func (m ListModel) View() string {
	var b strings.Builder

	// Header.
	b.WriteString(HeaderStyle.Width(m.width).Render("  " + m.title))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(StatusStyle.Render("  Loading…"))
		return b.String()
	}
	if m.err != nil {
		b.WriteString(fmt.Sprintf("  Error: %v", m.err))
		return b.String()
	}
	if len(m.items) == 0 {
		b.WriteString(StatusStyle.Render("  No stories found."))
		return b.String()
	}

	visible := m.visibleLines()
	end := m.offset + visible
	if end > len(m.items) {
		end = len(m.items)
	}

	for i := m.offset; i < end; i++ {
		item := m.items[i]
		selected := i == m.cursor

		// Line 1: index + title + score.
		idx := IndexStyle.Render(fmt.Sprintf("%d.", i+1))
		var titleStr string
		if selected {
			titleStr = SelectedTitleStyle.Render(item.Title)
		} else {
			titleStr = TitleStyle.Render(item.Title)
		}
		score := ScoreStyle.Render(fmt.Sprintf("▲ %d", item.Score))
		line1 := idx + " " + titleStr + "  " + score

		// Line 2: meta.
		host := ""
		if item.URL != "" {
			host = URLStyle.Render(hostname(item.URL))
		}
		meta := MetaStyle.Render(
			fmt.Sprintf("  %s comments · by %s · %s", commentsStr(item.Descendants), item.By, item.Age()),
		)
		if host != "" {
			meta = "    " + host + MetaStyle.Render(" · ") + meta[4:]
		}

		prefix := "  "
		if selected {
			prefix = lipgloss.NewStyle().Foreground(orange).Render("▶ ")
		}

		b.WriteString(prefix + line1 + "\n")
		b.WriteString("  " + meta + "\n")
	}

	// Help bar.
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("  ↑/↓ navigate · enter: comments · o: open url · c: open hn · q: quit"))

	return b.String()
}

func commentsStr(n int) string {
	if n == 1 {
		return "1"
	}
	return fmt.Sprintf("%d", n)
}

func hostname(rawURL string) string {
	// Trim scheme and www.
	s := rawURL
	for _, prefix := range []string{"https://", "http://"} {
		s = strings.TrimPrefix(s, prefix)
	}
	s = strings.TrimPrefix(s, "www.")
	if i := strings.IndexByte(s, '/'); i >= 0 {
		s = s[:i]
	}
	return s
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
