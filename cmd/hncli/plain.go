package main

import (
	"fmt"
	"strings"

	"github.com/hexadecimoose/hncli/internal/api"
	"github.com/hexadecimoose/hncli/internal/util"
)

// printStories prints a story list to stdout in plain text.
func printStories(items []*api.Item) {
	for i, item := range items {
		fmt.Printf("%d. %s (%d pts)\n", i+1, item.Title, item.Score)
		if item.URL != "" {
			fmt.Printf("   %s\n", item.URL)
		}
		fmt.Printf("   %d comments · by %s · %s · https://news.ycombinator.com/item?id=%d\n\n",
			item.Descendants, item.By, item.Age(), item.ID)
	}
}

// printItem prints a story and its comments to stdout in plain text.
func printItem(client *api.Client, id int) error {
	story, err := client.Item(id)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", story.Title)
	fmt.Printf("%s\n", strings.Repeat("─", len(story.Title)))
	if story.URL != "" {
		fmt.Printf("URL:      %s\n", story.URL)
	}
	fmt.Printf("Score:    %d\n", story.Score)
	fmt.Printf("Author:   %s\n", story.By)
	fmt.Printf("Posted:   %s\n", story.Age())
	fmt.Printf("Comments: %d\n", story.Descendants)
	fmt.Printf("HN:       https://news.ycombinator.com/item?id=%d\n", story.ID)
	if story.Text != "" {
		fmt.Printf("\n%s\n", util.StripHTML(story.Text))
	}

	if len(story.Kids) == 0 {
		return nil
	}

	limit := len(story.Kids)
	if limit > 50 {
		limit = 50
	}
	fmt.Printf("\n%s\n\n", strings.Repeat("─", 60))

	for _, kid := range story.Kids[:limit] {
		c, err := client.Item(kid)
		if err != nil || c == nil || c.Deleted || c.Dead {
			continue
		}
		fmt.Printf("%s  (%s)\n", c.By, c.Age())
		fmt.Printf("%s\n\n", util.StripHTML(c.Text))
	}
	return nil
}

// printUser prints a user profile to stdout in plain text.
func printUser(client *api.Client, username string) error {
	user, err := client.User(username)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user %q not found", username)
	}
	fmt.Printf("User:   %s\n", user.ID)
	fmt.Printf("Karma:  %d\n", user.Karma)
	if user.About != "" {
		fmt.Printf("About:  %s\n", util.StripHTML(user.About))
	}
	fmt.Printf("HN:     https://news.ycombinator.com/user?id=%s\n", user.ID)

	// Print up to 10 recent story submissions.
	count := 0
	fmt.Printf("\nRecent submissions:\n\n")
	for _, id := range user.Submitted {
		if count >= 10 {
			break
		}
		item, e := client.Item(id)
		if e != nil || item == nil || item.Type != "story" || item.Dead || item.Deleted {
			continue
		}
		fmt.Printf("  %s (%d pts · %d comments · %s)\n", item.Title, item.Score, item.Descendants, item.Age())
		if item.URL != "" {
			fmt.Printf("  %s\n", item.URL)
		}
		fmt.Println()
		count++
	}
	return nil
}
