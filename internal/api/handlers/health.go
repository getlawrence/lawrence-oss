package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/getlawrence/lawrence-oss/internal/services"
)

// HealthHandlers handles health check endpoints
type HealthHandlers struct {
	agentService     services.AgentService
	telemetryService services.TelemetryQueryService
	logger           *zap.Logger
}

// NewHealthHandlers creates a new health handlers instance
func NewHealthHandlers(agentService services.AgentService, telemetryService services.TelemetryQueryService, logger *zap.Logger) *HealthHandlers {
	return &HealthHandlers{
		agentService:     agentService,
		telemetryService: telemetryService,
		logger:           logger,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time        `json:"timestamp"`
	Version   string           `json:"version"`
	Services  map[string]string `json:"services"`
}

// handleHealth handles GET /health
func (h *HealthHandlers) HandleHealth(c *gin.Context) {
	// Check storage health
	sqliteHealthy := h.checkSQLiteHealth(c)
	duckdbHealthy := h.checkDuckDBHealth(c)

	// Determine overall status
	status := "healthy"
	if !sqliteHealthy || !duckdbHealthy {
		status = "unhealthy"
	}

	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Version:   "0.1.0",
		Services: map[string]string{
			"sqlite":  h.getHealthStatus(sqliteHealthy),
			"duckdb":  h.getHealthStatus(duckdbHealthy),
		},
	}

	// Set appropriate HTTP status code
	httpStatus := http.StatusOK
	if status == "unhealthy" {
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, response)
}

// checkSQLiteHealth checks if SQLite is healthy
func (h *HealthHandlers) checkSQLiteHealth(c *gin.Context) bool {
	// Try to get a simple count from agents table
	_, err := h.agentService.ListAgents(c.Request.Context())
	return err == nil
}

// checkDuckDBHealth checks if DuckDB is healthy
func (h *HealthHandlers) checkDuckDBHealth(c *gin.Context) bool {
	// Try to query a simple metric count
	query := services.MetricQuery{
		StartTime: time.Now().Add(-1 * time.Minute),
		EndTime:   time.Now(),
		Limit:     1,
	}
	_, err := h.telemetryService.QueryMetrics(c.Request.Context(), query)
	return err == nil
}

// getHealthStatus converts boolean to status string
func (h *HealthHandlers) getHealthStatus(healthy bool) string {
	if healthy {
		return "healthy"
	}
	return "unhealthy"
}
