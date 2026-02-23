package ui

import (
	"fmt"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/moose/hncli/internal/api"
)

// View is an enum for which screen is active.
type View int

const (
	ViewList View = iota
	ViewComments
	ViewUser
)

// App is the root bubbletea model for the interactive browser.
type App struct {
	apiClient *api.Client
	view      View
	list      ListModel
	comments  CommentsModel
	user      UserModel
}

// NewApp creates a new App ready to show the given story list.
func NewApp(client *api.Client, title string, loader func() ([]*api.Item, error)) *App {
	app := &App{
		apiClient: client,
		view:      ViewList,
		list:      NewListModel(title),
		comments:  NewCommentsModel(),
		user:      NewUserModel(),
	}
	app.list.loading = true
	app.comments.loading = false
	return app
}

// LoadCmd returns a command that fetches stories and sends StoriesLoaded.
func LoadCmd(loader func() ([]*api.Item, error)) tea.Cmd {
	return func() tea.Msg {
		items, err := loader()
		return StoriesLoaded{Items: items, Err: err}
	}
}

// LoadItemCmd fetches a story and its first-level comments in parallel.
func LoadItemCmd(client *api.Client, id int) tea.Cmd {
	return func() tea.Msg {
		story, err := client.Item(id)
		if err != nil {
			return ItemLoaded{Err: err}
		}
		if len(story.Kids) == 0 {
			return ItemLoaded{Story: story, Comments: nil}
		}

		// Fetch first-level comments in parallel (cap at 50).
		limit := len(story.Kids)
		if limit > 50 {
			limit = 50
		}
		kids := story.Kids[:limit]
		comments := make([]*api.Item, len(kids))
		var mu sync.Mutex
		var wg sync.WaitGroup
		sem := make(chan struct{}, 20)
		for i, kid := range kids {
			wg.Add(1)
			go func(i, kid int) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				c, e := client.Item(kid)
				if e == nil {
					mu.Lock()
					comments[i] = c
					mu.Unlock()
				}
			}(i, kid)
		}
		wg.Wait()
		return ItemLoaded{Story: story, Comments: comments}
	}
}

// LoadUserCmd fetches a user profile and their recent submissions.
func LoadUserCmd(client *api.Client, username string) tea.Cmd {
	return func() tea.Msg {
		user, err := client.User(username)
		if err != nil {
			return UserLoaded{Err: err}
		}
		// Fetch up to 10 recent story submissions.
		var items []*api.Item
		count := 0
		for _, id := range user.Submitted {
			if count >= 10 {
				break
			}
			item, e := client.Item(id)
			if e == nil && item != nil && item.Type == "story" && !item.Dead && !item.Deleted {
				items = append(items, item)
				count++
			}
		}
		return UserLoaded{User: user, Items: items}
	}
}

func (a *App) Init() tea.Cmd { return nil }

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" && a.view == ViewList {
			return a, tea.Quit
		}
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}

	case StoriesLoaded:
		m, cmd := a.list.Update(msg)
		a.list = m
		return a, cmd

	case ItemLoaded:
		a.view = ViewComments
		m, cmd := a.comments.Update(msg)
		a.comments = m
		return a, cmd

	case UserLoaded:
		a.view = ViewUser
		m, cmd := a.user.Update(msg)
		a.user = m
		return a, cmd

	case OpenItem:
		// Preserve terminal dimensions — NewCommentsModel() defaults to 80×24
		// which would cause buildLines() to wrap at 80 cols even on wider terminals.
		w, h := a.comments.width, a.comments.height
		a.comments = NewCommentsModel()
		a.comments.width, a.comments.height = w, h
		return a, LoadItemCmd(a.apiClient, msg.ID)

	case BackMsg:
		a.view = ViewList
		return a, nil

	case tea.WindowSizeMsg:
		m1, _ := a.list.Update(msg)
		a.list = m1
		m2, _ := a.comments.Update(msg)
		a.comments = m2
		m3, _ := a.user.Update(msg)
		a.user = m3
		return a, nil
	}

	// Route key events to the active view.
	switch a.view {
	case ViewList:
		m, cmd := a.list.Update(msg)
		a.list = m
		if cmd != nil {
			return a, cmd
		}
	case ViewComments:
		if msg, ok := msg.(tea.KeyMsg); ok {
			if msg.String() == "q" {
				return a, tea.Quit
			}
		}
		m, cmd := a.comments.Update(msg)
		a.comments = m
		if cmd != nil {
			return a, cmd
		}
	case ViewUser:
		if msg, ok := msg.(tea.KeyMsg); ok {
			if msg.String() == "q" {
				return a, tea.Quit
			}
		}
		m, cmd := a.user.Update(msg)
		a.user = m
		if cmd != nil {
			return a, cmd
		}
	}

	return a, nil
}

func (a *App) View() string {
	switch a.view {
	case ViewComments:
		return a.comments.View()
	case ViewUser:
		return a.user.View()
	default:
		return a.list.View()
	}
}

// Run starts the bubbletea program with the given loader.
func Run(client *api.Client, title string, loader func() ([]*api.Item, error)) error {
	app := NewApp(client, title, loader)
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// RunWithItems starts the TUI with a pre-built item list (e.g. search results).
func RunWithItems(client *api.Client, title string, items []*api.Item) error {
	app := NewApp(client, title, nil)
	app.list.loading = false
	app.list.items = items
	p := tea.NewProgram(app, tea.WithAltScreen())
	go func() { p.Send(StoriesLoaded{Items: items}) }()
	_, err := p.Run()
	return err
}

// RunWithLoader starts the TUI loading items async.
func RunWithLoader(client *api.Client, title string, loader func() ([]*api.Item, error)) error {
	app := NewApp(client, title, loader)
	p := tea.NewProgram(app, tea.WithAltScreen())
	go func() { p.Send(LoadCmd(loader)()) }()
	_, err := p.Run()
	return err
}

// RunItem opens a single item's comment view directly.
func RunItem(client *api.Client, id int) error {
	app := &App{
		apiClient: client,
		view:      ViewComments,
		list:      NewListModel(fmt.Sprintf("Item #%d", id)),
		comments:  NewCommentsModel(),
		user:      NewUserModel(),
	}
	p := tea.NewProgram(app, tea.WithAltScreen())
	go func() { p.Send(LoadItemCmd(client, id)()) }()
	_, err := p.Run()
	return err
}

// RunUser opens a user profile view directly.
func RunUser(client *api.Client, username string) error {
	app := &App{
		apiClient: client,
		view:      ViewUser,
		list:      NewListModel(""),
		comments:  NewCommentsModel(),
		user:      NewUserModel(),
	}
	p := tea.NewProgram(app, tea.WithAltScreen())
	go func() { p.Send(LoadUserCmd(client, username)()) }()
	_, err := p.Run()
	return err
}
