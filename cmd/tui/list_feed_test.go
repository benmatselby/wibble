package tui

import (
	"testing"

	"github.com/benmatselby/wibble/pkg/models"
)

func TestFeedItemTitle(t *testing.T) {
	item := feedItem{
		feed: models.Feed{
			Title:       "My Feed",
			UnreadCount: 3,
			TotalCount:  10,
		},
	}

	got := item.Title()
	want := "    [3/10] My Feed"

	if got != want {
		t.Errorf("Title() = %q, want %q", got, want)
	}
}

func TestFeedItemTitle_AllRead(t *testing.T) {
	item := feedItem{
		feed: models.Feed{
			Title:       "Read Feed",
			UnreadCount: 0,
			TotalCount:  5,
		},
	}

	got := item.Title()
	want := "     [0/5] Read Feed"

	if got != want {
		t.Errorf("Title() = %q, want %q", got, want)
	}
}

func TestFeedItemDescription(t *testing.T) {
	item := feedItem{
		feed: models.Feed{
			Title: "Some Feed",
		},
	}

	if got := item.Description(); got != "" {
		t.Errorf("Description() = %q, want empty string", got)
	}
}

func TestFeedItemFilterValue(t *testing.T) {
	item := feedItem{
		feed: models.Feed{
			Title: "Filter Feed",
			URL:   "https://example.com/feed",
		},
	}

	got := item.FilterValue()
	want := "Filter Feed"

	if got != want {
		t.Errorf("FilterValue() = %q, want %q", got, want)
	}
}

func TestFeedItemIsRead_True(t *testing.T) {
	item := feedItem{
		feed: models.Feed{
			UnreadCount: 0,
			TotalCount:  5,
		},
	}

	if !item.IsRead() {
		t.Error("IsRead() = false, want true")
	}
}

func TestFeedItemIsRead_False(t *testing.T) {
	item := feedItem{
		feed: models.Feed{
			UnreadCount: 2,
			TotalCount:  5,
		},
	}

	if item.IsRead() {
		t.Error("IsRead() = true, want false")
	}
}
