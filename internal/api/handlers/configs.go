package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/getlawrence/lawrence-oss/internal/storage"
	"github.com/getlawrence/lawrence-oss/internal/storage/interfaces"
)

// ConfigHandlers handles config-related API endpoints
type ConfigHandlers struct {
	storage *storage.Container
	logger  *zap.Logger
}

// NewConfigHandlers creates a new config handlers instance
func NewConfigHandlers(storage *storage.Container, logger *zap.Logger) *ConfigHandlers {
	return &ConfigHandlers{
		storage: storage,
		logger:  logger,
	}
}

// CreateConfigRequest represents the request for creating a config
type CreateConfigRequest struct {
	AgentID    *uuid.UUID `json:"agent_id,omitempty"`
	GroupID    *string    `json:"group_id,omitempty"`
	ConfigHash string     `json:"config_hash" binding:"required"`
	Content    string     `json:"content" binding:"required"`
	Version    int        `json:"version" binding:"required"`
}

// UpdateConfigRequest represents the request for updating a config
type UpdateConfigRequest struct {
	Content string `json:"content" binding:"required"`
	Version int    `json:"version" binding:"required"`
}

// handleGetConfigs handles GET /api/v1/configs
func (h *ConfigHandlers) HandleGetConfigs(c *gin.Context) {
	// Parse query parameters
	agentIDStr := c.Query("agent_id")
	groupIDStr := c.Query("group_id")
	limitStr := c.DefaultQuery("limit", "100")

	// Parse limit
	limit := 100
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}
	if limit > 1000 {
		limit = 1000
	}

	// Parse UUIDs
	var agentUUID *uuid.UUID
	var groupID *string
	var err error

	if agentIDStr != "" {
		parsed, err := uuid.Parse(agentIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID format"})
			return
		}
		agentUUID = &parsed
	}

	if groupIDStr != "" {
		groupID = &groupIDStr
	}

	// Build filter
	filter := interfaces.ConfigFilter{
		AgentID: agentUUID,
		GroupID: groupID,
		Limit:   limit,
	}

	// Get configs from storage
	configs, err := h.storage.App.ListConfigs(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to get configs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch configs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"configs": configs,
		"count":   len(configs),
	})
}

// handleCreateConfig handles POST /api/v1/configs
func (h *ConfigHandlers) HandleCreateConfig(c *gin.Context) {
	var req CreateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// Generate UUID for the config
	configID := uuid.New().String()

	// Create config
	config := &interfaces.Config{
		ID:         configID,
		AgentID:    req.AgentID,
		GroupID:    req.GroupID,
		ConfigHash: req.ConfigHash,
		Content:    req.Content,
		Version:    req.Version,
		CreatedAt:  time.Now(),
	}

	// Save config to storage
	err := h.storage.App.CreateConfig(c.Request.Context(), config)
	if err != nil {
		h.logger.Error("Failed to create config", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create config"})
		return
	}

	c.JSON(http.StatusCreated, config)
}

// handleGetConfig handles GET /api/v1/configs/:id
func (h *ConfigHandlers) HandleGetConfig(c *gin.Context) {
	configID := c.Param("id")
	if configID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Config ID is required"})
		return
	}

	// Get config from storage
	config, err := h.storage.App.GetConfig(c.Request.Context(), configID)
	if err != nil {
		h.logger.Error("Failed to get config", zap.String("config_id", configID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch config"})
		return
	}

	if config == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// handleUpdateConfig handles PUT /api/v1/configs/:id
func (h *ConfigHandlers) HandleUpdateConfig(c *gin.Context) {
	// Not implemented in current interface
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Config update not implemented"})
}

// handleDeleteConfig handles DELETE /api/v1/configs/:id
func (h *ConfigHandlers) HandleDeleteConfig(c *gin.Context) {
	// Not implemented in current interface
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Config deletion not implemented"})
}
