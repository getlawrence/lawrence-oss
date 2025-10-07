package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/getlawrence/lawrence-oss/internal/services"
)

// AgentHandlers handles agent-related API endpoints
type AgentHandlers struct {
	agentService services.AgentService
	logger       *zap.Logger
}

// NewAgentHandlers creates a new agent handlers instance
func NewAgentHandlers(agentService services.AgentService, logger *zap.Logger) *AgentHandlers {
	return &AgentHandlers{
		agentService: agentService,
		logger:       logger,
	}
}

// GetAgentsRequest represents the request for getting agents
type GetAgentsRequest struct {
	// No filters supported in current interface
}

// GetAgentsResponse represents the response for getting agents
type GetAgentsResponse struct {
	Agents        map[string]*services.Agent `json:"agents"`
	TotalCount    int                        `json:"totalCount"`
	ActiveCount   int                        `json:"activeCount"`
	InactiveCount int                        `json:"inactiveCount"`
}

// GetAgentStatsResponse represents agent statistics
type GetAgentStatsResponse struct {
	TotalAgents   int `json:"totalAgents"`
	OnlineAgents  int `json:"onlineAgents"`
	OfflineAgents int `json:"offlineAgents"`
	ErrorAgents   int `json:"errorAgents"`
	GroupsCount   int `json:"groupsCount"`
}

// UpdateAgentGroupRequest represents the request to update agent group
type UpdateAgentGroupRequest struct {
	GroupID *string `json:"group_id" binding:"omitempty,uuid"`
}

// handleGetAgents handles GET /api/v1/agents
func (h *AgentHandlers) HandleGetAgents(c *gin.Context) {
	// Get agents from service
	agents, err := h.agentService.ListAgents(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get agents", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agents"})
		return
	}

	// Convert to map format expected by frontend
	agentsMap := make(map[string]*services.Agent)
	activeCount := 0

	for _, agent := range agents {
		agentsMap[agent.ID.String()] = agent
		if agent.Status == services.AgentStatusOnline {
			activeCount++
		}
	}

	response := GetAgentsResponse{
		Agents:        agentsMap,
		TotalCount:    len(agents),
		ActiveCount:   activeCount,
		InactiveCount: len(agents) - activeCount,
	}

	c.JSON(http.StatusOK, response)
}

// handleGetAgent handles GET /api/v1/agents/:id
func (h *AgentHandlers) HandleGetAgent(c *gin.Context) {
	agentID := c.Param("id")
	if agentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Agent ID is required"})
		return
	}

	// Parse UUID
	agentUUID, err := uuid.Parse(agentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID format"})
		return
	}

	// Get agent from service
	agent, err := h.agentService.GetAgent(c.Request.Context(), agentUUID)
	if err != nil {
		h.logger.Error("Failed to get agent", zap.String("agent_id", agentID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agent"})
		return
	}

	if agent == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	c.JSON(http.StatusOK, agent)
}

// handleUpdateAgentGroup handles PATCH /api/v1/agents/:id/group
func (h *AgentHandlers) HandleUpdateAgentGroup(c *gin.Context) {
	// Not implemented in current interface
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Agent group update not implemented"})
}

// handleGetAgentStats handles GET /api/v1/agents/stats
func (h *AgentHandlers) HandleGetAgentStats(c *gin.Context) {
	// Get all agents
	agents, err := h.agentService.ListAgents(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get agents for stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agent statistics"})
		return
	}

	// Count agents by status
	stats := GetAgentStatsResponse{
		TotalAgents: len(agents),
	}

	for _, agent := range agents {
		switch agent.Status {
		case services.AgentStatusOnline:
			stats.OnlineAgents++
		case services.AgentStatusOffline:
			stats.OfflineAgents++
		case services.AgentStatusError:
			stats.ErrorAgents++
		}
	}

	// Get groups count
	groups, err := h.agentService.ListGroups(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get groups for stats", zap.Error(err))
		// Don't fail the request, just set groups count to 0
		stats.GroupsCount = 0
	} else {
		stats.GroupsCount = len(groups)
	}

	c.JSON(http.StatusOK, stats)
}

// Helper function to convert string pointer to string
func stringPtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
