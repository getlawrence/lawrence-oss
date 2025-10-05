package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/getlawrence/lawrence-oss/internal/storage"
	"github.com/getlawrence/lawrence-oss/internal/storage/interfaces"
)

// GroupHandlers handles group-related API endpoints
type GroupHandlers struct {
	storage *storage.Container
	logger  *zap.Logger
}

// NewGroupHandlers creates a new group handlers instance
func NewGroupHandlers(storage *storage.Container, logger *zap.Logger) *GroupHandlers {
	return &GroupHandlers{
		storage: storage,
		logger:  logger,
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
	groups, err := h.storage.App.ListGroups(c.Request.Context())
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
	group := &interfaces.Group{
		ID:        groupID,
		Name:      req.Name,
		Labels:    req.Labels,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save group to storage
	err := h.storage.App.CreateGroup(c.Request.Context(), group)
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
	group, err := h.storage.App.GetGroup(c.Request.Context(), groupID)
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
	err := h.storage.App.DeleteGroup(c.Request.Context(), groupID)
	if err != nil {
		h.logger.Error("Failed to delete group", zap.String("group_id", groupID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group deleted successfully"})
}
