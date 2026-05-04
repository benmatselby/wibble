// Package models defines the data structure for the application.
package models

import "time"

type Feed struct {
	ID          int64     `json:"id"`
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	AddedAt     time.Time `json:"added_at"`
	TotalCount  int
	UnreadCount int
	SortIndex   int
}

func (f Feed) Read() bool { return f.UnreadCount == 0 }
