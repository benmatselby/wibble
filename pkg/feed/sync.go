// Package feed provides functionality for syncing RSS feeds.
package feed

import (
	"fmt"
	"os"
	"time"

	"github.com/benmatselby/wibble/pkg/client"
	"github.com/benmatselby/wibble/pkg/config"
	"github.com/benmatselby/wibble/pkg/dao"
	"github.com/benmatselby/wibble/pkg/models"
)

// ProgressFunc is called by Sync for each feed as it is processed.
// title is the feed title (or URL if the title is not yet known), and err is
// non-nil if that feed failed to sync.
type ProgressFunc func(title string, err error)

// Sync reads feed URLs from the configuration file, ensures they exist in the
// database, then fetches and stores any new articles for each feed.
// progress is called for each feed as it is processed; it may be nil.
func Sync(db dao.DaoClient, c client.API, progress ProgressFunc) error {
	configFeeds, err := config.GetFeeds()
	if err != nil {
		return err
	}

	for feedIndex, url := range configFeeds {
		feed, articles, err := c.ParseURL(url)
		if err != nil {
			if progress != nil {
				progress(url, err)
			}
			continue
		}

		updatedFeed, err := db.AddFeed(models.Feed{URL: feed.URL, Title: feed.Title, AddedAt: time.Now(), SortIndex: feedIndex})
		if err != nil {
			if progress != nil {
				progress(feed.Title, err)
			}
			continue
		}

		if progress != nil {
			progress(updatedFeed.Title, nil)
		}

		for _, article := range articles {
			article.FeedID = updatedFeed.ID
			if err := db.AddArticle(article); err != nil {
				fmt.Fprintf(os.Stderr, "failed to store article %q: %v\n", article.Link, err)
			}
		}
	}

	return nil
}
