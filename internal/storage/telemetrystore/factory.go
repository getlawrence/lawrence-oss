package telemetrystore

import (
	"fmt"
	"io"

	"go.uber.org/zap"

	"github.com/getlawrence/lawrence-oss/internal/config"
	"github.com/getlawrence/lawrence-oss/internal/storage/telemetrystore/duckdb"
	"github.com/getlawrence/lawrence-oss/internal/storage/telemetrystore/types"
)

const (
	duckdbStorageType = "duckdb"
	// Add more storage types as needed
	// memoryStorageType = "memory"
	// postgresStorageType = "postgres"
)

// AllStorageTypes defines all available telemetry storage backends
var AllStorageTypes = []string{
	duckdbStorageType,
	// Add more storage types as they are implemented
}

// Factory implements TelemetryStoreFactory interface as a meta-factory for telemetry storage components.
// It provides a clean abstraction layer over concrete storage implementations, allowing easy switching
// between different storage backends (DuckDB, PostgreSQL, etc.) without changing the main application code.
type Factory struct {
	Config
	logger   *zap.Logger
	factories map[string]TelemetryStoreFactory
}

// NewFactory creates the meta-factory.
// It automatically creates and registers the factory for the configured storage type.
// Example usage:
//   config := telemetrystore.ConfigFrom(appConfig)
//   factory, err := telemetrystore.NewFactory(config)
//   if err != nil {
//       log.Fatal(err)
//   }
//   defer factory.Close()
func NewFactory(config Config) (*Factory, error) {
	f := &Factory{Config: config}
	f.factories = make(map[string]TelemetryStoreFactory)
	
	// Validate storage type
	if !IsStorageTypeSupported(config.Type) {
		return nil, fmt.Errorf("unsupported storage type %s. Supported types: %v", config.Type, AllStorageTypes)
	}
	
	// Initialize the factory for the configured storage type
	factory, err := f.getFactoryOfType(config.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to create factory for storage type %s: %w", config.Type, err)
	}
	f.factories[config.Type] = factory
	
	return f, nil
}

// getFactoryOfType creates a factory instance for the given storage type
// To add a new storage type:
// 1. Add a constant for the storage type (e.g., postgresStorageType = "postgres")
// 2. Add it to AllStorageTypes slice
// 3. Add a case in this switch statement
// 4. Implement the TelemetryStoreFactory interface for your storage type
func (f *Factory) getFactoryOfType(factoryType string) (TelemetryStoreFactory, error) {
	switch factoryType {
	case duckdbStorageType:
		return duckdb.NewFactory(f.Config.Path), nil
	// Add more storage types as they are implemented
	// case memoryStorageType:
	//     return memory.NewFactory(), nil
	// case postgresStorageType:
	//     return postgres.NewFactory(f.Config.Path), nil
	default:
		return nil, fmt.Errorf("unknown telemetry storage type %s. Valid types are %v", factoryType, AllStorageTypes)
	}
}

// Initialize initializes the meta factory and all underlying factories
func (f *Factory) Initialize(logger *zap.Logger) error {
	f.logger = logger
	
	// Initialize all registered factories
	for storageType, factory := range f.factories {
		if err := factory.Initialize(logger); err != nil {
			return fmt.Errorf("failed to initialize %s factory: %w", storageType, err)
		}
	}
	
	return nil
}

// CreateTelemetryReader creates a telemetry reader using the configured storage type
func (f *Factory) CreateTelemetryReader() (types.Reader, error) {
	factory, ok := f.factories[f.Config.Type]
	if !ok {
		return nil, fmt.Errorf("no %s backend registered for telemetry store", f.Config.Type)
	}
	return factory.CreateTelemetryReader()
}

// CreateTelemetryWriter creates a telemetry writer using the configured storage type
func (f *Factory) CreateTelemetryWriter() (types.Writer, error) {
	factory, ok := f.factories[f.Config.Type]
	if !ok {
		return nil, fmt.Errorf("no %s backend registered for telemetry store", f.Config.Type)
	}
	return factory.CreateTelemetryWriter()
}

// Close closes all underlying factories
func (f *Factory) Close() error {
	var errs []error
	for storageType, factory := range f.factories {
		if closer, ok := factory.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close %s factory: %w", storageType, err))
			}
		}
	}
	
	if len(errs) > 0 {
		return fmt.Errorf("errors closing factories: %v", errs)
	}
	
	return nil
}

// GetStorageType returns the configured storage type
func (f *Factory) GetStorageType() string {
	return f.Config.Type
}

// IsStorageTypeSupported checks if a storage type is supported
func IsStorageTypeSupported(storageType string) bool {
	for _, supportedType := range AllStorageTypes {
		if supportedType == storageType {
			return true
		}
	}
	return false
}

// NewFactoryFromAppConfig creates a new factory directly from app configuration
// This is a convenience function that combines ConfigFrom and NewFactory
func NewFactoryFromAppConfig(appConfig *config.Config) (*Factory, error) {
	config := ConfigFrom(appConfig)
	return NewFactory(config)
}