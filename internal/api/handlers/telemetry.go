package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/getlawrence/lawrence-oss/internal/storage"
)

// TelemetryHandlers handles telemetry-related API endpoints
type TelemetryHandlers struct {
	storage *storage.Container
	logger  *zap.Logger
}

// NewTelemetryHandlers creates a new telemetry handlers instance
func NewTelemetryHandlers(storage *storage.Container, logger *zap.Logger) *TelemetryHandlers {
	return &TelemetryHandlers{
		storage: storage,
		logger:  logger,
	}
}

// QueryMetricsRequest represents the request for querying metrics
type QueryMetricsRequest struct {
	AgentID     *string    `json:"agent_id" binding:"omitempty,uuid"`
	GroupID     *string    `json:"group_id" binding:"omitempty,uuid"`
	MetricName  *string    `json:"metric_name"`
	StartTime   time.Time  `json:"start_time" binding:"required"`
	EndTime     time.Time  `json:"end_time" binding:"required"`
	Limit       int        `json:"limit"`
	UseRollups  bool       `json:"use_rollups"`
}

// QueryLogsRequest represents the request for querying logs
type QueryLogsRequest struct {
	AgentID     *string    `json:"agent_id" binding:"omitempty,uuid"`
	GroupID     *string    `json:"group_id" binding:"omitempty,uuid"`
	Severity    *string    `json:"severity"`
	Search      *string    `json:"search"`
	StartTime   time.Time  `json:"start_time" binding:"required"`
	EndTime     time.Time  `json:"end_time" binding:"required"`
	Limit       int        `json:"limit"`
}

// QueryTracesRequest represents the request for querying traces
type QueryTracesRequest struct {
	AgentID     *string    `json:"agent_id" binding:"omitempty,uuid"`
	GroupID     *string    `json:"group_id" binding:"omitempty,uuid"`
	TraceID     *string    `json:"trace_id"`
	ServiceName *string    `json:"service_name"`
	StartTime   time.Time  `json:"start_time" binding:"required"`
	EndTime     time.Time  `json:"end_time" binding:"required"`
	Limit       int        `json:"limit"`
}

// TelemetryOverviewResponse represents the telemetry overview
type TelemetryOverviewResponse struct {
	TotalMetrics    int64     `json:"totalMetrics"`
	TotalLogs       int64     `json:"totalLogs"`
	TotalTraces     int64     `json:"totalTraces"`
	ActiveAgents    int       `json:"activeAgents"`
	Services        []string  `json:"services"`
	LastUpdated     time.Time `json:"lastUpdated"`
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

	// Build query
	query := `
		SELECT timestamp, agent_id, group_id, service_name, metric_name, value, metric_attributes
		FROM (
			SELECT timestamp, agent_id, group_id, service_name, metric_name, value, metric_attributes FROM metrics_sum
			UNION ALL
			SELECT timestamp, agent_id, group_id, service_name, metric_name, value, metric_attributes FROM metrics_gauge
		) AS all_metrics
		WHERE timestamp >= ? AND timestamp <= ?
	`
	args := []interface{}{req.StartTime, req.EndTime}

	if req.AgentID != nil {
		query += ` AND agent_id = ?`
		args = append(args, *req.AgentID)
	}

	if req.GroupID != nil {
		query += ` AND group_id = ?`
		args = append(args, *req.GroupID)
	}

	if req.MetricName != nil {
		query += ` AND metric_name = ?`
		args = append(args, *req.MetricName)
	}

	query += ` ORDER BY timestamp DESC LIMIT ?`
	args = append(args, req.Limit)

	// Execute query
	rows, err := h.storage.Telemetry.QueryRaw(c.Request.Context(), query, args...)
	if err != nil {
		h.logger.Error("Failed to query metrics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query metrics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics": rows,
		"count":   len(rows),
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

	// Build query
	query := `
		SELECT timestamp, agent_id, group_id, service_name, severity_text, severity_number, body, trace_id, span_id, log_attributes
		FROM logs
		WHERE timestamp >= ? AND timestamp <= ?
	`
	args := []interface{}{req.StartTime, req.EndTime}

	if req.AgentID != nil {
		query += ` AND agent_id = ?`
		args = append(args, *req.AgentID)
	}

	if req.GroupID != nil {
		query += ` AND group_id = ?`
		args = append(args, *req.GroupID)
	}

	if req.Severity != nil {
		query += ` AND severity_text = ?`
		args = append(args, *req.Severity)
	}

	if req.Search != nil && *req.Search != "" {
		query += ` AND body LIKE ?`
		args = append(args, "%"+*req.Search+"%")
	}

	query += ` ORDER BY timestamp DESC LIMIT ?`
	args = append(args, req.Limit)

	// Execute query
	rows, err := h.storage.Telemetry.QueryRaw(c.Request.Context(), query, args...)
	if err != nil {
		h.logger.Error("Failed to query logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  rows,
		"count": len(rows),
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

	// Build query
	query := `
		SELECT timestamp, agent_id, group_id, trace_id, span_id, parent_span_id,
		       service_name, span_name, duration, status_code, span_attributes
		FROM traces
		WHERE timestamp >= ? AND timestamp <= ?
	`
	args := []interface{}{req.StartTime, req.EndTime}

	if req.AgentID != nil {
		query += ` AND agent_id = ?`
		args = append(args, *req.AgentID)
	}

	if req.GroupID != nil {
		query += ` AND group_id = ?`
		args = append(args, *req.GroupID)
	}

	if req.TraceID != nil {
		query += ` AND trace_id = ?`
		args = append(args, *req.TraceID)
	}

	if req.ServiceName != nil {
		query += ` AND service_name = ?`
		args = append(args, *req.ServiceName)
	}

	query += ` ORDER BY timestamp DESC LIMIT ?`
	args = append(args, req.Limit)

	// Execute query
	rows, err := h.storage.Telemetry.QueryRaw(c.Request.Context(), query, args...)
	if err != nil {
		h.logger.Error("Failed to query traces", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query traces"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"traces": rows,
		"count":  len(rows),
	})
}

// handleGetTelemetryOverview handles GET /api/v1/telemetry/overview
func (h *TelemetryHandlers) HandleGetTelemetryOverview(c *gin.Context) {
	ctx := c.Request.Context()

	// Get counts from DuckDB
	var metricsCount, logsCount, tracesCount int64

	// Count metrics
	metricsQuery := `SELECT COUNT(*) FROM (SELECT 1 FROM metrics_sum UNION ALL SELECT 1 FROM metrics_gauge) AS all_metrics`
	if rows, err := h.storage.Telemetry.QueryRaw(ctx, metricsQuery); err == nil && len(rows) > 0 {
		if count, ok := rows[0]["count"].(int64); ok {
			metricsCount = count
		}
	}

	// Count logs
	logsQuery := `SELECT COUNT(*) FROM logs`
	if rows, err := h.storage.Telemetry.QueryRaw(ctx, logsQuery); err == nil && len(rows) > 0 {
		if count, ok := rows[0]["count"].(int64); ok {
			logsCount = count
		}
	}

	// Count traces
	tracesQuery := `SELECT COUNT(*) FROM traces`
	if rows, err := h.storage.Telemetry.QueryRaw(ctx, tracesQuery); err == nil && len(rows) > 0 {
		if count, ok := rows[0]["count"].(int64); ok {
			tracesCount = count
		}
	}

	// Get active agents count
	agents, _ := h.storage.App.ListAgents(ctx)
	activeAgents := 0
	for _, agent := range agents {
		if agent.Status == "online" {
			activeAgents++
		}
	}

	// Get unique services
	servicesQuery := `
		SELECT DISTINCT service_name FROM (
			SELECT service_name FROM metrics_sum
			UNION
			SELECT service_name FROM metrics_gauge
			UNION
			SELECT service_name FROM logs
			UNION
			SELECT service_name FROM traces
		) AS all_services
		WHERE service_name IS NOT NULL
		ORDER BY service_name
	`
	var services []string
	if rows, err := h.storage.Telemetry.QueryRaw(ctx, servicesQuery); err == nil {
		for _, row := range rows {
			if svc, ok := row["service_name"].(string); ok && svc != "" {
				services = append(services, svc)
			}
		}
	}

	response := TelemetryOverviewResponse{
		TotalMetrics: metricsCount,
		TotalLogs:    logsCount,
		TotalTraces:  tracesCount,
		ActiveAgents: activeAgents,
		Services:     services,
		LastUpdated:  time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// handleGetServices handles GET /api/v1/telemetry/services
func (h *TelemetryHandlers) HandleGetServices(c *gin.Context) {
	ctx := c.Request.Context()

	// Get unique services from all telemetry tables
	query := `
		SELECT DISTINCT service_name FROM (
			SELECT service_name FROM metrics_sum
			UNION
			SELECT service_name FROM metrics_gauge
			UNION
			SELECT service_name FROM logs
			UNION
			SELECT service_name FROM traces
		) AS all_services
		WHERE service_name IS NOT NULL AND service_name != ''
		ORDER BY service_name
	`

	rows, err := h.storage.Telemetry.QueryRaw(ctx, query)
	if err != nil {
		h.logger.Error("Failed to get services", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get services"})
		return
	}

	var services []string
	for _, row := range rows {
		if svc, ok := row["service_name"].(string); ok && svc != "" {
			services = append(services, svc)
		}
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
