// Copyright (c) 2024 Lawrence OSS Contributors
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"
	"testing"
	"time"

	"github.com/getlawrence/lawrence-oss/internal/storage/applicationstore/memory"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestStoreConfigForAgent_Success(t *testing.T) {
	// Setup
	store := memory.NewStore()
	logger := zap.NewNop()
	service := NewAgentService(store, logger)

	agentID := uuid.New()
	configContent := "receivers:\n  otlp:\n    protocols:\n      grpc:"

	// Create an online agent with capability
	agent := &Agent{
		ID:           agentID,
		Name:         "test-agent",
		Status:       AgentStatusOnline,
		Capabilities: []string{"accepts_remote_config"},
		Labels:       map[string]string{},
		LastSeen:     time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := service.CreateAgent(context.Background(), agent)
	require.NoError(t, err)

	// Execute
	config, err := service.StoreConfigForAgent(context.Background(), agentID, configContent)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, config)
	assert.Equal(t, configContent, config.Content)
	assert.Equal(t, 1, config.Version)

	// Verify config was stored
	storedConfig, err := service.GetLatestConfigForAgent(context.Background(), agentID)
	require.NoError(t, err)
	require.NotNil(t, storedConfig)
	assert.Equal(t, config.ID, storedConfig.ID)
}

func TestStoreConfigForAgent_AgentNotFound(t *testing.T) {
	store := memory.NewStore()
	logger := zap.NewNop()
	service := NewAgentService(store, logger)

	agentID := uuid.New()

	_, err := service.StoreConfigForAgent(context.Background(), agentID, "test-config")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent not found")
}

func TestStoreConfigForAgent_NoCapability(t *testing.T) {
	store := memory.NewStore()
	logger := zap.NewNop()
	service := NewAgentService(store, logger)

	agentID := uuid.New()

	// Create an online agent without remote config capability
	agent := &Agent{
		ID:           agentID,
		Name:         "test-agent",
		Status:       AgentStatusOnline,
		Capabilities: []string{"reports_status"}, // No remote config
		Labels:       map[string]string{},
		LastSeen:     time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := service.CreateAgent(context.Background(), agent)
	require.NoError(t, err)

	_, err = service.StoreConfigForAgent(context.Background(), agentID, "test-config")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not support remote config")
}

func TestStoreConfigForAgent_Versioning(t *testing.T) {
	store := memory.NewStore()
	logger := zap.NewNop()
	service := NewAgentService(store, logger)

	agentID := uuid.New()

	// Create an online agent with capability
	agent := &Agent{
		ID:           agentID,
		Name:         "test-agent",
		Status:       AgentStatusOnline,
		Capabilities: []string{"accepts_remote_config"},
		Labels:       map[string]string{},
		LastSeen:     time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := service.CreateAgent(context.Background(), agent)
	require.NoError(t, err)

	// Store first config
	config1, err := service.StoreConfigForAgent(context.Background(), agentID, "config-v1")
	require.NoError(t, err)
	assert.Equal(t, 1, config1.Version)

	// Store second config
	config2, err := service.StoreConfigForAgent(context.Background(), agentID, "config-v2")
	require.NoError(t, err)
	assert.Equal(t, 2, config2.Version)
	assert.Equal(t, "config-v2", config2.Content)
}
