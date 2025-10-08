package applicationstore

import (
	"context"
	"github.com/getlawrence/lawrence-oss/internal/storage/applicationstore/types"
	"go.uber.org/zap"
)

// ApplicationStoreFactory defines an interface for a factory that can create application store implementations.
type ApplicationStoreFactory interface {
	// CreateApplicationStore creates a types.ApplicationStore
	CreateApplicationStore() (types.ApplicationStore, error)

	// Initialize performs internal initialization of the factory, such as opening connections to the backend store.
	// It is called after all configuration of the factory itself has been done.
	Initialize(logger *zap.Logger) error
}

// Purger defines an interface that is capable of purging the storage.
// Only meant to be used from integration tests.
type Purger interface {
	// Purge removes all data from the storage.
	Purge(ctx context.Context) error
}

// Closer defines an interface for closing storage resources.
type Closer interface {
	// Close closes the storage and releases resources.
	Close() error
}
