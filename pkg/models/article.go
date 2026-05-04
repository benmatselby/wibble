package models

import "time"

type Article struct {
	ID        int64      `json:"id"`
	FeedID    int64      `json:"feed_id"`
	Title     string     `json:"title"`
	Link      string     `json:"link"`
	Published *time.Time `json:"published"`
	Summary   string     `json:"summary"`
	IsRead    bool
}
