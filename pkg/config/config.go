// Package config is a wrapper around viper
package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func GetFeeds() ([]string, error) {
	var configFeeds []string
	if err := viper.UnmarshalKey("feeds", &configFeeds); err != nil {
		return nil, fmt.Errorf("failed to read feeds from config")
	}

	return configFeeds, nil
}

// GetRefreshInterval returns the refresh interval in minutes from the configuration file.
func GetRefreshInterval() int {
	var refreshInterval int
	defaultRefreshInterval := 10
	if err := viper.UnmarshalKey("refresh_interval", &refreshInterval); err != nil {
		return defaultRefreshInterval
	}

	if refreshInterval <= 0 {
		return defaultRefreshInterval
	}

	return refreshInterval
}
