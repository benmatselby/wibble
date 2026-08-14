package tui

import "github.com/benmatselby/wibble/pkg/models"

type articlesLoadedMsg struct {
	articles []models.Article
	err      error
}

type feedsLoadedMsg struct {
	feeds []models.Feed
	err   error
}

type tagsLoadedMsg struct {
	tags []models.Tag
	err  error
}

// articleTagsLoadedMsg carries the tags currently attached to a specific
// article, used to populate the remove-tag picker.
type articleTagsLoadedMsg struct {
	tags []models.Tag
	err  error
}

type statusLevel int

const (
	statusInfo  statusLevel = iota
	statusError statusLevel = iota
)

// statusMsg sets the text shown in the status bar.
type statusMsg struct {
	text  string
	level statusLevel
}
