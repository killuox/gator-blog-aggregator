package rss

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func FetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(context.Background(), "GET", feedURL, nil)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("Error getting feed url: %s", err)
	}

	req.Header.Set("User-Agent", "gator")

	resp, err := client.Do(req)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("Error reading body: %s", err)
	}

	body, err := io.ReadAll(resp.Body)

	var response RSSFeed
	err = xml.Unmarshal(body, &response)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("Error unmarshalling XML: %s", err)
	}

	sanitizeHtml(&response)
	return &response, nil
}

func sanitizeHtml(feed *RSSFeed) {
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

	for idx, item := range feed.Channel.Item {
		feed.Channel.Item[idx].Title = html.UnescapeString(item.Title)
		feed.Channel.Item[idx].Description = html.UnescapeString(item.Description)
	}
}
