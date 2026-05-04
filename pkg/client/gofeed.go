package client

import (
	"github.com/benmatselby/wibble/pkg/models"
	"github.com/mmcdole/gofeed"
)

type GoFeedClient struct {
	Client *gofeed.Parser
}

// NewGoFeedClient creates a new GoFeed client for the RSS store.
func NewGoFeedClient() API {
	return &GoFeedClient{
		Client: gofeed.NewParser(),
	}
}

// ParseURL parses the RSS feed from the given URL and returns the feed and its articles.
func (c *GoFeedClient) ParseURL(url string) (*models.Feed, []models.Article, error) {
	feed, err := c.Client.ParseURL(url)
	if err != nil {
		return nil, nil, err
	}

	var items []models.Article
	for _, item := range feed.Items {
		items = append(items, models.Article{
			Title:     item.Title,
			Link:      item.Link,
			Published: item.PublishedParsed,
			Summary:   item.Description,
		})
	}

	modelFeed := &models.Feed{
		Title: feed.Title,
		URL:   url,
	}

	return modelFeed, items, nil
}
