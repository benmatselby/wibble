package models

// Tag represents a user-defined label that can be attached to articles.
type Tag struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	ArticleCount int    `json:"article_count"`
}
