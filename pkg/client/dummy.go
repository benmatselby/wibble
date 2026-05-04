package client

import (
	"time"

	"github.com/benmatselby/wibble/pkg/models"
)

// NewDummyClient creates a new dummy client for the RSS store. This is
// only used in the test suite.
func NewDummyClient() API {
	return &DummyClient{}
}

// DummyClient is a mock implementation of the API interface for testing purposes.
type DummyClient struct{}

// ParseURL returns a dummy feed and article for testing purposes.
func (c *DummyClient) ParseURL(url string) (*models.Feed, []models.Article, error) {
	feed := &models.Feed{
		ID:    0,
		URL:   "https://www.theguardian.com/uk/rss",
		Title: "The Guardian",
	}

	published, err := time.Parse("2006-01-02 15:04", "2022-07-24 08:04")
	if err != nil {
		return nil, nil, err
	}

	articles := []models.Article{
		{
			ID:        0,
			FeedID:    feed.ID,
			Title:     "Dummy Article",
			Link:      "https://www.theguardian.com/dummy-article",
			Published: &published,
			Summary:   "This is a dummy article for testing purposes.",
		},
	}
	return feed, articles, nil
}
