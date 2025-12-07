// Copyright (c) 2024 Lawrence OSS Contributors
// SPDX-License-Identifier: Apache-2.0

package sqlite

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/getlawrence/lawrence-oss/internal/storage/applicationstore/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func makeTempDB(t *testing.T) string {
	tmpDir := t.TempDir()
	return filepath.Join(tmpDir, "test.db")
}

func makeTestAgent(id uuid.UUID) *types.Agent {
	return &types.Agent{
		ID:   id,
		Name: "test-agent",
		Labels: map[string]string{
			"env":     "test",
			"service": "demo",
		},
		Status:       types.AgentStatusOnline,
		LastSeen:     time.Now().UTC(),
		Version:      "1.0.0",
		Capabilities: []string{"metrics", "logs", "traces"},
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
}

func makeTestGroup(id string) *types.Group {
	return &types.Group{
		ID:   id,
		Name: "test-group",
		Labels: map[string]string{
			"env": "test",
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}

func makeTestConfig(id string, agentID *uuid.UUID, groupID *string) *types.Config {
	return &types.Config{
		ID:         id,
		AgentID:    agentID,
		GroupID:    groupID,
		ConfigHash: "abc123",
		Content:    "receivers:\n  otlp:",
		Version:    1,
		CreatedAt:  time.Now().UTC(),
	}
}

func withSQLiteStore(t *testing.T, f func(store types.ApplicationStore)) {
	dbPath := makeTempDB(t)
	logger := zap.NewNop()

	store, err := NewSQLiteStorage(dbPath, logger)
	require.NoError(t, err)
	defer func() {
		if closer, ok := store.(*Storage); ok {
			closer.Close()
		}
		os.Remove(dbPath)
	}()

	f(store)
}

func withPopulatedSQLiteStore(t *testing.T, f func(store types.ApplicationStore, agentID uuid.UUID)) {
	withSQLiteStore(t, func(store types.ApplicationStore) {
		agentID := uuid.New()
		agent := makeTestAgent(agentID)
		err := store.CreateAgent(context.Background(), agent)
		require.NoError(t, err)
		f(store, agentID)
	})
}

// Agent tests

func TestSQLiteCreateAgent(t *testing.T) {
	withSQLiteStore(t, func(store types.ApplicationStore) {
		agentID := uuid.New()
		agent := makeTestAgent(agentID)

		err := store.CreateAgent(context.Background(), agent)
		require.NoError(t, err)

		// Verify it was created
		retrieved, err := store.GetAgent(context.Background(), agentID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, agent.ID, retrieved.ID)
		assert.Equal(t, agent.Name, retrieved.Name)
		assert.Equal(t, agent.Status, retrieved.Status)
	})
}

func TestSQLiteGetAgentNotFound(t *testing.T) {
	withSQLiteStore(t, func(store types.ApplicationStore) {
		retrieved, err := store.GetAgent(context.Background(), uuid.New())
		require.NoError(t, err)
		assert.Nil(t, retrieved)
	})
}

func TestSQLiteListAgents(t *testing.T) {
	withSQLiteStore(t, func(store types.ApplicationStore) {
		agent1 := makeTestAgent(uuid.New())
		agent2 := makeTestAgent(uuid.New())

		err := store.CreateAgent(context.Background(), agent1)
		require.NoError(t, err)
		err = store.CreateAgent(context.Background(), agent2)
		require.NoError(t, err)

		agents, err := store.ListAgents(context.Background())
		require.NoError(t, err)
		assert.Len(t, agents, 2)
	})
}

func TestSQLiteUpdateAgentStatus(t *testing.T) {
	withPopulatedSQLiteStore(t, func(store types.ApplicationStore, agentID uuid.UUID) {
		err := store.UpdateAgentStatus(context.Background(), agentID, types.AgentStatusOffline)
		require.NoError(t, err)

		// Verify the update
		agent, err := store.GetAgent(context.Background(), agentID)
		require.NoError(t, err)
		assert.Equal(t, types.AgentStatusOffline, agent.Status)
	})
}

func TestSQLiteUpdateAgentStatusNotFound(t *testing.T) {
	withSQLiteStore(t, func(store types.ApplicationStore) {
		err := store.UpdateAgentStatus(context.Background(), uuid.New(), types.AgentStatusOffline)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestSQLiteUpdateAgentLastSeen(t *testing.T) {
	withPopulatedSQLiteStore(t, func(store types.ApplicationStore, agentID uuid.UUID) {
		newTime := time.Now().Add(time.Hour).UTC()
		err := store.UpdateAgentLastSeen(context.Background(), agentID, newTime)
		require.NoError(t, err)

		// Verify the update
		agent, err := store.GetAgent(context.Background(), agentID)
		require.NoError(t, err)
		assert.Equal(t, newTime.Unix(), agent.LastSeen.Unix())
	})
}

func TestSQLiteDeleteAgent(t *testing.T) {
	withPopulatedSQLiteStore(t, func(store types.ApplicationStore, agentID uuid.UUID) {
		err := store.DeleteAgent(context.Background(), agentID)
		require.NoError(t, err)

		// Verify it was deleted
		agent, err := store.GetAgent(context.Background(), agentID)
		require.NoError(t, err)
		assert.Nil(t, agent)
	})
}

func TestSQLiteDeleteAgentNotFound(t *testing.T) {
	withSQLiteStore(t, func(store types.ApplicationStore) {
		err := store.DeleteAgent(context.Background(), uuid.New())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

// Group tests

func TestSQLiteCreateGroup(t *testing.T) {
	withSQLiteStore(t, func(store types.ApplicationStore) {
		group := makeTestGroup("test-group")

		err := store.CreateGroup(context.Background(), group)
		require.NoError(t, err)

		// Verify it was created
		retrieved, err := store.GetGroup(context.Background(), group.ID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, group.ID, retrieved.ID)
		assert.Equal(t, group.Name, retrieved.Name)
	})
}

func TestSQLiteGetGroupNotFound(t *testing.T) {
	withSQLiteStore(t, func(store types.ApplicationStore) {
		retrieved, err := store.GetGroup(context.Background(), "nonexistent")
		require.NoError(t, err)
		assert.Nil(t, retrieved)
	})
}

func TestSQLiteListGroups(t *testing.T) {
	withSQLiteStore(t, func(store types.ApplicationStore) {
		group1 := makeTestGroup("group-1")
		group2 := makeTestGroup("group-2")

		err := store.CreateGroup(context.Background(), group1)
		require.NoError(t, err)
		err = store.CreateGroup(context.Background(), group2)
		require.NoError(t, err)

		groups, err := store.ListGroups(context.Background())
		require.NoError(t, err)
		assert.Len(t, groups, 2)
	})
}

func TestSQLiteDeleteGroup(t *testing.T) {
	withSQLiteStore(t, func(store types.ApplicationStore) {
		group := makeTestGroup("test-group")
		err := store.CreateGroup(context.Background(), group)
		require.NoError(t, err)

		err = store.DeleteGroup(context.Background(), group.ID)
		require.NoError(t, err)

		// Verify it was deleted
		retrieved, err := store.GetGroup(context.Background(), group.ID)
		require.NoError(t, err)
		assert.Nil(t, retrieved)
	})
}

// Config tests

func TestSQLiteCreateConfig(t *testing.T) {
	withPopulatedSQLiteStore(t, func(store types.ApplicationStore, agentID uuid.UUID) {
		config := makeTestConfig("config-1", &agentID, nil)

		err := store.CreateConfig(context.Background(), config)
		require.NoError(t, err)

		// Verify it was created
		retrieved, err := store.GetConfig(context.Background(), config.ID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, config.ID, retrieved.ID)
		assert.Equal(t, config.Content, retrieved.Content)
	})
}

func TestSQLiteGetLatestConfigForAgent(t *testing.T) {
	withPopulatedSQLiteStore(t, func(store types.ApplicationStore, agentID uuid.UUID) {
		// Create multiple configs for the same agent
		config1 := makeTestConfig("config-1", &agentID, nil)
		config1.Version = 1
		config1.CreatedAt = time.Now().Add(-time.Hour).UTC()

		config2 := makeTestConfig("config-2", &agentID, nil)
		config2.Version = 2
		config2.CreatedAt = time.Now().UTC()

		err := store.CreateConfig(context.Background(), config1)
		require.NoError(t, err)
		err = store.CreateConfig(context.Background(), config2)
		require.NoError(t, err)

		// Get latest config
		latest, err := store.GetLatestConfigForAgent(context.Background(), agentID)
		require.NoError(t, err)
		require.NotNil(t, latest)
		assert.Equal(t, config2.ID, latest.ID)
		assert.Equal(t, 2, latest.Version)
	})
}

func TestSQLiteGetLatestConfigForGroup(t *testing.T) {
	withSQLiteStore(t, func(store types.ApplicationStore) {
		groupID := "test-group"
		group := makeTestGroup(groupID)
		err := store.CreateGroup(context.Background(), group)
		require.NoError(t, err)

		// Create multiple configs for the same group
		config1 := makeTestConfig("config-1", nil, &groupID)
		config1.Version = 1
		config1.CreatedAt = time.Now().Add(-time.Hour).UTC()

		config2 := makeTestConfig("config-2", nil, &groupID)
		config2.Version = 2
		config2.CreatedAt = time.Now().UTC()

		err = store.CreateConfig(context.Background(), config1)
		require.NoError(t, err)
		err = store.CreateConfig(context.Background(), config2)
		require.NoError(t, err)

		// Get latest config
		latest, err := store.GetLatestConfigForGroup(context.Background(), groupID)
		require.NoError(t, err)
		require.NotNil(t, latest)
		assert.Equal(t, config2.ID, latest.ID)
		assert.Equal(t, 2, latest.Version)
	})
}

func TestSQLiteListConfigsWithFilter(t *testing.T) {
	withSQLiteStore(t, func(store types.ApplicationStore) {
		agentID1 := uuid.New()
		agentID2 := uuid.New()
		groupID := "test-group"

		// Create agents and group
		agent1 := makeTestAgent(agentID1)
		agent2 := makeTestAgent(agentID2)
		group := makeTestGroup(groupID)

		err := store.CreateAgent(context.Background(), agent1)
		require.NoError(t, err)
		err = store.CreateAgent(context.Background(), agent2)
		require.NoError(t, err)
		err = store.CreateGroup(context.Background(), group)
		require.NoError(t, err)

		// Create configs
		config1 := makeTestConfig("config-1", &agentID1, nil)
		config2 := makeTestConfig("config-2", &agentID2, nil)
		config3 := makeTestConfig("config-3", nil, &groupID)

		err = store.CreateConfig(context.Background(), config1)
		require.NoError(t, err)
		err = store.CreateConfig(context.Background(), config2)
		require.NoError(t, err)
		err = store.CreateConfig(context.Background(), config3)
		require.NoError(t, err)

		// Filter by agent ID
		result, err := store.ListConfigs(context.Background(), types.ConfigFilter{
			AgentID: &agentID1,
		})
		require.NoError(t, err)
		assert.Len(t, result.Configs, 1)
		assert.Equal(t, config1.ID, result.Configs[0].ID)
		assert.Equal(t, 1, result.TotalCount)

		// Filter by group ID
		result, err = store.ListConfigs(context.Background(), types.ConfigFilter{
			GroupID: &groupID,
		})
		require.NoError(t, err)
		assert.Len(t, result.Configs, 1)
		assert.Equal(t, config3.ID, result.Configs[0].ID)
		assert.Equal(t, 1, result.TotalCount)

		// No filter
		result, err = store.ListConfigs(context.Background(), types.ConfigFilter{})
		require.NoError(t, err)
		assert.Len(t, result.Configs, 3)
		assert.Equal(t, 3, result.TotalCount)
	})
}

func TestSQLiteListConfigsWithLimit(t *testing.T) {
	withPopulatedSQLiteStore(t, func(store types.ApplicationStore, agentID uuid.UUID) {
		for i := 0; i < 5; i++ {
			config := makeTestConfig(uuid.New().String(), &agentID, nil)
			err := store.CreateConfig(context.Background(), config)
			require.NoError(t, err)
		}

		// List with limit
		result, err := store.ListConfigs(context.Background(), types.ConfigFilter{
			Limit: 3,
		})
		require.NoError(t, err)
		assert.Len(t, result.Configs, 3)
		assert.Equal(t, 5, result.TotalCount)
	})
}

// Schema migration tests

func TestSQLiteMigration(t *testing.T) {
	dbPath := makeTempDB(t)
	defer os.Remove(dbPath)

	logger := zap.NewNop()

	// Create first instance - should run migrations
	store1, err := NewSQLiteStorage(dbPath, logger)
	require.NoError(t, err)
	storage1 := store1.(*Storage)
	storage1.Close()

	// Create second instance - should not fail on existing schema
	store2, err := NewSQLiteStorage(dbPath, logger)
	require.NoError(t, err)
	storage2 := store2.(*Storage)
	defer storage2.Close()
}

// Concurrency test

func TestSQLiteConcurrentAccess(t *testing.T) {
	withSQLiteStore(t, func(store types.ApplicationStore) {
		// Create some test data concurrently
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func(idx int) {
				agent := makeTestAgent(uuid.New())
				err := store.CreateAgent(context.Background(), agent)
				assert.NoError(t, err)
				done <- true
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify all agents were created
		agents, err := store.ListAgents(context.Background())
		require.NoError(t, err)
		assert.Len(t, agents, 10)
	})
}
