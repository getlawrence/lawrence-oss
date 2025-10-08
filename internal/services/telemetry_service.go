package services

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// TelemetryQueryService defines the interface for telemetry query operations
type TelemetryQueryService interface {
	// Query operations
	QueryMetrics(ctx context.Context, query MetricQuery) ([]Metric, error)
	QueryLogs(ctx context.Context, query LogQuery) ([]Log, error)
	QueryTraces(ctx context.Context, query TraceQuery) ([]Trace, error)
	QueryRaw(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error)

	// Rollup operations
	CreateRollups(ctx context.Context, window time.Time, interval RollupInterval) error
	QueryRollups(ctx context.Context, query RollupQuery) ([]Rollup, error)

	// Cleanup operations
	CleanupOldData(ctx context.Context, retention time.Duration) error

	// Overview operations
	GetTelemetryOverview(ctx context.Context) (*TelemetryOverview, error)
	GetServices(ctx context.Context) ([]string, error)
}

// Metric represents a metric data point
type Metric struct {
	Timestamp        time.Time              `json:"timestamp"`
	AgentID          uuid.UUID              `json:"agent_id"`
	GroupID          *string                `json:"group_id,omitempty"`
	ServiceName      string                 `json:"service_name"`
	Name             string                 `json:"metric_name"`
	Value            float64                `json:"value"`
	MetricAttributes map[string]interface{} `json:"metric_attributes"`
	ConfigHash       *string                `json:"config_hash,omitempty"`
	Labels           map[string]string      `json:"labels,omitempty"`
	Type             MetricType             `json:"type,omitempty"`
}

// MetricType represents the type of metric
type MetricType string

const (
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeCounter   MetricType = "counter"
	MetricTypeHistogram MetricType = "histogram"
)

// Log represents a log entry
type Log struct {
	Timestamp      time.Time              `json:"timestamp"`
	AgentID        uuid.UUID              `json:"agent_id"`
	GroupID        *string                `json:"group_id,omitempty"`
	ServiceName    string                 `json:"service_name"`
	SeverityText   string                 `json:"severity_text"`
	SeverityNumber int                    `json:"severity_number"`
	Body           string                 `json:"body"`
	TraceID        *string                `json:"trace_id,omitempty"`
	SpanID         *string                `json:"span_id,omitempty"`
	LogAttributes  map[string]interface{} `json:"log_attributes"`
	ConfigHash     *string                `json:"config_hash,omitempty"`
	// Deprecated: use SeverityText instead
	Severity string `json:"severity,omitempty"`
	// Deprecated: use LogAttributes instead
	Attributes map[string]string `json:"attributes,omitempty"`
}

// Trace represents a trace span
type Trace struct {
	Timestamp     time.Time         `json:"timestamp"`
	AgentID       uuid.UUID         `json:"agent_id"`
	ConfigHash    *string           `json:"config_hash,omitempty"`
	TraceID       string            `json:"trace_id"`
	SpanID        string            `json:"span_id"`
	ParentSpanID  *string           `json:"parent_span_id,omitempty"`
	Name          string            `json:"name"`
	Duration      int64             `json:"duration"`
	StatusCode    string            `json:"status_code"`
	StatusMessage string            `json:"status_message"`
	Attributes    map[string]string `json:"attributes"`
}

// MetricQuery represents a query for metrics
type MetricQuery struct {
	AgentID    *uuid.UUID
	GroupID    *string
	MetricName *string
	StartTime  time.Time
	EndTime    time.Time
	Limit      int
}

// LogQuery represents a query for logs
type LogQuery struct {
	AgentID   *uuid.UUID
	GroupID   *string
	Severity  *string
	Search    *string
	StartTime time.Time
	EndTime   time.Time
	Limit     int
}

// TraceQuery represents a query for traces
type TraceQuery struct {
	AgentID   *uuid.UUID
	GroupID   *string
	TraceID   *string
	StartTime time.Time
	EndTime   time.Time
	Limit     int
}

// Rollup represents pre-aggregated data
type Rollup struct {
	WindowStart time.Time      `json:"window_start"`
	AgentID     *uuid.UUID     `json:"agent_id,omitempty"`
	GroupID     *string        `json:"group_id,omitempty"`
	MetricName  string         `json:"metric_name"`
	Count       int64          `json:"count"`
	Sum         float64        `json:"sum"`
	Avg         float64        `json:"avg"`
	Min         float64        `json:"min"`
	Max         float64        `json:"max"`
	Interval    RollupInterval `json:"interval"`
}

// RollupInterval represents the rollup time window
type RollupInterval string

const (
	RollupInterval1m RollupInterval = "1m"
	RollupInterval5m RollupInterval = "5m"
	RollupInterval1h RollupInterval = "1h"
	RollupInterval1d RollupInterval = "1d"
)

// RollupQuery represents a query for rollups
type RollupQuery struct {
	AgentID    *uuid.UUID
	GroupID    *string
	MetricName *string
	StartTime  time.Time
	EndTime    time.Time
	Interval   RollupInterval
}

// TelemetryOverview represents the telemetry overview
type TelemetryOverview struct {
	TotalMetrics int64     `json:"totalMetrics"`
	TotalLogs    int64     `json:"totalLogs"`
	TotalTraces  int64     `json:"totalTraces"`
	ActiveAgents int       `json:"activeAgents"`
	Services     []string  `json:"services"`
	LastUpdated  time.Time `json:"lastUpdated"`
}
