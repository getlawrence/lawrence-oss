package opamp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/open-telemetry/opamp-go/protobufs"
	"github.com/open-telemetry/opamp-go/server"
	"github.com/open-telemetry/opamp-go/server/types"
	"go.uber.org/zap"
)

// DefaultOTelConfig provides the default OpenTelemetry Collector configuration
const DefaultOTelConfig = `receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:

exporters:
  otlp:
    endpoint: localhost:4317
    tls:
      insecure: true

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp]
    metrics:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp]
    logs:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp]
`

type Server struct {
	logger      *zap.Logger
	opampServer server.OpAMPServer
	agents      *Agents
}

// zapToOpAmpLogger adapts zap.Logger to opamp's logger interface
type zapToOpAmpLogger struct {
	*zap.Logger
}

func (z *zapToOpAmpLogger) Debugf(ctx context.Context, format string, args ...interface{}) {
	z.Sugar().Debugf(format, args...)
}

func (z *zapToOpAmpLogger) Errorf(ctx context.Context, format string, args ...interface{}) {
	z.Sugar().Errorf(format, args...)
}

func NewServer(agents *Agents, logger *zap.Logger) (*Server, error) {
	s := &Server{
		logger: logger,
		agents: agents,
	}

	// Create the OpAMP server
	s.opampServer = server.New(&zapToOpAmpLogger{logger})

	return s, nil
}

func (s *Server) Start(port int) error {
	s.logger.Info("Starting OpAMP server...", zap.Int("port", port))

	settings := server.StartSettings{
		Settings: server.Settings{
			Callbacks: server.CallbacksStruct{
				OnConnectingFunc: func(request *http.Request) types.ConnectionResponse {
					return types.ConnectionResponse{
						Accept: true,
						ConnectionCallbacks: server.ConnectionCallbacksStruct{
							OnMessageFunc:         s.onMessage,
							OnConnectionCloseFunc: s.onDisconnect,
						},
					}
				},
			},
		},
		ListenEndpoint: fmt.Sprintf(":%d", port),
	}

	if err := s.opampServer.Start(settings); err != nil {
		return fmt.Errorf("failed to start OpAMP server: %w", err)
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping OpAMP server...")
	s.opampServer.Stop(ctx)
	return nil
}

func (s *Server) onDisconnect(conn types.Connection) {
	s.agents.RemoveConnection(conn)
}

func (s *Server) onMessage(ctx context.Context, conn types.Connection, msg *protobufs.AgentToServer) *protobufs.ServerToAgent {
	response := &protobufs.ServerToAgent{}
	instanceId := uuid.UUID(msg.InstanceUid)

	// Process the message
	agent := s.agents.FindOrCreateAgent(instanceId, conn)
	if agent == nil {
		return response
	}

	// Process agent grouping if agent description changed
	s.processAgentGrouping(ctx, agent, msg)

	agent.UpdateStatus(msg, response)
	return response
}

func (s *Server) GetEffectiveConfig(agentId uuid.UUID) (string, error) {
	agent := s.agents.FindAgent(agentId)
	if agent != nil {
		return agent.EffectiveConfig, nil
	}
	return "", fmt.Errorf("agent %s not found", agentId)
}

func (s *Server) UpdateConfig(agentId uuid.UUID, config map[string]interface{}, notifyNextStatusUpdate chan<- struct{}) error {
	agent := s.agents.FindAgent(agentId)
	if agent == nil {
		return fmt.Errorf("agent %s not found", agentId)
	}

	// Convert config to YAML or JSON string
	// For now, we'll use a simple string representation
	// In a real implementation, you'd marshal this to YAML
	configStr := DefaultOTelConfig

	configMap := &protobufs.AgentConfigMap{
		ConfigMap: map[string]*protobufs.AgentConfigFile{
			"": {Body: []byte(configStr)},
		},
	}

	s.agents.SetCustomConfigForAgent(agentId, configMap, notifyNextStatusUpdate)
	return nil
}

func (s *Server) ListAgents() map[uuid.UUID]*Agent {
	return s.agents.GetAllAgentsReadonlyClone()
}

// processAgentGrouping handles group resolution for agents
// In OSS version, this is simplified - no backend API calls
func (s *Server) processAgentGrouping(ctx context.Context, agent *Agent, msg *protobufs.AgentToServer) {
	// Only process if agent description is provided (indicates change or first connect)
	if msg.AgentDescription == nil {
		return
	}

	// Extract group information from agent description attributes
	groupID, groupName := s.extractGroupInfo(msg.AgentDescription)

	// Check if group information has changed
	groupChanged := false
	if agent.GroupID == nil && groupID != "" {
		groupChanged = true
	} else if agent.GroupID != nil && groupID != *agent.GroupID {
		groupChanged = true
	} else if agent.GroupName != nil && *agent.GroupName != groupName {
		groupChanged = true
	}

	// Update agent's group information
	agent.mux.Lock()
	previousGroupID := agent.GroupID
	agent.GroupID = &groupID
	agent.GroupName = &groupName
	agent.mux.Unlock()

	// Log group membership changes
	if previousGroupID == nil && groupID != "" {
		s.logger.Info("Agent joined group",
			zap.String("agentId", agent.InstanceIdStr),
			zap.String("groupId", groupID),
			zap.String("groupName", groupName))
	} else if previousGroupID != nil && groupID == "" {
		s.logger.Info("Agent left group",
			zap.String("agentId", agent.InstanceIdStr),
			zap.String("previousGroupId", *previousGroupID))
	} else if previousGroupID != nil && groupID != "" && *previousGroupID != groupID {
		s.logger.Info("Agent changed groups",
			zap.String("agentId", agent.InstanceIdStr),
			zap.String("previousGroupId", *previousGroupID),
			zap.String("newGroupId", groupID),
			zap.String("groupName", groupName))
	}

	// Set initial config based on group membership (or default)
	if groupChanged {
		config := s.getConfigForAgent(ctx, agent)
		if config != "" {
			agent.mux.Lock()
			agent.CustomInstanceConfig = config
			agent.calcRemoteConfig()
			agent.mux.Unlock()

			s.logger.Info("Set initial config for agent",
				zap.String("agentId", agent.InstanceIdStr),
				zap.String("groupId", groupID))
		}
	}
}

// extractGroupInfo extracts group ID and name from agent description
func (s *Server) extractGroupInfo(desc *protobufs.AgentDescription) (groupID string, groupName string) {
	if desc == nil {
		return "", ""
	}

	// Look for group information in identifying or non-identifying attributes
	attrs := append(desc.IdentifyingAttributes, desc.NonIdentifyingAttributes...)
	for _, attr := range attrs {
		if attr.Key == "group.id" || attr.Key == "service.group.id" {
			if attr.Value != nil && attr.Value.GetStringValue() != "" {
				groupID = attr.Value.GetStringValue()
			}
		}
		if attr.Key == "group.name" || attr.Key == "service.group.name" {
			if attr.Value != nil && attr.Value.GetStringValue() != "" {
				groupName = attr.Value.GetStringValue()
			}
		}
	}

	return groupID, groupName
}

// getConfigForAgent returns the configuration for an agent
// In OSS version, we just return the default config
func (s *Server) getConfigForAgent(ctx context.Context, agent *Agent) string {
	// For OSS version, always return default config
	// In the future, this could be extended to support group-specific configs
	// stored in the local SQLite database
	s.logger.Debug("Using default config for agent",
		zap.String("agentId", agent.InstanceIdStr))
	return DefaultOTelConfig
}
