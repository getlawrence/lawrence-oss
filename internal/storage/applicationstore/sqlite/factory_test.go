// Copyright (c) 2024 Lawrence OSS Contributors
// SPDX-License-Identifier: Apache-2.0

package sqlite

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestFactoryInitialize(t *testing.T) {
	dbPath := makeTempDB(t)
	defer os.Remove(dbPath)

	factory := NewFactory(dbPath)
	logger := zap.NewNop()

	err := factory.Initialize(logger)
	require.NoError(t, err)
	assert.NotNil(t, factory.store)
	assert.NotNil(t, factory.logger)

	defer factory.Close()
}

func TestFactoryCreateApplicationStore(t *testing.T) {
	dbPath := makeTempDB(t)
	defer os.Remove(dbPath)

	factory := NewFactory(dbPath)
	logger := zap.NewNop()

	err := factory.Initialize(logger)
	require.NoError(t, err)
	defer factory.Close()

	store, err := factory.CreateApplicationStore()
	require.NoError(t, err)
	assert.NotNil(t, store)
}

func TestFactoryPurge(t *testing.T) {
	dbPath := makeTempDB(t)
	defer os.Remove(dbPath)

	factory := NewFactory(dbPath)
	logger := zap.NewNop()

	err := factory.Initialize(logger)
	require.NoError(t, err)
	defer factory.Close()

	store, err := factory.CreateApplicationStore()
	require.NoError(t, err)

	// Add some data
	agent := makeTestAgent(uuid.New())
	err = store.CreateAgent(context.Background(), agent)
	require.NoError(t, err)

	// Purge
	err = factory.Purge(context.Background())
	require.NoError(t, err)

	// Verify data was removed
	agents, err := store.ListAgents(context.Background())
	require.NoError(t, err)
	assert.Empty(t, agents)
}

func TestFactoryClose(t *testing.T) {
	dbPath := makeTempDB(t)
	defer os.Remove(dbPath)

	factory := NewFactory(dbPath)
	logger := zap.NewNop()

	err := factory.Initialize(logger)
	require.NoError(t, err)

	err = factory.Close()
	require.NoError(t, err)
}
