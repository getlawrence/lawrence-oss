package app

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	OTLP      OTLPConfig      `yaml:"otlp"`
	Storage   StorageConfig   `yaml:"storage"`
	Retention RetentionConfig `yaml:"retention"`
	Rollups   RollupsConfig   `yaml:"rollups"`
	Logging   LoggingConfig   `yaml:"logging"`
}

// ServerConfig contains server configuration
type ServerConfig struct {
	HTTPPort  int `yaml:"http_port"`
	OpAMPPort int `yaml:"opamp_port"`
}

// OTLPConfig contains OTLP receiver configuration
type OTLPConfig struct {
	GRPCEndpoint string `yaml:"grpc_endpoint"`
	HTTPEndpoint string `yaml:"http_endpoint"`
}

// StorageConfig contains storage configuration
type StorageConfig struct {
	App       AppStorageConfig       `yaml:"app"`
	Telemetry TelemetryStorageConfig `yaml:"telemetry"`
}

// AppStorageConfig contains app storage configuration
type AppStorageConfig struct {
	Type string `yaml:"type"`
	Path string `yaml:"path"`
}

// TelemetryStorageConfig contains telemetry storage configuration
type TelemetryStorageConfig struct {
	Type string `yaml:"type"`
	Path string `yaml:"path"`
}

// RetentionConfig contains data retention configuration
type RetentionConfig struct {
	RawMetrics string `yaml:"raw_metrics"`
	RawLogs    string `yaml:"raw_logs"`
	Rollups1m  string `yaml:"rollups_1m"`
	Rollups5m  string `yaml:"rollups_5m"`
}

// RollupsConfig contains rollup configuration
type RollupsConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Interval1m string `yaml:"interval_1m"`
	Interval5m string `yaml:"interval_5m"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			HTTPPort:  8080,
			OpAMPPort: 4320,
		},
		OTLP: OTLPConfig{
			GRPCEndpoint: "0.0.0.0:4317",
			HTTPEndpoint: "0.0.0.0:4318",
		},
		Storage: StorageConfig{
			App: AppStorageConfig{
				Type: "sqlite",
				Path: "./data/app.db",
			},
			Telemetry: TelemetryStorageConfig{
				Type: "duckdb",
				Path: "./data/telemetry.db",
			},
		},
		Retention: RetentionConfig{
			RawMetrics: "24h",
			RawLogs:    "24h",
			Rollups1m:  "7d",
			Rollups5m:  "30d",
		},
		Rollups: RollupsConfig{
			Enabled:    true,
			Interval1m: "*/1 * * * *",
			Interval5m: "*/5 * * * *",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}
}

// ParseDuration parses a duration string like "24h", "7d", "30d"
func ParseDuration(s string) (time.Duration, error) {
	if len(s) < 2 {
		return 0, fmt.Errorf("invalid duration format: %s", s)
	}

	unit := s[len(s)-1:]
	value := s[:len(s)-1]

	var duration time.Duration
	switch unit {
	case "h":
		d, err := time.ParseDuration(value + "h")
		if err != nil {
			return 0, err
		}
		duration = d
	case "d":
		// Parse days as integer
		var days int
		if _, err := fmt.Sscanf(value, "%d", &days); err != nil {
			return 0, fmt.Errorf("invalid day value: %s", value)
		}
		duration = time.Duration(days*24) * time.Hour
	default:
		return time.ParseDuration(s)
	}

	return duration, nil
}
