package dao

import (
	"database/sql"
	"testing"
	"time"

	"github.com/benmatselby/wibble/pkg/models"
)

// newTestClient opens an in-memory SQLite database and initialises the schema.
func newTestClient(t *testing.T) *SQLiteClient {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}

	c := &SQLiteClient{db: db}

	if err := c.installFeed(); err != nil {
		t.Fatalf("installFeed: %v", err)
	}
	if err := c.installArticle(); err != nil {
		t.Fatalf("installArticle: %v", err)
	}

	t.Cleanup(func() { _ = c.Close() })

	return c
}

func TestInstallSchemaIsIdempotent(t *testing.T) {
	c := newTestClient(t)

	// Running install a second time should not return an error.
	if err := c.installFeed(); err != nil {
		t.Errorf("second installFeed call failed: %v", err)
	}
	if err := c.installArticle(); err != nil {
		t.Errorf("second installArticle call failed: %v", err)
	}
}

func TestAddFeed(t *testing.T) {
	c := newTestClient(t)

	feed := models.Feed{
		URL:       "https://example.com/feed",
		Title:     "Example Feed",
		AddedAt:   time.Now().UTC().Truncate(time.Second),
		SortIndex: 1,
	}

	got, err := c.AddFeed(feed)
	if err != nil {
		t.Fatalf("AddFeed: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("expected ID after insert, got %v", got.ID)
	}
	if got.URL != feed.URL {
		t.Errorf("URL: got %q, want %q", got.URL, feed.URL)
	}
	if got.Title != feed.Title {
		t.Errorf("Title: got %q, want %q", got.Title, feed.Title)
	}
	if got.SortIndex != feed.SortIndex {
		t.Errorf("SortIndex: got %d, want %d", got.SortIndex, feed.SortIndex)
	}
}

func TestAddFeed_DuplicateURLIsIgnored(t *testing.T) {
	c := newTestClient(t)

	feed := models.Feed{
		URL:     "https://example.com/feed",
		Title:   "Example Feed",
		AddedAt: time.Now().UTC().Truncate(time.Second),
	}

	first, err := c.AddFeed(feed)
	if err != nil {
		t.Fatalf("first AddFeed: %v", err)
	}

	second, err := c.AddFeed(feed)
	if err != nil {
		t.Fatalf("second AddFeed: %v", err)
	}

	if first.ID != second.ID {
		t.Errorf("expected same ID on duplicate insert: first=%d second=%d", first.ID, second.ID)
	}
}

func TestGetFeeds_Empty(t *testing.T) {
	c := newTestClient(t)

	feeds, err := c.GetFeeds()
	if err != nil {
		t.Fatalf("GetFeeds: %v", err)
	}

	if len(feeds) != 0 {
		t.Errorf("expected 0 feeds, got %d", len(feeds))
	}
}

func TestGetFeeds_ReturnsFeedsWithCounts(t *testing.T) {
	c := newTestClient(t)

	feed := models.Feed{
		URL:     "https://example.com/feed",
		Title:   "Example",
		AddedAt: time.Now().UTC().Truncate(time.Second),
	}
	inserted, err := c.AddFeed(feed)
	if err != nil {
		t.Fatalf("AddFeed: %v", err)
	}

	now := time.Now().UTC()
	articles := []models.Article{
		{FeedID: inserted.ID, Title: "Article 1", Link: "https://example.com/1", Published: &now},
		{FeedID: inserted.ID, Title: "Article 2", Link: "https://example.com/2", Published: &now},
	}
	for _, a := range articles {
		if err := c.AddArticle(a); err != nil {
			t.Fatalf("AddArticle: %v", err)
		}
	}

	feeds, err := c.GetFeeds()
	if err != nil {
		t.Fatalf("GetFeeds: %v", err)
	}

	if len(feeds) != 1 {
		t.Fatalf("expected 1 feed, got %d", len(feeds))
	}

	f := feeds[0]
	if f.TotalCount != 2 {
		t.Errorf("TotalCount: got %d, want 2", f.TotalCount)
	}
	if f.UnreadCount != 2 {
		t.Errorf("UnreadCount: got %d, want 2", f.UnreadCount)
	}
}

func TestAddArticle_DuplicateIsIgnored(t *testing.T) {
	c := newTestClient(t)

	feed, err := c.AddFeed(models.Feed{
		URL:     "https://example.com/feed",
		AddedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("AddFeed: %v", err)
	}

	now := time.Now().UTC()
	article := models.Article{
		FeedID:    feed.ID,
		Title:     "Article",
		Link:      "https://example.com/1",
		Published: &now,
	}

	if err := c.AddArticle(article); err != nil {
		t.Fatalf("first AddArticle: %v", err)
	}
	if err := c.AddArticle(article); err != nil {
		t.Errorf("second AddArticle (duplicate) should not error: %v", err)
	}

	articles, err := c.GetArticlesByFeedID(feed.ID)
	if err != nil {
		t.Fatalf("GetArticlesByFeedID: %v", err)
	}
	if len(articles) != 1 {
		t.Errorf("expected 1 article after duplicate insert, got %d", len(articles))
	}
}

func TestGetArticlesByFeedID_OrderedByPublishedDesc(t *testing.T) {
	c := newTestClient(t)

	feed, err := c.AddFeed(models.Feed{
		URL:     "https://example.com/feed",
		AddedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("AddFeed: %v", err)
	}

	older := time.Now().UTC().Add(-24 * time.Hour)
	newer := time.Now().UTC()

	if err := c.AddArticle(models.Article{FeedID: feed.ID, Title: "Older", Link: "https://example.com/older", Published: &older}); err != nil {
		t.Fatalf("AddArticle older: %v", err)
	}
	if err := c.AddArticle(models.Article{FeedID: feed.ID, Title: "Newer", Link: "https://example.com/newer", Published: &newer}); err != nil {
		t.Fatalf("AddArticle newer: %v", err)
	}

	articles, err := c.GetArticlesByFeedID(feed.ID)
	if err != nil {
		t.Fatalf("GetArticlesByFeedID: %v", err)
	}

	if len(articles) != 2 {
		t.Fatalf("expected 2 articles, got %d", len(articles))
	}
	if articles[0].Title != "Newer" {
		t.Errorf("expected first article to be 'Newer', got %q", articles[0].Title)
	}
}

func TestGetArticlesByFeedID_OnlyReturnsFeedArticles(t *testing.T) {
	c := newTestClient(t)

	feedA, _ := c.AddFeed(models.Feed{URL: "https://a.com/feed", AddedAt: time.Now().UTC()})
	feedB, _ := c.AddFeed(models.Feed{URL: "https://b.com/feed", AddedAt: time.Now().UTC()})

	now := time.Now().UTC()
	_ = c.AddArticle(models.Article{FeedID: feedA.ID, Title: "A1", Link: "https://a.com/1", Published: &now})
	_ = c.AddArticle(models.Article{FeedID: feedB.ID, Title: "B1", Link: "https://b.com/1", Published: &now})

	articles, err := c.GetArticlesByFeedID(feedA.ID)
	if err != nil {
		t.Fatalf("GetArticlesByFeedID: %v", err)
	}

	if len(articles) != 1 {
		t.Errorf("expected 1 article for feed A, got %d", len(articles))
	}
	if articles[0].Title != "A1" {
		t.Errorf("unexpected article title: %q", articles[0].Title)
	}
}

func TestGetArticleID(t *testing.T) {
	c := newTestClient(t)

	feed, err := c.AddFeed(models.Feed{
		URL:     "https://example.com/feed",
		AddedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("AddFeed: %v", err)
	}

	if err := c.AddArticle(models.Article{FeedID: feed.ID, Title: "New Blackadder season", Link: "https://example.com/older"}); err != nil {
		t.Fatalf("AddArticle: %v", err)
	}

	article, err := c.GetArticleByID(1)
	if err != nil {
		t.Fatalf("GetArticleByID: %v", err)
	}

	if article == nil {
		t.Fatal("expected article, got nil")
	}

	if article.Title != "New Blackadder season" {
		t.Errorf("expected title 'New Blackadder season', got %q", article.Title)
	}
}

func TestMarkArticleAsRead(t *testing.T) {
	c := newTestClient(t)

	feed, _ := c.AddFeed(models.Feed{URL: "https://example.com/feed", AddedAt: time.Now().UTC()})
	now := time.Now().UTC()
	_ = c.AddArticle(models.Article{FeedID: feed.ID, Title: "A", Link: "https://example.com/1", Published: &now})

	articles, err := c.GetArticlesByFeedID(feed.ID)
	if err != nil {
		t.Fatalf("GetArticlesByFeedID: %v", err)
	}

	if err := c.MarkArticleAsRead(articles[0].ID); err != nil {
		t.Fatalf("MarkArticleAsRead: %v", err)
	}

	articles, err = c.GetArticlesByFeedID(feed.ID)
	if err != nil {
		t.Fatalf("GetArticlesByFeedID after mark: %v", err)
	}
	if !articles[0].IsRead {
		t.Error("expected article to be marked as read")
	}

	feeds, err := c.GetFeeds()
	if err != nil {
		t.Fatalf("GetFeeds: %v", err)
	}
	if feeds[0].UnreadCount != 0 {
		t.Errorf("expected UnreadCount 0, got %d", feeds[0].UnreadCount)
	}
}

func TestMarkArticlesAsRead(t *testing.T) {
	c := newTestClient(t)

	feed, _ := c.AddFeed(models.Feed{URL: "https://example.com/feed", AddedAt: time.Now().UTC()})
	now := time.Now().UTC()
	_ = c.AddArticle(models.Article{FeedID: feed.ID, Title: "A", Link: "https://example.com/1", Published: &now})
	_ = c.AddArticle(models.Article{FeedID: feed.ID, Title: "B", Link: "https://example.com/2", Published: &now})

	if err := c.MarkArticlesAsRead(feed.ID); err != nil {
		t.Fatalf("MarkArticlesAsRead: %v", err)
	}

	articles, err := c.GetArticlesByFeedID(feed.ID)
	if err != nil {
		t.Fatalf("GetArticlesByFeedID: %v", err)
	}

	for _, a := range articles {
		if !a.IsRead {
			t.Errorf("expected article %q to be marked as read", a.Title)
		}
	}
}

func TestClose(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	c := &SQLiteClient{db: db}
	if err := c.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

func TestClose_NilDB(t *testing.T) {
	c := &SQLiteClient{db: nil}
	if err := c.Close(); err != nil {
		t.Errorf("Close with nil db should not error: %v", err)
	}
}

func TestDeleteFeedWithArticles(t *testing.T) {
	c := newTestClient(t)

	now := time.Now().UTC()

	keep, _ := c.AddFeed(models.Feed{URL: "https://keep.com/feed", Title: "Keep", AddedAt: now})
	stale, err := c.AddFeed(models.Feed{URL: "https://stale.com/feed", Title: "Stale", AddedAt: now})
	if err != nil {
		t.Fatalf("AddFeed: %v", err)
	}

	_ = c.AddArticle(models.Article{FeedID: stale.ID, Title: "S1", Link: "https://stale.com/1", Published: &now})
	_ = c.AddArticle(models.Article{FeedID: stale.ID, Title: "S2", Link: "https://stale.com/2", Published: &now})
	_ = c.AddArticle(models.Article{FeedID: keep.ID, Title: "K1", Link: "https://keep.com/1", Published: &now})

	count, err := c.DeleteFeedWithArticles(stale.ID)
	if err != nil {
		t.Fatalf("DeleteFeedWithArticles: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 articles deleted, got %d", count)
	}

	feeds, err := c.GetFeeds()
	if err != nil {
		t.Fatalf("GetFeeds: %v", err)
	}
	if len(feeds) != 1 {
		t.Errorf("expected 1 feed remaining, got %d", len(feeds))
	}
	if feeds[0].URL != "https://keep.com/feed" {
		t.Errorf("expected keep feed to remain, got %q", feeds[0].URL)
	}

	staleArticles, err := c.GetArticlesByFeedID(stale.ID)
	if err != nil {
		t.Fatalf("GetArticlesByFeedID for stale feed: %v", err)
	}
	if len(staleArticles) != 0 {
		t.Errorf("expected 0 articles for deleted feed, got %d", len(staleArticles))
	}

	keepArticles, err := c.GetArticlesByFeedID(keep.ID)
	if err != nil {
		t.Fatalf("GetArticlesByFeedID for keep feed: %v", err)
	}
	if len(keepArticles) != 1 {
		t.Errorf("expected keep feed articles to be untouched (1), got %d", len(keepArticles))
	}
}
