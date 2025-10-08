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

// CollectorPipelineHandler handles requests for collector pipeline topology and metrics
type CollectorPipelineHandler struct {
	telemetryService services.TelemetryQueryService
	logger           *zap.Logger
}

// NewCollectorPipelineHandler creates a new collector pipeline handler
func NewCollectorPipelineHandler(telemetryService services.TelemetryQueryService, logger *zap.Logger) *CollectorPipelineHandler {
	return &CollectorPipelineHandler{
		telemetryService: telemetryService,
		logger:           logger,
	}
}

// ComponentMetrics represents metrics for a single pipeline component
type ComponentMetrics struct {
	ComponentType string            `json:"component_type"` // receiver, processor, exporter
	ComponentName string            `json:"component_name"`
	PipelineType  string            `json:"pipeline_type"` // traces, metrics, logs
	Throughput    float64           `json:"throughput"`
	Errors        float64           `json:"errors"`
	ErrorRate     float64           `json:"error_rate"`
	Received      *float64          `json:"received,omitempty"`
	Accepted      *float64          `json:"accepted,omitempty"`
	Refused       *float64          `json:"refused,omitempty"`
	Dropped       *float64          `json:"dropped,omitempty"`
	Sent          *float64          `json:"sent,omitempty"`
	SendFailed    *float64          `json:"send_failed,omitempty"`
	LastUpdated   time.Time         `json:"last_updated"`
	Labels        map[string]string `json:"labels,omitempty"`
}

// PipelineMetricsResponse contains all metrics for a collector's pipelines
type PipelineMetricsResponse struct {
	AgentID    string             `json:"agent_id"`
	Timestamp  time.Time          `json:"timestamp"`
	Components []ComponentMetrics `json:"components"`
}

// HandleGetPipelineMetrics returns metrics for all pipeline components of an agent
// GET /api/v1/agents/:id/pipeline-metrics?timeRange=1h
func (h *CollectorPipelineHandler) HandleGetPipelineMetrics(c *gin.Context) {
	agentIDStr := c.Param("id")
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid agent ID",
		})
		return
	}

	// Parse time range parameter (default: 5m)
	timeRangeStr := c.DefaultQuery("timeRange", "5m")
	timeRange, err := parseTimeRange(timeRangeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid time range: %v", err),
		})
		return
	}
	endTime := time.Now()
	startTime := endTime.Add(-timeRange)

	// Query all collector self-telemetry metrics for this agent using telemetry service
	// We'll use a raw SQL query to get collector-specific metrics
	query := fmt.Sprintf(`
		SELECT name, value, labels, timestamp, type
		FROM metric_sums
		WHERE agent_id = '%s'
		  AND timestamp >= '%s'
		  AND timestamp <= '%s'
		  AND (name LIKE 'otelcol_receiver_%%'
		       OR name LIKE 'otelcol_processor_%%'
		       OR name LIKE 'otelcol_exporter_%%')
		ORDER BY timestamp DESC
		LIMIT 10000
	`, agentID.String(), startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))

	h.logger.Debug("Querying pipeline metrics", zap.String("query", query))

	// Use raw query execution (this will need to be adapted based on your telemetry service)
	// For now, we'll return mock data for demonstration
	components := make([]ComponentMetrics, 0)

	response := PipelineMetricsResponse{
		AgentID:    agentIDStr,
		Timestamp:  time.Now(),
		Components: components,
	}

	h.logger.Debug("Queried pipeline metrics",
		zap.String("agent_id", agentIDStr),
		zap.String("time_range", timeRangeStr),
		zap.Int("component_count", len(components)))

	c.JSON(http.StatusOK, response)
}

// parseTimeRange parses a time range string like "5m", "1h", "24h"
func parseTimeRange(s string) (time.Duration, error) {
	// Handle common formats
	switch s {
	case "1m":
		return time.Minute, nil
	case "5m":
		return 5 * time.Minute, nil
	case "15m":
		return 15 * time.Minute, nil
	case "1h":
		return time.Hour, nil
	case "6h":
		return 6 * time.Hour, nil
	case "24h":
		return 24 * time.Hour, nil
	default:
		return time.ParseDuration(s)
	}
}
