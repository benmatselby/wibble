package tui

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/benmatselby/wibble/pkg/utils"
)

func handleFeedsLoaded(msg feedsLoadedMsg, m model) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		return m, func() tea.Msg {
			return statusMsg{text: fmt.Sprintf("Failed to load feeds: %v", msg.err), level: statusError}
		}
	}
	items := make([]list.Item, len(msg.feeds))
	for i, f := range msg.feeds {
		items[i] = feedItem{feed: f}
	}
	cmd := m.feedsList.SetItems(items)
	return m, cmd
}

func handleTagsLoaded(msg tagsLoadedMsg, m model) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		return m, func() tea.Msg {
			return statusMsg{text: fmt.Sprintf("Failed to load tags: %v", msg.err), level: statusError}
		}
	}
	items := make([]list.Item, len(msg.tags))
	for i, t := range msg.tags {
		items[i] = tagItem{tag: t}
	}
	cmd := m.tagsList.SetItems(items)
	return m, cmd
}

func handleArticlesLoaded(msg articlesLoadedMsg, m model) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		// Show the error in the articles list title temporarily
		m.articlesTitle = fmt.Sprintf("Error: %v", msg.err)
		return m, nil
	}
	items := make([]list.Item, len(msg.articles))
	for i, e := range msg.articles {
		items[i] = articleItem{article: e}
	}
	cmd := m.articlesList.SetItems(items)
	m.focusedPane = paneArticles
	return m, cmd
}

func handleStatusMessage(m model, msg statusMsg) (tea.Model, tea.Cmd) {
	m.status = &msg
	return m, nil
}

func handleOpenFeed(m model) (tea.Model, tea.Cmd, bool) {
	sel := m.feedsList.SelectedItem()
	if sel == nil {
		return m, nil, true
	}
	fi := sel.(feedItem)
	m.currentFeedID = fi.feed.ID
	m.articlesFromTag = false
	m.articlesList.Select(0)
	m.articlesTitle = fi.feed.Title
	_ = m.articlesList.SetItems([]list.Item{})
	return m, fetchArticles(m.db, fi.feed.ID), true
}

// handleOpenTag loads all articles associated with the selected tag into the
// articles pane, across all feeds.
func handleOpenTag(m model) (tea.Model, tea.Cmd, bool) {
	sel := m.tagsList.SelectedItem()
	if sel == nil {
		return m, nil, true
	}
	ti := sel.(tagItem)
	m.currentTagID = ti.tag.ID
	m.articlesFromTag = true
	m.articlesList.Select(0)
	m.articlesTitle = fmt.Sprintf("#%s", ti.tag.Name)
	_ = m.articlesList.SetItems([]list.Item{})
	return m, fetchArticlesByTag(m.db, ti.tag.ID), true
}

// refreshArticlesCmd reloads the currently displayed article list, whether
// it originated from a feed or a tag.
func refreshArticlesCmd(m model) tea.Cmd {
	if m.articlesFromTag {
		return fetchArticlesByTag(m.db, m.currentTagID)
	}
	return fetchArticles(m.db, m.currentFeedID)
}

// handleToggleLeftPane swaps the left panel between the Feeds list and the
// Tags list.
func handleToggleLeftPane(m model) (tea.Model, tea.Cmd, bool) {
	if m.leftPaneMode == leftPaneFeeds {
		m.leftPaneMode = leftPaneTags
		return m, fetchTags(m.db), true
	}
	m.leftPaneMode = leftPaneFeeds
	return m, fetchFeeds(m.db), true
}

// handleViewArticle toggles the article view overlay and sets the current
// article ID.
func handleViewArticle(m model) (tea.Model, tea.Cmd, bool) {
	selectedItem := m.articlesList.SelectedItem()
	if selectedItem == nil {
		return m, nil, true
	}
	m.focusedPane = paneArticle
	m.currentArticleID = selectedItem.(articleItem).article.ID
	if err := m.db.MarkArticleAsRead(m.currentArticleID); err != nil {
		return m, func() tea.Msg {
			return statusMsg{text: fmt.Sprintf("Error marking article as read: %v", err), level: statusError}
		}, true
	}

	content := m.renderArticleModal()
	m.articleViewport.SetContent(content)
	m.articleViewport.GotoTop()

	return m, nil, true
}

// handleOpenArticle marks the article as read and opens the link in the browser.
func handleOpenArticle(m model) (tea.Model, tea.Cmd, bool) {
	sel := m.articlesList.SelectedItem()
	if sel == nil {
		return m, nil, true
	}
	ai := sel.(articleItem)
	if err := m.db.MarkArticleAsRead(ai.article.ID); err != nil {
		return m, func() tea.Msg {
			return statusMsg{text: fmt.Sprintf("Error marking article as read: %v", err), level: statusError}
		}, true
	}
	utils.OpenURL(ai.article.Link)
	return m, tea.Batch(
		refreshArticlesCmd(m),
		fetchFeeds(m.db),
	), true
}

func handleMarkItemAsRead(m model) (tea.Model, tea.Cmd, bool) {
	sel := m.articlesList.SelectedItem()
	if sel == nil {
		return m, nil, true
	}
	ai := sel.(articleItem)
	if ai.article.Link != "" {
		if err := m.db.MarkArticleAsRead(ai.article.ID); err != nil {
			return m, func() tea.Msg {
				return statusMsg{text: fmt.Sprintf("Error marking article as read: %v", err), level: statusError}
			}, true
		}
		return m, tea.Batch(
			refreshArticlesCmd(m),
			fetchFeeds(m.db),
		), true
	}
	return m, nil, true
}

func handleMarkAllAsRead(m model) (tea.Model, tea.Cmd, bool) {
	sel := m.feedsList.SelectedItem()
	if sel == nil {
		return m, nil, true
	}
	fi := sel.(feedItem)
	if err := m.db.MarkArticlesAsRead(fi.feed.ID); err != nil {
		return m, func() tea.Msg {
			return statusMsg{text: fmt.Sprintf("Error marking feed as read: %v", err), level: statusError}
		}, true
	}
	return m, tea.Batch(
		refreshArticlesCmd(m),
		fetchFeeds(m.db),
	), true
}

// currentArticleIDForTagging returns the article ID that tag actions should
// apply to, whichever pane (Articles list or Article viewport) is focused.
func currentArticleIDForTagging(m model) (int64, bool) {
	if m.focusedPane == paneArticle {
		return m.currentArticleID, true
	}
	sel := m.articlesList.SelectedItem()
	if sel == nil {
		return 0, false
	}
	return sel.(articleItem).article.ID, true
}

// handleStartAddTag opens the tag-name input overlay for the current article.
func handleStartAddTag(m model) (tea.Model, tea.Cmd, bool) {
	if _, ok := currentArticleIDForTagging(m); !ok {
		return m, nil, true
	}
	m.addingTag = true
	m.tagInput.SetValue("")
	cmd := m.tagInput.Focus()
	return m, cmd, true
}

// handleRemoveTagFromArticle removes the most recently added tag from the
// current article, one at a time per keypress.
func handleRemoveTagFromArticle(m model) (tea.Model, tea.Cmd, bool) {
	articleID, ok := currentArticleIDForTagging(m)
	if !ok {
		return m, nil, true
	}
	tags, err := m.db.GetTagsForArticle(articleID)
	if err != nil {
		return m, func() tea.Msg {
			return statusMsg{text: fmt.Sprintf("Error loading tags: %v", err), level: statusError}
		}, true
	}
	if len(tags) == 0 {
		return m, func() tea.Msg {
			return statusMsg{text: "Article has no tags", level: statusInfo}
		}, true
	}
	last := tags[len(tags)-1]
	if err := m.db.RemoveTagFromArticle(articleID, last.ID); err != nil {
		return m, func() tea.Msg {
			return statusMsg{text: fmt.Sprintf("Error removing tag: %v", err), level: statusError}
		}, true
	}
	return m, tea.Batch(
		func() tea.Msg {
			return statusMsg{text: fmt.Sprintf("Removed tag %q", last.Name), level: statusInfo}
		},
		fetchTags(m.db),
	), true
}

// handleTagInputKeypress processes keypresses while the tag-name input
// overlay is active (enter submits, esc cancels).
func handleTagInputKeypress(msg tea.KeyPressMsg, m model) (tea.Model, tea.Cmd, bool) {
	switch msg.String() {
	case "esc":
		m.addingTag = false
		m.tagInput.Blur()
		return m, nil, true
	case "enter":
		m.addingTag = false
		m.tagInput.Blur()
		name := m.tagInput.Value()
		if name == "" {
			return m, nil, true
		}
		articleID, ok := currentArticleIDForTagging(m)
		if !ok {
			return m, nil, true
		}
		tag, err := m.db.AddTag(name)
		if err != nil {
			return m, func() tea.Msg {
				return statusMsg{text: fmt.Sprintf("Error adding tag: %v", err), level: statusError}
			}, true
		}
		if err := m.db.AddTagToArticle(articleID, tag.ID); err != nil {
			return m, func() tea.Msg {
				return statusMsg{text: fmt.Sprintf("Error tagging article: %v", err), level: statusError}
			}, true
		}
		return m, tea.Batch(
			func() tea.Msg {
				return statusMsg{text: fmt.Sprintf("Tagged article %q", tag.Name), level: statusInfo}
			},
			fetchTags(m.db),
		), true
	}
	return m, nil, false
}

// handleKeypress processes keypresses.
// Returns the updated model, any command to run, and a boolean indicating
// whether the keypress was handled (true) or should be processed by the
// focused pane (false).
func handleKeypress(msg tea.KeyPressMsg, m model) (tea.Model, tea.Cmd, bool) {
	// Global quit (ctrl+c fires regardless of pane or filter state)
	if key.Matches(msg, m.keys.Quit) && msg.String() == "ctrl+c" {
		return m, func() tea.Msg { return tea.Quit() }, true
	}

	switch m.focusedPane {
	case paneFeeds:
		if m.leftPaneMode == leftPaneTags {
			if m.tagsList.FilterState() == list.Filtering {
				break
			}
			switch {
			case key.Matches(msg, m.keys.OpenTag):
				return handleOpenTag(m)
			case key.Matches(msg, m.keys.ToggleTagsPane):
				return handleToggleLeftPane(m)
			}
			break
		}

		// Don't intercept filter keys
		if m.feedsList.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.keys.OpenFeed):
			return handleOpenFeed(m)
		case key.Matches(msg, m.keys.MarkAllAsRead):
			return handleMarkAllAsRead(m)
		case key.Matches(msg, m.keys.ToggleTagsPane):
			return handleToggleLeftPane(m)
		}

	case paneArticles:
		if m.articlesList.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.keys.Back):
			m.focusedPane = paneFeeds
			return m, nil, true
		case key.Matches(msg, m.keys.ViewArticle):
			return handleViewArticle(m)
		case key.Matches(msg, m.keys.OpenArticle):
			return handleOpenArticle(m)
		case key.Matches(msg, m.keys.MarkAsRead):
			return handleMarkItemAsRead(m)
		case key.Matches(msg, m.keys.MarkAllAsRead):
			return handleMarkAllAsRead(m)
		case key.Matches(msg, m.keys.AddTag):
			return handleStartAddTag(m)
		case key.Matches(msg, m.keys.RemoveTag):
			return handleRemoveTagFromArticle(m)
		}

	case paneArticle:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit, true
		case key.Matches(msg, m.keys.Back):
			m.focusedPane = paneArticles
			return m, tea.Batch(
				fetchFeeds(m.db),
				refreshArticlesCmd(m),
			), true
		case key.Matches(msg, m.keys.OpenArticle):
			return handleOpenArticle(m)
		case key.Matches(msg, m.keys.AddTag):
			return handleStartAddTag(m)
		case key.Matches(msg, m.keys.RemoveTag):
			return handleRemoveTagFromArticle(m)
		}
	}
	return nil, nil, false
}
