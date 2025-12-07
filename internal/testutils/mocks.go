// Copyright (c) 2024 Lawrence OSS Contributors
// SPDX-License-Identifier: Apache-2.0

package testutils

import (
	"context"
	"sync"
	"time"

	"github.com/getlawrence/lawrence-oss/internal/services"
	"github.com/google/uuid"
)

// MockAgentService is a mock implementation of services.AgentService for testing
type MockAgentService struct {
	mu sync.RWMutex

	agents  map[uuid.UUID]*services.Agent
	groups  map[string]*services.Group
	configs map[string]*services.Config

	// Error flags for testing error cases
	CreateAgentErr                error
	GetAgentErr                   error
	ListAgentsErr                 error
	UpdateAgentStatusErr          error
	UpdateAgentLastSeenErr        error
	UpdateAgentEffectiveConfigErr error
	DeleteAgentErr                error
	CreateGroupErr                error
	GetGroupErr                   error
	GetGroupByNameErr             error
	ListGroupsErr                 error
	DeleteGroupErr                error
	CreateConfigErr               error
	GetConfigErr                  error
	GetLatestConfigForAgentErr    error
	GetLatestConfigForGroupErr    error
	ListConfigsErr                error
	StoreConfigForAgentErr        error
}

// NewMockAgentService creates a new mock agent service
func NewMockAgentService() *MockAgentService {
	return &MockAgentService{
		agents:  make(map[uuid.UUID]*services.Agent),
		groups:  make(map[string]*services.Group),
		configs: make(map[string]*services.Config),
	}
}

// CreateAgent implements services.AgentService
func (m *MockAgentService) CreateAgent(ctx context.Context, agent *services.Agent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.CreateAgentErr != nil {
		return m.CreateAgentErr
	}

	agentCopy := *agent
	m.agents[agent.ID] = &agentCopy
	return nil
}

// GetAgent implements services.AgentService
func (m *MockAgentService) GetAgent(ctx context.Context, id uuid.UUID) (*services.Agent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.GetAgentErr != nil {
		return nil, m.GetAgentErr
	}

	agent, exists := m.agents[id]
	if !exists {
		return nil, nil
	}

	agentCopy := *agent
	return &agentCopy, nil
}

// ListAgents implements services.AgentService
func (m *MockAgentService) ListAgents(ctx context.Context) ([]*services.Agent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.ListAgentsErr != nil {
		return nil, m.ListAgentsErr
	}

	agents := make([]*services.Agent, 0, len(m.agents))
	for _, agent := range m.agents {
		agentCopy := *agent
		agents = append(agents, &agentCopy)
	}

	return agents, nil
}

// UpdateAgentStatus implements services.AgentService
func (m *MockAgentService) UpdateAgentStatus(ctx context.Context, id uuid.UUID, status services.AgentStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.UpdateAgentStatusErr != nil {
		return m.UpdateAgentStatusErr
	}

	agent, exists := m.agents[id]
	if !exists {
		return nil
	}

	agent.Status = status
	agent.UpdatedAt = time.Now()
	return nil
}

// UpdateAgentLastSeen implements services.AgentService
func (m *MockAgentService) UpdateAgentLastSeen(ctx context.Context, id uuid.UUID, lastSeen time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.UpdateAgentLastSeenErr != nil {
		return m.UpdateAgentLastSeenErr
	}

	agent, exists := m.agents[id]
	if !exists {
		return nil
	}

	agent.LastSeen = lastSeen
	agent.UpdatedAt = time.Now()
	return nil
}

// UpdateAgentEffectiveConfig implements services.AgentService
func (m *MockAgentService) UpdateAgentEffectiveConfig(ctx context.Context, id uuid.UUID, effectiveConfig string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.UpdateAgentEffectiveConfigErr != nil {
		return m.UpdateAgentEffectiveConfigErr
	}

	agent, exists := m.agents[id]
	if !exists {
		return nil
	}

	agent.EffectiveConfig = effectiveConfig
	agent.UpdatedAt = time.Now()
	return nil
}

// DeleteAgent implements services.AgentService
func (m *MockAgentService) DeleteAgent(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.DeleteAgentErr != nil {
		return m.DeleteAgentErr
	}

	delete(m.agents, id)
	return nil
}

// CreateGroup implements services.AgentService
func (m *MockAgentService) CreateGroup(ctx context.Context, group *services.Group) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.CreateGroupErr != nil {
		return m.CreateGroupErr
	}

	groupCopy := *group
	m.groups[group.ID] = &groupCopy
	return nil
}

// GetGroup implements services.AgentService
func (m *MockAgentService) GetGroup(ctx context.Context, id string) (*services.Group, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.GetGroupErr != nil {
		return nil, m.GetGroupErr
	}

	group, exists := m.groups[id]
	if !exists {
		return nil, nil
	}

	groupCopy := *group
	return &groupCopy, nil
}

// GetGroupByName implements services.AgentService
func (m *MockAgentService) GetGroupByName(ctx context.Context, name string) (*services.Group, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.GetGroupByNameErr != nil {
		return nil, m.GetGroupByNameErr
	}

	for _, group := range m.groups {
		if group.Name == name {
			groupCopy := *group
			return &groupCopy, nil
		}
	}

	return nil, nil
}

// ListGroups implements services.AgentService
func (m *MockAgentService) ListGroups(ctx context.Context) ([]*services.Group, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.ListGroupsErr != nil {
		return nil, m.ListGroupsErr
	}

	groups := make([]*services.Group, 0, len(m.groups))
	for _, group := range m.groups {
		groupCopy := *group
		groups = append(groups, &groupCopy)
	}

	return groups, nil
}

// DeleteGroup implements services.AgentService
func (m *MockAgentService) DeleteGroup(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.DeleteGroupErr != nil {
		return m.DeleteGroupErr
	}

	delete(m.groups, id)
	return nil
}

// CreateConfig implements services.AgentService
func (m *MockAgentService) CreateConfig(ctx context.Context, config *services.Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.CreateConfigErr != nil {
		return m.CreateConfigErr
	}

	configCopy := *config
	m.configs[config.ID] = &configCopy
	return nil
}

// GetConfig implements services.AgentService
func (m *MockAgentService) GetConfig(ctx context.Context, id string) (*services.Config, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.GetConfigErr != nil {
		return nil, m.GetConfigErr
	}

	config, exists := m.configs[id]
	if !exists {
		return nil, nil
	}

	configCopy := *config
	return &configCopy, nil
}

// GetLatestConfigForAgent implements services.AgentService
func (m *MockAgentService) GetLatestConfigForAgent(ctx context.Context, agentID uuid.UUID) (*services.Config, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.GetLatestConfigForAgentErr != nil {
		return nil, m.GetLatestConfigForAgentErr
	}

	var latestConfig *services.Config
	for _, config := range m.configs {
		if config.AgentID != nil && *config.AgentID == agentID {
			if latestConfig == nil || config.Version > latestConfig.Version {
				latestConfig = config
			}
		}
	}

	if latestConfig == nil {
		return nil, nil
	}

	configCopy := *latestConfig
	return &configCopy, nil
}

// GetLatestConfigForGroup implements services.AgentService
func (m *MockAgentService) GetLatestConfigForGroup(ctx context.Context, groupID string) (*services.Config, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.GetLatestConfigForGroupErr != nil {
		return nil, m.GetLatestConfigForGroupErr
	}

	var latestConfig *services.Config
	for _, config := range m.configs {
		if config.GroupID != nil && *config.GroupID == groupID {
			if latestConfig == nil || config.Version > latestConfig.Version {
				latestConfig = config
			}
		}
	}

	if latestConfig == nil {
		return nil, nil
	}

	configCopy := *latestConfig
	return &configCopy, nil
}

// ListConfigs implements services.AgentService
func (m *MockAgentService) ListConfigs(ctx context.Context, filter services.ConfigFilter) (*services.ListConfigsResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.ListConfigsErr != nil {
		return nil, m.ListConfigsErr
	}

	configs := make([]*services.Config, 0)
	for _, config := range m.configs {
		// Apply filters
		if filter.AgentID != nil && (config.AgentID == nil || *config.AgentID != *filter.AgentID) {
			continue
		}
		if filter.GroupID != nil && (config.GroupID == nil || *config.GroupID != *filter.GroupID) {
			continue
		}

		configCopy := *config
		configs = append(configs, &configCopy)
	}

	// Get total count before applying pagination
	totalCount := len(configs)

	// Apply offset
	if filter.Offset > 0 && filter.Offset < len(configs) {
		configs = configs[filter.Offset:]
	} else if filter.Offset >= len(configs) {
		configs = []*services.Config{}
	}

	// Apply limit
	if filter.Limit > 0 && len(configs) > filter.Limit {
		configs = configs[:filter.Limit]
	}

	return &services.ListConfigsResult{
		Configs:    configs,
		TotalCount: totalCount,
	}, nil
}

// StoreConfigForAgent implements services.AgentService
func (m *MockAgentService) StoreConfigForAgent(ctx context.Context, agentID uuid.UUID, content string) (*services.Config, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.StoreConfigForAgentErr != nil {
		return nil, m.StoreConfigForAgentErr
	}

	// Get current version
	version := 1
	for _, config := range m.configs {
		if config.AgentID != nil && *config.AgentID == agentID {
			if config.Version >= version {
				version = config.Version + 1
			}
		}
	}

	// Create and store a mock config
	config := &services.Config{
		ID:        uuid.New().String(),
		AgentID:   &agentID,
		Content:   content,
		Version:   version,
		CreatedAt: time.Now(),
	}
	m.configs[config.ID] = config

	return config, nil
}
