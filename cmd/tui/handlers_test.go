package tui

import (
	"fmt"
	"testing"

	"charm.land/bubbles/v2/list"
	"github.com/benmatselby/wibble/pkg/models"
)

// mockDB is a minimal implementation of dao.DaoClient for testing.
type mockDB struct {
	articles []models.Article
	feeds    []models.Feed
}

func (m *mockDB) AddFeed(feed models.Feed) (models.Feed, error) { return feed, nil }
func (m *mockDB) GetFeeds() ([]models.Feed, error)              { return m.feeds, nil }
func (m *mockDB) AddArticle(article models.Article) error       { return nil }
func (m *mockDB) GetArticlesByFeedID(feedID int64) ([]models.Article, error) {
	return m.articles, nil
}
func (m *mockDB) GetArticleByID(articleID int64) (*models.Article, error) { return nil, nil }
func (m *mockDB) MarkArticleAsRead(articleID int64) error                        { return nil }
func (m *mockDB) MarkArticlesAsRead(feedID int64) error                          { return nil }
func (m *mockDB) DeleteFeedWithArticles(feedID int64) (int64, error)             { return 0, nil }
func (m *mockDB) Close() error                                                   { return nil }

// newTestModel constructs a minimal model suitable for handler tests.
func newTestModel(feedItems []list.Item) model {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false

	feedsList := list.New(feedItems, delegate, 80, 24)
	articlesList := list.New([]list.Item{}, delegate, 80, 24)

	return model{
		db:           &mockDB{},
		feedsList:    &feedsList,
		articlesList: &articlesList,
		keys:         DefaultKeyMap,
	}
}

func TestHandleStatusMessage_SetsStatus(t *testing.T) {
	m := newTestModel([]list.Item{})
	msg := statusMsg{text: "all good", level: statusInfo}

	result, cmd := handleStatusMessage(m, msg)

	if cmd != nil {
		t.Error("cmd should be nil")
	}
	updated, ok := result.(model)
	if !ok {
		t.Fatal("returned model is not of type model")
	}
	if updated.status == nil {
		t.Fatal("status should not be nil")
	}
	if updated.status.text != msg.text {
		t.Errorf("status.text = %q, want %q", updated.status.text, msg.text)
	}
	if updated.status.level != msg.level {
		t.Errorf("status.level = %v, want %v", updated.status.level, msg.level)
	}
}

func TestHandleStatusMessage_ErrorLevel(t *testing.T) {
	m := newTestModel([]list.Item{})
	msg := statusMsg{text: "something broke", level: statusError}

	result, _ := handleStatusMessage(m, msg)

	updated, ok := result.(model)
	if !ok {
		t.Fatal("returned model is not of type model")
	}
	if updated.status == nil {
		t.Fatal("status should not be nil")
	}
	if updated.status.level != statusError {
		t.Errorf("status.level = %v, want statusError", updated.status.level)
	}
}

func TestHandleOpenFeed_NilSelection(t *testing.T) {
	m := newTestModel([]list.Item{})

	result, cmd, handled := handleOpenFeed(m)

	if !handled {
		t.Error("handled = false, want true")
	}
	if cmd != nil {
		t.Error("cmd should be nil when no item is selected")
	}
	if result == nil {
		t.Error("returned model should not be nil")
	}
}

func TestHandleArticlesLoaded_Error(t *testing.T) {
	m := newTestModel([]list.Item{})
	msg := articlesLoadedMsg{err: fmt.Errorf("db failure")}

	result, cmd := handleArticlesLoaded(msg, m)

	if cmd != nil {
		t.Error("cmd should be nil on error")
	}
	updated, ok := result.(model)
	if !ok {
		t.Fatal("returned model is not of type model")
	}
	want := "Error: db failure"
	if updated.articlesTitle != want {
		t.Errorf("articlesTitle = %q, want %q", updated.articlesTitle, want)
	}
	if updated.focusedPane != paneFeeds {
		t.Errorf("focusedPane = %v, want paneFeeds", updated.focusedPane)
	}
}

func TestHandleArticlesLoaded_Success(t *testing.T) {
	m := newTestModel([]list.Item{})
	articles := []models.Article{
		{ID: 1, Title: "Article One"},
		{ID: 2, Title: "Article Two"},
	}
	msg := articlesLoadedMsg{articles: articles}

	result, _ := handleArticlesLoaded(msg, m)

	updated, ok := result.(model)
	if !ok {
		t.Fatal("returned model is not of type model")
	}
	if updated.focusedPane != paneArticles {
		t.Errorf("focusedPane = %v, want paneArticles", updated.focusedPane)
	}
	if got := len(updated.articlesList.Items()); got != len(articles) {
		t.Errorf("articlesList length = %d, want %d", got, len(articles))
	}
}

func TestHandleArticlesLoaded_Empty(t *testing.T) {
	m := newTestModel([]list.Item{})
	msg := articlesLoadedMsg{articles: []models.Article{}}

	result, _ := handleArticlesLoaded(msg, m)

	updated, ok := result.(model)
	if !ok {
		t.Fatal("returned model is not of type model")
	}
	if updated.focusedPane != paneArticles {
		t.Errorf("focusedPane = %v, want paneArticles", updated.focusedPane)
	}
	if got := len(updated.articlesList.Items()); got != 0 {
		t.Errorf("articlesList length = %d, want 0", got)
	}
}

func TestHandleOpenFeed_SelectsFeed(t *testing.T) {
	feed := models.Feed{
		ID:    42,
		Title: "Test Feed",
	}
	items := []list.Item{feedItem{feed: feed}}
	m := newTestModel(items)

	result, cmd, handled := handleOpenFeed(m)

	if !handled {
		t.Error("handled = false, want true")
	}
	if cmd == nil {
		t.Error("cmd should not be nil when a feed is selected")
	}

	updated, ok := result.(model)
	if !ok {
		t.Fatal("returned model is not of type model")
	}
	if updated.currentFeedID != feed.ID {
		t.Errorf("currentFeedID = %d, want %d", updated.currentFeedID, feed.ID)
	}
	if updated.articlesTitle != feed.Title {
		t.Errorf("articlesTitle = %q, want %q", updated.articlesTitle, feed.Title)
	}
}
