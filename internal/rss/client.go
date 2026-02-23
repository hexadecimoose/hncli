package rss

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const hnrssBase = "https://hnrss.org"

// Item is an RSS feed item from hnrss.org.
type Item struct {
	Title    string `xml:"title"`
	Link     string `xml:"link"`
	Comments string `xml:"comments"`
	PubDate  string `xml:"pubDate"`
	Creator  string `xml:"creator"`
	GUID     string `xml:"guid"`
}

type channel struct {
	Items []Item `xml:"item"`
}

type rss struct {
	Channel channel `xml:"channel"`
}

// Client is an hnrss.org client.
type Client struct {
	http *http.Client
}

// New returns a new Client.
func New() *Client {
	return &Client{http: &http.Client{Timeout: 10 * time.Second}}
}

func (c *Client) fetch(endpoint string, params url.Values) ([]Item, error) {
	u := fmt.Sprintf("%s/%s", hnrssBase, endpoint)
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	resp, err := c.http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var feed rss
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, err
	}
	return feed.Channel.Items, nil
}

// Search returns RSS items matching query q (optionally limited to n results).
func (c *Client) Search(q string, n int) ([]Item, error) {
	params := url.Values{"q": {q}, "count": {fmt.Sprint(n)}}
	return c.fetch("newest", params)
}

// Feed returns items from a named feed (frontpage, newest, ask, show, jobs, best).
func (c *Client) Feed(name string, n int) ([]Item, error) {
	params := url.Values{"count": {fmt.Sprint(n)}}
	return c.fetch(name, params)
}
