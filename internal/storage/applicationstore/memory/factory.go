// Copyright (c) 2024 Lawrence OSS Contributors
// SPDX-License-Identifier: Apache-2.0

package memory

import (
	"context"

	"go.uber.org/zap"

	"github.com/getlawrence/lawrence-oss/internal/storage/applicationstore/types"
)

// Factory implements ApplicationStoreFactory and creates storage components backed by memory store.
type Factory struct {
	logger *zap.Logger
	store  *Store
}

// NewFactory creates a new Factory.
func NewFactory() *Factory {
	return &Factory{}
}

// Initialize implements storage.ApplicationStoreFactory
func (f *Factory) Initialize(logger *zap.Logger) error {
	f.logger = logger
	f.store = NewStore()
	logger.Info("Memory application store initialized")
	return nil
}

// CreateApplicationStore implements storage.ApplicationStoreFactory
func (f *Factory) CreateApplicationStore() (types.ApplicationStore, error) {
	return f.store, nil
}

// Purge removes all data from the Factory's underlying memory store.
// This function is intended for testing purposes only and should not be used in production environments.
func (f *Factory) Purge(ctx context.Context) error {
	f.logger.Info("Purging data from memory application store")
	f.store.purge(ctx)
	return nil
}
