package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/benmatselby/wibble/pkg/models"
	"github.com/benmatselby/wibble/pkg/utils"
	"github.com/spf13/viper"
)

// NewSQLiteClient creates a new SQLite client for the RSS store.
func NewSQLiteClient() (DaoClient, error) {
	dbPath := os.ExpandEnv(viper.GetString("database"))

	// Ensure parent directories exist
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create directories for database: %w", err)
	}

	// Open the SQLite database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return NewSQLiteClientFromDB(db)
}

// NewSQLiteClientFromDB creates a SQLite client from an existing *sql.DB.
// Useful for testing with an in-memory database.
func NewSQLiteClientFromDB(db *sql.DB) (DaoClient, error) {
	c := &SQLiteClient{db: db}

	if err := c.installFeed(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := c.installArticle(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := c.installTag(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := c.installArticleTag(); err != nil {
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
		return fmt.Errorf("failed to create feeds table: %w", err)
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
		return fmt.Errorf("failed to create articles table: %w", err)
	}

	utils.Log("installArticle ended")
	return nil
}

// installTag initializes the tags table in the SQLite database.
func (c *SQLiteClient) installTag() error {
	utils.Log("installTag started")

	_, err := c.db.ExecContext(context.Background(), `
CREATE TABLE IF NOT EXISTS tags (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE
);`)
	if err != nil {
		return fmt.Errorf("failed to create tags table: %w", err)
	}

	utils.Log("installTag finished")
	return nil
}

// installArticleTag initializes the article_tags join table in the SQLite database.
func (c *SQLiteClient) installArticleTag() error {
	utils.Log("installArticleTag started")

	_, err := c.db.ExecContext(context.Background(), `
CREATE TABLE IF NOT EXISTS article_tags (
	article_id INTEGER NOT NULL,
	tag_id INTEGER NOT NULL,
	added_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE(article_id, tag_id),
	FOREIGN KEY(article_id) REFERENCES articles(id),
	FOREIGN KEY(tag_id) REFERENCES tags(id)
);`)
	if err != nil {
		return fmt.Errorf("failed to create article_tags table: %w", err)
	}

	if err := c.migrateArticleTagAddedAt(); err != nil {
		return err
	}

	utils.Log("installArticleTag finished")
	return nil
}

// migrateArticleTagAddedAt adds the added_at column to pre-existing
// article_tags tables that were created before the column was introduced.
func (c *SQLiteClient) migrateArticleTagAddedAt() error {
	rows, err := c.db.QueryContext(context.Background(), `PRAGMA table_info(article_tags)`)
	if err != nil {
		return fmt.Errorf("failed to inspect article_tags table: %w", err)
	}
	defer rows.Close()

	hasAddedAt := false
	for rows.Next() {
		var (
			cid       int
			name      string
			ctype     string
			notNull   int
			dfltValue sql.NullString
			pk        int
		)
		if err := rows.Scan(&cid, &name, &ctype, &notNull, &dfltValue, &pk); err != nil {
			return fmt.Errorf("failed to scan article_tags column info: %w", err)
		}
		if name == "added_at" {
			hasAddedAt = true
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating article_tags column info: %w", err)
	}

	if hasAddedAt {
		return nil
	}

	if _, err := c.db.ExecContext(context.Background(),
		`ALTER TABLE article_tags ADD COLUMN added_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP`); err != nil {
		return fmt.Errorf("failed to add added_at column to article_tags: %w", err)
	}

	return nil
}

// AddFeed inserts a new feed into the database and returns the model
func (c *SQLiteClient) AddFeed(feed models.Feed) (models.Feed, error) {
	utils.Log("AddFeed started")

	query := `
INSERT INTO feeds (url, title, added_at, sort_index)
VALUES (?, ?, ?, ?)
ON CONFLICT(url) DO UPDATE SET url=url, sort_index=excluded.sort_index
RETURNING id, url, title, added_at, sort_index;
`

	var result models.Feed
	ctx := context.Background()
	if err := c.
		db.QueryRowContext(ctx, query, feed.URL, feed.Title, feed.AddedAt, feed.SortIndex).
		Scan(&result.ID, &result.URL, &result.Title, &result.AddedAt, &result.SortIndex); err != nil {
		return feed, fmt.Errorf("failed to insert feed: %w", err)
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
		return nil, fmt.Errorf("failed to query feeds: %w", err)
	}
	defer rows.Close()

	var feeds []models.Feed
	for rows.Next() {
		var feed models.Feed
		if err := rows.Scan(&feed.ID, &feed.URL, &feed.Title, &feed.AddedAt, &feed.TotalCount, &feed.UnreadCount); err != nil {
			return nil, fmt.Errorf("failed to scan feed row: %w", err)
		}
		feeds = append(feeds, feed)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over feeds: %w", err)
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
		return fmt.Errorf("failed to insert article: %w", err)
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
		return nil, fmt.Errorf("failed to query articles: %w", err)
	}
	defer rows.Close()

	var articles []models.Article
	for rows.Next() {
		var article models.Article
		if err := rows.Scan(&article.ID, &article.FeedID, &article.Title, &article.Link, &article.Published, &article.Summary, &article.IsRead); err != nil {
			return nil, fmt.Errorf("failed to scan article row: %w", err)
		}
		articles = append(articles, article)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over articles: %w", err)
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
		return fmt.Errorf("failed to mark article ID %d as read: %w", articleID, err)
	}

	utils.Log(fmt.Sprintf("MarkArticleAsRead finished for article ID %d", articleID))
	return nil
}

func (c *SQLiteClient) MarkArticlesAsRead(feedID int64) error {
	utils.Log(fmt.Sprintf("MarkArticlesAsRead started for feed ID %d", feedID))

	_, err := c.db.ExecContext(context.Background(), "UPDATE articles SET is_read = 1 WHERE feed_id = ?", feedID)
	if err != nil {
		return fmt.Errorf("failed to mark articles as read for feed ID %d: %w", feedID, err)
	}
	utils.Log(fmt.Sprintf("MarkArticlesAsRead finished for feed ID %d", feedID))
	return nil
}

// DeleteFeedWithArticles deletes a feed and all its articles atomically within
// a single transaction. Returns the number of articles deleted.
func (c *SQLiteClient) DeleteFeedWithArticles(feedID int64) (int64, error) {
	utils.Log(fmt.Sprintf("DeleteFeedWithArticles started for feed ID %d", feedID))

	ctx := context.Background()

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction for feed ID %d: %w", feedID, err)
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.ExecContext(ctx, "DELETE FROM articles WHERE feed_id = ?", feedID)
	if err != nil {
		return 0, fmt.Errorf("failed to delete articles for feed ID %d: %w", feedID, err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected for feed ID %d: %w", feedID, err)
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM feeds WHERE id = ?", feedID); err != nil {
		return 0, fmt.Errorf("failed to delete feed ID %d: %w", feedID, err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction for feed ID %d: %w", feedID, err)
	}

	utils.Log(fmt.Sprintf("DeleteFeedWithArticles finished for feed ID %d, deleted %d articles", feedID, count))
	return count, nil
}

// AddTag inserts a new tag into the database if it doesn't already exist
// (case-insensitive match), and returns the resulting tag.
func (c *SQLiteClient) AddTag(name string) (models.Tag, error) {
	utils.Log(fmt.Sprintf("AddTag started for name %q", name))

	query := `
INSERT INTO tags (name)
VALUES (?)
ON CONFLICT(name) DO UPDATE SET name=name
RETURNING id, name;
`

	var result models.Tag
	ctx := context.Background()
	if err := c.
		db.QueryRowContext(ctx, query, name).
		Scan(&result.ID, &result.Name); err != nil {
		return result, fmt.Errorf("failed to insert tag: %w", err)
	}

	utils.Log(fmt.Sprintf("AddTag finished for name %q", name))
	return result, nil
}

// GetTags retrieves all tags from the database, along with the number of
// articles associated with each tag.
func (c *SQLiteClient) GetTags() ([]models.Tag, error) {
	utils.Log("GetTags started")

	rows, err := c.db.QueryContext(context.Background(), `
SELECT t.id, t.name, COUNT(at.article_id) AS article_count
FROM tags t
LEFT JOIN article_tags at ON at.tag_id = t.id
GROUP BY t.id, t.name
ORDER BY t.name COLLATE NOCASE`)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.ArticleCount); err != nil {
			return nil, fmt.Errorf("failed to scan tag row: %w", err)
		}
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over tags: %w", err)
	}

	utils.Log("GetTags finished")
	return tags, nil
}

// DeleteTag deletes a tag and its associations with articles.
func (c *SQLiteClient) DeleteTag(tagID int64) error {
	utils.Log(fmt.Sprintf("DeleteTag started for tag ID %d", tagID))

	ctx := context.Background()

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for tag ID %d: %w", tagID, err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, "DELETE FROM article_tags WHERE tag_id = ?", tagID); err != nil {
		return fmt.Errorf("failed to delete article associations for tag ID %d: %w", tagID, err)
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM tags WHERE id = ?", tagID); err != nil {
		return fmt.Errorf("failed to delete tag ID %d: %w", tagID, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction for tag ID %d: %w", tagID, err)
	}

	utils.Log(fmt.Sprintf("DeleteTag finished for tag ID %d", tagID))
	return nil
}

// AddTagToArticle associates a tag with an article, recording the time of
// association so the most recently added tag can be determined later.
func (c *SQLiteClient) AddTagToArticle(articleID, tagID int64) error {
	utils.Log(fmt.Sprintf("AddTagToArticle started for article ID %d, tag ID %d", articleID, tagID))

	// Use an explicit, nanosecond-precision timestamp rather than SQLite's
	// CURRENT_TIMESTAMP (which only has second resolution) so that tags
	// added in quick succession retain a stable, correct ordering.
	_, err := c.db.ExecContext(context.Background(),
		"INSERT OR IGNORE INTO article_tags (article_id, tag_id, added_at) VALUES (?, ?, ?)",
		articleID, tagID, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("failed to associate tag ID %d with article ID %d: %w", tagID, articleID, err)
	}

	utils.Log(fmt.Sprintf("AddTagToArticle finished for article ID %d, tag ID %d", articleID, tagID))
	return nil
}

// RemoveTagFromArticle removes the association between a tag and an article.
func (c *SQLiteClient) RemoveTagFromArticle(articleID, tagID int64) error {
	utils.Log(fmt.Sprintf("RemoveTagFromArticle started for article ID %d, tag ID %d", articleID, tagID))

	_, err := c.db.ExecContext(context.Background(),
		"DELETE FROM article_tags WHERE article_id = ? AND tag_id = ?",
		articleID, tagID)
	if err != nil {
		return fmt.Errorf("failed to remove tag ID %d from article ID %d: %w", tagID, articleID, err)
	}

	utils.Log(fmt.Sprintf("RemoveTagFromArticle finished for article ID %d, tag ID %d", articleID, tagID))
	return nil
}

// GetTagsForArticle retrieves all tags associated with a given article,
// ordered from least to most recently added, so the last element of the
// returned slice is the most recently added tag.
func (c *SQLiteClient) GetTagsForArticle(articleID int64) ([]models.Tag, error) {
	utils.Log(fmt.Sprintf("GetTagsForArticle started for article ID %d", articleID))

	rows, err := c.db.QueryContext(context.Background(), `
SELECT t.id, t.name
FROM tags t
JOIN article_tags at ON at.tag_id = t.id
WHERE at.article_id = ?
ORDER BY at.added_at ASC, t.id ASC`, articleID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags for article ID %d: %w", articleID, err)
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		if err := rows.Scan(&tag.ID, &tag.Name); err != nil {
			return nil, fmt.Errorf("failed to scan tag row: %w", err)
		}
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over tags for article ID %d: %w", articleID, err)
	}

	utils.Log(fmt.Sprintf("GetTagsForArticle finished for article ID %d", articleID))
	return tags, nil
}

// GetArticlesByTagID retrieves all articles associated with a given tag,
// ordered by published date descending.
func (c *SQLiteClient) GetArticlesByTagID(tagID int64) ([]models.Article, error) {
	utils.Log(fmt.Sprintf("GetArticlesByTagID started for tag ID %d", tagID))

	rows, err := c.db.QueryContext(context.Background(), `
SELECT a.id, a.feed_id, a.title, a.link, a.published, a.summary, a.is_read
FROM articles a
JOIN article_tags at ON at.article_id = a.id
WHERE at.tag_id = ?
ORDER BY a.published DESC`, tagID)
	if err != nil {
		return nil, fmt.Errorf("failed to query articles for tag ID %d: %w", tagID, err)
	}
	defer rows.Close()

	var articles []models.Article
	for rows.Next() {
		var article models.Article
		if err := rows.Scan(&article.ID, &article.FeedID, &article.Title, &article.Link, &article.Published, &article.Summary, &article.IsRead); err != nil {
			return nil, fmt.Errorf("failed to scan article row: %w", err)
		}
		articles = append(articles, article)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over articles for tag ID %d: %w", tagID, err)
	}

	utils.Log(fmt.Sprintf("GetArticlesByTagID finished for tag ID %d", tagID))
	return articles, nil
}

// Close closes the database connection.
func (c *SQLiteClient) Close() error {
	utils.Log("Close started")
	if c.db != nil {
		if err := c.db.Close(); err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
	}

	utils.Log("Close finished")
	return nil
}
