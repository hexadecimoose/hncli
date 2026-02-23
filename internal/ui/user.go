package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/moose/hncli/internal/api"
	"github.com/moose/hncli/internal/util"
)

// UserLoaded is sent when a user profile has been fetched.
type UserLoaded struct {
	User  *api.User
	Items []*api.Item
	Err   error
}

// UserModel is a bubbletea model for a user profile view.
type UserModel struct {
	user    *api.User
	items   []*api.Item
	scroll  int
	height  int
	width   int
	loading bool
	err     error
}

// NewUserModel returns a loading user model.
func NewUserModel() UserModel {
	return UserModel{loading: true, height: 24, width: 80}
}

func (m UserModel) Init() tea.Cmd { return nil }

func (m UserModel) Update(msg tea.Msg) (UserModel, tea.Cmd) {
	switch msg := msg.(type) {
	case UserLoaded:
		m.loading = false
		m.err = msg.Err
		m.user = msg.User
		m.items = msg.Items
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
		case "o":
			if m.user != nil {
				util.OpenBrowser(fmt.Sprintf("https://news.ycombinator.com/user?id=%s", m.user.ID)) //nolint:errcheck
			}
		case "q", "backspace", "esc", "left", "h":
			return m, func() tea.Msg { return BackMsg{} }
		}
	}
	return m, nil
}

func (m UserModel) View() string {
	var b strings.Builder

	title := "User Profile"
	if m.user != nil {
		title = "User: " + m.user.ID
	}
	b.WriteString(HeaderStyle.Width(m.width).Render("  " + title))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(StatusStyle.Render("  Loading…"))
		return b.String()
	}
	if m.err != nil {
		b.WriteString(fmt.Sprintf("  Error: %v\n", m.err))
		return b.String()
	}
	if m.user == nil {
		b.WriteString(StatusStyle.Render("  User not found."))
		return b.String()
	}

	u := m.user
	b.WriteString(fmt.Sprintf("  %s  %s  karma: %s  joined: %s\n\n",
		UserNameStyle.Render(u.ID),
		Sep(),
		UserKarmaStyle.Render(fmt.Sprint(u.Karma)),
		MetaStyle.Render(time.Unix(u.Created, 0).Format("Jan 2006")),
	))

	if u.About != "" {
		b.WriteString(wrapText(util.StripHTML(u.About), m.width-4, "  ") + "\n\n")
	}

	if len(m.items) > 0 {
		b.WriteString(TitleStyle.Render("  Recent submissions:") + "\n\n")
		end := m.scroll + m.height - 8
		if end > len(m.items) {
			end = len(m.items)
		}
		start := m.scroll
		if start > len(m.items) {
			start = len(m.items)
		}
		for i, item := range m.items[start:end] {
			if item.Title != "" {
				b.WriteString(fmt.Sprintf("  %s %s\n",
					IndexStyle.Render(fmt.Sprintf("%d.", start+i+1)),
					TitleStyle.Render(item.Title),
				))
				b.WriteString(fmt.Sprintf("     %s%s%s\n\n",
					MetaStyle.Render(fmt.Sprintf("▲ %d · %s comments · %s", item.Score, commentsStr(item.Descendants), item.Age())),
					Sep(),
					URLStyle.Render(hostname(item.URL)),
				))
			}
		}
	}

	b.WriteString(HelpStyle.Render("  ↑/↓ scroll · o: open in browser · ←/esc: back · q: quit"))
	return b.String()
}
