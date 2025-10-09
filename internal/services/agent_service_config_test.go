// Copyright (c) 2024 Lawrence OSS Contributors
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/getlawrence/lawrence-oss/internal/storage/applicationstore/memory"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockOpAMPSender is a mock implementation of OpAMPConfigSender
type MockOpAMPSender struct {
	mock.Mock
}

func (m *MockOpAMPSender) SendConfigToAgent(agentId uuid.UUID, configContent string) error {
	args := m.Called(agentId, configContent)
	return args.Error(0)
}

func TestSendConfigToAgent_Success(t *testing.T) {
	// Setup
	store := memory.NewStore()
	mockOpAMP := new(MockOpAMPSender)
	logger := zap.NewNop()
	service := NewAgentService(store, logger)
	service.(*AgentServiceImpl).SetConfigSender(mockOpAMP)

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

	// Mock expectations
	mockOpAMP.On("SendConfigToAgent", agentID, configContent).Return(nil)

	// Execute
	err = service.SendConfigToAgent(context.Background(), agentID, configContent)

	// Assert
	require.NoError(t, err)
	mockOpAMP.AssertExpectations(t)

	// Verify config was stored
	config, err := service.GetLatestConfigForAgent(context.Background(), agentID)
	require.NoError(t, err)
	require.NotNil(t, config)
	assert.Equal(t, configContent, config.Content)
	assert.Equal(t, 1, config.Version)
}

func TestSendConfigToAgent_AgentNotFound(t *testing.T) {
	store := memory.NewStore()
	mockOpAMP := new(MockOpAMPSender)
	logger := zap.NewNop()
	service := NewAgentService(store, logger)
	service.(*AgentServiceImpl).SetConfigSender(mockOpAMP)

	agentID := uuid.New()

	err := service.SendConfigToAgent(context.Background(), agentID, "test-config")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent not found")
	mockOpAMP.AssertNotCalled(t, "SendConfigToAgent")
}

func TestSendConfigToAgent_AgentOffline(t *testing.T) {
	store := memory.NewStore()
	mockOpAMP := new(MockOpAMPSender)
	logger := zap.NewNop()
	service := NewAgentService(store, logger)
	service.(*AgentServiceImpl).SetConfigSender(mockOpAMP)

	agentID := uuid.New()

	// Create an offline agent with remote config capability
	agent := &Agent{
		ID:           agentID,
		Name:         "test-agent",
		Status:       AgentStatusOffline,
		Capabilities: []string{"accepts_remote_config"},
		Labels:       map[string]string{},
		LastSeen:     time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := service.CreateAgent(context.Background(), agent)
	require.NoError(t, err)

	// Mock OpAMP should accept the config even for offline agent (supervisor can queue it)
	mockOpAMP.On("SendConfigToAgent", agentID, "test-config").Return(nil)

	err = service.SendConfigToAgent(context.Background(), agentID, "test-config")

	// Should succeed - supervisors can queue config for offline agents
	require.NoError(t, err)
	mockOpAMP.AssertCalled(t, "SendConfigToAgent", agentID, "test-config")
}

func TestSendConfigToAgent_NoCapability(t *testing.T) {
	store := memory.NewStore()
	mockOpAMP := new(MockOpAMPSender)
	logger := zap.NewNop()
	service := NewAgentService(store, logger)
	service.(*AgentServiceImpl).SetConfigSender(mockOpAMP)

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

	err = service.SendConfigToAgent(context.Background(), agentID, "test-config")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not support remote config")
	mockOpAMP.AssertNotCalled(t, "SendConfigToAgent")
}

func TestSendConfigToAgent_OpAMPSendFails(t *testing.T) {
	store := memory.NewStore()
	mockOpAMP := new(MockOpAMPSender)
	logger := zap.NewNop()
	service := NewAgentService(store, logger)
	service.(*AgentServiceImpl).SetConfigSender(mockOpAMP)

	agentID := uuid.New()
	configContent := "test-config"

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

	// Mock OpAMP send failure
	mockOpAMP.On("SendConfigToAgent", agentID, configContent).Return(fmt.Errorf("timeout waiting for agent"))

	err = service.SendConfigToAgent(context.Background(), agentID, configContent)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "timeout waiting for agent")
	mockOpAMP.AssertExpectations(t)

	// Verify config was NOT stored (since send failed)
	config, err := service.GetLatestConfigForAgent(context.Background(), agentID)
	require.NoError(t, err)
	assert.Nil(t, config)
}

func TestSendConfigToAgent_Versioning(t *testing.T) {
	store := memory.NewStore()
	mockOpAMP := new(MockOpAMPSender)
	logger := zap.NewNop()
	service := NewAgentService(store, logger)
	service.(*AgentServiceImpl).SetConfigSender(mockOpAMP)

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

	// Send first config
	mockOpAMP.On("SendConfigToAgent", agentID, "config-v1").Return(nil)
	err = service.SendConfigToAgent(context.Background(), agentID, "config-v1")
	require.NoError(t, err)

	config1, err := service.GetLatestConfigForAgent(context.Background(), agentID)
	require.NoError(t, err)
	assert.Equal(t, 1, config1.Version)

	// Send second config
	mockOpAMP.On("SendConfigToAgent", agentID, "config-v2").Return(nil)
	err = service.SendConfigToAgent(context.Background(), agentID, "config-v2")
	require.NoError(t, err)

	config2, err := service.GetLatestConfigForAgent(context.Background(), agentID)
	require.NoError(t, err)
	assert.Equal(t, 2, config2.Version)
	assert.Equal(t, "config-v2", config2.Content)

	mockOpAMP.AssertExpectations(t)
}
