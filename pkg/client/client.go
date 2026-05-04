// Package client provides mechanism for fetching RSS data
package client

import (
	"github.com/benmatselby/wibble/pkg/models"
)

// API defines the interface for the client package.
type API interface {
	// ParseURL takes a URL and returns a Feed and a slice of Articles.
	ParseURL(url string) (*models.Feed, []models.Article, error)
}
