package telemetrystore

import (
	"github.com/getlawrence/lawrence-oss/internal/config"
)

// Config represents the configuration for the telemetry store meta factory
type Config struct {
	Type string `yaml:"type"`
	Path string `yaml:"path"`
}

// ConfigFrom creates a Config from the app storage config
func ConfigFrom(appConfig *config.Config) Config {
	return Config{
		Type: appConfig.Storage.Telemetry.Type,
		Path: appConfig.Storage.Telemetry.Path,
	}
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		Type: "duckdb",
		Path: "./data/telemetry.db",
	}
}
