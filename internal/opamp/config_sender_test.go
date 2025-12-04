// Copyright (c) 2024 Lawrence OSS Contributors
// SPDX-License-Identifier: Apache-2.0

package opamp

import (
	"context"
	"net"
	"testing"

	"github.com/google/uuid"
	"github.com/open-telemetry/opamp-go/protobufs"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// mockConnection is a simple mock implementation of types.Connection for testing
type mockConnection struct{}

func (m *mockConnection) Send(ctx context.Context, msg *protobufs.ServerToAgent) error {
	// No-op for testing
	return nil
}

func (m *mockConnection) Connection() net.Conn {
	// Return a dummy connection
	conn, _ := net.Pipe()
	return conn
}

func (m *mockConnection) Disconnect() error {
	// No-op for testing
	return nil
}

// TestSendConfigToAgentsInGroup_FindsCorrectAgents tests that
// SendConfigToAgentsInGroup correctly identifies agents in the target group
// and attempts to send config to them (regardless of whether they succeed)
func TestSendConfigToAgentsInGroup_FindsCorrectAgents(t *testing.T) {
	logger := zap.NewNop()
	agents := NewAgents(logger)
	configSender := NewConfigSender(agents, logger)

	// Create group IDs
	groupID := "test-group-1"
	otherGroupID := "test-group-2"

	// Create mock connections for agents
	mockConn1 := &mockConnection{}
	mockConn2 := &mockConnection{}
	mockConn3 := &mockConnection{}
	mockConn4 := &mockConnection{}

	// Create agents in the target group using NewAgent to properly initialize connection
	agent1ID := uuid.New()
	agent1 := NewAgent(agent1ID, mockConn1)
	agent1.GroupID = &groupID
	agent1.Status = &protobufs.AgentToServer{
		Capabilities: uint64(protobufs.AgentCapabilities_AgentCapabilities_AcceptsRemoteConfig),
	}

	agent2ID := uuid.New()
	agent2 := NewAgent(agent2ID, mockConn2)
	agent2.GroupID = &groupID
	agent2.Status = &protobufs.AgentToServer{
		Capabilities: uint64(protobufs.AgentCapabilities_AgentCapabilities_AcceptsRemoteConfig),
	}

	// Create an agent in a different group (should not receive config)
	agent3ID := uuid.New()
	agent3 := NewAgent(agent3ID, mockConn3)
	agent3.GroupID = &otherGroupID
	agent3.Status = &protobufs.AgentToServer{
		Capabilities: uint64(protobufs.AgentCapabilities_AgentCapabilities_AcceptsRemoteConfig),
	}

	// Create an agent with no group (should not receive config)
	agent4ID := uuid.New()
	agent4 := NewAgent(agent4ID, mockConn4)
	agent4.GroupID = nil
	agent4.Status = &protobufs.AgentToServer{
		Capabilities: uint64(protobufs.AgentCapabilities_AgentCapabilities_AcceptsRemoteConfig),
	}

	// Add agents to the agents collection
	agents.agentsById[agent1ID] = agent1
	agents.agentsById[agent2ID] = agent2
	agents.agentsById[agent3ID] = agent3
	agents.agentsById[agent4ID] = agent4

	// Set initial config and remote config so that setting the same config again will notify immediately
	// This allows the test to complete without waiting for agent status updates
	initialConfig := "initial-config"
	agent1.CustomInstanceConfig = initialConfig
	agent2.CustomInstanceConfig = initialConfig

	// Set remote config to match so calcRemoteConfig returns false (no change)
	// This triggers immediate notification
	configMap := &protobufs.AgentConfigMap{
		ConfigMap: map[string]*protobufs.AgentConfigFile{
			"": {Body: []byte(initialConfig)},
		},
	}
	agent1.SetRemoteConfig(&protobufs.AgentRemoteConfig{Config: configMap})
	agent2.SetRemoteConfig(&protobufs.AgentRemoteConfig{Config: configMap})

	// Send the same config again - this will trigger immediate notification
	// because the config hasn't changed
	configContent := initialConfig
	updatedAgents, errors := configSender.SendConfigToAgentsInGroup(groupID, configContent)

	// KEY ASSERTIONS:
	// 1. Only agents in the target group should be attempted
	assert.Len(t, updatedAgents, 2, "Should update 2 agents in the group")
	assert.Contains(t, updatedAgents, agent1ID, "Agent 1 should be updated")
	assert.Contains(t, updatedAgents, agent2ID, "Agent 2 should be updated")
	assert.NotContains(t, updatedAgents, agent3ID, "Agent 3 (different group) should NOT be updated")
	assert.NotContains(t, updatedAgents, agent4ID, "Agent 4 (no group) should NOT be updated")
	assert.Empty(t, errors, "Should have no errors")
}

// TestSendConfigToAgentsInGroup_EmptyGroup tests behavior with empty group
func TestSendConfigToAgentsInGroup_EmptyGroup(t *testing.T) {
	logger := zap.NewNop()
	agents := NewAgents(logger)
	configSender := NewConfigSender(agents, logger)

	// Send config to non-existent group
	configContent := "test-config-content"
	updatedAgents, errors := configSender.SendConfigToAgentsInGroup("non-existent-group", configContent)

	// Should return empty results, no errors
	assert.Empty(t, updatedAgents, "Should have no updated agents for non-existent group")
	assert.Empty(t, errors, "Should have no errors for empty group")
}

// TestSendConfigToAgentsInGroup_AgentWithoutCapability tests that agents
// without remote config capability are handled correctly (error returned)
func TestSendConfigToAgentsInGroup_AgentWithoutCapability(t *testing.T) {
	logger := zap.NewNop()
	agents := NewAgents(logger)
	configSender := NewConfigSender(agents, logger)

	groupID := "test-group-1"

	// Create mock connections
	mockConn1 := &mockConnection{}
	mockConn2 := &mockConnection{}

	// Create agent without remote config capability
	agent1ID := uuid.New()
	agent1 := NewAgent(agent1ID, mockConn1)
	agent1.GroupID = &groupID
	agent1.Status = &protobufs.AgentToServer{
		Capabilities: 0, // No capabilities
	}

	// Create agent with capability
	agent2ID := uuid.New()
	agent2 := NewAgent(agent2ID, mockConn2)
	agent2.GroupID = &groupID
	agent2.Status = &protobufs.AgentToServer{
		Capabilities: uint64(protobufs.AgentCapabilities_AgentCapabilities_AcceptsRemoteConfig),
	}

	agents.agentsById[agent1ID] = agent1
	agents.agentsById[agent2ID] = agent2

	// Set initial config and remote config for agent2 so it notifies immediately
	initialConfig := "initial-config"
	agent2.CustomInstanceConfig = initialConfig
	configMap := &protobufs.AgentConfigMap{
		ConfigMap: map[string]*protobufs.AgentConfigFile{
			"": {Body: []byte(initialConfig)},
		},
	}
	agent2.SetRemoteConfig(&protobufs.AgentRemoteConfig{Config: configMap})

	// Send config to group
	configContent := initialConfig
	updatedAgents, errors := configSender.SendConfigToAgentsInGroup(groupID, configContent)

	// KEY ASSERTIONS:
	// 1. Only agent with capability should be updated
	assert.Len(t, updatedAgents, 1, "Should update only 1 agent (the one with capability)")
	assert.Contains(t, updatedAgents, agent2ID, "Agent 2 (with capability) should be updated")
	assert.NotContains(t, updatedAgents, agent1ID, "Agent 1 (without capability) should NOT be updated")
	// 2. Should have one error for the agent without capability
	assert.Len(t, errors, 1, "Should have 1 error for agent without capability")
	assert.Contains(t, errors[0].Error(), agent1ID.String(), "Error should mention agent 1")
}

// TestSendConfigToAgentsInGroup_PartialFailure tests that partial failures
// are handled correctly - some agents succeed, some fail
func TestSendConfigToAgentsInGroup_PartialFailure(t *testing.T) {
	logger := zap.NewNop()
	agents := NewAgents(logger)
	configSender := NewConfigSender(agents, logger)

	groupID := "test-group-1"

	// Create mock connections
	mockConn1 := &mockConnection{}
	mockConn2 := &mockConnection{}
	mockConn3 := &mockConnection{}

	// Create multiple agents - some will succeed, some will fail
	agent1ID := uuid.New()
	agent1 := NewAgent(agent1ID, mockConn1)
	agent1.GroupID = &groupID
	agent1.Status = &protobufs.AgentToServer{
		Capabilities: uint64(protobufs.AgentCapabilities_AgentCapabilities_AcceptsRemoteConfig),
	}

	agent2ID := uuid.New()
	agent2 := NewAgent(agent2ID, mockConn2)
	agent2.GroupID = &groupID
	agent2.Status = &protobufs.AgentToServer{
		Capabilities: 0, // No capability - will fail
	}

	agent3ID := uuid.New()
	agent3 := NewAgent(agent3ID, mockConn3)
	agent3.GroupID = &groupID
	agent3.Status = &protobufs.AgentToServer{
		Capabilities: uint64(protobufs.AgentCapabilities_AgentCapabilities_AcceptsRemoteConfig),
	}

	agents.agentsById[agent1ID] = agent1
	agents.agentsById[agent2ID] = agent2
	agents.agentsById[agent3ID] = agent3

	// Set initial config and remote config for successful agents so they notify immediately
	initialConfig := "initial-config"
	agent1.CustomInstanceConfig = initialConfig
	agent3.CustomInstanceConfig = initialConfig
	configMap := &protobufs.AgentConfigMap{
		ConfigMap: map[string]*protobufs.AgentConfigFile{
			"": {Body: []byte(initialConfig)},
		},
	}
	agent1.SetRemoteConfig(&protobufs.AgentRemoteConfig{Config: configMap})
	agent3.SetRemoteConfig(&protobufs.AgentRemoteConfig{Config: configMap})

	// Send config to group
	configContent := initialConfig
	updatedAgents, errors := configSender.SendConfigToAgentsInGroup(groupID, configContent)

	// KEY ASSERTIONS:
	// 1. Should have 2 successful updates
	assert.Len(t, updatedAgents, 2, "Should update 2 agents successfully")
	assert.Contains(t, updatedAgents, agent1ID, "Agent 1 should be updated")
	assert.Contains(t, updatedAgents, agent3ID, "Agent 3 should be updated")
	// 2. Should have 1 error
	assert.Len(t, errors, 1, "Should have 1 error for agent without capability")
	assert.Contains(t, errors[0].Error(), agent2ID.String(), "Error should mention agent 2")
}
