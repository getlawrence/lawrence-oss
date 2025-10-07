package opamp

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/open-telemetry/opamp-go/protobufs"
	"github.com/open-telemetry/opamp-go/server"
	"github.com/open-telemetry/opamp-go/server/types"
	"go.uber.org/zap"

	"github.com/getlawrence/lawrence-oss/internal/metrics"
	"github.com/getlawrence/lawrence-oss/internal/services"
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
	logger       *zap.Logger
	opampServer  server.OpAMPServer
	agents       *Agents
	agentService services.AgentService
	metrics      *metrics.OpAMPMetrics
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

func NewServer(agents *Agents, agentService services.AgentService, metricsInstance *metrics.OpAMPMetrics, logger *zap.Logger) (*Server, error) {
	s := &Server{
		logger:       logger,
		agents:       agents,
		agentService: agentService,
		metrics:      metricsInstance,
	}

	// Create the OpAMP server
	s.opampServer = server.New(&zapToOpAmpLogger{logger})

	return s, nil
}

func (s *Server) Start(port int) error {
	s.logger.Info("Starting OpAMP server...", zap.Int("port", port))

	// Record server start time
	if s.metrics != nil {
		s.metrics.ServerStartTime.Update(time.Now().Unix())
	}

	settings := server.StartSettings{
		Settings: server.Settings{
			Callbacks: server.CallbacksStruct{
				OnConnectingFunc: func(request *http.Request) types.ConnectionResponse {
					// Track connection attempts
					if s.metrics != nil {
						s.metrics.AgentConnectionsTotal.Inc(1)
					}
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
	_ = s.opampServer.Stop(ctx)
	return nil
}

func (s *Server) onDisconnect(conn types.Connection) {
	// Track disconnections
	if s.metrics != nil {
		s.metrics.AgentDisconnectsTotal.Inc(1)
	}

	// Get agents before removing connection
	s.agents.mux.Lock()
	agentsToMarkOffline := s.agents.connections[conn]
	s.agents.mux.Unlock()

	// Mark all agents on this connection as offline in storage
	if s.agentService != nil {
		ctx := context.Background()
		for agentId := range agentsToMarkOffline {
			if err := s.agentService.UpdateAgentStatus(ctx, agentId, services.AgentStatusOffline); err != nil {
				s.logger.Error("Failed to mark agent offline on disconnect",
					zap.String("agentId", agentId.String()),
					zap.Error(err))
			}
		}
	}

	s.agents.RemoveConnection(conn)

	// Update current connections gauge
	if s.metrics != nil {
		s.metrics.AgentConnections.Update(int64(len(s.agents.GetAllAgentsReadonlyClone())))
	}
}

func (s *Server) onMessage(ctx context.Context, conn types.Connection, msg *protobufs.AgentToServer) *protobufs.ServerToAgent {
	start := time.Now()
	response := &protobufs.ServerToAgent{}
	instanceId := uuid.UUID(msg.InstanceUid)

	// Track message received
	if s.metrics != nil {
		s.metrics.MessagesReceived.Inc(1)
	}

	// Process the message
	agent := s.agents.FindOrCreateAgent(instanceId, conn)
	if agent == nil {
		if s.metrics != nil {
			s.metrics.MessageErrors.Inc(1)
		}
		return response
	}

	// Update connections gauge
	if s.metrics != nil {
		s.metrics.AgentConnections.Update(int64(len(s.agents.GetAllAgentsReadonlyClone())))
	}

	// Track status update if present
	if msg.AgentDescription != nil || msg.RemoteConfigStatus != nil {
		if s.metrics != nil {
			s.metrics.StatusUpdateReceived.Inc(1)
		}
	}

	// Track health report if present
	if msg.Health != nil {
		if s.metrics != nil {
			s.metrics.HealthReportReceived.Inc(1)
		}
	}

	// Process agent grouping if agent description changed
	s.processAgentGrouping(ctx, agent, msg)

	agent.UpdateStatus(msg, response)

	// Persist agent to storage
	if s.agentService != nil {
		s.persistAgent(ctx, agent, msg)
	}

	// Track message sent
	if s.metrics != nil {
		s.metrics.MessagesSent.Inc(1)
		s.metrics.MessageProcessDuration.Record(time.Since(start))
	}

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
		if attr.Key == "group.id" || attr.Key == "service.group.id" || attr.Key == "agent.group_id" {
			if attr.Value != nil && attr.Value.GetStringValue() != "" {
				groupID = attr.Value.GetStringValue()
			}
		}
		if attr.Key == "group.name" || attr.Key == "service.group.name" || attr.Key == "agent.group_name" {
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

// persistAgent persists agent information to storage
func (s *Server) persistAgent(ctx context.Context, agent *Agent, msg *protobufs.AgentToServer) {
	// Check if agent already exists in storage
	existingAgent, err := s.agentService.GetAgent(ctx, agent.InstanceId)
	if err != nil {
		s.logger.Debug("Error checking existing agent",
			zap.String("agentId", agent.InstanceIdStr),
			zap.Error(err))
	}

	now := time.Now()

	// Extract agent details
	name := s.extractAgentName(msg.AgentDescription)
	labels := s.extractAgentLabels(msg.AgentDescription)
	version := s.extractAgentVersion(msg.AgentDescription)
	capabilities := s.extractAgentCapabilities(msg.Capabilities)
	status := s.determineAgentStatus(msg)

	if existingAgent == nil {
		// Auto-create group if it doesn't exist
		if agent.GroupName != nil && *agent.GroupName != "" {
			existingGroup, err := s.agentService.GetGroupByName(ctx, *agent.GroupName)
			if err != nil {
				s.logger.Debug("Error checking existing group",
					zap.String("groupName", *agent.GroupName),
					zap.Error(err))
			}

			if existingGroup == nil {
				// Group doesn't exist, create it
				newGroup := &services.Group{
					ID:        uuid.New().String(),
					Name:      *agent.GroupName,
					Labels:    make(map[string]string),
					CreatedAt: now,
					UpdatedAt: now,
				}

				if err := s.agentService.CreateGroup(ctx, newGroup); err != nil {
					s.logger.Error("Failed to auto-create group",
						zap.String("groupName", *agent.GroupName),
						zap.Error(err))
				} else {
					s.logger.Info("Auto-created group for agent",
						zap.String("groupName", *agent.GroupName),
						zap.String("groupId", newGroup.ID))
					// Update agent's GroupID
					agent.GroupID = &newGroup.ID
				}
			} else {
				// Group exists, set GroupID
				agent.GroupID = &existingGroup.ID
			}
		}

		// Create new agent
		serviceAgent := &services.Agent{
			ID:           agent.InstanceId,
			Name:         name,
			Labels:       labels,
			Status:       services.AgentStatus(status),
			LastSeen:     now,
			GroupID:      agent.GroupID,
			GroupName:    agent.GroupName,
			Version:      version,
			Capabilities: capabilities,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		if err := s.agentService.CreateAgent(ctx, serviceAgent); err != nil {
			s.logger.Error("Failed to create agent in storage",
				zap.String("agentId", agent.InstanceIdStr),
				zap.Error(err))
		} else {
			s.logger.Info("Agent persisted to storage",
				zap.String("agentId", agent.InstanceIdStr),
				zap.String("name", name))
		}
	} else {
		// Update existing agent
		if err := s.agentService.UpdateAgentStatus(ctx, agent.InstanceId, services.AgentStatus(status)); err != nil {
			s.logger.Error("Failed to update agent status",
				zap.String("agentId", agent.InstanceIdStr),
				zap.Error(err))
		}

		if err := s.agentService.UpdateAgentLastSeen(ctx, agent.InstanceId, now); err != nil {
			s.logger.Error("Failed to update agent last seen",
				zap.String("agentId", agent.InstanceIdStr),
				zap.Error(err))
		}
	}
}

// extractAgentName extracts the agent name from agent description
func (s *Server) extractAgentName(desc *protobufs.AgentDescription) string {
	if desc == nil {
		return "unknown"
	}

	// Look for service.name or agent.name
	attrs := append(desc.IdentifyingAttributes, desc.NonIdentifyingAttributes...)
	for _, attr := range attrs {
		if attr.Key == "service.name" || attr.Key == "agent.name" {
			if attr.Value != nil && attr.Value.GetStringValue() != "" {
				return attr.Value.GetStringValue()
			}
		}
	}

	return "unknown"
}

// extractAgentLabels extracts labels from agent description
func (s *Server) extractAgentLabels(desc *protobufs.AgentDescription) map[string]string {
	labels := make(map[string]string)

	if desc == nil {
		return labels
	}

	// Extract all non-identifying attributes as labels
	for _, attr := range desc.NonIdentifyingAttributes {
		if attr.Value != nil {
			labels[attr.Key] = attr.Value.GetStringValue()
		}
	}

	return labels
}

// extractAgentVersion extracts version from agent description
func (s *Server) extractAgentVersion(desc *protobufs.AgentDescription) string {
	if desc == nil {
		return "unknown"
	}

	// Look for service.version or agent.version
	attrs := append(desc.IdentifyingAttributes, desc.NonIdentifyingAttributes...)
	for _, attr := range attrs {
		if attr.Key == "service.version" || attr.Key == "agent.version" {
			if attr.Value != nil && attr.Value.GetStringValue() != "" {
				return attr.Value.GetStringValue()
			}
		}
	}

	return "unknown"
}

// extractAgentCapabilities extracts capabilities from OpAMP message
func (s *Server) extractAgentCapabilities(caps uint64) []string {
	capabilities := []string{}

	// Map OpAMP capabilities to strings
	if caps&uint64(protobufs.AgentCapabilities_AgentCapabilities_ReportsStatus) != 0 {
		capabilities = append(capabilities, "reports_status")
	}
	if caps&uint64(protobufs.AgentCapabilities_AgentCapabilities_AcceptsRemoteConfig) != 0 {
		capabilities = append(capabilities, "accepts_remote_config")
	}
	if caps&uint64(protobufs.AgentCapabilities_AgentCapabilities_ReportsEffectiveConfig) != 0 {
		capabilities = append(capabilities, "reports_effective_config")
	}
	if caps&uint64(protobufs.AgentCapabilities_AgentCapabilities_ReportsOwnTraces) != 0 {
		capabilities = append(capabilities, "reports_own_traces")
	}
	if caps&uint64(protobufs.AgentCapabilities_AgentCapabilities_ReportsOwnMetrics) != 0 {
		capabilities = append(capabilities, "reports_own_metrics")
	}
	if caps&uint64(protobufs.AgentCapabilities_AgentCapabilities_ReportsOwnLogs) != 0 {
		capabilities = append(capabilities, "reports_own_logs")
	}
	if caps&uint64(protobufs.AgentCapabilities_AgentCapabilities_AcceptsPackages) != 0 {
		capabilities = append(capabilities, "accepts_packages")
	}
	if caps&uint64(protobufs.AgentCapabilities_AgentCapabilities_ReportsPackageStatuses) != 0 {
		capabilities = append(capabilities, "reports_package_statuses")
	}
	if caps&uint64(protobufs.AgentCapabilities_AgentCapabilities_ReportsHealth) != 0 {
		capabilities = append(capabilities, "reports_health")
	}
	if caps&uint64(protobufs.AgentCapabilities_AgentCapabilities_ReportsRemoteConfig) != 0 {
		capabilities = append(capabilities, "reports_remote_config")
	}

	return capabilities
}

// determineAgentStatus determines agent status from OpAMP message
func (s *Server) determineAgentStatus(msg *protobufs.AgentToServer) services.AgentStatus {
	// If we're receiving a message, the agent is connected
	// Check health status if provided
	if msg.Health != nil {
		if msg.Health.Healthy {
			return services.AgentStatusOnline
		}
		return services.AgentStatusError
	}

	// No health info means agent is connected but not reporting health
	// This is normal for initial connections, so mark as online
	return services.AgentStatusOnline
}
