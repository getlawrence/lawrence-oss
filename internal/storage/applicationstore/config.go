package applicationstore

import (
	"github.com/getlawrence/lawrence-oss/internal/app"
)

// FactoryConfig represents the configuration for the application store meta factory
type FactoryConfig struct {
	Type string `yaml:"type"`
	Path string `yaml:"path"`
}

// ConfigFrom creates a FactoryConfig from the app storage config
func ConfigFrom(appConfig *app.Config) FactoryConfig {
	return FactoryConfig{
		Type: appConfig.Storage.App.Type,
		Path: appConfig.Storage.App.Path,
	}
}

// DefaultConfig returns a default configuration
func DefaultConfig() FactoryConfig {
	return FactoryConfig{
		Type: "sqlite",
		Path: "./data/app.db",
	}
}
