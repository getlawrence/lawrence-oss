package storage

import (
	"go.uber.org/zap"

	"github.com/getlawrence/lawrence-oss/internal/storage/applicationstore"
	"github.com/getlawrence/lawrence-oss/internal/storage/applicationstore/sqlite"
	"github.com/getlawrence/lawrence-oss/internal/storage/telemetrystore"
	"github.com/getlawrence/lawrence-oss/internal/storage/telemetrystore/duckdb"
)

// SQLiteApplicationStoreFactory creates SQLite-based application store instances
type SQLiteApplicationStoreFactory struct {
	dbPath string
	logger *zap.Logger
}

// NewSQLiteApplicationStoreFactory creates a new SQLite application store factory
func NewSQLiteApplicationStoreFactory(dbPath string, logger *zap.Logger) *SQLiteApplicationStoreFactory {
	return &SQLiteApplicationStoreFactory{
		dbPath: dbPath,
		logger: logger,
	}
}

// Initialize implements ApplicationStoreFactory
func (f *SQLiteApplicationStoreFactory) Initialize(logger *zap.Logger) error {
	f.logger = logger
	return nil
}

// CreateApplicationStore creates an applicationstore.ApplicationStore
func (f *SQLiteApplicationStoreFactory) CreateApplicationStore() (applicationstore.ApplicationStore, error) {
	return sqlite.NewSQLiteStorage(f.dbPath, f.logger)
}

// DuckDBTelemetryStoreFactory creates DuckDB-based telemetry store instances
type DuckDBTelemetryStoreFactory struct {
	dbPath  string
	logger  *zap.Logger
	factory *duckdb.Factory
}

// NewDuckDBTelemetryStoreFactory creates a new DuckDB telemetry store factory
func NewDuckDBTelemetryStoreFactory(dbPath string, logger *zap.Logger) *DuckDBTelemetryStoreFactory {
	factory := duckdb.NewFactory(dbPath)
	if err := factory.Initialize(logger); err != nil {
		logger.Fatal("Failed to initialize DuckDB factory", zap.Error(err))
	}
	return &DuckDBTelemetryStoreFactory{
		dbPath:  dbPath,
		logger:  logger,
		factory: factory,
	}
}

// Initialize implements TelemetryStoreFactory
func (f *DuckDBTelemetryStoreFactory) Initialize(logger *zap.Logger) error {
	f.logger = logger
	return nil
}

// CreateTelemetryReader creates a telemetrystore.Reader
func (f *DuckDBTelemetryStoreFactory) CreateTelemetryReader() (telemetrystore.Reader, error) {
	return f.factory.CreateTelemetryReader()
}

// CreateTelemetryWriter creates a telemetrystore.Writer
func (f *DuckDBTelemetryStoreFactory) CreateTelemetryWriter() (telemetrystore.Writer, error) {
	return f.factory.CreateTelemetryWriter()
}

// CreateWriterAdapter creates a writer adapter for OTLP receivers
func (f *DuckDBTelemetryStoreFactory) CreateWriterAdapter() *duckdb.WriterAdapter {
	return f.factory.CreateWriterAdapter()
}
