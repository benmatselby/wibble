package tui

import (
	"fmt"

	"github.com/benmatselby/wibble/pkg/models"
)

// articleItem adapts models.Article to the list.Item interface.
type articleItem struct {
	article models.Article
}

func (e articleItem) Title() string {
	title := e.article.Title

	published := ""
	if e.article.Published != nil {
		published = e.article.Published.Format("02 Jan 2006 15:04:05")
	}
	return fmt.Sprintf("%-22s %s", published, title)
}

func (e articleItem) Description() string { return "" }
func (e articleItem) FilterValue() string { return e.article.Title }
func (e articleItem) IsRead() bool        { return e.article.IsRead }
