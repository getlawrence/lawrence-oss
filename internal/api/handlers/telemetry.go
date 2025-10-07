package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/getlawrence/lawrence-oss/internal/services"
)

// TelemetryHandlers handles telemetry-related API endpoints
type TelemetryHandlers struct {
	telemetryService services.TelemetryQueryService
	logger           *zap.Logger
}

// NewTelemetryHandlers creates a new telemetry handlers instance
func NewTelemetryHandlers(telemetryService services.TelemetryQueryService, logger *zap.Logger) *TelemetryHandlers {
	return &TelemetryHandlers{
		telemetryService: telemetryService,
		logger:           logger,
	}
}

// QueryMetricsRequest represents the request for querying metrics
type QueryMetricsRequest struct {
	AgentID    *string   `json:"agent_id" binding:"omitempty,uuid"`
	GroupID    *string   `json:"group_id" binding:"omitempty,uuid"`
	MetricName *string   `json:"metric_name"`
	StartTime  time.Time `json:"start_time" binding:"required"`
	EndTime    time.Time `json:"end_time" binding:"required"`
	Limit      int       `json:"limit"`
	UseRollups bool      `json:"use_rollups"`
}

// QueryLogsRequest represents the request for querying logs
type QueryLogsRequest struct {
	AgentID   *string   `json:"agent_id" binding:"omitempty,uuid"`
	GroupID   *string   `json:"group_id" binding:"omitempty,uuid"`
	Severity  *string   `json:"severity"`
	Search    *string   `json:"search"`
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time" binding:"required"`
	Limit     int       `json:"limit"`
}

// QueryTracesRequest represents the request for querying traces
type QueryTracesRequest struct {
	AgentID     *string   `json:"agent_id" binding:"omitempty,uuid"`
	GroupID     *string   `json:"group_id" binding:"omitempty,uuid"`
	TraceID     *string   `json:"trace_id"`
	ServiceName *string   `json:"service_name"`
	StartTime   time.Time `json:"start_time" binding:"required"`
	EndTime     time.Time `json:"end_time" binding:"required"`
	Limit       int       `json:"limit"`
}

// TelemetryOverviewResponse represents the telemetry overview
type TelemetryOverviewResponse struct {
	TotalMetrics int64     `json:"totalMetrics"`
	TotalLogs    int64     `json:"totalLogs"`
	TotalTraces  int64     `json:"totalTraces"`
	ActiveAgents int       `json:"activeAgents"`
	Services     []string  `json:"services"`
	LastUpdated  time.Time `json:"lastUpdated"`
}

// ServicesResponse represents the services list
type ServicesResponse struct {
	Services []string `json:"services"`
	Count    int      `json:"count"`
}

// handleQueryMetrics handles POST /api/v1/telemetry/metrics/query
func (h *TelemetryHandlers) HandleQueryMetrics(c *gin.Context) {
	var req QueryMetricsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// Set default limit if not provided
	if req.Limit == 0 {
		req.Limit = 1000
	}
	if req.Limit > 10000 {
		req.Limit = 10000
	}

	// Convert agent ID from string to UUID
	var agentID *uuid.UUID
	if req.AgentID != nil {
		parsedID, err := uuid.Parse(*req.AgentID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID format"})
			return
		}
		agentID = &parsedID
	}

	// Convert request to service query
	query := services.MetricQuery{
		AgentID:    agentID,
		GroupID:    req.GroupID,
		MetricName: req.MetricName,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		Limit:      req.Limit,
	}

	// Execute query through service
	metrics, err := h.telemetryService.QueryMetrics(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to query metrics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query metrics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
		"count":   len(metrics),
	})
}

// handleQueryLogs handles POST /api/v1/telemetry/logs/query
func (h *TelemetryHandlers) HandleQueryLogs(c *gin.Context) {
	var req QueryLogsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// Set default limit if not provided
	if req.Limit == 0 {
		req.Limit = 1000
	}
	if req.Limit > 10000 {
		req.Limit = 10000
	}

	// Convert agent ID from string to UUID
	var agentID *uuid.UUID
	if req.AgentID != nil {
		parsedID, err := uuid.Parse(*req.AgentID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID format"})
			return
		}
		agentID = &parsedID
	}

	// Convert request to service query
	query := services.LogQuery{
		AgentID:   agentID,
		GroupID:   req.GroupID,
		Severity:  req.Severity,
		Search:    req.Search,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Limit:     req.Limit,
	}

	// Execute query through service
	logs, err := h.telemetryService.QueryLogs(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to query logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"count": len(logs),
	})
}

// handleQueryTraces handles POST /api/v1/telemetry/traces/query
func (h *TelemetryHandlers) HandleQueryTraces(c *gin.Context) {
	var req QueryTracesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// Set default limit if not provided
	if req.Limit == 0 {
		req.Limit = 1000
	}
	if req.Limit > 10000 {
		req.Limit = 10000
	}

	// Convert agent ID from string to UUID
	var agentID *uuid.UUID
	if req.AgentID != nil {
		parsedID, err := uuid.Parse(*req.AgentID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID format"})
			return
		}
		agentID = &parsedID
	}

	// Convert request to service query
	query := services.TraceQuery{
		AgentID:   agentID,
		GroupID:   req.GroupID,
		TraceID:   req.TraceID,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Limit:     req.Limit,
	}

	// Execute query through service
	traces, err := h.telemetryService.QueryTraces(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to query traces", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query traces"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"traces": traces,
		"count":  len(traces),
	})
}

// handleGetTelemetryOverview handles GET /api/v1/telemetry/overview
func (h *TelemetryHandlers) HandleGetTelemetryOverview(c *gin.Context) {
	ctx := c.Request.Context()

	// Get overview from service
	overview, err := h.telemetryService.GetTelemetryOverview(ctx)
	if err != nil {
		h.logger.Error("Failed to get telemetry overview", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get telemetry overview"})
		return
	}

	response := TelemetryOverviewResponse{
		TotalMetrics: overview.TotalMetrics,
		TotalLogs:    overview.TotalLogs,
		TotalTraces:  overview.TotalTraces,
		ActiveAgents: overview.ActiveAgents,
		Services:     overview.Services,
		LastUpdated:  overview.LastUpdated,
	}

	c.JSON(http.StatusOK, response)
}

// handleGetServices handles GET /api/v1/telemetry/services
func (h *TelemetryHandlers) HandleGetServices(c *gin.Context) {
	ctx := c.Request.Context()

	// Get services from service
	services, err := h.telemetryService.GetServices(ctx)
	if err != nil {
		h.logger.Error("Failed to get services", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get services"})
		return
	}

	response := ServicesResponse{
		Services: services,
		Count:    len(services),
	}

	c.JSON(http.StatusOK, response)
}

// parseTimeRange parses time range string to duration
func parseTimeRange(timeRange string) (time.Duration, error) {
	switch timeRange {
	case "5m":
		return 5 * time.Minute, nil
	case "15m":
		return 15 * time.Minute, nil
	case "1h":
		return 1 * time.Hour, nil
	case "6h":
		return 6 * time.Hour, nil
	case "24h":
		return 24 * time.Hour, nil
	case "7d":
		return 7 * 24 * time.Hour, nil
	default:
		// Try to parse as duration string
		return time.ParseDuration(timeRange)
	}
}
