// Copyright (c) 2024 Lawrence OSS Contributors
// SPDX-License-Identifier: Apache-2.0

package memory

import (
	"context"
	"testing"
	"time"

	"github.com/getlawrence/lawrence-oss/internal/storage/applicationstore"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testAgentID   = uuid.New()
	testGroupID   = "test-group"
	testConfigID  = "test-config"
	testTimestamp = time.Now().UTC()
)

func makeTestAgent() *applicationstore.Agent {
	return &applicationstore.Agent{
		ID:   testAgentID,
		Name: "test-agent",
		Labels: map[string]string{
			"env":     "test",
			"service": "demo",
		},
		Status:       applicationstore.AgentStatusOnline,
		LastSeen:     testTimestamp,
		Version:      "1.0.0",
		Capabilities: []string{"metrics", "logs", "traces"},
		CreatedAt:    testTimestamp,
		UpdatedAt:    testTimestamp,
	}
}

func makeTestGroup() *applicationstore.Group {
	return &applicationstore.Group{
		ID:   testGroupID,
		Name: "test-group",
		Labels: map[string]string{
			"env": "test",
		},
		CreatedAt: testTimestamp,
		UpdatedAt: testTimestamp,
	}
}

func makeTestConfig(agentID *uuid.UUID, groupID *string) *applicationstore.Config {
	return &applicationstore.Config{
		ID:         testConfigID,
		AgentID:    agentID,
		GroupID:    groupID,
		ConfigHash: "abc123",
		Content:    "receivers:\n  otlp:",
		Version:    1,
		CreatedAt:  testTimestamp,
	}
}

func withMemoryStore(f func(store *Store)) {
	f(NewStore())
}

func withPopulatedMemoryStore(f func(store *Store)) {
	store := NewStore()
	agent := makeTestAgent()
	_ = store.CreateAgent(context.Background(), agent)
	f(store)
}

// Agent tests

func TestStoreCreateAgent(t *testing.T) {
	withMemoryStore(func(store *Store) {
		agent := makeTestAgent()
		err := store.CreateAgent(context.Background(), agent)
		require.NoError(t, err)

		// Verify it was created
		retrieved, err := store.GetAgent(context.Background(), agent.ID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, agent.ID, retrieved.ID)
		assert.Equal(t, agent.Name, retrieved.Name)
		assert.Equal(t, agent.Labels, retrieved.Labels)
		assert.Equal(t, agent.Status, retrieved.Status)
	})
}

func TestStoreCreateAgentDuplicate(t *testing.T) {
	withMemoryStore(func(store *Store) {
		agent := makeTestAgent()
		err := store.CreateAgent(context.Background(), agent)
		require.NoError(t, err)

		// Try to create duplicate
		err = store.CreateAgent(context.Background(), agent)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}

func TestStoreGetAgentNotFound(t *testing.T) {
	withMemoryStore(func(store *Store) {
		retrieved, err := store.GetAgent(context.Background(), uuid.New())
		require.NoError(t, err)
		assert.Nil(t, retrieved)
	})
}

func TestStoreGetAgentIsolation(t *testing.T) {
	withPopulatedMemoryStore(func(store *Store) {
		retrieved, err := store.GetAgent(context.Background(), testAgentID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)

		// Modify the retrieved agent
		retrieved.Name = "modified"
		retrieved.Labels["new"] = "label"

		// Get the agent again and verify it wasn't modified
		retrieved2, err := store.GetAgent(context.Background(), testAgentID)
		require.NoError(t, err)
		require.NotNil(t, retrieved2)
		assert.Equal(t, "test-agent", retrieved2.Name)
		assert.NotContains(t, retrieved2.Labels, "new")
	})
}

func TestStoreListAgents(t *testing.T) {
	withMemoryStore(func(store *Store) {
		// Create multiple agents
		agent1 := makeTestAgent()
		agent2 := makeTestAgent()
		agent2.ID = uuid.New()
		agent2.Name = "test-agent-2"

		err := store.CreateAgent(context.Background(), agent1)
		require.NoError(t, err)
		err = store.CreateAgent(context.Background(), agent2)
		require.NoError(t, err)

		// List agents
		agents, err := store.ListAgents(context.Background())
		require.NoError(t, err)
		assert.Len(t, agents, 2)
	})
}

func TestStoreUpdateAgentStatus(t *testing.T) {
	withPopulatedMemoryStore(func(store *Store) {
		err := store.UpdateAgentStatus(context.Background(), testAgentID, applicationstore.AgentStatusOffline)
		require.NoError(t, err)

		// Verify the update
		agent, err := store.GetAgent(context.Background(), testAgentID)
		require.NoError(t, err)
		assert.Equal(t, applicationstore.AgentStatusOffline, agent.Status)
	})
}

func TestStoreUpdateAgentStatusNotFound(t *testing.T) {
	withMemoryStore(func(store *Store) {
		err := store.UpdateAgentStatus(context.Background(), uuid.New(), applicationstore.AgentStatusOffline)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestStoreUpdateAgentLastSeen(t *testing.T) {
	withPopulatedMemoryStore(func(store *Store) {
		newTime := time.Now().Add(time.Hour)
		err := store.UpdateAgentLastSeen(context.Background(), testAgentID, newTime)
		require.NoError(t, err)

		// Verify the update
		agent, err := store.GetAgent(context.Background(), testAgentID)
		require.NoError(t, err)
		assert.Equal(t, newTime.Unix(), agent.LastSeen.Unix())
	})
}

func TestStoreDeleteAgent(t *testing.T) {
	withPopulatedMemoryStore(func(store *Store) {
		err := store.DeleteAgent(context.Background(), testAgentID)
		require.NoError(t, err)

		// Verify it was deleted
		agent, err := store.GetAgent(context.Background(), testAgentID)
		require.NoError(t, err)
		assert.Nil(t, agent)
	})
}

func TestStoreDeleteAgentNotFound(t *testing.T) {
	withMemoryStore(func(store *Store) {
		err := store.DeleteAgent(context.Background(), uuid.New())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

// Group tests

func TestStoreCreateGroup(t *testing.T) {
	withMemoryStore(func(store *Store) {
		group := makeTestGroup()
		err := store.CreateGroup(context.Background(), group)
		require.NoError(t, err)

		// Verify it was created
		retrieved, err := store.GetGroup(context.Background(), group.ID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, group.ID, retrieved.ID)
		assert.Equal(t, group.Name, retrieved.Name)
		assert.Equal(t, group.Labels, retrieved.Labels)
	})
}

func TestStoreCreateGroupDuplicate(t *testing.T) {
	withMemoryStore(func(store *Store) {
		group := makeTestGroup()
		err := store.CreateGroup(context.Background(), group)
		require.NoError(t, err)

		// Try to create duplicate
		err = store.CreateGroup(context.Background(), group)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}

func TestStoreGetGroupNotFound(t *testing.T) {
	withMemoryStore(func(store *Store) {
		retrieved, err := store.GetGroup(context.Background(), "nonexistent")
		require.NoError(t, err)
		assert.Nil(t, retrieved)
	})
}

func TestStoreListGroups(t *testing.T) {
	withMemoryStore(func(store *Store) {
		group1 := makeTestGroup()
		group2 := makeTestGroup()
		group2.ID = "test-group-2"
		group2.Name = "test-group-2"

		err := store.CreateGroup(context.Background(), group1)
		require.NoError(t, err)
		err = store.CreateGroup(context.Background(), group2)
		require.NoError(t, err)

		groups, err := store.ListGroups(context.Background())
		require.NoError(t, err)
		assert.Len(t, groups, 2)
	})
}

func TestStoreDeleteGroup(t *testing.T) {
	withMemoryStore(func(store *Store) {
		group := makeTestGroup()
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

func TestStoreCreateConfig(t *testing.T) {
	withMemoryStore(func(store *Store) {
		agentID := testAgentID
		config := makeTestConfig(&agentID, nil)
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

func TestStoreGetLatestConfigForAgent(t *testing.T) {
	withMemoryStore(func(store *Store) {
		agentID := testAgentID

		// Create multiple configs for the same agent
		config1 := makeTestConfig(&agentID, nil)
		config1.ID = "config-1"
		config1.Version = 1
		config1.CreatedAt = time.Now().Add(-time.Hour)

		config2 := makeTestConfig(&agentID, nil)
		config2.ID = "config-2"
		config2.Version = 2
		config2.CreatedAt = time.Now()

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

func TestStoreGetLatestConfigForGroup(t *testing.T) {
	withMemoryStore(func(store *Store) {
		groupID := testGroupID

		// Create multiple configs for the same group
		config1 := makeTestConfig(nil, &groupID)
		config1.ID = "config-1"
		config1.Version = 1
		config1.CreatedAt = time.Now().Add(-time.Hour)

		config2 := makeTestConfig(nil, &groupID)
		config2.ID = "config-2"
		config2.Version = 2
		config2.CreatedAt = time.Now()

		err := store.CreateConfig(context.Background(), config1)
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

func TestStoreListConfigsWithFilter(t *testing.T) {
	withMemoryStore(func(store *Store) {
		agentID1 := testAgentID
		agentID2 := uuid.New()
		groupID := testGroupID

		config1 := makeTestConfig(&agentID1, nil)
		config1.ID = "config-1"

		config2 := makeTestConfig(&agentID2, nil)
		config2.ID = "config-2"

		config3 := makeTestConfig(nil, &groupID)
		config3.ID = "config-3"

		err := store.CreateConfig(context.Background(), config1)
		require.NoError(t, err)
		err = store.CreateConfig(context.Background(), config2)
		require.NoError(t, err)
		err = store.CreateConfig(context.Background(), config3)
		require.NoError(t, err)

		// Filter by agent ID
		configs, err := store.ListConfigs(context.Background(), applicationstore.ConfigFilter{
			AgentID: &agentID1,
		})
		require.NoError(t, err)
		assert.Len(t, configs, 1)
		assert.Equal(t, config1.ID, configs[0].ID)

		// Filter by group ID
		configs, err = store.ListConfigs(context.Background(), applicationstore.ConfigFilter{
			GroupID: &groupID,
		})
		require.NoError(t, err)
		assert.Len(t, configs, 1)
		assert.Equal(t, config3.ID, configs[0].ID)

		// No filter
		configs, err = store.ListConfigs(context.Background(), applicationstore.ConfigFilter{})
		require.NoError(t, err)
		assert.Len(t, configs, 3)
	})
}

func TestStoreListConfigsWithLimit(t *testing.T) {
	withMemoryStore(func(store *Store) {
		agentID := testAgentID

		for i := 0; i < 5; i++ {
			config := makeTestConfig(&agentID, nil)
			config.ID = uuid.New().String()
			err := store.CreateConfig(context.Background(), config)
			require.NoError(t, err)
		}

		// List with limit
		configs, err := store.ListConfigs(context.Background(), applicationstore.ConfigFilter{
			Limit: 3,
		})
		require.NoError(t, err)
		assert.Len(t, configs, 3)
	})
}

// Purge test

func TestStorePurge(t *testing.T) {
	withMemoryStore(func(store *Store) {
		// Add some data
		agent := makeTestAgent()
		group := makeTestGroup()
		agentID := testAgentID
		config := makeTestConfig(&agentID, nil)

		err := store.CreateAgent(context.Background(), agent)
		require.NoError(t, err)
		err = store.CreateGroup(context.Background(), group)
		require.NoError(t, err)
		err = store.CreateConfig(context.Background(), config)
		require.NoError(t, err)

		// Purge
		store.purge(context.Background())

		// Verify all data was removed
		agents, err := store.ListAgents(context.Background())
		require.NoError(t, err)
		assert.Empty(t, agents)

		groups, err := store.ListGroups(context.Background())
		require.NoError(t, err)
		assert.Empty(t, groups)

		configs, err := store.ListConfigs(context.Background(), applicationstore.ConfigFilter{})
		require.NoError(t, err)
		assert.Empty(t, configs)
	})
}
