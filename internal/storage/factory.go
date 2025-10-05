package storage

import (
	"fmt"

	"github.com/getlawrence/lawrence-oss/internal/storage/duckdb"
	"github.com/getlawrence/lawrence-oss/internal/storage/interfaces"
	"github.com/getlawrence/lawrence-oss/internal/storage/sqlite"
	"go.uber.org/zap"
)

// Container holds both app and telemetry storage instances
type Container struct {
	App       interfaces.AppStorage
	Telemetry interfaces.TelemetryStorage
	logger    *zap.Logger
}

// NewContainer creates a new storage container with both SQLite and DuckDB
func NewContainer(sqlitePath, duckdbPath string, logger *zap.Logger) (*Container, error) {
	// Initialize SQLite for app data
	appStorage, err := sqlite.NewStorage(sqlitePath, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize SQLite storage: %w", err)
	}

	// Initialize DuckDB for telemetry data
	telemetryStorage, err := duckdb.NewStorage(duckdbPath, logger)
	if err != nil {
		appStorage.Close() // Clean up on error
		return nil, fmt.Errorf("failed to initialize DuckDB storage: %w", err)
	}

	container := &Container{
		App:       appStorage,
		Telemetry: telemetryStorage,
		logger:    logger,
	}

	logger.Info("Storage container initialized",
		zap.String("sqlite_path", sqlitePath),
		zap.String("duckdb_path", duckdbPath))

	return container, nil
}

// Close closes both storage connections
func (c *Container) Close() error {
	var appErr, telemetryErr error

	if c.App != nil {
		appErr = c.App.Close()
	}

	if c.Telemetry != nil {
		telemetryErr = c.Telemetry.Close()
	}

	if appErr != nil {
		return fmt.Errorf("failed to close app storage: %w", appErr)
	}
	if telemetryErr != nil {
		return fmt.Errorf("failed to close telemetry storage: %w", telemetryErr)
	}

	c.logger.Info("Storage container closed")
	return nil
}

// GetTelemetryWriter returns a TelemetryWriter implementation for OTLP receivers
func (c *Container) GetTelemetryWriter() interface{} {
	if duckdbStorage, ok := c.Telemetry.(*duckdb.Storage); ok {
		return duckdb.NewWriterAdapter(duckdbStorage)
	}
	return nil
}
