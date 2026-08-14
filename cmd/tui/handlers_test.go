package tui

import (
	"fmt"
	"testing"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/benmatselby/wibble/pkg/models"
)

// mockDB is a minimal implementation of dao.DaoClient for testing.
type mockDB struct {
	articles      []models.Article
	feeds         []models.Feed
	tags          []models.Tag
	tagged        map[int64][]models.Tag
	addTagFn      func(name string) (models.Tag, error)
	deleteTagFn   func(tagID int64) error
	deletedTagIDs []int64
	removedTags   []int64
}

func (m *mockDB) AddFeed(feed models.Feed) (models.Feed, error) { return feed, nil }
func (m *mockDB) GetFeeds() ([]models.Feed, error)              { return m.feeds, nil }
func (m *mockDB) AddArticle(article models.Article) error       { return nil }
func (m *mockDB) GetArticlesByFeedID(feedID int64) ([]models.Article, error) {
	return m.articles, nil
}
func (m *mockDB) GetArticleByID(articleID int64) (*models.Article, error) { return nil, nil }
func (m *mockDB) MarkArticleAsRead(articleID int64) error                 { return nil }
func (m *mockDB) MarkArticlesAsRead(feedID int64) error                   { return nil }
func (m *mockDB) DeleteFeedWithArticles(feedID int64) (int64, error)      { return 0, nil }
func (m *mockDB) Close() error                                            { return nil }

func (m *mockDB) AddTag(name string) (models.Tag, error) {
	if m.addTagFn != nil {
		return m.addTagFn(name)
	}
	return models.Tag{ID: 1, Name: name}, nil
}
func (m *mockDB) GetTags() ([]models.Tag, error) { return m.tags, nil }
func (m *mockDB) DeleteTag(tagID int64) error {
	m.deletedTagIDs = append(m.deletedTagIDs, tagID)
	if m.deleteTagFn != nil {
		return m.deleteTagFn(tagID)
	}
	return nil
}
func (m *mockDB) AddTagToArticle(articleID, tagID int64) error {
	if m.tagged == nil {
		m.tagged = map[int64][]models.Tag{}
	}
	m.tagged[articleID] = append(m.tagged[articleID], models.Tag{ID: tagID})
	return nil
}
func (m *mockDB) RemoveTagFromArticle(articleID, tagID int64) error {
	m.removedTags = append(m.removedTags, tagID)
	tags := m.tagged[articleID]
	for i, t := range tags {
		if t.ID == tagID {
			m.tagged[articleID] = append(tags[:i], tags[i+1:]...)
			break
		}
	}
	return nil
}
func (m *mockDB) GetTagsForArticle(articleID int64) ([]models.Tag, error) {
	return m.tagged[articleID], nil
}
func (m *mockDB) GetArticlesByTagID(tagID int64) ([]models.Article, error) {
	return m.articles, nil
}

// newTestModel constructs a minimal model suitable for handler tests.
func newTestModel(feedItems []list.Item) model {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false

	feedsList := list.New(feedItems, delegate, 80, 24)
	tagsList := list.New([]list.Item{}, delegate, 80, 24)
	tagPickerList := list.New([]list.Item{}, delegate, 80, 24)
	removeTagList := list.New([]list.Item{}, delegate, 80, 24)
	articlesList := list.New([]list.Item{}, delegate, 80, 24)

	return model{
		db:            &mockDB{},
		feedsList:     &feedsList,
		tagsList:      &tagsList,
		tagPickerList: &tagPickerList,
		removeTagList: &removeTagList,
		articlesList:  &articlesList,
		tagInput:      textinput.New(),
		keys:          DefaultKeyMap,
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

func TestHandleOpenTag_SelectsTag(t *testing.T) {
	tag := models.Tag{ID: 7, Name: "research"}
	m := newTestModel([]list.Item{})
	items := []list.Item{tagItem{tag: tag}}
	_ = m.tagsList.SetItems(items)

	result, cmd, handled := handleOpenTag(m)

	if !handled {
		t.Error("handled = false, want true")
	}
	if cmd == nil {
		t.Error("cmd should not be nil when a tag is selected")
	}

	updated, ok := result.(model)
	if !ok {
		t.Fatal("returned model is not of type model")
	}
	if updated.currentTagID != tag.ID {
		t.Errorf("currentTagID = %d, want %d", updated.currentTagID, tag.ID)
	}
	if !updated.articlesFromTag {
		t.Error("articlesFromTag should be true after opening a tag")
	}
}

func TestHandleOpenTag_NilSelection(t *testing.T) {
	m := newTestModel([]list.Item{})

	result, cmd, handled := handleOpenTag(m)

	if !handled {
		t.Error("handled = false, want true")
	}
	if cmd != nil {
		t.Error("cmd should be nil when no tag is selected")
	}
	if result == nil {
		t.Error("returned model should not be nil")
	}
}

func TestHandleDeleteTag_ArmsConfirmation(t *testing.T) {
	tag := models.Tag{ID: 7, Name: "research"}
	m := newTestModel([]list.Item{})
	db := m.db.(*mockDB)
	items := []list.Item{tagItem{tag: tag}}
	_ = m.tagsList.SetItems(items)

	result, cmd, handled := handleDeleteTag(m)

	if !handled {
		t.Error("handled = false, want true")
	}
	if cmd != nil {
		t.Error("cmd should be nil; deletion is deferred until confirmation")
	}
	updated, ok := result.(model)
	if !ok {
		t.Fatal("returned model is not of type model")
	}
	if updated.confirmDeleteTag == nil || updated.confirmDeleteTag.ID != tag.ID {
		t.Errorf("confirmDeleteTag = %v, want tag ID %d", updated.confirmDeleteTag, tag.ID)
	}
	if len(db.deletedTagIDs) != 0 {
		t.Errorf("deletedTagIDs = %v, want none until confirmed", db.deletedTagIDs)
	}
}

func TestHandleDeleteTag_NilSelection(t *testing.T) {
	m := newTestModel([]list.Item{})
	db := m.db.(*mockDB)

	result, cmd, handled := handleDeleteTag(m)

	if !handled {
		t.Error("handled = false, want true")
	}
	if cmd != nil {
		t.Error("cmd should be nil when no tag is selected")
	}
	if result == nil {
		t.Error("returned model should not be nil")
	}
	if len(db.deletedTagIDs) != 0 {
		t.Errorf("deletedTagIDs = %v, want none", db.deletedTagIDs)
	}
}

func TestHandleConfirmDeleteTagKeypress_YConfirmsDeletion(t *testing.T) {
	tag := models.Tag{ID: 7, Name: "research"}
	m := newTestModel([]list.Item{})
	db := m.db.(*mockDB)
	m.confirmDeleteTag = &tag

	msg := tea.KeyPressMsg(tea.Key{Code: 'y', Text: "y"})
	result, cmd := handleConfirmDeleteTagKeypress(msg, m)

	if cmd == nil {
		t.Error("cmd should not be nil when confirmed")
	}
	updated, ok := result.(model)
	if !ok {
		t.Fatal("returned model is not of type model")
	}
	if updated.confirmDeleteTag != nil {
		t.Error("confirmDeleteTag should be cleared after confirmation")
	}
	if len(db.deletedTagIDs) != 1 || db.deletedTagIDs[0] != tag.ID {
		t.Errorf("deletedTagIDs = %v, want [%d]", db.deletedTagIDs, tag.ID)
	}
}

func TestHandleConfirmDeleteTagKeypress_OtherKeyCancels(t *testing.T) {
	tag := models.Tag{ID: 7, Name: "research"}
	m := newTestModel([]list.Item{})
	db := m.db.(*mockDB)
	m.confirmDeleteTag = &tag

	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyEsc})
	result, cmd := handleConfirmDeleteTagKeypress(msg, m)

	if cmd != nil {
		t.Error("cmd should be nil when cancelled")
	}
	updated, ok := result.(model)
	if !ok {
		t.Fatal("returned model is not of type model")
	}
	if updated.confirmDeleteTag != nil {
		t.Error("confirmDeleteTag should be cleared after cancelling")
	}
	if len(db.deletedTagIDs) != 0 {
		t.Errorf("deletedTagIDs = %v, want none", db.deletedTagIDs)
	}
}

func TestHandleConfirmDeleteTagKeypress_Error(t *testing.T) {
	tag := models.Tag{ID: 9, Name: "broken"}
	m := newTestModel([]list.Item{})
	db := m.db.(*mockDB)
	db.deleteTagFn = func(tagID int64) error {
		return fmt.Errorf("db failure")
	}
	m.confirmDeleteTag = &tag

	msg := tea.KeyPressMsg(tea.Key{Code: 'y', Text: "y"})
	result, cmd := handleConfirmDeleteTagKeypress(msg, m)

	if cmd == nil {
		t.Fatal("cmd should not be nil on error")
	}
	msg2 := cmd()
	statusMessage, ok := msg2.(statusMsg)
	if !ok {
		t.Fatalf("expected statusMsg, got %T", msg2)
	}
	if statusMessage.level != statusError {
		t.Errorf("status.level = %v, want statusError", statusMessage.level)
	}
	if result == nil {
		t.Error("returned model should not be nil")
	}
}

func TestHandleToggleLeftPane_TogglesMode(t *testing.T) {
	m := newTestModel([]list.Item{})

	result, cmd, handled := handleToggleLeftPane(m)
	if !handled {
		t.Error("handled = false, want true")
	}
	if cmd == nil {
		t.Error("cmd should not be nil")
	}
	updated := result.(model)
	if updated.leftPaneMode != leftPaneTags {
		t.Errorf("leftPaneMode = %v, want leftPaneTags", updated.leftPaneMode)
	}

	result2, _, _ := handleToggleLeftPane(updated)
	updated2 := result2.(model)
	if updated2.leftPaneMode != leftPaneFeeds {
		t.Errorf("leftPaneMode = %v, want leftPaneFeeds", updated2.leftPaneMode)
	}
}

func TestHandleStartAddTag_NoSelection(t *testing.T) {
	m := newTestModel([]list.Item{})
	m.focusedPane = paneArticles

	result, cmd, handled := handleStartAddTag(m)
	if !handled {
		t.Error("handled = false, want true")
	}
	if cmd != nil {
		t.Error("cmd should be nil when no article is selected")
	}
	updated := result.(model)
	if updated.addingTag {
		t.Error("addingTag should remain false when no article is selected")
	}
}

func TestHandleStartAddTag_OpensOverlay(t *testing.T) {
	m := newTestModel([]list.Item{})
	m.focusedPane = paneArticles
	_ = m.articlesList.SetItems([]list.Item{articleItem{article: models.Article{ID: 1, Title: "A1"}}})

	result, _, handled := handleStartAddTag(m)
	if !handled {
		t.Error("handled = false, want true")
	}
	updated := result.(model)
	if !updated.addingTag {
		t.Error("addingTag should be true after starting add-tag")
	}
}

func TestHandleStartRemoveTag_FetchesArticleTags(t *testing.T) {
	m := newTestModel([]list.Item{})
	m.focusedPane = paneArticles
	_ = m.articlesList.SetItems([]list.Item{articleItem{article: models.Article{ID: 1, Title: "A1"}}})

	result, cmd, handled := handleStartRemoveTag(m)
	if !handled {
		t.Error("handled = false, want true")
	}
	if cmd == nil {
		t.Error("cmd should not be nil (should fetch article tags)")
	}

	updated := result.(model)
	if !updated.removingTag {
		t.Error("removingTag should be true after starting remove-tag")
	}
}

func TestHandleStartRemoveTag_NoSelection(t *testing.T) {
	m := newTestModel([]list.Item{})
	m.focusedPane = paneArticles

	result, cmd, handled := handleStartRemoveTag(m)
	if !handled {
		t.Error("handled = false, want true")
	}
	if cmd != nil {
		t.Error("cmd should be nil when no article is selected")
	}
	_ = result
}

func TestHandleArticleTagsLoaded_PopulatesRemoveTagList(t *testing.T) {
	m := newTestModel([]list.Item{})
	msg := articleTagsLoadedMsg{tags: []models.Tag{{ID: 10, Name: "zebra"}, {ID: 20, Name: "apple"}}}

	result, cmd := handleArticleTagsLoaded(msg, m)
	_ = cmd
	updated := result.(model)

	if len(updated.removeTagList.Items()) != 2 {
		t.Fatalf("expected 2 items in removeTagList, got %d", len(updated.removeTagList.Items()))
	}
}

func TestHandleRemoveTagKeypress_EnterRemovesSelectedTag(t *testing.T) {
	m := newTestModel([]list.Item{})
	m.focusedPane = paneArticles
	_ = m.articlesList.SetItems([]list.Item{articleItem{article: models.Article{ID: 1, Title: "A1"}}})

	db := m.db.(*mockDB)
	db.tagged = map[int64][]models.Tag{
		1: {{ID: 10, Name: "zebra"}, {ID: 20, Name: "apple"}},
	}
	_ = m.removeTagList.SetItems([]list.Item{
		tagItem{tag: models.Tag{ID: 10, Name: "zebra"}},
		tagItem{tag: models.Tag{ID: 20, Name: "apple"}},
	})
	m.removeTagList.Select(1) // highlight "apple"

	result, cmd := handleRemoveTagKeypress(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}), m)
	if cmd == nil {
		t.Error("cmd should not be nil")
	}
	_ = result

	if len(db.removedTags) != 1 || db.removedTags[0] != 20 {
		t.Errorf("expected tag ID 20 (apple) to be removed, got %v", db.removedTags)
	}
	remaining := db.tagged[1]
	if len(remaining) != 1 || remaining[0].Name != "zebra" {
		t.Errorf("expected zebra to remain, got %v", remaining)
	}
}

func TestHandleRemoveTagKeypress_NoSelection(t *testing.T) {
	m := newTestModel([]list.Item{})
	m.focusedPane = paneArticles
	_ = m.articlesList.SetItems([]list.Item{articleItem{article: models.Article{ID: 1, Title: "A1"}}})

	result, cmd := handleRemoveTagKeypress(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}), m)
	if cmd != nil {
		t.Error("cmd should be nil when no tag is selected")
	}
	_ = result

	db := m.db.(*mockDB)
	if len(db.removedTags) != 0 {
		t.Errorf("expected no tags removed, got %v", db.removedTags)
	}
}

func TestHandleRemoveTagKeypress_EscClosesOverlay(t *testing.T) {
	m := newTestModel([]list.Item{})
	m.removingTag = true

	result, cmd := handleRemoveTagKeypress(tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}), m)
	if cmd != nil {
		t.Error("cmd should be nil on esc")
	}

	updated := result.(model)
	if updated.removingTag {
		t.Error("removingTag should be false after esc")
	}
}

func TestHandleTagInputKeypress_EnterAddsTag(t *testing.T) {
	m := newTestModel([]list.Item{})
	m.focusedPane = paneArticles
	_ = m.articlesList.SetItems([]list.Item{articleItem{article: models.Article{ID: 1, Title: "A1"}}})
	m.addingTag = true
	m.tagInput.SetValue("research")

	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter})

	result, cmd, handled := handleTagInputKeypress(msg, m)
	if !handled {
		t.Error("handled = false, want true")
	}
	if cmd == nil {
		t.Error("cmd should not be nil")
	}
	updated := result.(model)
	if updated.addingTag {
		t.Error("addingTag should be false after submitting")
	}
}
