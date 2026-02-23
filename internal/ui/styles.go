package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colours.
	orange    = lipgloss.Color("#FF6600") // HN orange
	subtleGray = lipgloss.Color("#6C7D8C")
	dimGray   = lipgloss.Color("#3D4B56")
	white     = lipgloss.Color("#FFFAF0")
	green     = lipgloss.Color("#72C472")
	yellow    = lipgloss.Color("#E8C547")

	// Story list styles.
	TitleStyle = lipgloss.NewStyle().
			Foreground(white).
			Bold(true)

	SelectedTitleStyle = lipgloss.NewStyle().
				Foreground(orange).
				Bold(true)

	MetaStyle = lipgloss.NewStyle().
			Foreground(subtleGray)

	ScoreStyle = lipgloss.NewStyle().
			Foreground(orange).
			Bold(true)

	IndexStyle = lipgloss.NewStyle().
			Foreground(dimGray).
			Width(4).
			Align(lipgloss.Right)

	// Comment styles.
	CommentAuthorStyle = lipgloss.NewStyle().
				Foreground(orange).
				Bold(true)

	CommentTimeStyle = lipgloss.NewStyle().
				Foreground(subtleGray)

	CommentTextStyle = lipgloss.NewStyle().
				Foreground(white)

	IndentStyle = lipgloss.NewStyle().
			Foreground(dimGray)

	// Header / title bar.
	HeaderStyle = lipgloss.NewStyle().
			Background(orange).
			Foreground(lipgloss.Color("#000000")).
			Bold(true).
			Padding(0, 1)

	// Status bar.
	StatusStyle = lipgloss.NewStyle().
			Foreground(subtleGray).
			Italic(true)

	// User profile.
	UserNameStyle = lipgloss.NewStyle().
			Foreground(orange).
			Bold(true).
			Underline(true)

	UserKarmaStyle = lipgloss.NewStyle().
			Foreground(yellow).
			Bold(true)

	// Help bar.
	HelpStyle = lipgloss.NewStyle().
			Foreground(dimGray)

	// Separator.
	SepStyle = lipgloss.NewStyle().
			Foreground(dimGray).
			SetString("  ·  ")

	// URL.
	URLStyle = lipgloss.NewStyle().
			Foreground(green).
			Italic(true)
)

// Sep renders the separator bullet.
func Sep() string { return SepStyle.Render() }
