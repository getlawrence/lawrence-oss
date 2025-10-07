// Copyright (c) 2024 Lawrence OSS Contributors
// SPDX-License-Identifier: Apache-2.0

package duckdb

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/getlawrence/lawrence-oss/internal/storage/telemetrystore"
)

// Factory implements TelemetryStoreFactory and creates storage components backed by DuckDB.
type Factory struct {
	dbPath  string
	logger  *zap.Logger
	storage *Storage
}

// NewFactory creates a new Factory with the given database path.
func NewFactory(dbPath string) *Factory {
	return &Factory{
		dbPath: dbPath,
	}
}

// Initialize implements TelemetryStoreFactory
func (f *Factory) Initialize(logger *zap.Logger) error {
	f.logger = logger
	storage, err := NewStorage(f.dbPath, logger)
	if err != nil {
		return err
	}
	f.storage = storage
	return nil
}

// CreateTelemetryReader implements TelemetryStoreFactory
func (f *Factory) CreateTelemetryReader() (telemetrystore.Reader, error) {
	return &Reader{Storage: f.storage}, nil
}

// CreateTelemetryWriter implements TelemetryStoreFactory
func (f *Factory) CreateTelemetryWriter() (telemetrystore.Writer, error) {
	return &Writer{Storage: f.storage}, nil
}

// CreateWriterAdapter creates a writer adapter for OTLP receivers
func (f *Factory) CreateWriterAdapter() *WriterAdapter {
	return NewWriterAdapter(f.storage)
}

// Purge removes all data from the Factory's underlying DuckDB store.
// This function is intended for testing purposes only and should not be used in production environments.
func (f *Factory) Purge(ctx context.Context) error {
	f.logger.Info("Purging data from DuckDB telemetry store")

	tables := []string{
		"metrics_sum",
		"metrics_gauge",
		"metrics_histogram",
		"logs",
		"traces",
		"rollups_1m",
		"rollups_5m",
		"rollups_1h",
		"rollups_1d",
	}

	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s", table)
		_, err := f.storage.db.ExecContext(ctx, query)
		if err != nil {
			// Ignore errors for tables that don't exist
			f.logger.Warn("Failed to purge table", zap.String("table", table), zap.Error(err))
		}
	}

	return nil
}

// Close closes the storage connection
func (f *Factory) Close() error {
	if f.storage != nil {
		return f.storage.Close()
	}
	return nil
}
