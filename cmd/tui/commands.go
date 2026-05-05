package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/benmatselby/wibble/pkg/dao"
)

func (m model) Init() tea.Cmd {
	return nil
}

// fetchArticles is a Bubbletea command that loads articles for a given feed.
func fetchArticles(db dao.DaoClient, feedID int64) tea.Cmd {
	return func() tea.Msg {
		articles, err := db.GetArticlesByFeedID(feedID)
		return articlesLoadedMsg{articles: articles, err: err}
	}
}

// fetchFeeds is a Bubbletea command that reloads all feeds from the database.
func fetchFeeds(db dao.DaoClient) tea.Cmd {
	return func() tea.Msg {
		feeds, err := db.GetFeeds()
		return feedsLoadedMsg{feeds: feeds, err: err}
	}
}
