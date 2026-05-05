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
	if len(items) == 0 {
		m.articlesTitle = "Articles (none)"
	} else {
		m.articlesTitle = fmt.Sprintf("Articles (%d)", len(items))
	}
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
	m.articlesList.Select(0)
	m.articlesTitle = "Loading..."
	_ = m.articlesList.SetItems([]list.Item{})
	return m, fetchArticles(m.db, fi.feed.ID), true
}

// handleViewArticle toggles the article view overlay and sets the current
// article ID.
func handleViewArticle(m model) (tea.Model, tea.Cmd, bool) {
	m.focusedPane = paneArticle

	selectedItem := m.articlesList.SelectedItem()
	if selectedItem == nil {
		return m, nil, true
	}
	m.currentArticleID = selectedItem.(articleItem).article.ID
	if err := m.db.MarkArticleAsRead(m.currentArticleID); err != nil {
		return m, func() tea.Msg {
			return statusMsg{text: fmt.Sprintf("Error marking article as read: %v", err), level: statusError}
		}, true
	}
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
		fetchArticles(m.db, m.currentFeedID),
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
			fetchArticles(m.db, m.currentFeedID),
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
		fetchFeeds(m.db),
	), true
}

// handkleKeypress processes keypresses.
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
		// Don't intercept filter keys
		if m.feedsList.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.keys.OpenFeed):
			return handleOpenFeed(m)
		case key.Matches(msg, m.keys.MarkAllAsRead):
			return handleMarkAllAsRead(m)
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
		}

	case paneArticle:
		switch {
		case key.Matches(msg, m.keys.Back):
			m.focusedPane = paneArticles
			return m, tea.Batch(
				fetchFeeds(m.db),
				fetchArticles(m.db, m.currentFeedID),
			), true
		case key.Matches(msg, m.keys.OpenArticle):
			return handleOpenArticle(m)
		}
	}
	return nil, nil, false
}
