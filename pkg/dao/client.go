// Package dao provides data access operations for the application.
package dao

import (
	"github.com/benmatselby/wibble/pkg/models"
	_ "github.com/mattn/go-sqlite3"
)

// DaoClient is an interface for data access operations. Allows us to swap out
// implementations if needed and helps with testing.
type DaoClient interface {
	// Feeds
	AddFeed(feed models.Feed) (models.Feed, error)
	GetFeeds() ([]models.Feed, error)
	DeleteFeedWithArticles(feedID int64) (int64, error)

	// Articles
	AddArticle(article models.Article) error
	GetArticlesByFeedID(feedID int64) ([]models.Article, error)
	GetArticleByID(articleID int64) (*models.Article, error)
	MarkArticleAsRead(articleID int64) error
	MarkArticlesAsRead(feedID int64) error

	// Maintenance
	Close() error
}
