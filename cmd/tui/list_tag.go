package tui

import (
	"fmt"

	"github.com/benmatselby/wibble/pkg/models"
)

// tagItem adapts models.Tag to the list.Item interface.
type tagItem struct {
	tag models.Tag
}

func (t tagItem) Title() string {
	return fmt.Sprintf("%6s %s", fmt.Sprintf("[%d]", t.tag.ArticleCount), t.tag.Name)
}
func (t tagItem) Description() string { return "" }
func (t tagItem) FilterValue() string { return t.tag.Name }
