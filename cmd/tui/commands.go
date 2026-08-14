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

// fetchTags is a Bubbletea command that reloads all tags from the database.
func fetchTags(db dao.DaoClient) tea.Cmd {
	return func() tea.Msg {
		tags, err := db.GetTags()
		return tagsLoadedMsg{tags: tags, err: err}
	}
}

// fetchArticleTags is a Bubbletea command that loads the tags currently
// attached to a given article, used by the remove-tag picker.
func fetchArticleTags(db dao.DaoClient, articleID int64) tea.Cmd {
	return func() tea.Msg {
		tags, err := db.GetTagsForArticle(articleID)
		return articleTagsLoadedMsg{tags: tags, err: err}
	}
}

// fetchArticlesByTag is a Bubbletea command that loads articles for a given tag.
func fetchArticlesByTag(db dao.DaoClient, tagID int64) tea.Cmd {
	return func() tea.Msg {
		articles, err := db.GetArticlesByTagID(tagID)
		return articlesLoadedMsg{articles: articles, err: err}
	}
}
