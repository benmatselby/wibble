package config_test

import (
	"testing"

	"github.com/benmatselby/wibble/pkg/config"
	"github.com/spf13/viper"
)

func TestGetFeeds_ReturnsFeedURLs(t *testing.T) {
	viper.Reset()
	viper.Set("feeds", []string{"https://example.com/feed1.xml", "https://example.com/feed2.xml"})

	feeds, err := config.GetFeeds()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(feeds) != 2 {
		t.Fatalf("expected 2 feeds, got %d", len(feeds))
	}
}

func TestGetFeeds_ReturnsEmptyWhenNotSet(t *testing.T) {
	viper.Reset()

	feeds, err := config.GetFeeds()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(feeds) != 0 {
		t.Fatalf("expected 0 feeds, got %d", len(feeds))
	}
}

func TestGetRefreshInterval_ReturnsConfiguredValue(t *testing.T) {
	viper.Reset()
	viper.Set("refresh_interval", 30)

	interval := config.GetRefreshInterval()
	if interval != 30 {
		t.Errorf("expected 30, got %d", interval)
	}
}

func TestGetRefreshInterval_ReturnsDefaultWhenNotSet(t *testing.T) {
	viper.Reset()

	interval := config.GetRefreshInterval()
	if interval != 10 {
		t.Errorf("expected default 10, got %d", interval)
	}
}

func TestGetRefreshInterval_ReturnsDefaultForZero(t *testing.T) {
	viper.Reset()
	viper.Set("refresh_interval", 0)

	interval := config.GetRefreshInterval()
	if interval != 10 {
		t.Errorf("expected default 10 for zero value, got %d", interval)
	}
}

func TestGetRefreshInterval_ReturnsDefaultForNegative(t *testing.T) {
	viper.Reset()
	viper.Set("refresh_interval", -5)

	interval := config.GetRefreshInterval()
	if interval != 10 {
		t.Errorf("expected default 10 for negative value, got %d", interval)
	}
}
