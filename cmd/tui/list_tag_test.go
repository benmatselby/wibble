package tui

import (
	"testing"

	"github.com/benmatselby/wibble/pkg/models"
)

func TestTagItemTitle(t *testing.T) {
	item := tagItem{
		tag: models.Tag{
			Name:         "research",
			ArticleCount: 3,
		},
	}

	got := item.Title()
	want := "   [3] research"

	if got != want {
		t.Errorf("Title() = %q, want %q", got, want)
	}
}

func TestTagItemDescription(t *testing.T) {
	item := tagItem{
		tag: models.Tag{
			Name: "research",
		},
	}

	if got := item.Description(); got != "" {
		t.Errorf("Description() = %q, want empty string", got)
	}
}

func TestTagItemFilterValue(t *testing.T) {
	item := tagItem{
		tag: models.Tag{
			Name: "research",
		},
	}

	got := item.FilterValue()
	want := "research"

	if got != want {
		t.Errorf("FilterValue() = %q, want %q", got, want)
	}
}
