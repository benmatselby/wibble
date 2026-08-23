package tui

import (
	"fmt"
	"testing"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/benmatselby/wibble/pkg/dao"
	"github.com/benmatselby/wibble/pkg/models"
	"go.uber.org/mock/gomock"
)

// newTestModelWithDB constructs a minimal model suitable for handler tests,
// backed by a gomock-generated dao.MockDaoClient.
func newTestModelWithDB(t *testing.T, feedItems []list.Item, optionalArticleItems ...[]list.Item) (model, *dao.MockDaoClient) {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	articleItems := []list.Item{}
	if len(optionalArticleItems) > 0 {
		articleItems = optionalArticleItems[0]
	}

	feedsList := list.New(feedItems, delegate, 80, 24)
	tagsList := list.New([]list.Item{}, delegate, 80, 24)
	tagPickerList := list.New([]list.Item{}, delegate, 80, 24)
	removeTagList := list.New([]list.Item{}, delegate, 80, 24)
	articlesList := list.New(articleItems, delegate, 80, 24)

	ctrl := gomock.NewController(t)

	db := dao.NewMockDaoClient(ctrl)

	return model{
		db:            db,
		feedsList:     &feedsList,
		tagsList:      &tagsList,
		tagPickerList: &tagPickerList,
		removeTagList: &removeTagList,
		articlesList:  &articlesList,
		tagInput:      textinput.New(),
		keys:          DefaultKeyMap,
	}, db
}

func TestHandleStatusMessage_SetsStatus(t *testing.T) {
	m, _ := newTestModelWithDB(t, []list.Item{})
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
	m, _ := newTestModelWithDB(t, []list.Item{})
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
	m, _ := newTestModelWithDB(t, []list.Item{})

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
	m, _ := newTestModelWithDB(t, []list.Item{})
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
	m, _ := newTestModelWithDB(t, []list.Item{})
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
	m, _ := newTestModelWithDB(t, []list.Item{})
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
	m, _ := newTestModelWithDB(t, items)

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
	m, _ := newTestModelWithDB(t, []list.Item{})
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
	m, _ := newTestModelWithDB(t, []list.Item{})

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
	m, _ := newTestModelWithDB(t, []list.Item{})
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
	// No mock expectations are set up, so a strict-mode call to DeleteTag
	// here (before confirmation) would fail the test on its own.
}

func TestHandleDeleteTag_NilSelection(t *testing.T) {
	m, _ := newTestModelWithDB(t, []list.Item{})

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
}

func TestHandleConfirmDeleteTagKeypress_YConfirmsDeletion(t *testing.T) {
	tag := models.Tag{ID: 7, Name: "research"}
	m, db := newTestModelWithDB(t, []list.Item{})
	db.EXPECT().DeleteTag(tag.ID).Return(nil).Times(1)
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
}

func TestHandleConfirmDeleteTagKeypress_OtherKeyCancels(t *testing.T) {
	tag := models.Tag{ID: 7, Name: "research"}
	m, _ := newTestModelWithDB(t, []list.Item{})
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
	// No mock expectations are set up, so a strict-mode call to DeleteTag
	// would fail the test on its own.
}

func TestHandleConfirmDeleteTagKeypress_Error(t *testing.T) {
	tag := models.Tag{ID: 9, Name: "broken"}
	m, db := newTestModelWithDB(t, []list.Item{})
	db.EXPECT().DeleteTag(tag.ID).Return(fmt.Errorf("db failure")).Times(1)
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
	m, _ := newTestModelWithDB(t, []list.Item{})

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
	m, _ := newTestModelWithDB(t, []list.Item{})
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
	m, _ := newTestModelWithDB(t, []list.Item{})
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
	m, _ := newTestModelWithDB(t, []list.Item{})
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
	m, _ := newTestModelWithDB(t, []list.Item{})
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

func TestHandleTagsLoaded_ShowsArticleCountBadge(t *testing.T) {
	m, _ := newTestModelWithDB(t, []list.Item{})
	msg := tagsLoadedMsg{tags: []models.Tag{{ID: 1, Name: "research", ArticleCount: 3}}}

	result, _ := handleTagsLoaded(msg, m)
	updated := result.(model)

	items := updated.tagsList.Items()
	if len(items) != 1 {
		t.Fatalf("expected 1 item in tagsList, got %d", len(items))
	}
	if got, want := items[0].(tagItem).Title(), "   [3] research"; got != want {
		t.Errorf("Title() = %q, want %q (count badge should show in Tags panel)", got, want)
	}
}

func TestHandleArticleTagsLoaded_PopulatesRemoveTagList(t *testing.T) {
	m, _ := newTestModelWithDB(t, []list.Item{})
	// ArticleCount is deliberately nonzero here to prove the remove-tag
	// picker doesn't render a "[N]" count badge, even when the field is
	// populated: GetTagsForArticle never populates it in practice, and
	// rendering it in this list would be misleading regardless (see
	// tagItem's showCount field).
	msg := articleTagsLoadedMsg{tags: []models.Tag{
		{ID: 10, Name: "zebra", ArticleCount: 5},
		{ID: 20, Name: "apple", ArticleCount: 2},
	}}

	result, cmd := handleArticleTagsLoaded(msg, m)
	_ = cmd
	updated := result.(model)

	items := updated.removeTagList.Items()
	if len(items) != 2 {
		t.Fatalf("expected 2 items in removeTagList, got %d", len(items))
	}
	if got, want := items[0].(tagItem).Title(), "zebra"; got != want {
		t.Errorf("Title() = %q, want %q (no count badge in remove-tag picker)", got, want)
	}
	if got, want := items[1].(tagItem).Title(), "apple"; got != want {
		t.Errorf("Title() = %q, want %q (no count badge in remove-tag picker)", got, want)
	}
}

func TestHandleRemoveTagKeypress_EnterRemovesSelectedTag(t *testing.T) {
	m, db := newTestModelWithDB(t, []list.Item{})
	m.focusedPane = paneArticles
	_ = m.articlesList.SetItems([]list.Item{articleItem{article: models.Article{ID: 1, Title: "A1"}}})

	db.EXPECT().RemoveTagFromArticle(int64(1), int64(20)).Return(nil).Times(1)

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
}

func TestHandleRemoveTagKeypress_NoSelection(t *testing.T) {
	m, _ := newTestModelWithDB(t, []list.Item{})
	m.focusedPane = paneArticles
	_ = m.articlesList.SetItems([]list.Item{articleItem{article: models.Article{ID: 1, Title: "A1"}}})

	result, cmd := handleRemoveTagKeypress(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}), m)
	if cmd != nil {
		t.Error("cmd should be nil when no tag is selected")
	}
	_ = result
	// No mock expectations are set up, so a strict-mode call to
	// RemoveTagFromArticle would fail the test on its own.
}

func TestHandleRemoveTagKeypress_EscClosesOverlay(t *testing.T) {
	m, _ := newTestModelWithDB(t, []list.Item{})
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
	m, db := newTestModelWithDB(t, []list.Item{})
	m.focusedPane = paneArticles
	_ = m.articlesList.SetItems([]list.Item{articleItem{article: models.Article{ID: 1, Title: "A1"}}})
	m.addingTag = true
	m.tagInput.SetValue("research")

	db.EXPECT().AddTag("research").Return(models.Tag{ID: 1, Name: "research"}, nil).Times(1)
	db.EXPECT().AddTagToArticle(int64(1), int64(1)).Return(nil).Times(1)

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

func TestHandleMarkItemAsRead_NoOpWhenArticleLinkIsEmpty(t *testing.T) {
	article := models.Article{ID: 1, Title: "Test Article"}
	m, db := newTestModelWithDB(t, []list.Item{}, []list.Item{articleItem{article: article}})

	db.EXPECT().MarkArticleAsRead(article.ID).MaxTimes(0)

	_, command, wasHandled := handleMarkItemAsRead(m)

	if command != nil {
		t.Fatal("expected command to be nil, got command")
	}

	if wasHandled != true {
		t.Fatal("expected wasHandled to be true, got false")
	}
}

func TestHandleMarkItemAsRead_ReturnsErrorMessageWhenArticleCannotBeMarkedAsRead(t *testing.T) {
	article := models.Article{ID: 1, Title: "Test Article", Link: "http://example.com"}
	m, db := newTestModelWithDB(t, []list.Item{}, []list.Item{articleItem{article: article}})

	db.EXPECT().MarkArticleAsRead(article.ID).Return(fmt.Errorf("db failure")).Times(1)

	_, command, wasHandled := handleMarkItemAsRead(m)

	if command == nil {
		t.Fatal("expected command to not be nil, got nil")
	}

	message := command()
	if _, ok := message.(statusMsg); !ok {
		t.Fatal("expected message to be of type statusMsg")
	}

	if wasHandled != true {
		t.Fatal("expected wasHandled to be true, got false")
	}
}

func TestHandleMarkItemAsRead_PicksNextArticleItemToRead(t *testing.T) {
	article := models.Article{ID: 1, Title: "Test Article", Link: "http://example.com"}
	nextArticle := models.Article{ID: 2, Title: "Next Article", Link: "http://example.com/next"}
	m, db := newTestModelWithDB(t, []list.Item{}, []list.Item{articleItem{article: article}, articleItem{article: nextArticle}})

	db.EXPECT().MarkArticleAsRead(article.ID).MaxTimes(1).Return(nil)

	if m.articlesList.GlobalIndex() != 0 {
		t.Fatalf("expected initial global index to be 0, got %d", m.articlesList.GlobalIndex())
	}

	_, _, _ = handleMarkItemAsRead(m)

	if m.articlesList.GlobalIndex() != 1 {
		t.Fatalf("expected global index to be 1 after marking as read, got %d", m.articlesList.GlobalIndex())
	}
}

func TestHandleMarkItemAsRead_StaysOnLastItemIfItsTheLastItem(t *testing.T) {
	article := models.Article{ID: 1, Title: "Test Article", Link: "http://example.com"}
	m, db := newTestModelWithDB(t, []list.Item{}, []list.Item{articleItem{article: article}})

	db.EXPECT().MarkArticleAsRead(article.ID).MaxTimes(1).Return(nil)

	if m.articlesList.GlobalIndex() != 0 {
		t.Fatalf("expected initial global index to be 0, got %d", m.articlesList.GlobalIndex())
	}

	_, _, _ = handleMarkItemAsRead(m)

	if m.articlesList.GlobalIndex() != 0 {
		t.Fatalf("expected global index to be 0 after marking as read, got %d", m.articlesList.GlobalIndex())
	}
}
