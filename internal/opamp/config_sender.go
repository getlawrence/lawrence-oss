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
