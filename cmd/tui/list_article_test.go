package tui

import (
	"testing"
	"time"

	"github.com/benmatselby/wibble/pkg/models"
)

func TestArticleItemTitle_WithPublishedDate(t *testing.T) {
	published := time.Date(2026, 5, 3, 14, 30, 0, 0, time.UTC)
	item := articleItem{
		article: models.Article{
			Title:     "Test Article",
			Published: &published,
			IsRead:    true,
		},
	}

	got := item.Title()
	want := "   03 May 2026 14:30:00   Test Article"

	if got != want {
		t.Errorf("Title() = %q, want %q", got, want)
	}
}

func TestArticleItemTitle_WithoutPublishedDate(t *testing.T) {
	item := articleItem{
		article: models.Article{
			Title:     "No Date Article",
			Published: nil,
			IsRead:    true,
		},
	}

	got := item.Title()
	want := "                          No Date Article"

	if got != want {
		t.Errorf("Title() = %q, want %q", got, want)
	}
}

func TestArticleItemTitle_Unread(t *testing.T) {
	published := time.Date(2026, 5, 3, 14, 30, 0, 0, time.UTC)
	item := articleItem{
		article: models.Article{
			Title:     "News rocks the world",
			Published: &published,
			IsRead:    false,
		},
	}

	got := item.Title()
	want := "⏺  03 May 2026 14:30:00   News rocks the world"

	if got != want {
		t.Errorf("Title() = %q, want %q", got, want)
	}
}

func TestArticleItemDescription(t *testing.T) {
	item := articleItem{
		article: models.Article{
			Title: "Some Article",
		},
	}

	if got := item.Description(); got != "" {
		t.Errorf("Description() = %q, want empty string", got)
	}
}

func TestArticleItemFilterValue(t *testing.T) {
	item := articleItem{
		article: models.Article{
			Title: "Filterable Title",
		},
	}

	if got := item.FilterValue(); got != "Filterable Title" {
		t.Errorf("FilterValue() = %q, want %q", got, "Filterable Title")
	}
}

func TestArticleItemIsRead_True(t *testing.T) {
	item := articleItem{
		article: models.Article{
			IsRead: true,
		},
	}

	if !item.IsRead() {
		t.Error("IsRead() = false, want true")
	}
}

func TestArticleItemIsRead_False(t *testing.T) {
	item := articleItem{
		article: models.Article{
			IsRead: false,
		},
	}

	if item.IsRead() {
		t.Error("IsRead() = true, want false")
	}
}
