package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/getlawrence/lawrence-oss/internal/storage"
	"github.com/getlawrence/lawrence-oss/internal/storage/interfaces"
)

// TopologyHandlers handles topology-related API endpoints
type TopologyHandlers struct {
	storage *storage.Container
	logger  *zap.Logger
}

// NewTopologyHandlers creates a new topology handlers instance
func NewTopologyHandlers(storage *storage.Container, logger *zap.Logger) *TopologyHandlers {
	return &TopologyHandlers{
		storage: storage,
		logger:  logger,
	}
}

// TopologyNode represents a node in the topology graph
type TopologyNode struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // "agent", "group", "service"
	Name        string                 `json:"name"`
	Status      string                 `json:"status"`
	GroupID     *string                `json:"group_id,omitempty"`
	GroupName   *string                `json:"group_name,omitempty"`
	Labels      map[string]string      `json:"labels"`
	Metrics     *NodeMetrics           `json:"metrics,omitempty"`
	LastSeen    *time.Time             `json:"last_seen,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TopologyEdge represents a connection between nodes
type TopologyEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"` // "belongs_to", "sends_to", etc.
	Label  string `json:"label,omitempty"`
}

// NodeMetrics represents metrics for a topology node
type NodeMetrics struct {
	MetricCount   int64   `json:"metric_count"`
	LogCount      int64   `json:"log_count"`
	TraceCount    int64   `json:"trace_count"`
	ErrorRate     float64 `json:"error_rate"`
	Latency       float64 `json:"latency"`
	ThroughputRPS float64 `json:"throughput_rps"`
}

// TopologyResponse represents the complete topology graph
type TopologyResponse struct {
	Nodes      []TopologyNode `json:"nodes"`
	Edges      []TopologyEdge `json:"edges"`
	Groups     []GroupSummary `json:"groups"`
	Services   []string       `json:"services"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// GroupSummary represents a group with agent count
type GroupSummary struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	AgentCount int    `json:"agent_count"`
	Status     string `json:"status"`
}

// AgentTopologyRequest represents request for agent topology
type AgentTopologyRequest struct {
	AgentID   *string    `json:"agent_id"`
	GroupID   *string    `json:"group_id"`
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
}

// HandleGetTopology handles GET /api/v1/topology
func (h *TopologyHandlers) HandleGetTopology(c *gin.Context) {
	ctx := c.Request.Context()

	// Get all agents
	agents, err := h.storage.App.ListAgents(ctx)
	if err != nil {
		h.logger.Error("Failed to get agents for topology", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch topology"})
		return
	}

	// Get all groups
	groups, err := h.storage.App.ListGroups(ctx)
	if err != nil {
		h.logger.Error("Failed to get groups for topology", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch topology"})
		return
	}

	// Build topology
	var nodes []TopologyNode
	var edges []TopologyEdge
	var groupSummaries []GroupSummary

	// Create group nodes and summaries
	groupAgentCount := make(map[string]int)
	for _, group := range groups {
		nodes = append(nodes, TopologyNode{
			ID:       "group-" + group.ID,
			Type:     "group",
			Name:     group.Name,
			Status:   "active",
			Labels:   group.Labels,
			Metadata: map[string]interface{}{
				"created_at": group.CreatedAt,
			},
		})
		groupAgentCount[group.ID] = 0
	}

	// Create agent nodes
	for _, agent := range agents {
		agentID := agent.ID.String()

		// Get metrics for this agent (last 5 minutes)
		metrics := h.getAgentMetrics(ctx, agent.ID)

		node := TopologyNode{
			ID:        "agent-" + agentID,
			Type:      "agent",
			Name:      agent.Name,
			Status:    string(agent.Status),
			GroupID:   agent.GroupID,
			GroupName: agent.GroupName,
			Labels:    agent.Labels,
			Metrics:   metrics,
			LastSeen:  &agent.LastSeen,
			Metadata: map[string]interface{}{
				"version":      agent.Version,
				"capabilities": agent.Capabilities,
			},
		}
		nodes = append(nodes, node)

		// Create edge from agent to group if assigned
		if agent.GroupID != nil && *agent.GroupID != "" {
			edges = append(edges, TopologyEdge{
				Source: "agent-" + agentID,
				Target: "group-" + *agent.GroupID,
				Type:   "belongs_to",
				Label:  "member of",
			})
			groupAgentCount[*agent.GroupID]++
		}
	}

	// Build group summaries
	for _, group := range groups {
		summary := GroupSummary{
			ID:         group.ID,
			Name:       group.Name,
			AgentCount: groupAgentCount[group.ID],
			Status:     "active",
		}
		if summary.AgentCount == 0 {
			summary.Status = "empty"
		}
		groupSummaries = append(groupSummaries, summary)
	}

	// Get unique services
	servicesQuery := `
		SELECT DISTINCT service_name FROM (
			SELECT service_name FROM metrics_sum
			UNION
			SELECT service_name FROM metrics_gauge
			UNION
			SELECT service_name FROM logs
			UNION
			SELECT service_name FROM traces
		) AS all_services
		WHERE service_name IS NOT NULL AND service_name != ''
		ORDER BY service_name
	`
	var services []string
	if rows, err := h.storage.Telemetry.QueryRaw(ctx, servicesQuery); err == nil {
		for _, row := range rows {
			if svc, ok := row["service_name"].(string); ok && svc != "" {
				services = append(services, svc)
			}
		}
	}

	response := TopologyResponse{
		Nodes:     nodes,
		Edges:     edges,
		Groups:    groupSummaries,
		Services:  services,
		UpdatedAt: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// HandleGetAgentTopology handles GET /api/v1/topology/agent/:id
func (h *TopologyHandlers) HandleGetAgentTopology(c *gin.Context) {
	agentIDStr := c.Param("id")
	if agentIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Agent ID is required"})
		return
	}

	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID format"})
		return
	}

	ctx := c.Request.Context()

	// Get agent
	agent, err := h.storage.App.GetAgent(ctx, agentID)
	if err != nil {
		h.logger.Error("Failed to get agent", zap.String("agent_id", agentIDStr), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agent"})
		return
	}

	if agent == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	// Get metrics for this agent
	metrics := h.getAgentMetrics(ctx, agentID)

	// Get pipeline info (from config if available)
	config, _ := h.storage.App.GetLatestConfigForAgent(ctx, agentID)
	var pipelineInfo map[string]interface{}
	if config != nil {
		pipelineInfo = map[string]interface{}{
			"config_id":      config.ID,
			"config_version": config.Version,
			"config_hash":    config.ConfigHash,
		}
	}

	response := gin.H{
		"agent":    agent,
		"metrics":  metrics,
		"pipeline": pipelineInfo,
	}

	c.JSON(http.StatusOK, response)
}

// HandleGetGroupTopology handles GET /api/v1/topology/group/:id
func (h *TopologyHandlers) HandleGetGroupTopology(c *gin.Context) {
	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group ID is required"})
		return
	}

	ctx := c.Request.Context()

	// Get group
	group, err := h.storage.App.GetGroup(ctx, groupID)
	if err != nil {
		h.logger.Error("Failed to get group", zap.String("group_id", groupID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch group"})
		return
	}

	if group == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// Get all agents in this group
	allAgents, err := h.storage.App.ListAgents(ctx)
	if err != nil {
		h.logger.Error("Failed to get agents", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agents"})
		return
	}

	var groupAgents []*interfaces.Agent
	for _, agent := range allAgents {
		if agent.GroupID != nil && *agent.GroupID == groupID {
			groupAgents = append(groupAgents, agent)
		}
	}

	// Get aggregated metrics for the group (last 5 minutes)
	endTime := time.Now()
	startTime := endTime.Add(-5 * time.Minute)

	metricsQuery := `
		SELECT COUNT(*) as count FROM (
			SELECT 1 FROM metrics_sum WHERE group_id = ? AND timestamp >= ? AND timestamp <= ?
			UNION ALL
			SELECT 1 FROM metrics_gauge WHERE group_id = ? AND timestamp >= ? AND timestamp <= ?
		) AS all_metrics
	`
	var metricCount int64
	if rows, err := h.storage.Telemetry.QueryRaw(ctx, metricsQuery, groupID, startTime, endTime, groupID, startTime, endTime); err == nil && len(rows) > 0 {
		if count, ok := rows[0]["count"].(int64); ok {
			metricCount = count
		}
	}

	response := gin.H{
		"group":        group,
		"agents":       groupAgents,
		"agent_count":  len(groupAgents),
		"metric_count": metricCount,
	}

	c.JSON(http.StatusOK, response)
}

// getAgentMetrics retrieves metrics for an agent (last 5 minutes)
func (h *TopologyHandlers) getAgentMetrics(ctx context.Context, agentID uuid.UUID) *NodeMetrics {
	endTime := time.Now()
	startTime := endTime.Add(-5 * time.Minute)

	// Count metrics
	metricsQuery := `
		SELECT COUNT(*) as count FROM (
			SELECT 1 FROM metrics_sum WHERE agent_id = ? AND timestamp >= ? AND timestamp <= ?
			UNION ALL
			SELECT 1 FROM metrics_gauge WHERE agent_id = ? AND timestamp >= ? AND timestamp <= ?
		) AS all_metrics
	`
	var metricCount int64
	agentIDStr := agentID.String()
	if rows, err := h.storage.Telemetry.QueryRaw(ctx, metricsQuery, agentIDStr, startTime, endTime, agentIDStr, startTime, endTime); err == nil && len(rows) > 0 {
		if count, ok := rows[0]["count"].(int64); ok {
			metricCount = count
		}
	}

	// Count logs
	logsQuery := `SELECT COUNT(*) as count FROM logs WHERE agent_id = ? AND timestamp >= ? AND timestamp <= ?`
	var logCount int64
	if rows, err := h.storage.Telemetry.QueryRaw(ctx, logsQuery, agentIDStr, startTime, endTime); err == nil && len(rows) > 0 {
		if count, ok := rows[0]["count"].(int64); ok {
			logCount = count
		}
	}

	// Count traces
	tracesQuery := `SELECT COUNT(*) as count FROM traces WHERE agent_id = ? AND timestamp >= ? AND timestamp <= ?`
	var traceCount int64
	if rows, err := h.storage.Telemetry.QueryRaw(ctx, tracesQuery, agentIDStr, startTime, endTime); err == nil && len(rows) > 0 {
		if count, ok := rows[0]["count"].(int64); ok {
			traceCount = count
		}
	}

	return &NodeMetrics{
		MetricCount:   metricCount,
		LogCount:      logCount,
		TraceCount:    traceCount,
		ErrorRate:     0, // TODO: Calculate from logs/traces
		Latency:       0, // TODO: Calculate from traces
		ThroughputRPS: float64(metricCount) / 300.0, // rough estimate over 5 minutes
	}
}
