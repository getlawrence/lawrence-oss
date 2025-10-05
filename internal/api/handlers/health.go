package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/getlawrence/lawrence-oss/internal/storage"
	"github.com/getlawrence/lawrence-oss/internal/storage/interfaces"
)

// HealthHandlers handles health check endpoints
type HealthHandlers struct {
	storage *storage.Container
	logger  *zap.Logger
}

// NewHealthHandlers creates a new health handlers instance
func NewHealthHandlers(storage *storage.Container, logger *zap.Logger) *HealthHandlers {
	return &HealthHandlers{
		storage: storage,
		logger:  logger,
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
	sqliteHealthy := h.checkSQLiteHealth()
	duckdbHealthy := h.checkDuckDBHealth()

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
func (h *HealthHandlers) checkSQLiteHealth() bool {
	// Try to get a simple count from agents table
	_, err := h.storage.App.ListAgents(nil)
	return err == nil
}

// checkDuckDBHealth checks if DuckDB is healthy
func (h *HealthHandlers) checkDuckDBHealth() bool {
	// Try to query a simple metric count
	query := interfaces.MetricQuery{
		StartTime: time.Now().Add(-1 * time.Minute),
		EndTime:   time.Now(),
		Limit:     1,
	}
	_, err := h.storage.Telemetry.QueryMetrics(nil, query)
	return err == nil
}

// getHealthStatus converts boolean to status string
func (h *HealthHandlers) getHealthStatus(healthy bool) string {
	if healthy {
		return "healthy"
	}
	return "unhealthy"
}
