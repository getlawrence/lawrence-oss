package interfaces

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AppStorage handles application metadata storage (agents, groups, configs)
type AppStorage interface {
	// Agent management
	CreateAgent(ctx context.Context, agent *Agent) error
	GetAgent(ctx context.Context, id uuid.UUID) (*Agent, error)
	ListAgents(ctx context.Context) ([]*Agent, error)
	UpdateAgentStatus(ctx context.Context, id uuid.UUID, status AgentStatus) error
	UpdateAgentLastSeen(ctx context.Context, id uuid.UUID, lastSeen time.Time) error
	DeleteAgent(ctx context.Context, id uuid.UUID) error

	// Group management
	CreateGroup(ctx context.Context, group *Group) error
	GetGroup(ctx context.Context, id string) (*Group, error)
	ListGroups(ctx context.Context) ([]*Group, error)
	DeleteGroup(ctx context.Context, id string) error

	// Config management
	CreateConfig(ctx context.Context, config *Config) error
	GetConfig(ctx context.Context, id string) (*Config, error)
	GetLatestConfigForAgent(ctx context.Context, agentID uuid.UUID) (*Config, error)
	GetLatestConfigForGroup(ctx context.Context, groupID string) (*Config, error)
	ListConfigs(ctx context.Context, filter ConfigFilter) ([]*Config, error)

	// Close the storage connection
	Close() error
}

// TelemetryStorage handles telemetry data storage (metrics, logs, traces)
type TelemetryStorage interface {
	// Raw telemetry writes
	WriteMetrics(ctx context.Context, metrics []Metric) error
	WriteLogs(ctx context.Context, logs []Log) error
	WriteTraces(ctx context.Context, traces []Trace) error

	// Queries
	QueryMetrics(ctx context.Context, query MetricQuery) ([]Metric, error)
	QueryLogs(ctx context.Context, query LogQuery) ([]Log, error)
	QueryTraces(ctx context.Context, query TraceQuery) ([]Trace, error)

	// Rollups
	CreateRollups(ctx context.Context, window time.Time, interval RollupInterval) error
	QueryRollups(ctx context.Context, query RollupQuery) ([]Rollup, error)

	// Cleanup
	CleanupOldData(ctx context.Context, retention time.Duration) error

	// Close the storage connection
	Close() error
}

// Agent represents an OpenTelemetry agent
type Agent struct {
	ID           uuid.UUID         `json:"id"`
	Name         string            `json:"name"`
	Labels       map[string]string `json:"labels"`
	Status       AgentStatus       `json:"status"`
	LastSeen     time.Time         `json:"last_seen"`
	GroupID      *string           `json:"group_id,omitempty"`
	GroupName    *string           `json:"group_name,omitempty"`
	Version      string            `json:"version"`
	Capabilities []string          `json:"capabilities"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// AgentStatus represents the status of an agent
type AgentStatus string

const (
	AgentStatusOnline  AgentStatus = "online"
	AgentStatusOffline AgentStatus = "offline"
	AgentStatusError   AgentStatus = "error"
)

// Group represents a group of agents
type Group struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Labels    map[string]string `json:"labels"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// Config represents an agent configuration
type Config struct {
	ID         string    `json:"id"`
	AgentID    *uuid.UUID `json:"agent_id,omitempty"`
	GroupID    *string   `json:"group_id,omitempty"`
	ConfigHash string    `json:"config_hash"`
	Content    string    `json:"content"`
	Version    int       `json:"version"`
	CreatedAt  time.Time `json:"created_at"`
}

// ConfigFilter represents filters for listing configs
type ConfigFilter struct {
	AgentID *uuid.UUID
	GroupID *string
	Limit   int
}

// Metric represents a metric data point
type Metric struct {
	Timestamp  time.Time         `json:"timestamp"`
	AgentID    uuid.UUID         `json:"agent_id"`
	ConfigHash *string           `json:"config_hash,omitempty"`
	Name       string            `json:"name"`
	Value      float64           `json:"value"`
	Labels     map[string]string `json:"labels"`
	Type       MetricType        `json:"type"`
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
	Timestamp  time.Time         `json:"timestamp"`
	AgentID    uuid.UUID         `json:"agent_id"`
	ConfigHash *string           `json:"config_hash,omitempty"`
	Severity   string            `json:"severity"`
	Body       string            `json:"body"`
	Attributes map[string]string `json:"attributes"`
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
	AgentID   *uuid.UUID
	GroupID   *string
	MetricName *string
	StartTime time.Time
	EndTime   time.Time
	Limit     int
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
	WindowStart time.Time `json:"window_start"`
	AgentID     *uuid.UUID `json:"agent_id,omitempty"`
	GroupID     *string   `json:"group_id,omitempty"`
	MetricName  string    `json:"metric_name"`
	Count       int64     `json:"count"`
	Sum         float64   `json:"sum"`
	Avg         float64   `json:"avg"`
	Min         float64   `json:"min"`
	Max         float64   `json:"max"`
	Interval    RollupInterval `json:"interval"`
}

// RollupInterval represents the rollup time window
type RollupInterval string

const (
	RollupInterval1m  RollupInterval = "1m"
	RollupInterval5m  RollupInterval = "5m"
	RollupInterval1h  RollupInterval = "1h"
	RollupInterval1d  RollupInterval = "1d"
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
