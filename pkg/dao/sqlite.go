package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/benmatselby/wibble/pkg/models"
	"github.com/benmatselby/wibble/pkg/utils"
	"github.com/spf13/viper"
)

// NewSQLiteClient creates a new SQLite client for the RSS store.
func NewSQLiteClient() (DaoClient, error) {
	dbPath := os.ExpandEnv(viper.GetString("database"))

	// Ensure parent directories exist
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create directories for database: %v", err)
	}

	// Open the SQLite database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	c := &SQLiteClient{db: db}

	if err := c.installFeed(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := c.installArticle(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return c, nil
}

type SQLiteClient struct {
	db *sql.DB
}

// installFeed initializes the feeds table in the SQLite database.
func (c *SQLiteClient) installFeed() error {
	utils.Log("installFeed started")

	_, err := c.db.ExecContext(context.Background(), `
CREATE TABLE IF NOT EXISTS feeds (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	url TEXT NOT NULL UNIQUE,
	title TEXT,
	added_at DATETIME NOT NULL,
	sort_index INTEGER NOT NULL DEFAULT 0
);`)
	if err != nil {
		return fmt.Errorf("failed to create feeds table: %v", err)
	}

	utils.Log("installFeed finished")
	return nil
}

// installArticle initializes the articles table in the SQLite database.
func (c *SQLiteClient) installArticle() error {
	utils.Log("installArticle started")

	_, err := c.db.ExecContext(context.Background(), `
CREATE TABLE IF NOT EXISTS articles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    feed_id INTEGER NOT NULL,
    title TEXT,
    link TEXT,
    published DATETIME,
    summary TEXT,
    is_read INTEGER NOT NULL DEFAULT 0,
    UNIQUE(feed_id, link),
    FOREIGN KEY(feed_id) REFERENCES feeds(id)
);`)
	if err != nil {
		return fmt.Errorf("failed to create articles table: %v", err)
	}

	utils.Log("installArticle ended")
	return nil
}

// AddFeed inserts a new feed into the database and returns the model
func (c *SQLiteClient) AddFeed(feed models.Feed) (models.Feed, error) {
	utils.Log("AddFeed started")

	query := `
INSERT INTO feeds (url, title, added_at, sort_index)
VALUES (?, ?, ?, ?)
ON CONFLICT(url) DO UPDATE SET url=url
RETURNING id, url, title, added_at, sort_index;
`

	var result models.Feed
	ctx := context.Background()
	if err := c.
		db.QueryRowContext(ctx, query, feed.URL, feed.Title, feed.AddedAt, feed.SortIndex).
		Scan(&result.ID, &result.URL, &result.Title, &result.AddedAt, &result.SortIndex); err != nil {
		return feed, fmt.Errorf("failed to insert feed: %v", err)
	}

	utils.Log("AddFeed finished")

	return result, nil
}

// GetFeeds retrieves all feeds from the database.
func (c *SQLiteClient) GetFeeds() ([]models.Feed, error) {
	utils.Log("GetFeeds started")

	rows, err := c.db.QueryContext(context.Background(), `
SELECT f.id, f.url, f.title, f.added_at,
       COUNT(a.id) AS total_count,
       SUM(CASE WHEN a.is_read = 0 THEN 1 ELSE 0 END) AS unread_count
FROM feeds f
LEFT JOIN articles a ON a.feed_id = f.id
GROUP BY f.id, f.url, f.title, f.added_at
ORDER BY sort_index`)
	if err != nil {
		return nil, fmt.Errorf("failed to query feeds: %v", err)
	}
	defer rows.Close()

	var feeds []models.Feed
	for rows.Next() {
		var feed models.Feed
		if err := rows.Scan(&feed.ID, &feed.URL, &feed.Title, &feed.AddedAt, &feed.TotalCount, &feed.UnreadCount); err != nil {
			return nil, fmt.Errorf("failed to scan feed row: %v", err)
		}
		feeds = append(feeds, feed)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over feeds: %v", err)
	}

	utils.Log("GetFeeds finished")
	return feeds, nil
}

// AddArticle inserts a new article into the database.
func (c *SQLiteClient) AddArticle(article models.Article) error {
	utils.Log(fmt.Sprintf("AddArticle started for feed ID %d", article.FeedID))

	_, err := c.db.ExecContext(context.Background(),
		"INSERT OR IGNORE INTO articles (feed_id, title, link, published, summary) VALUES (?, ?, ?, ?, ?)",
		article.FeedID, article.Title, article.Link, article.Published, article.Summary)
	if err != nil {
		return fmt.Errorf("failed to insert article: %v", err)
	}

	utils.Log(fmt.Sprintf("AddArticle finished for feed ID %d", article.FeedID))
	return nil
}

// GetArticlesByFeedID retrieves all articles for a given feed from the database, ordered by published date descending.
func (c *SQLiteClient) GetArticlesByFeedID(feedID int64) ([]models.Article, error) {
	utils.Log(fmt.Sprintf("GetArticlesByFeedID started for feed ID %d", feedID))

	rows, err := c.db.QueryContext(context.Background(),
		"SELECT id, feed_id, title, link, published, summary, is_read FROM articles WHERE feed_id = ? ORDER BY published DESC",
		feedID)
	if err != nil {
		return nil, fmt.Errorf("failed to query articles: %v", err)
	}
	defer rows.Close()

	var articles []models.Article
	for rows.Next() {
		var article models.Article
		if err := rows.Scan(&article.ID, &article.FeedID, &article.Title, &article.Link, &article.Published, &article.Summary, &article.IsRead); err != nil {
			return nil, fmt.Errorf("failed to scan article row: %v", err)
		}
		articles = append(articles, article)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over articles: %v", err)
	}

	utils.Log(fmt.Sprintf("GetArticlesByFeedID finished for feed ID %d", feedID))
	return articles, nil
}

// GetArticleByID retrieves a single article by its ID from the database.
func (c *SQLiteClient) GetArticleByID(articleID int64) (*models.Article, error) {
	utils.Log(fmt.Sprintf("GetArticleByID started for article ID %d", articleID))

	row := c.db.QueryRowContext(context.Background(),
		"SELECT id, feed_id, title, link, published, summary, is_read FROM articles WHERE id = ?",
		articleID)

	var article models.Article
	if err := row.Scan(&article.ID, &article.FeedID, &article.Title, &article.Link, &article.Published, &article.Summary, &article.IsRead); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("article %d not found", articleID)
		}
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	utils.Log(fmt.Sprintf("GetArticleByID finished for article ID %d", articleID))
	return &article, nil
}

// MarkArticleAsRead updates the read status of an article to true in the database.
func (c *SQLiteClient) MarkArticleAsRead(articleID int64) error {
	utils.Log(fmt.Sprintf("MarkArticleAsRead started for article ID %d", articleID))

	_, err := c.db.ExecContext(context.Background(), "UPDATE articles SET is_read = 1 WHERE id = ?", articleID)
	if err != nil {
		return fmt.Errorf("failed to mark article ID %d as read: %v", articleID, err)
	}

	utils.Log(fmt.Sprintf("MarkArticleAsRead finished for article ID %d", articleID))
	return nil
}

func (c *SQLiteClient) MarkArticlesAsRead(feedID int64) error {
	utils.Log(fmt.Sprintf("MarkArticlesAsRead started for feed ID %d", feedID))

	_, err := c.db.ExecContext(context.Background(), "UPDATE articles SET is_read = 1 WHERE feed_id = ?", feedID)
	if err != nil {
		return fmt.Errorf("failed to mark articles as read for feed ID %d: %v", feedID, err)
	}
	utils.Log(fmt.Sprintf("MarkArticlesAsRead finished for feed ID %d", feedID))
	return nil
}

// Close closes the database connection.
func (c *SQLiteClient) Close() error {
	utils.Log("Close started")
	if c.db != nil {
		if err := c.db.Close(); err != nil {
			return fmt.Errorf("failed to close database: %v", err)
		}
	}

	utils.Log("Close finished")
	return nil
}
