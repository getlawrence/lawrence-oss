// Copyright (c) 2024 Lawrence OSS Contributors
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"

	"go.uber.org/zap"

	"github.com/getlawrence/lawrence-oss/internal/storage/applicationstore"
	"github.com/getlawrence/lawrence-oss/internal/storage/telemetrystore"
)

// BaseApplicationStoreFactory is the factory interface for application store without Initialize.
type BaseApplicationStoreFactory interface {
	// CreateApplicationStore creates an applicationstore.ApplicationStore
	CreateApplicationStore() (applicationstore.ApplicationStore, error)
}

// ApplicationStoreFactory defines an interface for a factory that can create application store implementations.
type ApplicationStoreFactory interface {
	BaseApplicationStoreFactory
	// Initialize performs internal initialization of the factory, such as opening connections to the backend store.
	// It is called after all configuration of the factory itself has been done.
	Initialize(logger *zap.Logger) error
}

// BaseTelemetryStoreFactory is the factory interface for telemetry store without Initialize.
type BaseTelemetryStoreFactory interface {
	// CreateTelemetryReader creates a telemetrystore.Reader
	CreateTelemetryReader() (telemetrystore.Reader, error)

	// CreateTelemetryWriter creates a telemetrystore.Writer
	CreateTelemetryWriter() (telemetrystore.Writer, error)
}

// TelemetryStoreFactory defines an interface for a factory that can create telemetry store implementations.
type TelemetryStoreFactory interface {
	BaseTelemetryStoreFactory
	// Initialize performs internal initialization of the factory, such as opening connections to the backend store.
	// It is called after all configuration of the factory itself has been done.
	Initialize(logger *zap.Logger) error
}

// Purger defines an interface that is capable of purging the storage.
// Only meant to be used from integration tests.
type Purger interface {
	// Purge removes all data from the storage.
	Purge(context.Context) error
}

// Closer defines an interface for closing storage resources.
type Closer interface {
	// Close closes the storage and releases resources.
	Close() error
}
