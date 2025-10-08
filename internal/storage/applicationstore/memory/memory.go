// Copyright (c) 2024 Lawrence OSS Contributors
// SPDX-License-Identifier: Apache-2.0

package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/getlawrence/lawrence-oss/internal/storage/applicationstore/types"
	"github.com/google/uuid"
)

// Store is an in-memory implementation of ApplicationStore
type Store struct {
	mu      sync.RWMutex
	agents  map[uuid.UUID]*types.Agent
	groups  map[string]*types.Group
	configs map[string]*types.Config
}

// NewStore creates a new in-memory store
func NewStore() *Store {
	return &Store{
		agents:  make(map[uuid.UUID]*types.Agent),
		groups:  make(map[string]*types.Group),
		configs: make(map[string]*types.Config),
	}
}

// Agent management

func (s *Store) CreateAgent(ctx context.Context, agent *types.Agent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.agents[agent.ID]; exists {
		return fmt.Errorf("agent already exists: %s", agent.ID)
	}

	// Deep copy to prevent external modifications
	agentCopy := *agent
	if agent.Labels != nil {
		agentCopy.Labels = make(map[string]string, len(agent.Labels))
		for k, v := range agent.Labels {
			agentCopy.Labels[k] = v
		}
	}
	if agent.Capabilities != nil {
		agentCopy.Capabilities = make([]string, len(agent.Capabilities))
		copy(agentCopy.Capabilities, agent.Capabilities)
	}

	s.agents[agent.ID] = &agentCopy
	return nil
}

func (s *Store) GetAgent(ctx context.Context, id uuid.UUID) (*types.Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	agent, exists := s.agents[id]
	if !exists {
		return nil, nil
	}

	// Deep copy to prevent external modifications
	agentCopy := *agent
	if agent.Labels != nil {
		agentCopy.Labels = make(map[string]string, len(agent.Labels))
		for k, v := range agent.Labels {
			agentCopy.Labels[k] = v
		}
	}
	if agent.Capabilities != nil {
		agentCopy.Capabilities = make([]string, len(agent.Capabilities))
		copy(agentCopy.Capabilities, agent.Capabilities)
	}

	return &agentCopy, nil
}

func (s *Store) ListAgents(ctx context.Context) ([]*types.Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	agents := make([]*types.Agent, 0, len(s.agents))
	for _, agent := range s.agents {
		// Deep copy
		agentCopy := *agent
		if agent.Labels != nil {
			agentCopy.Labels = make(map[string]string, len(agent.Labels))
			for k, v := range agent.Labels {
				agentCopy.Labels[k] = v
			}
		}
		if agent.Capabilities != nil {
			agentCopy.Capabilities = make([]string, len(agent.Capabilities))
			copy(agentCopy.Capabilities, agent.Capabilities)
		}
		agents = append(agents, &agentCopy)
	}

	return agents, nil
}

func (s *Store) UpdateAgentStatus(ctx context.Context, id uuid.UUID, status types.AgentStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	agent, exists := s.agents[id]
	if !exists {
		return fmt.Errorf("agent not found: %s", id)
	}

	agent.Status = status
	agent.UpdatedAt = time.Now()
	return nil
}

func (s *Store) UpdateAgentLastSeen(ctx context.Context, id uuid.UUID, lastSeen time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	agent, exists := s.agents[id]
	if !exists {
		return fmt.Errorf("agent not found: %s", id)
	}

	agent.LastSeen = lastSeen
	agent.UpdatedAt = time.Now()
	return nil
}

func (s *Store) UpdateAgentEffectiveConfig(ctx context.Context, id uuid.UUID, effectiveConfig string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	agent, exists := s.agents[id]
	if !exists {
		return fmt.Errorf("agent not found: %s", id)
	}

	agent.EffectiveConfig = effectiveConfig
	agent.UpdatedAt = time.Now()
	return nil
}

func (s *Store) DeleteAgent(ctx context.Context, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.agents[id]; !exists {
		return fmt.Errorf("agent not found: %s", id)
	}

	delete(s.agents, id)
	return nil
}

// Group management

func (s *Store) CreateGroup(ctx context.Context, group *types.Group) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.groups[group.ID]; exists {
		return fmt.Errorf("group already exists: %s", group.ID)
	}

	// Deep copy
	groupCopy := *group
	if group.Labels != nil {
		groupCopy.Labels = make(map[string]string, len(group.Labels))
		for k, v := range group.Labels {
			groupCopy.Labels[k] = v
		}
	}

	s.groups[group.ID] = &groupCopy
	return nil
}

func (s *Store) GetGroup(ctx context.Context, id string) (*types.Group, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	group, exists := s.groups[id]
	if !exists {
		return nil, nil
	}

	// Deep copy
	groupCopy := *group
	if group.Labels != nil {
		groupCopy.Labels = make(map[string]string, len(group.Labels))
		for k, v := range group.Labels {
			groupCopy.Labels[k] = v
		}
	}

	return &groupCopy, nil
}

func (s *Store) ListGroups(ctx context.Context) ([]*types.Group, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	groups := make([]*types.Group, 0, len(s.groups))
	for _, group := range s.groups {
		// Deep copy
		groupCopy := *group
		if group.Labels != nil {
			groupCopy.Labels = make(map[string]string, len(group.Labels))
			for k, v := range group.Labels {
				groupCopy.Labels[k] = v
			}
		}
		groups = append(groups, &groupCopy)
	}

	return groups, nil
}

func (s *Store) DeleteGroup(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.groups[id]; !exists {
		return fmt.Errorf("group not found: %s", id)
	}

	delete(s.groups, id)
	return nil
}

// Config management

func (s *Store) CreateConfig(ctx context.Context, config *types.Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.configs[config.ID]; exists {
		return fmt.Errorf("config already exists: %s", config.ID)
	}

	// Deep copy
	configCopy := *config
	s.configs[config.ID] = &configCopy
	return nil
}

func (s *Store) GetConfig(ctx context.Context, id string) (*types.Config, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	config, exists := s.configs[id]
	if !exists {
		return nil, nil
	}

	// Deep copy
	configCopy := *config
	return &configCopy, nil
}

func (s *Store) GetLatestConfigForAgent(ctx context.Context, agentID uuid.UUID) (*types.Config, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var latestConfig *types.Config
	for _, config := range s.configs {
		if config.AgentID != nil && *config.AgentID == agentID {
			if latestConfig == nil || config.Version > latestConfig.Version ||
				(config.Version == latestConfig.Version && config.CreatedAt.After(latestConfig.CreatedAt)) {
				latestConfig = config
			}
		}
	}

	if latestConfig == nil {
		return nil, nil
	}

	// Deep copy
	configCopy := *latestConfig
	return &configCopy, nil
}

func (s *Store) GetLatestConfigForGroup(ctx context.Context, groupID string) (*types.Config, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var latestConfig *types.Config
	for _, config := range s.configs {
		if config.GroupID != nil && *config.GroupID == groupID {
			if latestConfig == nil || config.Version > latestConfig.Version ||
				(config.Version == latestConfig.Version && config.CreatedAt.After(latestConfig.CreatedAt)) {
				latestConfig = config
			}
		}
	}

	if latestConfig == nil {
		return nil, nil
	}

	// Deep copy
	configCopy := *latestConfig
	return &configCopy, nil
}

func (s *Store) ListConfigs(ctx context.Context, filter types.ConfigFilter) ([]*types.Config, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	configs := make([]*types.Config, 0)
	for _, config := range s.configs {
		// Apply filters
		if filter.AgentID != nil && (config.AgentID == nil || *config.AgentID != *filter.AgentID) {
			continue
		}
		if filter.GroupID != nil && (config.GroupID == nil || *config.GroupID != *filter.GroupID) {
			continue
		}

		// Deep copy
		configCopy := *config
		configs = append(configs, &configCopy)
	}

	// Apply limit
	if filter.Limit > 0 && len(configs) > filter.Limit {
		configs = configs[:filter.Limit]
	}

	return configs, nil
}

// purge removes all data from the store (for testing)
func (s *Store) purge(context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.agents = make(map[uuid.UUID]*types.Agent)
	s.groups = make(map[string]*types.Group)
	s.configs = make(map[string]*types.Config)
}
