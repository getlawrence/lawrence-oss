package types

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ApplicationStore interface for managing application data
type ApplicationStore interface {
	CreateAgent(ctx context.Context, agent *Agent) error
	GetAgent(ctx context.Context, id uuid.UUID) (*Agent, error)
	ListAgents(ctx context.Context) ([]*Agent, error)
	UpdateAgentStatus(ctx context.Context, id uuid.UUID, status AgentStatus) error
	UpdateAgentLastSeen(ctx context.Context, id uuid.UUID, lastSeen time.Time) error
	UpdateAgentEffectiveConfig(ctx context.Context, id uuid.UUID, effectiveConfig string) error
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
	ListConfigs(ctx context.Context, filter ConfigFilter) (*ListConfigsResult, error)

	// Workflow management
	CreateWorkflow(ctx context.Context, workflow *Workflow) error
	GetWorkflow(ctx context.Context, id string) (*Workflow, error)
	ListWorkflows(ctx context.Context, filter WorkflowFilter) ([]*Workflow, error)
	UpdateWorkflow(ctx context.Context, workflow *Workflow) error
	DeleteWorkflow(ctx context.Context, id string) error

	// Workflow execution tracking
	CreateWorkflowExecution(ctx context.Context, execution *WorkflowExecution) error
	GetWorkflowExecution(ctx context.Context, id string) (*WorkflowExecution, error)
	ListWorkflowExecutions(ctx context.Context, filter WorkflowExecutionFilter) ([]*WorkflowExecution, error)
	UpdateWorkflowExecution(ctx context.Context, execution *WorkflowExecution) error

	// Delayed action queue
	CreateDelayedAction(ctx context.Context, action *DelayedActionQueue) error
	GetDelayedAction(ctx context.Context, id string) (*DelayedActionQueue, error)
	ListPendingDelayedActions(ctx context.Context) ([]*DelayedActionQueue, error)
	UpdateDelayedAction(ctx context.Context, action *DelayedActionQueue) error
	DeleteDelayedAction(ctx context.Context, id string) error

	// Workflow triggers (normalized)
	CreateWorkflowTrigger(ctx context.Context, trigger *WorkflowTrigger) error
	GetWorkflowTrigger(ctx context.Context, workflowID string) (*WorkflowTrigger, error)
	UpdateWorkflowTrigger(ctx context.Context, trigger *WorkflowTrigger) error
	DeleteWorkflowTrigger(ctx context.Context, workflowID string) error

	// Workflow steps (normalized)
	CreateWorkflowStep(ctx context.Context, step *WorkflowStep) error
	GetWorkflowStep(ctx context.Context, id string) (*WorkflowStep, error)
	ListWorkflowSteps(ctx context.Context, workflowID string) ([]*WorkflowStep, error)
	UpdateWorkflowStep(ctx context.Context, step *WorkflowStep) error
	DeleteWorkflowStep(ctx context.Context, id string) error
	DeleteWorkflowSteps(ctx context.Context, workflowID string) error

	// Step executions
	CreateStepExecution(ctx context.Context, execution *StepExecution) error
	GetStepExecution(ctx context.Context, id string) (*StepExecution, error)
	ListStepExecutions(ctx context.Context, filter StepExecutionFilter) ([]*StepExecution, error)
	UpdateStepExecution(ctx context.Context, execution *StepExecution) error

	// Step execution logs
	CreateStepExecutionLog(ctx context.Context, log *StepExecutionLog) error
	ListStepExecutionLogs(ctx context.Context, stepExecutionID string, limit int) ([]*StepExecutionLog, error)
}

// Agent represents an OpenTelemetry agent
type Agent struct {
	ID              uuid.UUID         `json:"id"`
	Name            string            `json:"name"`
	Labels          map[string]string `json:"labels"`
	Status          AgentStatus       `json:"status"`
	LastSeen        time.Time         `json:"last_seen"`
	GroupID         *string           `json:"group_id,omitempty"`
	GroupName       *string           `json:"group_name,omitempty"`
	Version         string            `json:"version"`
	Capabilities    []string          `json:"capabilities"`
	EffectiveConfig string            `json:"effective_config,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
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
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	AgentID    *uuid.UUID `json:"agent_id,omitempty"`
	GroupID    *string    `json:"group_id,omitempty"`
	ConfigHash string     `json:"config_hash"`
	Content    string     `json:"content"`
	Version    int        `json:"version"`
	CreatedAt  time.Time  `json:"created_at"`
}

// ConfigFilter represents filters for listing configs
type ConfigFilter struct {
	AgentID *uuid.UUID
	GroupID *string
	Limit   int
	Offset  int
}

// ListConfigsResult represents the result of listing configs with total count
type ListConfigsResult struct {
	Configs    []*Config
	TotalCount int
}

// Workflow represents an automation workflow for config updates
// Note: Workflow structure is now stored in normalized tables (workflow_triggers, workflow_steps)
// Use GetWorkflowTrigger and ListWorkflowSteps to retrieve the full workflow structure
type Workflow struct {
	ID            string              `json:"id"`
	Name          string              `json:"name"`
	Description   string              `json:"description"`
	Type          WorkflowTriggerType `json:"type"`
	Status        WorkflowStatus      `json:"status"`
	Schedule      *ScheduleConfig     `json:"schedule,omitempty"`
	WebhookURL    string              `json:"webhook_url,omitempty"`
	WebhookSecret string              `json:"webhook_secret,omitempty"`

	CreatedBy  string     `json:"created_by,omitempty"`
	LastRun    *time.Time `json:"last_run,omitempty"`
	NextRun    *time.Time `json:"next_run,omitempty"`
	RunCount   int        `json:"run_count"`
	ErrorCount int        `json:"error_count"`
	LastError  string     `json:"last_error,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// WorkflowTriggerType represents the type of trigger that starts a workflow
type WorkflowTriggerType string

const (
	WorkflowTriggerTypeWebhook   WorkflowTriggerType = "webhook"
	WorkflowTriggerTypeSchedule  WorkflowTriggerType = "schedule"
	WorkflowTriggerTypeManual    WorkflowTriggerType = "manual"
	WorkflowTriggerTypeTelemetry WorkflowTriggerType = "telemetry"
)

// WorkflowStatus represents the status of a workflow
type WorkflowStatus string

const (
	WorkflowStatusActive   WorkflowStatus = "active"
	WorkflowStatusInactive WorkflowStatus = "inactive"
	WorkflowStatusError    WorkflowStatus = "error"
)

// ScheduleConfig represents cron-based scheduling configuration
type ScheduleConfig struct {
	CronExpression string `json:"cron_expression"`
	Timezone       string `json:"timezone"`
}

// TelemetryTriggerType represents the type of telemetry trigger
type TelemetryTriggerType string

const (
	TelemetryTriggerTypeLog    TelemetryTriggerType = "log"
	TelemetryTriggerTypeMetric TelemetryTriggerType = "metric"
)

// TelemetryTriggerConfig represents telemetry-based trigger configuration
type TelemetryTriggerConfig struct {
	Type TelemetryTriggerType `json:"type"` // "log" or "metric"

	// Log trigger fields
	Severity    string `json:"severity,omitempty"`     // "error", "warn", "info" (optional)
	Pattern     string `json:"pattern,omitempty"`      // Pattern to match in log body (optional, supports regex)
	AgentID     string `json:"agent_id,omitempty"`     // Filter by agent ID (optional)
	ServiceName string `json:"service_name,omitempty"` // Filter by service name (optional)

	// Metric trigger fields
	MetricName string  `json:"metric_name,omitempty"` // Name of the metric to monitor
	Operator   string  `json:"operator,omitempty"`    // ">", "<", ">=", "<="
	Threshold  float64 `json:"threshold,omitempty"`   // Threshold value
	TimeWindow string  `json:"time_window,omitempty"` // Duration string like "5m", "1h"
}

// WorkflowAction represents an action to execute when a workflow runs
type WorkflowAction struct {
	Type          ActionType          `json:"type"`
	TargetType    string              `json:"target_type"`
	TargetID      string              `json:"target_id"`
	ConfigUpdate  *ConfigUpdateAction `json:"config_update,omitempty"`
	DelayedAction *DelayedAction      `json:"delayed_action,omitempty"`
	TailSampling  *TailSamplingAction `json:"tail_sampling,omitempty"`
}

// ActionType represents the type of action
type ActionType string

const (
	ActionTypeConfigUpdate  ActionType = "config_update"
	ActionTypeDelayedAction ActionType = "delayed_action"
	ActionTypeTailSampling  ActionType = "tail_sampling"
)

// ConfigUpdateAction represents a config update action
type ConfigUpdateAction struct {
	Operation string      `json:"operation"`
	Template  string      `json:"template,omitempty"`
	YAMLPath  string      `json:"yaml_path,omitempty"`
	Value     interface{} `json:"value,omitempty"`
}

// DelayedAction represents an action that executes after a delay
type DelayedAction struct {
	Delay  string          `json:"delay"`  // Duration string like "24h", "30m"
	Action *WorkflowAction `json:"action"` // The action to execute after delay
}

// TailSamplingAction represents a simplified tail sampling update
type TailSamplingAction struct {
	ServiceName        string  `json:"service_name"`           // Service to update (supports ${variable})
	SamplingPercentage float64 `json:"sampling_percentage"`    // New sampling percentage (0-100)
	PolicyName         string  `json:"policy_name,omitempty"`  // Optional: specific policy name to update
	RevertAfter        string  `json:"revert_after,omitempty"` // Optional: duration to revert (e.g. "24h")
	RevertTo           float64 `json:"revert_to,omitempty"`    // Value to revert to
}

// WorkflowCondition represents a condition for workflow execution
type WorkflowCondition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

// WorkflowExecution tracks individual workflow executions
type WorkflowExecution struct {
	ID               string            `json:"id"`
	WorkflowID       string            `json:"workflow_id"`
	WorkflowName     string            `json:"workflow_name"`
	Status           ExecutionStatus   `json:"status"`
	StartedAt        time.Time         `json:"started_at"`
	CompletedAt      *time.Time        `json:"completed_at,omitempty"`
	DurationMs       *int64            `json:"duration_ms,omitempty"`
	ActionsExecuted  int               `json:"actions_executed"`
	ActionsSucceeded int               `json:"actions_succeeded"`
	ActionsFailed    int               `json:"actions_failed"`
	ConfigsCreated   []string          `json:"configs_created"`
	Error            string            `json:"error,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
}

// ExecutionStatus represents the status of a workflow execution
type ExecutionStatus string

const (
	ExecutionStatusRunning ExecutionStatus = "running"
	ExecutionStatusSuccess ExecutionStatus = "success"
	ExecutionStatusFailed  ExecutionStatus = "failed"
	ExecutionStatusPartial ExecutionStatus = "partial"
)

// WorkflowFilter represents filters for listing workflows
type WorkflowFilter struct {
	Type   *WorkflowTriggerType
	Status *WorkflowStatus
	Limit  int
}

// WorkflowExecutionFilter represents filters for listing workflow executions
type WorkflowExecutionFilter struct {
	WorkflowID *string
	Status     *ExecutionStatus
	Limit      int
}

// DelayedActionQueue represents a delayed action waiting to execute
type DelayedActionQueue struct {
	ID           string            `json:"id"`
	WorkflowID   string            `json:"workflow_id"`
	ExecutionID  string            `json:"execution_id"`
	Action       WorkflowAction    `json:"action"`
	ScheduledFor time.Time         `json:"scheduled_for"`
	Status       string            `json:"status"` // "pending", "executing", "completed", "failed"
	Metadata     map[string]string `json:"metadata,omitempty"`
	Error        string            `json:"error,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	CompletedAt  *time.Time        `json:"completed_at,omitempty"`
}

// WorkflowTrigger represents a normalized workflow trigger
type WorkflowTrigger struct {
	WorkflowID      string                  `json:"workflow_id"`
	Type            WorkflowTriggerType     `json:"type"`
	Schedule        *ScheduleConfig         `json:"schedule,omitempty"`
	WebhookURL      string                  `json:"webhook_url,omitempty"`
	WebhookSecret   string                  `json:"webhook_secret,omitempty"`
	TelemetryConfig *TelemetryTriggerConfig `json:"telemetry_config,omitempty"`
	Enabled         bool                    `json:"enabled"`
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
}

// FlowNodeType represents the type of flow node
type FlowNodeType string

const (
	FlowNodeTypeTrigger      FlowNodeType = "trigger"
	FlowNodeTypeCondition    FlowNodeType = "condition"
	FlowNodeTypeAction       FlowNodeType = "action"
	FlowNodeTypeParallel     FlowNodeType = "parallel"
	FlowNodeTypeSequential   FlowNodeType = "sequential"
	FlowNodeTypeLoop         FlowNodeType = "loop"
	FlowNodeTypeDelay        FlowNodeType = "delay"
	FlowNodeTypeNotification FlowNodeType = "notification"
	FlowNodeTypeGroup        FlowNodeType = "group"
	FlowNodeTypeVariable     FlowNodeType = "variable"
	FlowNodeTypeErrorHandler FlowNodeType = "error-handler"
	FlowNodeTypeBranch       FlowNodeType = "branch"
)

// ActionNodeData represents the data for an action node (used when unmarshaling step config)
type ActionNodeData struct {
	Label           string         `json:"label"`
	Action          WorkflowAction `json:"action"`
	TargetName      string         `json:"targetName,omitempty"`
	Description     string         `json:"description,omitempty"`
	ContinueOnError bool           `json:"continueOnError,omitempty"`
}

// ConditionNodeData represents the data for a condition node (used when unmarshaling step config)
type ConditionNodeData struct {
	Label       string              `json:"label"`
	Conditions  []WorkflowCondition `json:"conditions"`
	Logic       string              `json:"logic,omitempty"` // "and" or "or"
	Description string              `json:"description,omitempty"`
}

// ParallelNodeData represents the data for a parallel execution node
type ParallelNodeData struct {
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	WaitForAll  bool   `json:"waitForAll,omitempty"` // Wait for all branches to complete
	Timeout     int    `json:"timeout,omitempty"`    // Timeout in seconds
}

// SequentialNodeData represents the data for a sequential execution node
type SequentialNodeData struct {
	Label        string `json:"label"`
	Description  string `json:"description,omitempty"`
	DelayBetween int    `json:"delayBetween,omitempty"` // Delay in seconds between actions
}

// LoopNodeData represents the data for a loop node
type LoopNodeData struct {
	Label             string `json:"label"`
	Description       string `json:"description,omitempty"`
	LoopType          string `json:"loopType"`         // "agents", "groups", "range"
	Filter            string `json:"filter,omitempty"` // Filter expression
	MaxIterations     int    `json:"maxIterations,omitempty"`
	ParallelExecution bool   `json:"parallelExecution,omitempty"`
}

// DelayNodeData represents the data for a delay node
type DelayNodeData struct {
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	Duration    int    `json:"duration"` // Duration value
	Unit        string `json:"unit"`     // "seconds", "minutes", "hours"
}

// NotificationNodeData represents the data for a notification node
type NotificationNodeData struct {
	Label       string   `json:"label"`
	Description string   `json:"description,omitempty"`
	Channel     string   `json:"channel"` // "email", "slack", "webhook", "log"
	Recipients  []string `json:"recipients,omitempty"`
	Message     string   `json:"message"`
	Severity    string   `json:"severity,omitempty"` // "info", "warning", "error", "success"
}

// VariableNodeData represents the data for a variable node
type VariableNodeData struct {
	Label        string      `json:"label"`
	Description  string      `json:"description,omitempty"`
	Operation    string      `json:"operation"` // "set", "get", "increment", "append"
	VariableName string      `json:"variableName"`
	Value        interface{} `json:"value,omitempty"`
}

// ErrorHandlerNodeData represents the data for an error handler node
type ErrorHandlerNodeData struct {
	Label       string   `json:"label"`
	Description string   `json:"description,omitempty"`
	CatchAll    bool     `json:"catchAll,omitempty"` // Catch all errors
	ErrorTypes  []string `json:"errorTypes,omitempty"`
	RetryCount  int      `json:"retryCount,omitempty"`
	RetryDelay  int      `json:"retryDelay,omitempty"` // Delay in seconds between retries
}

// BranchNodeData represents the data for a multi-way branch node
type BranchNodeData struct {
	Label       string         `json:"label"`
	Description string         `json:"description,omitempty"`
	Branches    []BranchOption `json:"branches"`
}

// BranchOption represents a single branch option
type BranchOption struct {
	Name       string              `json:"name"`
	Conditions []WorkflowCondition `json:"condition"`
	IsDefault  bool                `json:"isDefault,omitempty"` // Default branch if no conditions match
}

// WorkflowStep represents a normalized workflow step (node in the flow graph)
type WorkflowStep struct {
	ID          string       `json:"id"`
	WorkflowID  string       `json:"workflow_id"`
	Type        FlowNodeType `json:"type"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Order       int          `json:"order"`                // Execution order
	PositionX   *float64     `json:"position_x,omitempty"` // UI position
	PositionY   *float64     `json:"position_y,omitempty"` // UI position

	// Step-specific configuration (JSON)
	ConfigJSON string `json:"config_json"` // Serialized node data

	// Retry configuration
	RetryEnabled    bool `json:"retry_enabled"`
	RetryCount      int  `json:"retry_count"`       // Max retries
	RetryDelayMs    int  `json:"retry_delay_ms"`    // Delay between retries in ms
	ContinueOnError bool `json:"continue_on_error"` // Continue workflow if step fails

	// Dependencies (which steps must complete before this one)
	DependsOn []string `json:"depends_on,omitempty"` // Step IDs this step depends on

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// StepExecution represents the execution of a single workflow step
type StepExecution struct {
	ID                  string                 `json:"id"`
	StepID              string                 `json:"step_id"`
	WorkflowID          string                 `json:"workflow_id"`
	WorkflowExecutionID string                 `json:"workflow_execution_id"`
	StepName            string                 `json:"step_name"`
	StepType            FlowNodeType           `json:"step_type"`
	Status              StepStatus             `json:"status"`
	StartedAt           time.Time              `json:"started_at"`
	CompletedAt         *time.Time             `json:"completed_at,omitempty"`
	DurationMs          *int64                 `json:"duration_ms,omitempty"`
	RetryAttempt        int                    `json:"retry_attempt"` // Current retry attempt (0 = first try)
	Error               string                 `json:"error,omitempty"`
	ErrorCode           string                 `json:"error_code,omitempty"`
	InputData           map[string]interface{} `json:"input_data,omitempty"`  // Input data for the step
	OutputData          map[string]interface{} `json:"output_data,omitempty"` // Output data from the step
	Metadata            map[string]string      `json:"metadata,omitempty"`
	CreatedAt           time.Time              `json:"created_at"`
}

// StepStatus represents the status of a step execution
type StepStatus string

const (
	StepStatusPending   StepStatus = "pending"
	StepStatusRunning   StepStatus = "running"
	StepStatusSuccess   StepStatus = "success"
	StepStatusFailed    StepStatus = "failed"
	StepStatusSkipped   StepStatus = "skipped"
	StepStatusRetrying  StepStatus = "retrying"
	StepStatusCancelled StepStatus = "cancelled"
)

// StepExecutionLog represents a log entry for a step execution
type StepExecutionLog struct {
	ID              string    `json:"id"`
	StepExecutionID string    `json:"step_execution_id"`
	Level           LogLevel  `json:"level"`
	Message         string    `json:"message"`
	Data            string    `json:"data,omitempty"` // JSON data
	Timestamp       time.Time `json:"timestamp"`
}

// LogLevel represents the log level
type LogLevel string

const (
	LogLevelDebug   LogLevel = "debug"
	LogLevelInfo    LogLevel = "info"
	LogLevelWarning LogLevel = "warning"
	LogLevelError   LogLevel = "error"
)

// StepExecutionFilter represents filters for listing step executions
type StepExecutionFilter struct {
	WorkflowID          *string
	WorkflowExecutionID *string
	StepID              *string
	Status              *StepStatus
	Limit               int
}
