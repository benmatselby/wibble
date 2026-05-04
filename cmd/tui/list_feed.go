package tui

import (
	"fmt"

	"github.com/benmatselby/wibble/pkg/models"
)

type feedItem struct {
	feed models.Feed
}

func (f feedItem) Title() string {
	counts := fmt.Sprintf("[%d/%d]", f.feed.UnreadCount, f.feed.TotalCount)
	return fmt.Sprintf("%10s %s", counts, f.feed.Title)
}
func (f feedItem) Description() string { return "" }
func (f feedItem) FilterValue() string { return f.feed.Title }
func (f feedItem) IsRead() bool        { return f.feed.Read() }
