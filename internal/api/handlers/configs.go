package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/getlawrence/lawrence-oss/internal/services"
)

// ConfigHandlers handles config-related API endpoints
type ConfigHandlers struct {
	agentService services.AgentService
	logger       *zap.Logger
}

// NewConfigHandlers creates a new config handlers instance
func NewConfigHandlers(agentService services.AgentService, logger *zap.Logger) *ConfigHandlers {
	return &ConfigHandlers{
		agentService: agentService,
		logger:       logger,
	}
}

// CreateConfigRequest represents the request for creating a config
type CreateConfigRequest struct {
	Name       string     `json:"name,omitempty"`
	AgentID    *uuid.UUID `json:"agent_id,omitempty"`
	GroupID    *string    `json:"group_id,omitempty"`
	ConfigHash string     `json:"config_hash" binding:"required"`
	Content    string     `json:"content" binding:"required"`
	Version    int        `json:"version" binding:"required"`
}

// UpdateConfigRequest represents the request for updating a config
type UpdateConfigRequest struct {
	Name    string `json:"name,omitempty"`
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
	filter := services.ConfigFilter{
		AgentID: agentUUID,
		GroupID: groupID,
		Limit:   limit,
	}

	// Get configs from service
	configs, err := h.agentService.ListConfigs(c.Request.Context(), filter)
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
	config := &services.Config{
		ID:         configID,
		Name:       req.Name,
		AgentID:    req.AgentID,
		GroupID:    req.GroupID,
		ConfigHash: req.ConfigHash,
		Content:    req.Content,
		Version:    req.Version,
		CreatedAt:  time.Now(),
	}

	// Save config to service
	err := h.agentService.CreateConfig(c.Request.Context(), config)
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

	// Get config from service
	config, err := h.agentService.GetConfig(c.Request.Context(), configID)
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
	configID := c.Param("id")
	if configID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Config ID is required"})
		return
	}

	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// Get existing config
	existingConfig, err := h.agentService.GetConfig(c.Request.Context(), configID)
	if err != nil {
		h.logger.Error("Failed to get config", zap.String("config_id", configID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch config"})
		return
	}

	if existingConfig == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		return
	}

	// Validate YAML content
	if err := validateYAMLConfig(req.Content); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid YAML configuration", "details": err.Error()})
		return
	}

	// Create new version of config
	newConfigID := uuid.New().String()
	configHash := hashConfig(req.Content)

	// Use the new name if provided, otherwise keep the existing name
	name := req.Name
	if name == "" {
		name = existingConfig.Name
	}

	newConfig := &services.Config{
		ID:         newConfigID,
		Name:       name,
		AgentID:    existingConfig.AgentID,
		GroupID:    existingConfig.GroupID,
		ConfigHash: configHash,
		Content:    req.Content,
		Version:    req.Version,
		CreatedAt:  time.Now(),
	}

	// Save new config version
	err = h.agentService.CreateConfig(c.Request.Context(), newConfig)
	if err != nil {
		h.logger.Error("Failed to create config version", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create config version"})
		return
	}

	c.JSON(http.StatusOK, newConfig)
}

// handleDeleteConfig handles DELETE /api/v1/configs/:id
func (h *ConfigHandlers) HandleDeleteConfig(c *gin.Context) {
	// Note: In production, you may want to soft-delete or prevent deletion
	// of configs that are in use
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Config deletion not implemented - configs are versioned and immutable"})
}

// handleValidateConfig handles POST /api/v1/configs/validate
func (h *ConfigHandlers) HandleValidateConfig(c *gin.Context) {
	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// Validate YAML syntax
	if err := validateYAMLConfig(req.Content); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"valid":  false,
			"errors": []string{err.Error()},
		})
		return
	}

	// Additional validation (check required fields, etc.)
	warnings, err := validateOTelConfig(req.Content)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"valid":  false,
			"errors": []string{err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":    true,
		"warnings": warnings,
	})
}

// handleGetConfigVersions handles GET /api/v1/configs/:id/versions
func (h *ConfigHandlers) HandleGetConfigVersions(c *gin.Context) {
	// Get query parameters
	agentIDStr := c.Query("agent_id")
	groupIDStr := c.Query("group_id")

	if agentIDStr == "" && groupIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Either agent_id or group_id is required"})
		return
	}

	// Build filter
	filter := services.ConfigFilter{
		Limit: 100,
	}

	if agentIDStr != "" {
		agentUUID, err := uuid.Parse(agentIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID format"})
			return
		}
		filter.AgentID = &agentUUID
	}

	if groupIDStr != "" {
		filter.GroupID = &groupIDStr
	}

	// Get config versions
	configs, err := h.agentService.ListConfigs(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to get config versions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch config versions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"versions": configs,
		"count":    len(configs),
	})
}

// Validation helper functions

// validateYAMLConfig validates YAML syntax
func validateYAMLConfig(content string) error {
	var config map[string]interface{}
	if err := yaml.Unmarshal([]byte(content), &config); err != nil {
		return fmt.Errorf("invalid YAML syntax: %w", err)
	}
	return nil
}

// validateOTelConfig performs OpenTelemetry-specific validation
func validateOTelConfig(content string) ([]string, error) {
	var config map[string]interface{}
	if err := yaml.Unmarshal([]byte(content), &config); err != nil {
		return nil, fmt.Errorf("invalid YAML: %w", err)
	}

	var warnings []string

	// Check for required top-level sections
	requiredSections := []string{"receivers", "processors", "exporters", "service"}
	for _, section := range requiredSections {
		if _, exists := config[section]; !exists {
			warnings = append(warnings, fmt.Sprintf("missing recommended section: %s", section))
		}
	}

	// Check service.pipelines
	if service, ok := config["service"].(map[string]interface{}); ok {
		if pipelines, ok := service["pipelines"].(map[string]interface{}); !ok || len(pipelines) == 0 {
			warnings = append(warnings, "no pipelines defined in service section")
		}
	}

	return warnings, nil
}

// hashConfig creates a hash of the config content
func hashConfig(content string) string {
	// Normalize whitespace and newlines for consistent hashing
	normalized := strings.TrimSpace(content)
	hash := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(hash[:])
}
