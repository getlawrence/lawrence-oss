package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	collectorconfigschema "github.com/pavolloffay/opentelemetry-collector-config-schema"
	"go.uber.org/zap"
)

// SchemaHandlers handles schema-related API endpoints
type SchemaHandlers struct {
	logger        *zap.Logger
	schemaManager *collectorconfigschema.SchemaManager
}

// NewSchemaHandlers creates a new schema handlers instance
func NewSchemaHandlers(logger *zap.Logger) *SchemaHandlers {
	return &SchemaHandlers{
		logger:        logger,
		schemaManager: collectorconfigschema.NewSchemaManager(),
	}
}

// ComponentType represents the type of OpenTelemetry collector component
type ComponentType string

const (
	ComponentTypeReceiver  ComponentType = "receiver"
	ComponentTypeProcessor ComponentType = "processor"
	ComponentTypeExporter  ComponentType = "exporter"
	ComponentTypeConnector ComponentType = "connector"
	ComponentTypeExtension ComponentType = "extension"
)

// ComponentInfo represents information about a component
type ComponentInfo struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// HandleGetComponentSchemas handles GET /api/v1/schemas/components
// Returns list of available components, optionally filtered by type
func (h *SchemaHandlers) HandleGetComponentSchemas(c *gin.Context) {
	componentType := c.Query("type")

	// Get latest version
	version, err := h.schemaManager.GetLatestVersion()
	if err != nil {
		h.logger.Error("Failed to get latest schema version", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get schema version",
		})
		return
	}

	// Get all available components
	allComponents, err := h.schemaManager.ListAvailableComponents(version)
	if err != nil {
		h.logger.Error("Failed to list available components", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list components",
		})
		return
	}

	// Convert to ComponentInfo slice
	components := []ComponentInfo{}
	for compType, compNames := range allComponents {
		// Filter by type if provided
		if componentType != "" && string(compType) != componentType {
			continue
		}

		for _, compName := range compNames {
			components = append(components, ComponentInfo{
				Type: string(compType),
				Name: compName,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"components": components,
	})
}

// HandleGetComponentSchema handles GET /api/v1/schemas/components/:type/:name
// Returns JSON schema for a specific component
func (h *SchemaHandlers) HandleGetComponentSchema(c *gin.Context) {
	componentTypeStr := c.Param("type")
	componentName := c.Param("name")

	// Convert string to ComponentType
	var componentType collectorconfigschema.ComponentType
	switch componentTypeStr {
	case "receiver":
		componentType = collectorconfigschema.ComponentTypeReceiver
	case "processor":
		componentType = collectorconfigschema.ComponentTypeProcessor
	case "exporter":
		componentType = collectorconfigschema.ComponentTypeExporter
	case "connector":
		componentType = collectorconfigschema.ComponentTypeConnector
	case "extension":
		componentType = collectorconfigschema.ComponentTypeExtension
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid component type. Must be one of: receiver, processor, exporter, connector, extension",
		})
		return
	}

	// Get latest version
	version, err := h.schemaManager.GetLatestVersion()
	if err != nil {
		h.logger.Error("Failed to get latest schema version", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get schema version",
		})
		return
	}

	// Get component schema JSON
	schemaJSON, err := h.schemaManager.GetComponentSchemaJSON(componentType, componentName, version)
	if err != nil {
		h.logger.Error("Failed to get component schema", zap.Error(err), zap.String("type", componentTypeStr), zap.String("name", componentName))
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Component schema not found: %s", err.Error()),
		})
		return
	}

	// Parse JSON to return as structured response
	var schema map[string]interface{}
	if err := json.Unmarshal(schemaJSON, &schema); err != nil {
		h.logger.Error("Failed to parse schema JSON", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse schema",
		})
		return
	}

	c.JSON(http.StatusOK, schema)
}

// ValidateComponentConfigRequest represents the request for validating a component config
type ValidateComponentConfigRequest struct {
	Type   string      `json:"type" binding:"required"`
	Name   string      `json:"name" binding:"required"`
	Config interface{} `json:"config" binding:"required"`
}

// ValidateComponentConfigResponse represents the validation response
type ValidateComponentConfigResponse struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// HandleValidateComponentConfig handles POST /api/v1/schemas/validate
// Validates a component configuration against its schema
func (h *SchemaHandlers) HandleValidateComponentConfig(c *gin.Context) {
	var req ValidateComponentConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// Validate component type
	validTypes := map[string]bool{
		"receiver":  true,
		"processor": true,
		"exporter":  true,
		"connector": true,
		"extension": true,
	}
	if !validTypes[req.Type] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid component type. Must be one of: receiver, processor, exporter, connector, extension",
		})
		return
	}

	// Convert string to ComponentType
	var componentType collectorconfigschema.ComponentType
	switch req.Type {
	case "receiver":
		componentType = collectorconfigschema.ComponentTypeReceiver
	case "processor":
		componentType = collectorconfigschema.ComponentTypeProcessor
	case "exporter":
		componentType = collectorconfigschema.ComponentTypeExporter
	case "connector":
		componentType = collectorconfigschema.ComponentTypeConnector
	case "extension":
		componentType = collectorconfigschema.ComponentTypeExtension
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid component type. Must be one of: receiver, processor, exporter, connector, extension",
		})
		return
	}

	// Marshal config to JSON
	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid config format: " + err.Error(),
		})
		return
	}

	// Get latest version
	version, err := h.schemaManager.GetLatestVersion()
	if err != nil {
		h.logger.Error("Failed to get latest schema version", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get schema version",
		})
		return
	}

	// Validate component config
	validationResult, err := h.schemaManager.ValidateComponentJSON(componentType, req.Name, version, configJSON)
	if err != nil {
		h.logger.Error("Failed to validate component config", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Validation failed: " + err.Error(),
		})
		return
	}

	// Extract errors from validation result
	response := ValidateComponentConfigResponse{
		Valid:    validationResult.Valid(),
		Errors:   []string{},
		Warnings: []string{},
	}

	if !validationResult.Valid() {
		for _, resultError := range validationResult.Errors() {
			errorMsg := resultError.Description()
			if resultError.Field() != "" {
				errorMsg = fmt.Sprintf("%s: %s", resultError.Field(), errorMsg)
			}
			response.Errors = append(response.Errors, errorMsg)
		}
	}

	c.JSON(http.StatusOK, response)
}
