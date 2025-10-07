package handlers

import (
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
	logger       *zap.Logger
}

// NewGroupHandlers creates a new group handlers instance
func NewGroupHandlers(agentService services.AgentService, logger *zap.Logger) *GroupHandlers {
	return &GroupHandlers{
		agentService: agentService,
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
