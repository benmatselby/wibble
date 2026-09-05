package tui

import (
	"fmt"
	"strings"

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
		items[i] = tagItem{tag: t, showCount: true}
	}
	cmd := m.tagsList.SetItems(items)

	// tagPickerList is rendered outside the normal focused-pane routing (it's
	// a modal), so the async re-filter command that SetItems would otherwise
	// return never gets dispatched back to it. Recompute its filtered view
	// synchronously instead, using whatever's currently typed in tagInput.
	_ = m.tagPickerList.SetItems(items)
	m.tagPickerList.SetFilterText(strings.TrimSpace(m.tagInput.Value()))

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

// handleDeleteTag opens a confirmation prompt for deleting the currently
// selected tag. The actual deletion is performed by
// handleConfirmDeleteTagKeypress once the user confirms with "y"; any other
// key cancels.
func handleDeleteTag(m model) (tea.Model, tea.Cmd, bool) {
	sel := m.tagsList.SelectedItem()
	if sel == nil {
		return m, nil, true
	}
	ti := sel.(tagItem)
	tag := ti.tag
	m.confirmDeleteTag = &tag
	return m, nil, true
}

// handleConfirmDeleteTagKeypress processes the keypress following a
// handleDeleteTag prompt. Pressing "y" deletes the tag (along with all of
// its article associations) and refreshes the tags list; any other key
// cancels without deleting anything.
func handleConfirmDeleteTagKeypress(msg tea.KeyPressMsg, m model) (tea.Model, tea.Cmd) {
	tag := m.confirmDeleteTag
	m.confirmDeleteTag = nil

	if msg.String() != "y" {
		return m, nil
	}

	if err := m.db.DeleteTag(tag.ID); err != nil {
		return m, func() tea.Msg {
			return statusMsg{text: fmt.Sprintf("Error deleting tag: %v", err), level: statusError}
		}
	}
	return m, tea.Batch(
		func() tea.Msg {
			return statusMsg{text: fmt.Sprintf("Deleted tag %q", tag.Name), level: statusInfo}
		},
		fetchTags(m.db),
	)
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

// handleMarkItemAsRead marks the currently selected article as read and
// moves the list selection to the next article, so the user can skip
// over their articles quicker whilst marking as read.
func handleMarkItemAsRead(m model) (tea.Model, tea.Cmd, bool) {
	selectedArticle := m.articlesList.SelectedItem()
	if selectedArticle == nil {
		return m, nil, true
	}

	article := selectedArticle.(articleItem)
	if article.article.Link != "" {
		if err := m.db.MarkArticleAsRead(article.article.ID); err != nil {
			return m, func() tea.Msg {
				return statusMsg{text: fmt.Sprintf("Error marking article as read: %v", err), level: statusError}
			}, true
		}

		// Once we have read the article we want to move to the next items
		// so the user can `r` their way through the list, rather than `r + j`.
		if m.articlesList.GlobalIndex()+1 < len(m.articlesList.Items()) {
			m.articlesList.Select(m.articlesList.GlobalIndex() + 1)
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
	m.tagPickerList.SetFilterText("")
	m.tagPickerList.Select(0)
	cmd := tea.Batch(m.tagInput.Focus(), fetchTags(m.db))
	return m, cmd, true
}

// handleStartRemoveTag opens the remove-tag picker overlay for the current
// article, populated with the tags currently attached to it.
func handleStartRemoveTag(m model) (tea.Model, tea.Cmd, bool) {
	articleID, ok := currentArticleIDForTagging(m)
	if !ok {
		return m, nil, true
	}
	m.removingTag = true
	return m, fetchArticleTags(m.db, articleID), true
}

// handleArticleTagsLoaded populates the remove-tag picker with the tags
// currently attached to the article being tagged.
func handleArticleTagsLoaded(msg articleTagsLoadedMsg, m model) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		return m, func() tea.Msg {
			return statusMsg{text: fmt.Sprintf("Failed to load article tags: %v", msg.err), level: statusError}
		}
	}
	items := make([]list.Item, len(msg.tags))
	for i, t := range msg.tags {
		items[i] = tagItem{tag: t}
	}
	cmd := m.removeTagList.SetItems(items)
	m.removeTagList.Select(0)
	return m, cmd
}

// handleRemoveTagKeypress processes keypresses while the remove-tag picker
// overlay is active. Up/down navigate the list of the article's current
// tags, enter removes the highlighted tag (and refreshes the picker so
// further tags can be removed), and esc closes the overlay. Every key is
// handled while the overlay is open since there's no text input to fall
// through to.
func handleRemoveTagKeypress(msg tea.KeyPressMsg, m model) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.removingTag = false
		return m, nil
	case "up":
		m.removeTagList.CursorUp()
		return m, nil
	case "down":
		m.removeTagList.CursorDown()
		return m, nil
	case "enter":
		articleID, ok := currentArticleIDForTagging(m)
		if !ok {
			m.removingTag = false
			return m, nil
		}
		sel := m.removeTagList.SelectedItem()
		if sel == nil {
			m.removingTag = false
			return m, nil
		}
		tag := sel.(tagItem).tag
		if err := m.db.RemoveTagFromArticle(articleID, tag.ID); err != nil {
			return m, func() tea.Msg {
				return statusMsg{text: fmt.Sprintf("Error removing tag: %v", err), level: statusError}
			}
		}
		return m, tea.Batch(
			func() tea.Msg {
				return statusMsg{text: fmt.Sprintf("Removed tag %q", tag.Name), level: statusInfo}
			},
			fetchArticleTags(m.db, articleID),
			fetchTags(m.db),
		)
	}
	// Swallow any other key while the picker is open; there's no text input
	// to fall through to here.
	return m, nil
}

// handleTagInputKeypress processes keypresses while the tag-name input
// overlay is active. Up/down navigate the existing-tags picker list, enter
// submits (either the typed name, or the highlighted existing tag if the
// input is empty), and esc cancels.
func handleTagInputKeypress(msg tea.KeyPressMsg, m model) (tea.Model, tea.Cmd, bool) {
	switch msg.String() {
	case "esc":
		m.addingTag = false
		m.tagInput.Blur()
		return m, nil, true
	case "up":
		m.tagPickerList.CursorUp()
		return m, nil, true
	case "down":
		m.tagPickerList.CursorDown()
		return m, nil, true
	case "enter":
		m.addingTag = false
		m.tagInput.Blur()

		articleID, ok := currentArticleIDForTagging(m)
		if !ok {
			return m, nil, true
		}

		name := strings.TrimSpace(m.tagInput.Value())
		if name == "" {
			// No name typed: fall back to whatever tag is highlighted in
			// the picker list, if any.
			sel := m.tagPickerList.SelectedItem()
			if sel == nil {
				return m, nil, true
			}
			tag := sel.(tagItem).tag
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
			case key.Matches(msg, m.keys.DeleteTag):
				return handleDeleteTag(m)
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
			return handleStartRemoveTag(m)
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
			return handleStartRemoveTag(m)
		case key.Matches(msg, m.keys.NextArticle):
			m.articlesList.CursorDown()
			return handleViewArticle(m)
		case key.Matches(msg, m.keys.PreviousArticle):
			m.articlesList.CursorUp()
			return handleViewArticle(m)
		}
	}
	return nil, nil, false
}
