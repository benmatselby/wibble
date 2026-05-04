package models_test

import (
	"testing"

	"github.com/benmatselby/wibble/pkg/models"
)

func TestFeed_Read(t *testing.T) {
	tests := []struct {
		name        string // description of this test case
		want        bool
		totalCount  int
		unreadCount int
	}{
		{
			name:        "feed with no unread articles is read",
			want:        true,
			totalCount:  5,
			unreadCount: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f models.Feed
			f.TotalCount = tt.totalCount
			f.UnreadCount = tt.unreadCount
			got := f.Read()
			if got != tt.want {
				t.Errorf("Read() = %v, want %v", got, tt.want)
			}
		})
	}
}
