package opamp

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/open-telemetry/opamp-go/protobufs"
	"go.uber.org/zap"
)

// ConfigSender handles sending configurations to agents via OpAMP
type ConfigSender struct {
	agents *Agents
	logger *zap.Logger
}

// NewConfigSender creates a new config sender
func NewConfigSender(agents *Agents, logger *zap.Logger) *ConfigSender {
	return &ConfigSender{
		agents: agents,
		logger: logger,
	}
}

// SendConfigToAgent sends a configuration to a specific agent
// Returns an error if the agent doesn't exist, is not online, or doesn't support remote config
func (cs *ConfigSender) SendConfigToAgent(agentId uuid.UUID, configContent string) error {
	agent := cs.agents.FindAgent(agentId)
	if agent == nil {
		return fmt.Errorf("agent not found")
	}

	// Check if agent has capability to accept remote config
	if !agent.hasCapability(protobufs.AgentCapabilities_AgentCapabilities_AcceptsRemoteConfig) {
		return fmt.Errorf("agent does not support remote config")
	}

	// Create config map
	configMap := &protobufs.AgentConfigMap{
		ConfigMap: map[string]*protobufs.AgentConfigFile{
			"": {Body: []byte(configContent)},
		},
	}

	// Send config with notification channel
	notifyChannel := make(chan struct{}, 1)
	agent.SetCustomConfig(configMap, notifyChannel)

	// Optional: wait for confirmation with timeout
	select {
	case <-notifyChannel:
		cs.logger.Info("Config successfully applied to agent",
			zap.String("agentId", agentId.String()))
		return nil
	case <-time.After(30 * time.Second):
		return fmt.Errorf("timeout waiting for agent to apply config")
	}
}

// RestartAgent sends a restart command to a specific agent
// Returns an error if the agent doesn't exist or doesn't support restart
func (cs *ConfigSender) RestartAgent(agentId uuid.UUID) error {
	agent := cs.agents.FindAgent(agentId)
	if agent == nil {
		return fmt.Errorf("agent not found")
	}

	// Check if agent has capability to accept restart command
	if !agent.hasCapability(protobufs.AgentCapabilities_AgentCapabilities_AcceptsRestartCommand) {
		return fmt.Errorf("agent does not support restart command")
	}

	agent.SendRestartCommand()
	cs.logger.Info("Restart command sent to agent", zap.String("agentId", agentId.String()))
	return nil
}

// SendConfigToAgentsInGroup sends a configuration to all agents in a group
// Returns the list of agent IDs that were successfully updated and any errors encountered
func (cs *ConfigSender) SendConfigToAgentsInGroup(groupId string, configContent string) ([]uuid.UUID, []error) {
	var updatedAgents []uuid.UUID
	var errors []error

	// Get all agents
	allAgents := cs.agents.GetAllAgentsReadonlyClone()

	// Find agents in this group
	for agentId, agent := range allAgents {
		if agent.GroupID != nil && *agent.GroupID == groupId {
			cs.logger.Info("Attempting to send config to agent in group",
				zap.String("agentId", agentId.String()),
				zap.String("groupId", groupId))

			// Try to send config to this agent
			if err := cs.SendConfigToAgent(agentId, configContent); err != nil {
				cs.logger.Error("Failed to send config to agent",
					zap.String("agentId", agentId.String()),
					zap.Error(err))
				errors = append(errors, fmt.Errorf("agent %s: %w", agentId.String(), err))
			} else {
				cs.logger.Info("Successfully sent config to agent",
					zap.String("agentId", agentId.String()))
				updatedAgents = append(updatedAgents, agentId)
			}
		}
	}

	cs.logger.Info("Group config update completed",
		zap.String("groupId", groupId),
		zap.Int("updated", len(updatedAgents)),
		zap.Int("failed", len(errors)))

	return updatedAgents, errors
}

// RestartAgentsInGroup sends restart commands to all agents in a group
// Returns the list of agent IDs that were successfully restarted and any errors encountered
func (cs *ConfigSender) RestartAgentsInGroup(groupId string) ([]uuid.UUID, []error) {
	var restartedAgents []uuid.UUID
	var errors []error

	// Get all agents
	allAgents := cs.agents.GetAllAgentsReadonlyClone()

	// Find agents in this group
	for agentId, agent := range allAgents {
		if agent.GroupID != nil && *agent.GroupID == groupId {
			// Try to restart this agent
			if err := cs.RestartAgent(agentId); err != nil {
				errors = append(errors, fmt.Errorf("agent %s: %w", agentId.String(), err))
			} else {
				restartedAgents = append(restartedAgents, agentId)
			}
		}
	}

	cs.logger.Info("Group restart command completed",
		zap.String("groupId", groupId),
		zap.Int("restarted", len(restartedAgents)),
		zap.Int("failed", len(errors)))

	return restartedAgents, errors
}
