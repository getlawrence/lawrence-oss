// Copyright (c) 2024 Lawrence OSS Contributors
// SPDX-License-Identifier: Apache-2.0

package testutils

import (
	"time"

	"github.com/getlawrence/lawrence-oss/internal/services"
	"github.com/google/uuid"
)

// Test fixture data

// MakeTestAgent creates a test agent with the given ID
func MakeTestAgent(id uuid.UUID) *services.Agent {
	return &services.Agent{
		ID:   id,
		Name: "test-agent",
		Labels: map[string]string{
			"env":     "test",
			"service": "demo",
		},
		Status:       services.AgentStatusOnline,
		LastSeen:     time.Now().UTC(),
		Version:      "1.0.0",
		Capabilities: []string{"metrics", "logs", "traces"},
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
}

// MakeTestAgentWithStatus creates a test agent with the given ID and status
func MakeTestAgentWithStatus(id uuid.UUID, status services.AgentStatus) *services.Agent {
	agent := MakeTestAgent(id)
	agent.Status = status
	return agent
}

// MakeTestGroup creates a test group with the given ID
func MakeTestGroup(id string) *services.Group {
	return &services.Group{
		ID:   id,
		Name: "test-group",
		Labels: map[string]string{
			"env": "test",
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}

// MakeTestConfig creates a test config
func MakeTestConfig(id string, agentID *uuid.UUID, groupID *string) *services.Config {
	return &services.Config{
		ID:         id,
		AgentID:    agentID,
		GroupID:    groupID,
		ConfigHash: "abc123",
		Content:    "receivers:\n  otlp:\n    protocols:\n      grpc:\n      http:\n",
		Version:    1,
		CreatedAt:  time.Now().UTC(),
	}
}

// MakeTestConfigWithVersion creates a test config with a specific version
func MakeTestConfigWithVersion(id string, agentID *uuid.UUID, groupID *string, version int) *services.Config {
	config := MakeTestConfig(id, agentID, groupID)
	config.Version = version
	return config
}

// StringPtr returns a pointer to the given string
func StringPtr(s string) *string {
	return &s
}

// UUIDPtr returns a pointer to the given UUID
func UUIDPtr(id uuid.UUID) *uuid.UUID {
	return &id
}
