package services

import (
	"context"

	"github.com/getlawrence/lawrence-oss/internal/storage/applicationstore/types"
)

// WorkflowService defines the interface for trigger management operations
type WorkflowService interface {
	// Trigger CRUD
	CreateWorkflow(ctx context.Context, trigger *types.Workflow) error
	GetWorkflow(ctx context.Context, id string) (*types.Workflow, error)
	ListWorkflows(ctx context.Context, filter types.WorkflowFilter) ([]*types.Workflow, error)
	UpdateWorkflow(ctx context.Context, trigger *types.Workflow) error
	DeleteWorkflow(ctx context.Context, id string) error

	// Execution
	ExecuteWorkflow(ctx context.Context, workflowID string, metadata map[string]string) (*types.WorkflowExecution, error)
	GetWorkflowExecution(ctx context.Context, id string) (*types.WorkflowExecution, error)
	ListWorkflowExecutions(ctx context.Context, filter types.WorkflowExecutionFilter) ([]*types.WorkflowExecution, error)

	// Webhook validation
	ValidateWebhookSecret(ctx context.Context, workflowID, secret string) bool
}
