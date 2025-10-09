package services

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/getlawrence/lawrence-oss/internal/storage/applicationstore"
)

// AgentServiceImpl implements the AgentService interface
type AgentServiceImpl struct {
	appStore applicationstore.ApplicationStore
	logger   *zap.Logger
}

// NewAgentService creates a new agent service
func NewAgentService(appStore applicationstore.ApplicationStore, logger *zap.Logger) AgentService {
	return &AgentServiceImpl{
		appStore: appStore,
		logger:   logger,
	}
}

// CreateAgent creates an agent
func (s *AgentServiceImpl) CreateAgent(ctx context.Context, agent *Agent) error {
	storageAgent := &applicationstore.Agent{
		ID:           agent.ID,
		Name:         agent.Name,
		Labels:       agent.Labels,
		Status:       applicationstore.AgentStatus(agent.Status),
		LastSeen:     agent.LastSeen,
		GroupID:      agent.GroupID,
		GroupName:    agent.GroupName,
		Version:      agent.Version,
		Capabilities: agent.Capabilities,
		CreatedAt:    agent.CreatedAt,
		UpdatedAt:    agent.UpdatedAt,
	}
	return s.appStore.CreateAgent(ctx, storageAgent)
}

// GetAgent gets an agent by ID
func (s *AgentServiceImpl) GetAgent(ctx context.Context, id uuid.UUID) (*Agent, error) {
	agent, err := s.appStore.GetAgent(ctx, id)
	if err != nil {
		return nil, err
	}

	if agent == nil {
		return nil, nil
	}

	return &Agent{
		ID:              agent.ID,
		Name:            agent.Name,
		Labels:          agent.Labels,
		Status:          AgentStatus(agent.Status),
		LastSeen:        agent.LastSeen,
		GroupID:         agent.GroupID,
		GroupName:       agent.GroupName,
		Version:         agent.Version,
		Capabilities:    agent.Capabilities,
		EffectiveConfig: agent.EffectiveConfig,
		CreatedAt:       agent.CreatedAt,
		UpdatedAt:       agent.UpdatedAt,
	}, nil
}

// ListAgents lists all agents
func (s *AgentServiceImpl) ListAgents(ctx context.Context) ([]*Agent, error) {
	agents, err := s.appStore.ListAgents(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*Agent, len(agents))
	for i, agent := range agents {
		result[i] = &Agent{
			ID:              agent.ID,
			Name:            agent.Name,
			Labels:          agent.Labels,
			Status:          AgentStatus(agent.Status),
			LastSeen:        agent.LastSeen,
			GroupID:         agent.GroupID,
			GroupName:       agent.GroupName,
			Version:         agent.Version,
			Capabilities:    agent.Capabilities,
			EffectiveConfig: agent.EffectiveConfig,
			CreatedAt:       agent.CreatedAt,
			UpdatedAt:       agent.UpdatedAt,
		}
	}

	return result, nil
}

// UpdateAgentStatus updates agent status
func (s *AgentServiceImpl) UpdateAgentStatus(ctx context.Context, id uuid.UUID, status AgentStatus) error {
	return s.appStore.UpdateAgentStatus(ctx, id, applicationstore.AgentStatus(status))
}

// UpdateAgentLastSeen updates agent last seen timestamp
func (s *AgentServiceImpl) UpdateAgentLastSeen(ctx context.Context, id uuid.UUID, lastSeen time.Time) error {
	return s.appStore.UpdateAgentLastSeen(ctx, id, lastSeen)
}

// UpdateAgentEffectiveConfig updates agent effective config
func (s *AgentServiceImpl) UpdateAgentEffectiveConfig(ctx context.Context, id uuid.UUID, effectiveConfig string) error {
	return s.appStore.UpdateAgentEffectiveConfig(ctx, id, effectiveConfig)
}

// DeleteAgent deletes an agent
func (s *AgentServiceImpl) DeleteAgent(ctx context.Context, id uuid.UUID) error {
	return s.appStore.DeleteAgent(ctx, id)
}

// CreateGroup creates a group
func (s *AgentServiceImpl) CreateGroup(ctx context.Context, group *Group) error {
	storageGroup := &applicationstore.Group{
		ID:        group.ID,
		Name:      group.Name,
		Labels:    group.Labels,
		CreatedAt: group.CreatedAt,
		UpdatedAt: group.UpdatedAt,
	}
	return s.appStore.CreateGroup(ctx, storageGroup)
}

// GetGroup gets a group by ID
func (s *AgentServiceImpl) GetGroup(ctx context.Context, id string) (*Group, error) {
	group, err := s.appStore.GetGroup(ctx, id)
	if err != nil {
		return nil, err
	}

	if group == nil {
		return nil, nil
	}

	return &Group{
		ID:        group.ID,
		Name:      group.Name,
		Labels:    group.Labels,
		CreatedAt: group.CreatedAt,
		UpdatedAt: group.UpdatedAt,
	}, nil
}

// GetGroupByName gets a group by name
func (s *AgentServiceImpl) GetGroupByName(ctx context.Context, name string) (*Group, error) {
	groups, err := s.appStore.ListGroups(ctx)
	if err != nil {
		return nil, err
	}

	for _, group := range groups {
		if group.Name == name {
			return &Group{
				ID:        group.ID,
				Name:      group.Name,
				Labels:    group.Labels,
				CreatedAt: group.CreatedAt,
				UpdatedAt: group.UpdatedAt,
			}, nil
		}
	}

	return nil, nil
}

// ListGroups lists all groups
func (s *AgentServiceImpl) ListGroups(ctx context.Context) ([]*Group, error) {
	groups, err := s.appStore.ListGroups(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*Group, len(groups))
	for i, group := range groups {
		result[i] = &Group{
			ID:        group.ID,
			Name:      group.Name,
			Labels:    group.Labels,
			CreatedAt: group.CreatedAt,
			UpdatedAt: group.UpdatedAt,
		}
	}

	return result, nil
}

// DeleteGroup deletes a group
func (s *AgentServiceImpl) DeleteGroup(ctx context.Context, id string) error {
	return s.appStore.DeleteGroup(ctx, id)
}

// CreateConfig creates a configuration
func (s *AgentServiceImpl) CreateConfig(ctx context.Context, config *Config) error {
	storageConfig := &applicationstore.Config{
		ID:         config.ID,
		AgentID:    config.AgentID,
		GroupID:    config.GroupID,
		ConfigHash: config.ConfigHash,
		Content:    config.Content,
		Version:    config.Version,
		CreatedAt:  config.CreatedAt,
	}
	return s.appStore.CreateConfig(ctx, storageConfig)
}

// GetConfig gets a configuration by ID
func (s *AgentServiceImpl) GetConfig(ctx context.Context, id string) (*Config, error) {
	config, err := s.appStore.GetConfig(ctx, id)
	if err != nil {
		return nil, err
	}

	if config == nil {
		return nil, nil
	}

	return &Config{
		ID:         config.ID,
		AgentID:    config.AgentID,
		GroupID:    config.GroupID,
		ConfigHash: config.ConfigHash,
		Content:    config.Content,
		Version:    config.Version,
		CreatedAt:  config.CreatedAt,
	}, nil
}

// GetLatestConfigForAgent gets the latest configuration for an agent
func (s *AgentServiceImpl) GetLatestConfigForAgent(ctx context.Context, agentID uuid.UUID) (*Config, error) {
	config, err := s.appStore.GetLatestConfigForAgent(ctx, agentID)
	if err != nil {
		return nil, err
	}

	if config == nil {
		return nil, nil
	}

	return &Config{
		ID:         config.ID,
		AgentID:    config.AgentID,
		GroupID:    config.GroupID,
		ConfigHash: config.ConfigHash,
		Content:    config.Content,
		Version:    config.Version,
		CreatedAt:  config.CreatedAt,
	}, nil
}

// GetLatestConfigForGroup gets the latest configuration for a group
func (s *AgentServiceImpl) GetLatestConfigForGroup(ctx context.Context, groupID string) (*Config, error) {
	config, err := s.appStore.GetLatestConfigForGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	if config == nil {
		return nil, nil
	}

	return &Config{
		ID:         config.ID,
		AgentID:    config.AgentID,
		GroupID:    config.GroupID,
		ConfigHash: config.ConfigHash,
		Content:    config.Content,
		Version:    config.Version,
		CreatedAt:  config.CreatedAt,
	}, nil
}

// ListConfigs lists configurations with filters
func (s *AgentServiceImpl) ListConfigs(ctx context.Context, filter ConfigFilter) ([]*Config, error) {
	storageFilter := applicationstore.ConfigFilter{
		AgentID: filter.AgentID,
		GroupID: filter.GroupID,
		Limit:   filter.Limit,
	}

	configs, err := s.appStore.ListConfigs(ctx, storageFilter)
	if err != nil {
		return nil, err
	}

	result := make([]*Config, len(configs))
	for i, config := range configs {
		result[i] = &Config{
			ID:         config.ID,
			AgentID:    config.AgentID,
			GroupID:    config.GroupID,
			ConfigHash: config.ConfigHash,
			Content:    config.Content,
			Version:    config.Version,
			CreatedAt:  config.CreatedAt,
		}
	}

	return result, nil
}

// StoreConfigForAgent validates and stores configuration for an agent (storage only, no delivery)
func (s *AgentServiceImpl) StoreConfigForAgent(ctx context.Context, agentID uuid.UUID, content string) (*Config, error) {
	// 1. Validate agent exists and has remote config capability
	agent, err := s.GetAgent(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	if agent == nil {
		return nil, fmt.Errorf("agent not found")
	}

	// 2. Check if agent has remote config capability
	hasCapability := false
	for _, cap := range agent.Capabilities {
		if cap == "accepts_remote_config" {
			hasCapability = true
			break
		}
	}
	if !hasCapability {
		return nil, fmt.Errorf("agent does not support remote config")
	}

	// 3. Store config in database with versioning
	configHash := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))

	// Get latest version for this agent
	latestConfig, _ := s.GetLatestConfigForAgent(ctx, agentID)
	version := 1
	if latestConfig != nil {
		version = latestConfig.Version + 1
	}

	newConfig := &Config{
		ID:         uuid.New().String(),
		AgentID:    &agentID,
		ConfigHash: configHash,
		Content:    content,
		Version:    version,
		CreatedAt:  time.Now(),
	}

	if err := s.CreateConfig(ctx, newConfig); err != nil {
		return nil, fmt.Errorf("failed to store config: %w", err)
	}

	s.logger.Info("Configuration stored for agent",
		zap.String("agent_id", agentID.String()),
		zap.String("config_id", newConfig.ID),
		zap.Int("version", version))

	return newConfig, nil
}
