package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const baseURL = "https://hacker-news.firebaseio.com/v0"

// Client is an HN Firebase API client.
type Client struct {
	http *http.Client
}

// New returns a new Client.
func New() *Client {
	return &Client{
		http: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) get(url string, v any) error {
	resp, err := c.http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HN API: %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

// Item fetches a single item by ID.
func (c *Client) Item(id int) (*Item, error) {
	var item Item
	if err := c.get(fmt.Sprintf("%s/item/%d.json", baseURL, id), &item); err != nil {
		return nil, err
	}
	return &item, nil
}

// User fetches a user by username.
func (c *Client) User(username string) (*User, error) {
	var user User
	if err := c.get(fmt.Sprintf("%s/user/%s.json", baseURL, username), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// list fetches a named list of item IDs (e.g. topstories, newstories).
func (c *Client) list(name string) ([]int, error) {
	var ids []int
	if err := c.get(fmt.Sprintf("%s/%s.json", baseURL, name), &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// Stories fetches the top N items from a named list, in parallel.
func (c *Client) Stories(listName string, n int) ([]*Item, error) {
	ids, err := c.list(listName)
	if err != nil {
		return nil, err
	}
	if n > len(ids) {
		n = len(ids)
	}
	ids = ids[:n]

	items := make([]*Item, n)
	errs := make([]error, n)
	var wg sync.WaitGroup
	// Limit concurrency to avoid overwhelming the API.
	sem := make(chan struct{}, 20)
	for i, id := range ids {
		wg.Add(1)
		go func(i, id int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			item, err := c.Item(id)
			items[i] = item
			errs[i] = err
		}(i, id)
	}
	wg.Wait()

	// Filter errors: return first non-nil error but still return partial results.
	var firstErr error
	result := make([]*Item, 0, n)
	for i, item := range items {
		if errs[i] != nil && firstErr == nil {
			firstErr = errs[i]
		}
		if item != nil {
			result = append(result, item)
		}
	}
	return result, firstErr
}

// TopStories returns the top N front-page stories.
func (c *Client) TopStories(n int) ([]*Item, error) { return c.Stories("topstories", n) }

// NewStories returns the N newest stories.
func (c *Client) NewStories(n int) ([]*Item, error) { return c.Stories("newstories", n) }

// BestStories returns the N best stories.
func (c *Client) BestStories(n int) ([]*Item, error) { return c.Stories("beststories", n) }

// AskStories returns the N latest Ask HN stories.
func (c *Client) AskStories(n int) ([]*Item, error) { return c.Stories("askstories", n) }

// ShowStories returns the N latest Show HN stories.
func (c *Client) ShowStories(n int) ([]*Item, error) { return c.Stories("showstories", n) }

// JobStories returns the N latest job stories.
func (c *Client) JobStories(n int) ([]*Item, error) { return c.Stories("jobstories", n) }
