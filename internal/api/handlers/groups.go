package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/getlawrence/lawrence-oss/internal/services"
)

// GroupHandlers handles group-related API endpoints
type GroupHandlers struct {
	agentService services.AgentService
	commander    AgentCommander
	logger       *zap.Logger
}

// NewGroupHandlers creates a new group handlers instance
func NewGroupHandlers(agentService services.AgentService, commander AgentCommander, logger *zap.Logger) *GroupHandlers {
	return &GroupHandlers{
		agentService: agentService,
		commander:    commander,
		logger:       logger,
	}
}

// CreateGroupRequest represents the request for creating a group
type CreateGroupRequest struct {
	Name   string            `json:"name" binding:"required"`
	Labels map[string]string `json:"labels,omitempty"`
}

// handleGetGroups handles GET /api/v1/groups
func (h *GroupHandlers) HandleGetGroups(c *gin.Context) {
	// Get groups from storage (no filters supported in current interface)
	groups, err := h.agentService.ListGroups(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get groups", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch groups"})
		return
	}

	// Get all agents to count them per group
	allAgents, err := h.agentService.ListAgents(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get agents", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agents"})
		return
	}

	// Count agents per group
	agentCountByGroup := make(map[string]int)
	for _, agent := range allAgents {
		if agent.GroupID != nil {
			agentCountByGroup[*agent.GroupID]++
		}
	}

	// Enrich groups with agent count and config name
	for _, group := range groups {
		group.AgentCount = agentCountByGroup[group.ID]

		// Get latest config for the group
		config, err := h.agentService.GetLatestConfigForGroup(c.Request.Context(), group.ID)
		if err == nil && config != nil {
			// Extract a simple name from the config (first line or ID)
			group.ConfigName = config.ID
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"groups": groups,
		"count":  len(groups),
	})
}

// handleCreateGroup handles POST /api/v1/groups
func (h *GroupHandlers) HandleCreateGroup(c *gin.Context) {
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// Generate UUID for the group
	groupID := uuid.New().String()

	// Create group
	group := &services.Group{
		ID:        groupID,
		Name:      req.Name,
		Labels:    req.Labels,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save group to storage
	err := h.agentService.CreateGroup(c.Request.Context(), group)
	if err != nil {
		h.logger.Error("Failed to create group", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group"})
		return
	}

	c.JSON(http.StatusCreated, group)
}

// handleGetGroup handles GET /api/v1/groups/:id
func (h *GroupHandlers) HandleGetGroup(c *gin.Context) {
	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group ID is required"})
		return
	}

	// Get group from storage
	group, err := h.agentService.GetGroup(c.Request.Context(), groupID)
	if err != nil {
		h.logger.Error("Failed to get group", zap.String("group_id", groupID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch group"})
		return
	}

	if group == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	c.JSON(http.StatusOK, group)
}

// handleUpdateGroup handles PUT /api/v1/groups/:id
func (h *GroupHandlers) HandleUpdateGroup(c *gin.Context) {
	// Not implemented in current interface
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Group update not implemented"})
}

// handleDeleteGroup handles DELETE /api/v1/groups/:id
func (h *GroupHandlers) HandleDeleteGroup(c *gin.Context) {
	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group ID is required"})
		return
	}

	// Delete group from storage
	err := h.agentService.DeleteGroup(c.Request.Context(), groupID)
	if err != nil {
		h.logger.Error("Failed to delete group", zap.String("group_id", groupID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group deleted successfully"})
}

// AssignConfigRequest represents the request to assign a config to a group
type AssignConfigRequest struct {
	ConfigID string `json:"config_id" binding:"required"`
}

// handleAssignConfig handles POST /api/v1/groups/:id/config
func (h *GroupHandlers) HandleAssignConfig(c *gin.Context) {
	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group ID is required"})
		return
	}

	var req AssignConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// Verify group exists
	group, err := h.agentService.GetGroup(c.Request.Context(), groupID)
	if err != nil || group == nil {
		h.logger.Error("Failed to get group", zap.String("group_id", groupID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// Verify config exists
	config, err := h.agentService.GetConfig(c.Request.Context(), req.ConfigID)
	if err != nil || config == nil {
		h.logger.Error("Failed to get config", zap.String("config_id", req.ConfigID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		return
	}

	// Update config to be assigned to this group
	newConfig := &services.Config{
		ID:         uuid.New().String(),
		GroupID:    &groupID,
		ConfigHash: config.ConfigHash,
		Content:    config.Content,
		Version:    config.Version + 1,
		CreatedAt:  time.Now(),
	}

	err = h.agentService.CreateConfig(c.Request.Context(), newConfig)
	if err != nil {
		h.logger.Error("Failed to assign config to group", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign config"})
		return
	}

	h.logger.Info("Assigned config to group", zap.String("group_id", groupID), zap.String("config_id", newConfig.ID))

	// Send config to all agents in the group
	updatedAgents, errors := h.commander.SendConfigToAgentsInGroup(groupID, newConfig.Content)
	
	// Log the results
	if len(errors) > 0 {
		h.logger.Warn("Some agents failed to receive group config",
			zap.String("group_id", groupID),
			zap.Int("updated", len(updatedAgents)),
			zap.Int("failed", len(errors)))
	} else {
		h.logger.Info("Group config sent to all agents",
			zap.String("group_id", groupID),
			zap.Int("updated", len(updatedAgents)))
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Config assigned to group successfully",
		"config":  newConfig,
	})
}

// handleGetGroupConfig handles GET /api/v1/groups/:id/config
func (h *GroupHandlers) HandleGetGroupConfig(c *gin.Context) {
	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group ID is required"})
		return
	}

	// Get latest config for group
	config, err := h.agentService.GetLatestConfigForGroup(c.Request.Context(), groupID)
	if err != nil {
		h.logger.Error("Failed to get group config", zap.String("group_id", groupID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch group config"})
		return
	}

	if config == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No config assigned to this group"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// handleGetGroupAgents handles GET /api/v1/groups/:id/agents
func (h *GroupHandlers) HandleGetGroupAgents(c *gin.Context) {
	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group ID is required"})
		return
	}

	// Get all agents
	allAgents, err := h.agentService.ListAgents(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get agents", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agents"})
		return
	}

	// Filter agents by group
	var groupAgents []*services.Agent
	for _, agent := range allAgents {
		if agent.GroupID != nil && *agent.GroupID == groupID {
			groupAgents = append(groupAgents, agent)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"agents": groupAgents,
		"count":  len(groupAgents),
	})
}

// RestartGroupResponse represents the response after restarting agents in a group
type RestartGroupResponse struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	RestartedCount int    `json:"restarted_count"`
	FailedCount    int    `json:"failed_count"`
}

// HandleRestartGroup handles POST /api/v1/groups/:id/restart
func (h *GroupHandlers) HandleRestartGroup(c *gin.Context) {
	// 1. Parse group ID from URL
	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group ID is required"})
		return
	}

	// 2. Verify group exists
	group, err := h.agentService.GetGroup(c.Request.Context(), groupID)
	if err != nil {
		h.logger.Error("Failed to get group", zap.String("group_id", groupID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch group"})
		return
	}

	if group == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// 3. Send restart commands to all agents in the group
	restartedAgents, errors := h.commander.RestartAgentsInGroup(groupID)

	// 4. Build response
	restartedCount := len(restartedAgents)
	failedCount := len(errors)
	totalAttempted := restartedCount + failedCount

	h.logger.Info("Group restart command completed",
		zap.String("group_id", groupID),
		zap.Int("restarted", restartedCount),
		zap.Int("failed", failedCount))

	// Determine success and message
	var success bool
	var message string

	if failedCount == 0 {
		if restartedCount == 0 {
			success = false
			message = "No agents found in group"
		} else {
			success = true
			message = fmt.Sprintf("Successfully restarted all %d agent(s)", restartedCount)
		}
	} else if restartedCount > 0 {
		success = true
		message = fmt.Sprintf("Partially successful: restarted %d/%d agent(s)", restartedCount, totalAttempted)
	} else {
		success = false
		message = fmt.Sprintf("Failed to restart all %d agent(s)", failedCount)
	}

	// Return appropriate status code
	statusCode := http.StatusOK
	if !success {
		if restartedCount == 0 && failedCount > 0 {
			statusCode = http.StatusBadRequest
		}
	}

	c.JSON(statusCode, RestartGroupResponse{
		Success:        success,
		Message:        message,
		RestartedCount: restartedCount,
		FailedCount:    failedCount,
	})
}
