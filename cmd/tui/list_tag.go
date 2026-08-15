package tui

import (
	"fmt"

	"github.com/benmatselby/wibble/pkg/models"
)

// tagItem adapts models.Tag to the list.Item interface. showCount controls
// whether the "[N]" article-count badge is rendered in the title: it's
// meaningful in the Tags panel and the add-tag picker (both backed by
// GetTags, which populates ArticleCount), but not in the remove-tag picker
// (backed by GetTagsForArticle, which doesn't populate it, and where the
// count of an article's own tags isn't relevant anyway).
type tagItem struct {
	tag       models.Tag
	showCount bool
}

func (t tagItem) Title() string {
	if !t.showCount {
		return t.tag.Name
	}
	return fmt.Sprintf("%6s %s", fmt.Sprintf("[%d]", t.tag.ArticleCount), t.tag.Name)
}
func (t tagItem) Description() string { return "" }
func (t tagItem) FilterValue() string { return t.tag.Name }
