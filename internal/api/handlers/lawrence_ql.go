// Copyright (c) 2024 Lawrence OSS Contributors
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/getlawrence/lawrence-oss/internal/query"
	"github.com/getlawrence/lawrence-oss/internal/services"
)

// LawrenceQLHandlers handles Lawrence QL query endpoints
type LawrenceQLHandlers struct {
	telemetryService services.TelemetryQueryService
	executor         *query.Executor
	logger           *zap.Logger
}

// NewLawrenceQLHandlers creates a new Lawrence QL handlers instance
func NewLawrenceQLHandlers(telemetryService services.TelemetryQueryService, logger *zap.Logger) *LawrenceQLHandlers {
	return &LawrenceQLHandlers{
		telemetryService: telemetryService,
		executor:         query.NewExecutor(telemetryService, logger),
		logger:           logger,
	}
}

// LawrenceQLRequest represents a Lawrence QL query request
type LawrenceQLRequest struct {
	Query     string     `json:"query" binding:"required"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	AgentID   *string    `json:"agent_id,omitempty"`
	GroupID   *string    `json:"group_id,omitempty"`
}

// LawrenceQLResponse represents a Lawrence QL query response
type LawrenceQLResponse struct {
	Results []query.QueryResult `json:"results"`
	Meta    query.QueryMeta     `json:"meta"`
}

// ValidateQueryRequest represents a query validation request
type ValidateQueryRequest struct {
	Query string `json:"query" binding:"required"`
}

// ValidateQueryResponse represents a query validation response
type ValidateQueryResponse struct {
	Valid   bool   `json:"valid"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// SuggestionsRequest represents a request for query suggestions
type SuggestionsRequest struct {
	Query      string `json:"query"`
	CursorPos  int    `json:"cursor_pos"`
}

// SuggestionsResponse represents query suggestions response
type SuggestionsResponse struct {
	Suggestions []string `json:"suggestions"`
}

// HandleLawrenceQLQuery handles POST /api/v1/telemetry/query
func (h *LawrenceQLHandlers) HandleLawrenceQLQuery(c *gin.Context) {
	var req LawrenceQLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// Parse the query
	parser := query.NewParser(req.Query)
	parsedQuery, err := parser.Parse()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query syntax", "details": err.Error()})
		return
	}

	// Build execution context
	execCtx := &query.ExecutionContext{
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Limit:     req.Limit,
	}

	// Parse agent ID if provided
	if req.AgentID != nil {
		agentID, err := uuid.Parse(*req.AgentID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID format"})
			return
		}
		execCtx.AgentID = &agentID
	}

	// Set group ID if provided
	if req.GroupID != nil {
		execCtx.GroupID = req.GroupID
	}

	// Set default limit if not provided
	if execCtx.Limit == 0 {
		execCtx.Limit = 1000
	}
	if execCtx.Limit > 10000 {
		execCtx.Limit = 10000
	}

	// Execute the query
	results, meta, err := h.executor.Execute(c.Request.Context(), parsedQuery, execCtx)
	if err != nil {
		h.logger.Error("Failed to execute Lawrence QL query",
			zap.Error(err),
			zap.String("query", req.Query))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute query", "details": err.Error()})
		return
	}

	response := LawrenceQLResponse{
		Results: results,
		Meta:    *meta,
	}

	c.JSON(http.StatusOK, response)
}

// HandleValidateQuery handles POST /api/v1/telemetry/query/validate
func (h *LawrenceQLHandlers) HandleValidateQuery(c *gin.Context) {
	var req ValidateQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// Validate the query
	err := query.ValidateQuery(req.Query)
	if err != nil {
		c.JSON(http.StatusOK, ValidateQueryResponse{
			Valid: false,
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ValidateQueryResponse{
		Valid:   true,
		Message: "Query syntax is valid",
	})
}

// HandleGetSuggestions handles POST /api/v1/telemetry/query/suggestions
func (h *LawrenceQLHandlers) HandleGetSuggestions(c *gin.Context) {
	var req SuggestionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// Get suggestions
	suggestions := query.GetQuerySuggestions(req.Query, req.CursorPos)

	// Add function suggestions
	functionSuggestions := query.GetFunctionSuggestions()
	suggestions = append(suggestions, functionSuggestions...)

	c.JSON(http.StatusOK, SuggestionsResponse{
		Suggestions: suggestions,
	})
}

// TemplateInfo represents a query template
type TemplateInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Query       string `json:"query"`
	Category    string `json:"category"`
}

// TemplatesResponse represents query templates response
type TemplatesResponse struct {
	Templates []TemplateInfo `json:"templates"`
}

// HandleGetTemplates handles GET /api/v1/telemetry/query/templates
func (h *LawrenceQLHandlers) HandleGetTemplates(c *gin.Context) {
	templates := []TemplateInfo{
		{
			ID:          "metrics-5m",
			Name:        "Recent Metrics (5m)",
			Description: "Query all metrics from the last 5 minutes",
			Query:       `metrics{} [5m]`,
			Category:    "metrics",
		},
		{
			ID:          "metrics-by-service",
			Name:        "Metrics by Service",
			Description: "Query metrics for a specific service",
			Query:       `metrics{service="api"} [1h]`,
			Category:    "metrics",
		},
		{
			ID:          "logs-errors",
			Name:        "Error Logs",
			Description: "Query error logs from the last hour",
			Query:       `logs{severity="error"} [1h]`,
			Category:    "logs",
		},
		{
			ID:          "logs-search",
			Name:        "Search Logs",
			Description: "Search logs by message content",
			Query:       `logs{body=~".*error.*"} [1h]`,
			Category:    "logs",
		},
		{
			ID:          "traces-service",
			Name:        "Traces by Service",
			Description: "Query traces for a specific service",
			Query:       `traces{service="api"} [30m]`,
			Category:    "traces",
		},
		{
			ID:          "metrics-sum",
			Name:        "Sum Metrics",
			Description: "Calculate sum of metric values",
			Query:       `sum(metrics{metric="requests_total"} [5m])`,
			Category:    "aggregation",
		},
		{
			ID:          "metrics-avg-by-service",
			Name:        "Average by Service",
			Description: "Calculate average metric value grouped by service",
			Query:       `avg(metrics{metric="response_time"} [1h]) by (service)`,
			Category:    "aggregation",
		},
		{
			ID:          "metrics-rate",
			Name:        "Request Rate",
			Description: "Calculate per-second rate of requests",
			Query:       `rate(metrics{metric="requests_total"} [5m])`,
			Category:    "functions",
		},
		{
			ID:          "metrics-increase",
			Name:        "Total Increase",
			Description: "Calculate total increase over time range",
			Query:       `increase(metrics{metric="requests_total"} [1h])`,
			Category:    "functions",
		},
		{
			ID:          "metrics-histogram",
			Name:        "Histogram Quantiles",
			Description: "Calculate histogram quantiles (p50, p95, p99)",
			Query:       `histogram_quantile(metrics{metric="response_time"} [5m])`,
			Category:    "functions",
		},
	}

	c.JSON(http.StatusOK, TemplatesResponse{
		Templates: templates,
	})
}

// FunctionInfo represents a function description
type FunctionInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Example     string `json:"example"`
}

// FunctionsResponse represents available functions response
type FunctionsResponse struct {
	Functions []FunctionInfo `json:"functions"`
}

// HandleGetFunctions handles GET /api/v1/telemetry/query/functions
func (h *LawrenceQLHandlers) HandleGetFunctions(c *gin.Context) {
	functions := []FunctionInfo{
		{
			Name:        "sum",
			Description: "Calculates the sum of values",
			Example:     "sum(metrics{metric=\"cpu_usage\"} [5m])",
		},
		{
			Name:        "avg",
			Description: "Calculates the average of values",
			Example:     "avg(metrics{metric=\"memory_usage\"} [5m])",
		},
		{
			Name:        "min",
			Description: "Returns the minimum value",
			Example:     "min(metrics{metric=\"response_time\"} [5m])",
		},
		{
			Name:        "max",
			Description: "Returns the maximum value",
			Example:     "max(metrics{metric=\"response_time\"} [5m])",
		},
		{
			Name:        "count",
			Description: "Counts the number of results",
			Example:     "count(logs{severity=\"error\"} [1h])",
		},
		{
			Name:        "rate",
			Description: "Calculates the per-second rate of increase",
			Example:     "rate(metrics{metric=\"requests_total\"} [5m])",
		},
		{
			Name:        "increase",
			Description: "Calculates the total increase over the time range",
			Example:     "increase(metrics{metric=\"requests_total\"} [1h])",
		},
		{
			Name:        "histogram_quantile",
			Description: "Calculates histogram quantiles (p50, p90, p95, p99)",
			Example:     "histogram_quantile(metrics{metric=\"response_time\"} [5m])",
		},
	}

	c.JSON(http.StatusOK, FunctionsResponse{
		Functions: functions,
	})
}
