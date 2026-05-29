package cmd

import (
	"bytes"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/benmatselby/wibble/pkg/dao"
	"github.com/benmatselby/wibble/pkg/models"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
)

// newCleanTestClient opens an in-memory SQLite database, installs the schema,
// and registers cleanup — mirroring the pattern in pkg/dao/sqlite_test.go.
func newCleanTestClient(t *testing.T) dao.DaoClient {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}

	c, err := dao.NewSQLiteClientFromDB(db)
	if err != nil {
		t.Fatalf("NewSQLiteClientFromDB: %v", err)
	}

	t.Cleanup(func() { _ = c.Close() })

	return c
}

// runClean sets up viper with the given config feed URLs, then calls the clean
// command's RunE directly, returning captured stdout and any error.
func runClean(t *testing.T, db dao.DaoClient, configURLs []string, args []string) (string, error) {
	t.Helper()

	viper.Reset()
	viper.Set("feeds", configURLs)

	cmd := NewCleanCommand(&db)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return buf.String(), err
}

func TestCleanCommand_NoStaleFeedsDoesNothing(t *testing.T) {
	c := newCleanTestClient(t)

	now := time.Now().UTC()
	_, err := c.AddFeed(models.Feed{URL: "https://a.com/feed", Title: "Feed A", AddedAt: now})
	if err != nil {
		t.Fatalf("AddFeed: %v", err)
	}

	out, err := runClean(t, c, []string{"https://a.com/feed"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "Removed 0 feed(s) and 0 article(s).") {
		t.Errorf("unexpected output: %q", out)
	}

	feeds, _ := c.GetFeeds()
	if len(feeds) != 1 {
		t.Errorf("expected feed to be retained, got %d feeds", len(feeds))
	}
}

func TestCleanCommand_RemovesStaleFeed(t *testing.T) {
	c := newCleanTestClient(t)

	now := time.Now().UTC()
	_, _ = c.AddFeed(models.Feed{URL: "https://keep.com/feed", Title: "Keep Feed", AddedAt: now})
	stale, err := c.AddFeed(models.Feed{URL: "https://stale.com/feed", Title: "Stale Feed", AddedAt: now})
	if err != nil {
		t.Fatalf("AddFeed: %v", err)
	}

	_ = c.AddArticle(models.Article{FeedID: stale.ID, Title: "A1", Link: "https://stale.com/1", Published: &now})
	_ = c.AddArticle(models.Article{FeedID: stale.ID, Title: "A2", Link: "https://stale.com/2", Published: &now})

	out, err := runClean(t, c, []string{"https://keep.com/feed"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "Removing feed: Stale Feed (https://stale.com/feed)") {
		t.Errorf("expected removal message in output, got: %q", out)
	}
	if !strings.Contains(out, "Removed 1 feed(s) and 2 article(s).") {
		t.Errorf("unexpected summary in output: %q", out)
	}

	feeds, _ := c.GetFeeds()
	if len(feeds) != 1 {
		t.Errorf("expected 1 feed remaining, got %d", len(feeds))
	}
	if feeds[0].URL != "https://keep.com/feed" {
		t.Errorf("expected retained feed URL %q, got %q", "https://keep.com/feed", feeds[0].URL)
	}

	articles, _ := c.GetArticlesByFeedID(stale.ID)
	if len(articles) != 0 {
		t.Errorf("expected 0 articles for deleted feed, got %d", len(articles))
	}
}

func TestCleanCommand_RemovesMultipleStaleFeeds(t *testing.T) {
	c := newCleanTestClient(t)

	now := time.Now().UTC()
	_, _ = c.AddFeed(models.Feed{URL: "https://keep.com/feed", Title: "Keep Feed", AddedAt: now})
	s1, _ := c.AddFeed(models.Feed{URL: "https://stale1.com/feed", Title: "Stale One", AddedAt: now})
	s2, _ := c.AddFeed(models.Feed{URL: "https://stale2.com/feed", Title: "Stale Two", AddedAt: now})

	_ = c.AddArticle(models.Article{FeedID: s1.ID, Title: "S1A1", Link: "https://stale1.com/1", Published: &now})
	_ = c.AddArticle(models.Article{FeedID: s2.ID, Title: "S2A1", Link: "https://stale2.com/1", Published: &now})
	_ = c.AddArticle(models.Article{FeedID: s2.ID, Title: "S2A2", Link: "https://stale2.com/2", Published: &now})

	out, err := runClean(t, c, []string{"https://keep.com/feed"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "Removed 2 feed(s) and 3 article(s).") {
		t.Errorf("unexpected summary: %q", out)
	}

	feeds, _ := c.GetFeeds()
	if len(feeds) != 1 {
		t.Errorf("expected 1 feed remaining, got %d", len(feeds))
	}
}

func TestCleanCommand_DryRunDoesNotDelete(t *testing.T) {
	c := newCleanTestClient(t)

	now := time.Now().UTC()
	stale, _ := c.AddFeed(models.Feed{URL: "https://stale.com/feed", Title: "Stale Feed", AddedAt: now})
	_ = c.AddArticle(models.Article{FeedID: stale.ID, Title: "A1", Link: "https://stale.com/1", Published: &now})

	out, err := runClean(t, c, []string{}, []string{"--dry-run"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "Would remove 1 feed(s) and 1 article(s) (dry run).") {
		t.Errorf("unexpected dry-run output: %q", out)
	}

	feeds, _ := c.GetFeeds()
	if len(feeds) != 1 {
		t.Errorf("expected feed to still exist after dry-run, got %d feeds", len(feeds))
	}

	articles, _ := c.GetArticlesByFeedID(stale.ID)
	if len(articles) != 1 {
		t.Errorf("expected articles to still exist after dry-run, got %d articles", len(articles))
	}
}

func TestCleanCommand_DryRunWithNoStaleFeeds(t *testing.T) {
	c := newCleanTestClient(t)

	now := time.Now().UTC()
	_, _ = c.AddFeed(models.Feed{URL: "https://a.com/feed", Title: "Feed A", AddedAt: now})

	out, err := runClean(t, c, []string{"https://a.com/feed"}, []string{"--dry-run"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "Would remove 0 feed(s) and 0 article(s) (dry run).") {
		t.Errorf("unexpected dry-run output: %q", out)
	}
}

func TestCleanCommand_DBUnavailableReturnsError(t *testing.T) {
	// Close the DB before running to force (*db).GetFeeds() to fail.
	c := newCleanTestClient(t)
	_ = c.Close()

	_, err := runClean(t, c, []string{}, nil)
	if err == nil {
		t.Fatal("expected an error when DB is closed, got nil")
	}
}
