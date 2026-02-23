package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
	"github.com/hexadecimoose/hncli/internal/api"
	"github.com/hexadecimoose/hncli/internal/ui"
)

var version = "dev" // set by -ldflags at build time

var (
	count  int
	plain  bool
	client *api.Client
)

// isPlain returns true if plain mode is active (flag set, or stdout is not a TTY).
func isPlain() bool {
	return plain || !term.IsTerminal(int(os.Stdout.Fd()))
}

func main() {
	client = api.New()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:     "hncli",
	Short:   "A Hacker News CLI",
	Version: version,
	Long: `hncli — browse Hacker News from your terminal.

Run without arguments to launch the interactive TUI browser.
Use subcommands for quick access to specific feeds.
Use --plain / -p (or pipe output) for plain text output.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if isPlain() {
			items, err := client.TopStories(count)
			if err != nil {
				return err
			}
			printStories(items)
			return nil
		}
		return ui.RunWithLoader(client, "Hacker News · Top Stories", func() ([]*api.Item, error) {
			return client.TopStories(count)
		})
	},
}

func init() {
	rootCmd.PersistentFlags().IntVarP(&count, "count", "n", 30, "number of stories to fetch")
	rootCmd.PersistentFlags().BoolVarP(&plain, "plain", "p", false, "plain text output (no TUI); auto-enabled when stdout is not a TTY")

	rootCmd.AddCommand(topCmd, newCmd, bestCmd, askCmd, showCmd, jobsCmd, itemCmd, userCmd, searchCmd)
}

var topCmd = &cobra.Command{
	Use:   "top",
	Short: "Top stories",
	RunE: func(cmd *cobra.Command, args []string) error {
		if isPlain() {
			items, err := client.TopStories(count)
			if err != nil {
				return err
			}
			printStories(items)
			return nil
		}
		return ui.RunWithLoader(client, "Hacker News · Top Stories", func() ([]*api.Item, error) {
			return client.TopStories(count)
		})
	},
}

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Newest stories",
	RunE: func(cmd *cobra.Command, args []string) error {
		if isPlain() {
			items, err := client.NewStories(count)
			if err != nil {
				return err
			}
			printStories(items)
			return nil
		}
		return ui.RunWithLoader(client, "Hacker News · New Stories", func() ([]*api.Item, error) {
			return client.NewStories(count)
		})
	},
}

var bestCmd = &cobra.Command{
	Use:   "best",
	Short: "Best stories",
	RunE: func(cmd *cobra.Command, args []string) error {
		if isPlain() {
			items, err := client.BestStories(count)
			if err != nil {
				return err
			}
			printStories(items)
			return nil
		}
		return ui.RunWithLoader(client, "Hacker News · Best Stories", func() ([]*api.Item, error) {
			return client.BestStories(count)
		})
	},
}

var askCmd = &cobra.Command{
	Use:   "ask",
	Short: "Ask HN stories",
	RunE: func(cmd *cobra.Command, args []string) error {
		if isPlain() {
			items, err := client.AskStories(count)
			if err != nil {
				return err
			}
			printStories(items)
			return nil
		}
		return ui.RunWithLoader(client, "Hacker News · Ask HN", func() ([]*api.Item, error) {
			return client.AskStories(count)
		})
	},
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show HN stories",
	RunE: func(cmd *cobra.Command, args []string) error {
		if isPlain() {
			items, err := client.ShowStories(count)
			if err != nil {
				return err
			}
			printStories(items)
			return nil
		}
		return ui.RunWithLoader(client, "Hacker News · Show HN", func() ([]*api.Item, error) {
			return client.ShowStories(count)
		})
	},
}

var jobsCmd = &cobra.Command{
	Use:   "jobs",
	Short: "Job postings",
	RunE: func(cmd *cobra.Command, args []string) error {
		if isPlain() {
			items, err := client.JobStories(count)
			if err != nil {
				return err
			}
			printStories(items)
			return nil
		}
		return ui.RunWithLoader(client, "Hacker News · Jobs", func() ([]*api.Item, error) {
			return client.JobStories(count)
		})
	},
}

var itemCmd = &cobra.Command{
	Use:   "item <id>",
	Short: "View a story and its comments",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid item ID: %q", args[0])
		}
		if isPlain() {
			return printItem(client, id)
		}
		return ui.RunItem(client, id)
	},
}

var userCmd = &cobra.Command{
	Use:   "user <username>",
	Short: "View a user's profile and recent submissions",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if isPlain() {
			return printUser(client, args[0])
		}
		return ui.RunUser(client, args[0])
	},
}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search Hacker News via Algolia",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		q := strings.Join(args, " ")
		items, err := client.Search(q, count)
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}
		if isPlain() {
			for i, item := range items {
				fmt.Printf("%d. %s\n   %s\n   https://news.ycombinator.com/item?id=%d\n\n",
					i+1, item.Title, item.URL, item.ID)
			}
			return nil
		}
		return ui.RunWithItems(client, fmt.Sprintf("Search: %q", q), items)
	},
}

