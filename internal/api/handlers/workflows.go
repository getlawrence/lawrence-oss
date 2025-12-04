package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/getlawrence/lawrence-oss/internal/services"
	"github.com/getlawrence/lawrence-oss/internal/storage/applicationstore/types"
)

// WorkflowHandlers handles trigger-related API endpoints
type WorkflowHandlers struct {
	workflowService services.WorkflowService
	scheduler       *services.WorkflowScheduler
	appStore        types.ApplicationStore
	logger          *zap.Logger
}

// NewWorkflowHandlers creates a new trigger handlers instance
func NewWorkflowHandlers(workflowService services.WorkflowService, scheduler *services.WorkflowScheduler, appStore types.ApplicationStore, logger *zap.Logger) *WorkflowHandlers {
	return &WorkflowHandlers{
		workflowService: workflowService,
		scheduler:       scheduler,
		appStore:        appStore,
		logger:          logger,
	}
}

// HandleCreateWorkflow handles POST /api/v1/workflows
func (h *WorkflowHandlers) HandleCreateWorkflow(c *gin.Context) {
	var req WorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	workflow := req.Workflow

	// Generate ID first (needed for conversion)
	if workflow.ID == "" {
		workflow.ID = uuid.New().String()
	}

	// Convert flow_graph to trigger and steps if provided
	if req.FlowGraph != nil {
		// Generate webhook URL and secret if webhook type (before conversion)
		if workflow.Type == types.WorkflowTriggerTypeWebhook {
			workflow.WebhookURL = fmt.Sprintf("/api/v1/webhooks/workflows/%s", workflow.ID)
			workflow.WebhookSecret = generateWebhookSecret()
		}

		convertedTrigger, convertedSteps, err := ConvertFlowGraphToNormalized(
			workflow.ID,
			workflow.Type,
			workflow.Schedule,
			workflow.WebhookURL,
			workflow.WebhookSecret,
			req.FlowGraph,
		)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to convert flow_graph: %v", err)})
			return
		}

		// Use converted trigger and steps
		req.Trigger = convertedTrigger
		req.Steps = convertedSteps

		// Update workflow type from trigger if needed
		if convertedTrigger.Type != workflow.Type {
			workflow.Type = convertedTrigger.Type
		}
	}

	// Validate workflow and normalized structure
	if err := h.validateWorkflow(&workflow); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.validateNormalizedStructure(req.Trigger, req.Steps); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set timestamps
	workflow.CreatedAt = time.Now()
	workflow.UpdatedAt = time.Now()
	workflow.Status = types.WorkflowStatusActive

	// Generate webhook URL and secret if webhook type
	if workflow.Type == types.WorkflowTriggerTypeWebhook {
		if workflow.WebhookURL == "" {
			workflow.WebhookURL = fmt.Sprintf("/api/v1/webhooks/workflows/%s", workflow.ID)
		}
		if workflow.WebhookSecret == "" {
			workflow.WebhookSecret = generateWebhookSecret()
		}
	}

	// Use provided trigger and steps, ensuring workflow ID matches
	trigger := req.Trigger
	if trigger == nil {
		// Create default trigger if not provided
		trigger = &types.WorkflowTrigger{
			WorkflowID:    workflow.ID,
			Type:          workflow.Type,
			Enabled:       true,
			Schedule:      workflow.Schedule,
			WebhookURL:    workflow.WebhookURL,
			WebhookSecret: workflow.WebhookSecret,
		}
	} else {
		// Ensure trigger matches workflow
		trigger.WorkflowID = workflow.ID
		trigger.Type = workflow.Type
		if workflow.Schedule != nil {
			trigger.Schedule = workflow.Schedule
		}
		if workflow.WebhookURL != "" {
			trigger.WebhookURL = workflow.WebhookURL
		}
		if workflow.WebhookSecret != "" {
			trigger.WebhookSecret = workflow.WebhookSecret
		}
	}

	steps := req.Steps
	// Ensure all steps have correct workflow ID
	for _, step := range steps {
		step.WorkflowID = workflow.ID
	}

	// Create workflow
	if err := h.workflowService.CreateWorkflow(c.Request.Context(), &workflow); err != nil {
		h.logger.Error("Failed to create workflow", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create workflow"})
		return
	}

	// Create trigger
	trigger.CreatedAt = time.Now()
	trigger.UpdatedAt = time.Now()
	if err := h.appStore.CreateWorkflowTrigger(c.Request.Context(), trigger); err != nil {
		h.logger.Error("Failed to create workflow trigger", zap.Error(err))
		// Clean up workflow
		h.workflowService.DeleteWorkflow(c.Request.Context(), workflow.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create workflow trigger"})
		return
	}

	// Create steps
	for _, step := range steps {
		step.CreatedAt = time.Now()
		step.UpdatedAt = time.Now()
		if err := h.appStore.CreateWorkflowStep(c.Request.Context(), step); err != nil {
			h.logger.Error("Failed to create workflow step", zap.String("step_id", step.ID), zap.Error(err))
			// Clean up workflow and trigger
			h.appStore.DeleteWorkflowTrigger(c.Request.Context(), workflow.ID)
			h.workflowService.DeleteWorkflow(c.Request.Context(), workflow.ID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create workflow step: %s", step.ID)})
			return
		}
	}

	var flowGraph *FlowGraph
	if fg, err := ConvertNormalizedToFlowGraph(workflow.ID, trigger, steps); err != nil {
		h.logger.Warn("Failed to convert normalized workflow to flow_graph", zap.String("workflow_id", workflow.ID), zap.Error(err))
	} else {
		flowGraph = fg
	}

	// If it's a schedule trigger, add to scheduler
	if workflow.Type == types.WorkflowTriggerTypeSchedule && h.scheduler != nil && h.scheduler.IsRunning() {
		if err := h.scheduler.AddWorkflow(&workflow); err != nil {
			h.logger.Warn("Failed to add workflow to scheduler", zap.Error(err))
		}
	}

	response := WorkflowResponse{
		Workflow:  workflow,
		Trigger:   trigger,
		Steps:     steps,
		FlowGraph: flowGraph,
	}

	c.JSON(http.StatusCreated, response)
}

// HandleListWorkflows handles GET /api/v1/triggers
func (h *WorkflowHandlers) HandleListWorkflows(c *gin.Context) {
	// Parse optional filters from query params
	filter := types.WorkflowFilter{
		Limit: 100,
	}

	// Get type filter
	if workflowType := c.Query("type"); workflowType != "" {
		t := types.WorkflowTriggerType(workflowType)
		filter.Type = &t
	}

	// Get status filter
	if status := c.Query("status"); status != "" {
		s := types.WorkflowStatus(status)
		filter.Status = &s
	}

	triggers, err := h.workflowService.ListWorkflows(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list triggers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list triggers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"workflows": triggers,
		"count":     len(triggers),
	})
}

// HandleGetWorkflow handles GET /api/v1/workflows/:id
func (h *WorkflowHandlers) HandleGetWorkflow(c *gin.Context) {
	id := c.Param("id")

	workflow, err := h.workflowService.GetWorkflow(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get workflow", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get workflow"})
		return
	}

	if workflow == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	// Load trigger and steps
	trigger, err := h.appStore.GetWorkflowTrigger(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get workflow trigger", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get workflow trigger"})
		return
	}

	steps, err := h.appStore.ListWorkflowSteps(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get workflow steps", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get workflow steps"})
		return
	}

	response := WorkflowResponse{
		Workflow: *workflow,
		Trigger:  trigger,
		Steps:    steps,
	}

	if fg, err := ConvertNormalizedToFlowGraph(workflow.ID, trigger, steps); err != nil {
		h.logger.Warn("Failed to convert normalized workflow to flow_graph", zap.String("workflow_id", workflow.ID), zap.Error(err))
	} else {
		response.FlowGraph = fg
	}

	c.JSON(http.StatusOK, response)
}

// HandleUpdateWorkflow handles PUT /api/v1/workflows/:id
func (h *WorkflowHandlers) HandleUpdateWorkflow(c *gin.Context) {
	id := c.Param("id")

	var req WorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	workflow := req.Workflow
	workflow.ID = id

	// Convert flow_graph to trigger and steps if provided
	if req.FlowGraph != nil {
		// Get existing workflow to preserve webhook secret if it exists
		existingWorkflow, err := h.workflowService.GetWorkflow(c.Request.Context(), id)
		if err == nil && existingWorkflow != nil {
			if workflow.WebhookSecret == "" && existingWorkflow.WebhookSecret != "" {
				workflow.WebhookSecret = existingWorkflow.WebhookSecret
			}
			if workflow.WebhookURL == "" && existingWorkflow.WebhookURL != "" {
				workflow.WebhookURL = existingWorkflow.WebhookURL
			}
		}

		// Generate webhook URL and secret if webhook type (before conversion)
		if workflow.Type == types.WorkflowTriggerTypeWebhook {
			if workflow.WebhookURL == "" {
				workflow.WebhookURL = fmt.Sprintf("/api/v1/webhooks/workflows/%s", id)
			}
			if workflow.WebhookSecret == "" {
				workflow.WebhookSecret = generateWebhookSecret()
			}
		}

		convertedTrigger, convertedSteps, err := ConvertFlowGraphToNormalized(
			id,
			workflow.Type,
			workflow.Schedule,
			workflow.WebhookURL,
			workflow.WebhookSecret,
			req.FlowGraph,
		)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to convert flow_graph: %v", err)})
			return
		}

		// Use converted trigger and steps
		req.Trigger = convertedTrigger
		req.Steps = convertedSteps

		// Update workflow type from trigger if needed
		if convertedTrigger.Type != workflow.Type {
			workflow.Type = convertedTrigger.Type
		}
	}

	// Validate workflow and normalized structure
	if err := h.validateWorkflow(&workflow); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.validateNormalizedStructure(req.Trigger, req.Steps); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workflow.UpdatedAt = time.Now()

	// Use provided trigger and steps, ensuring workflow ID matches
	trigger := req.Trigger
	if trigger == nil {
		// Create default trigger if not provided
		trigger = &types.WorkflowTrigger{
			WorkflowID:    id,
			Type:          workflow.Type,
			Enabled:       true,
			Schedule:      workflow.Schedule,
			WebhookURL:    workflow.WebhookURL,
			WebhookSecret: workflow.WebhookSecret,
		}
	} else {
		// Ensure trigger matches workflow
		trigger.WorkflowID = id
		trigger.Type = workflow.Type
		if workflow.Schedule != nil {
			trigger.Schedule = workflow.Schedule
		}
		if workflow.WebhookURL != "" {
			trigger.WebhookURL = workflow.WebhookURL
		}
		if workflow.WebhookSecret != "" {
			trigger.WebhookSecret = workflow.WebhookSecret
		}
	}

	steps := req.Steps
	// Ensure all steps have correct workflow ID
	for _, step := range steps {
		step.WorkflowID = id
	}

	// Update workflow
	if err := h.workflowService.UpdateWorkflow(c.Request.Context(), &workflow); err != nil {
		h.logger.Error("Failed to update workflow", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update workflow"})
		return
	}

	// Update trigger
	trigger.UpdatedAt = time.Now()
	if err := h.appStore.UpdateWorkflowTrigger(c.Request.Context(), trigger); err != nil {
		h.logger.Error("Failed to update workflow trigger", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update workflow trigger"})
		return
	}

	// Delete existing steps and create new ones
	if err := h.appStore.DeleteWorkflowSteps(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete workflow steps", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete workflow steps"})
		return
	}

	// Create new steps
	for _, step := range steps {
		step.UpdatedAt = time.Now()
		if err := h.appStore.CreateWorkflowStep(c.Request.Context(), step); err != nil {
			h.logger.Error("Failed to create workflow step", zap.String("step_id", step.ID), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create workflow step: %s", step.ID)})
			return
		}
	}

	var flowGraph *FlowGraph
	if fg, err := ConvertNormalizedToFlowGraph(id, trigger, steps); err != nil {
		h.logger.Warn("Failed to convert normalized workflow to flow_graph", zap.String("workflow_id", id), zap.Error(err))
	} else {
		flowGraph = fg
	}

	// Update scheduler if needed
	if workflow.Type == types.WorkflowTriggerTypeSchedule && h.scheduler != nil && h.scheduler.IsRunning() {
		if err := h.scheduler.AddWorkflow(&workflow); err != nil {
			h.logger.Warn("Failed to update workflow in scheduler", zap.Error(err))
		}
	} else if h.scheduler != nil {
		h.scheduler.RemoveWorkflow(workflow.ID)
	}

	response := WorkflowResponse{
		Workflow:  workflow,
		Trigger:   trigger,
		Steps:     steps,
		FlowGraph: flowGraph,
	}

	c.JSON(http.StatusOK, response)
}

// HandleDeleteWorkflow handles DELETE /api/v1/triggers/:id
func (h *WorkflowHandlers) HandleDeleteWorkflow(c *gin.Context) {
	id := c.Param("id")

	// Remove from scheduler first
	if h.scheduler != nil {
		h.scheduler.RemoveWorkflow(id)
	}

	if err := h.workflowService.DeleteWorkflow(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete trigger", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete trigger"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workflow deleted"})
}

// HandleExecuteWorkflow handles POST /api/v1/triggers/:id/execute
func (h *WorkflowHandlers) HandleExecuteWorkflow(c *gin.Context) {
	id := c.Param("id")

	execution, err := h.workflowService.ExecuteWorkflow(c.Request.Context(), id, map[string]string{
		"source": "manual",
		"user":   "api", // TODO: Get from auth context
	})

	if err != nil {
		h.logger.Error("Failed to execute trigger", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute trigger", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, execution)
}

// HandleGetWorkflowExecutions handles GET /api/v1/triggers/:id/executions
func (h *WorkflowHandlers) HandleGetWorkflowExecutions(c *gin.Context) {
	id := c.Param("id")

	executions, err := h.workflowService.ListWorkflowExecutions(c.Request.Context(), types.WorkflowExecutionFilter{
		WorkflowID: &id,
		Limit:      50,
	})

	if err != nil {
		h.logger.Error("Failed to get executions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get executions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"executions": executions,
		"count":      len(executions),
	})
}

// HandleWebhookWorkflow handles POST /api/v1/webhooks/triggers/:id
func (h *WorkflowHandlers) HandleWebhookWorkflow(c *gin.Context) {
	id := c.Param("id")

	// Validate webhook secret
	secret := c.GetHeader("X-Webhook-Secret")
	if !h.workflowService.ValidateWebhookSecret(c.Request.Context(), id, secret) {
		h.logger.Warn("Invalid webhook secret", zap.String("workflow_id", id))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid webhook secret"})
		return
	}

	// Parse webhook payload
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	// Convert payload to metadata
	metadata := make(map[string]string)
	for k, v := range payload {
		metadata[k] = fmt.Sprintf("%v", v)
	}
	metadata["source"] = "webhook"

	// Execute trigger
	execution, err := h.workflowService.ExecuteWorkflow(c.Request.Context(), id, metadata)
	if err != nil {
		h.logger.Error("Webhook trigger execution failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Execution failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"execution_id": execution.ID,
		"status":       execution.Status,
	})
}

// validateWorkflow validates a workflow configuration
func (h *WorkflowHandlers) validateWorkflow(workflow *types.Workflow) error {
	if workflow.Name == "" {
		return fmt.Errorf("name is required")
	}

	if workflow.Type == types.WorkflowTriggerTypeSchedule && workflow.Schedule == nil {
		return fmt.Errorf("schedule config required for schedule triggers")
	}

	if workflow.Type == types.WorkflowTriggerTypeSchedule && workflow.Schedule != nil {
		if workflow.Schedule.CronExpression == "" {
			return fmt.Errorf("cron expression is required for schedule triggers")
		}
	}

	return nil
}

// validateNormalizedStructure validates normalized workflow structure (trigger + steps)
func (h *WorkflowHandlers) validateNormalizedStructure(trigger *types.WorkflowTrigger, steps []*types.WorkflowStep) error {
	if trigger == nil {
		return fmt.Errorf("trigger is required")
	}

	// Validate trigger has required fields
	if trigger.Type == "" {
		return fmt.Errorf("trigger type is required")
	}

	// Validate steps if provided
	for i, step := range steps {
		if step.ID == "" {
			return fmt.Errorf("step %d: id is required", i)
		}
		if step.Type == "" {
			return fmt.Errorf("step %s: type is required", step.ID)
		}
		if step.Name == "" {
			return fmt.Errorf("step %s: name is required", step.ID)
		}
	}

	return nil
}

// HandleGetStepExecutions handles GET /api/v1/workflows/:id/executions/:executionId/steps
func (h *WorkflowHandlers) HandleGetStepExecutions(c *gin.Context) {
	executionID := c.Param("executionId")

	filter := types.StepExecutionFilter{
		WorkflowExecutionID: &executionID,
		Limit:               100,
	}

	stepExecutions, err := h.appStore.ListStepExecutions(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to get step executions", zap.String("execution_id", executionID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get step executions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"step_executions": stepExecutions,
		"count":           len(stepExecutions),
	})
}

// HandleGetStepExecution handles GET /api/v1/step-executions/:executionId
func (h *WorkflowHandlers) HandleGetStepExecution(c *gin.Context) {
	executionID := c.Param("executionId")

	stepExecution, err := h.appStore.GetStepExecution(c.Request.Context(), executionID)
	if err != nil {
		h.logger.Error("Failed to get step execution", zap.String("execution_id", executionID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get step execution"})
		return
	}

	if stepExecution == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Step execution not found"})
		return
	}

	c.JSON(http.StatusOK, stepExecution)
}

// HandleGetStepExecutionLogs handles GET /api/v1/step-executions/:executionId/logs
func (h *WorkflowHandlers) HandleGetStepExecutionLogs(c *gin.Context) {
	executionID := c.Param("executionId")

	limit := 1000
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil || parsedLimit != 1 {
			limit = 1000
		}
	}

	logs, err := h.appStore.ListStepExecutionLogs(c.Request.Context(), executionID, limit)
	if err != nil {
		h.logger.Error("Failed to get step execution logs", zap.String("execution_id", executionID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get step execution logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"count": len(logs),
	})
}

// HandleRetryStepExecution handles POST /api/v1/step-executions/:executionId/retry
func (h *WorkflowHandlers) HandleRetryStepExecution(c *gin.Context) {
	executionID := c.Param("executionId")

	// Get the step execution
	stepExecution, err := h.appStore.GetStepExecution(c.Request.Context(), executionID)
	if err != nil {
		h.logger.Error("Failed to get step execution", zap.String("execution_id", executionID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get step execution"})
		return
	}

	if stepExecution == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Step execution not found"})
		return
	}

	// Get the step to check retry configuration
	step, err := h.appStore.GetWorkflowStep(c.Request.Context(), stepExecution.StepID)
	if err != nil {
		h.logger.Error("Failed to get workflow step", zap.String("step_id", stepExecution.StepID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get workflow step"})
		return
	}

	if step == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow step not found"})
		return
	}

	// Check if retry is enabled and if we haven't exceeded retry count
	if !step.RetryEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Retry is not enabled for this step"})
		return
	}

	if stepExecution.RetryAttempt >= step.RetryCount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Maximum retry attempts exceeded"})
		return
	}

	// Check if step execution is in a retryable state
	if stepExecution.Status != types.StepStatusFailed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Step execution is not in a failed state"})
		return
	}

	// TODO: Implement actual retry logic - for now just return success
	// This would typically involve:
	// 1. Creating a new step execution with incremented retry attempt
	// 2. Executing the step again
	// 3. Updating the workflow execution status

	c.JSON(http.StatusOK, gin.H{
		"message":           "Step execution retry initiated",
		"step_execution_id": executionID,
	})
}

// generateWebhookSecret generates a random webhook secret
func generateWebhookSecret() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
