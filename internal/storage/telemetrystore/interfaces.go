package telemetrystore

import (
	"github.com/getlawrence/lawrence-oss/internal/storage/telemetrystore/types"
	"go.uber.org/zap"
)

// TelemetryStoreFactory defines an interface for a factory that can create telemetry store implementations.
type TelemetryStoreFactory interface {
	// CreateTelemetryReader creates a types.Reader
	CreateTelemetryReader() (types.Reader, error)

	// CreateTelemetryWriter creates a types.Writer
	CreateTelemetryWriter() (types.Writer, error)

	// Initialize performs internal initialization of the factory, such as opening connections to the backend store.
	// It is called after all configuration of the factory itself has been done.
	Initialize(logger *zap.Logger) error
}
