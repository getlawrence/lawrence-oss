package services

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/getlawrence/lawrence-oss/internal/storage/applicationstore/types"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// workflowServiceImpl implements the WorkflowService interface
type workflowServiceImpl struct {
	appStore     types.ApplicationStore
	agentService AgentService
	logger       *zap.Logger
}

// NewWorkflowService creates a new workflow service
func NewWorkflowService(appStore types.ApplicationStore, agentService AgentService, logger *zap.Logger) WorkflowService {
	return &workflowServiceImpl{
		appStore:     appStore,
		agentService: agentService,
		logger:       logger,
	}
}

// CreateWorkflow creates a new workflow
func (s *workflowServiceImpl) CreateWorkflow(ctx context.Context, workflow *types.Workflow) error {
	if err := s.appStore.CreateWorkflow(ctx, workflow); err != nil {
		return fmt.Errorf("failed to update workflow: %w", err)
	}

	s.logger.Info("Workflow created",
		zap.String("id", workflow.ID),
		zap.String("name", workflow.Name),
		zap.String("type", string(workflow.Type)))

	return nil
}

// GetWorkflow retrieves a workflow by ID
func (s *workflowServiceImpl) GetWorkflow(ctx context.Context, id string) (*types.Workflow, error) {
	return s.appStore.GetWorkflow(ctx, id)
}

// ListWorkflows lists all workflows with optional filtering
func (s *workflowServiceImpl) ListWorkflows(ctx context.Context, filter types.WorkflowFilter) ([]*types.Workflow, error) {
	return s.appStore.ListWorkflows(ctx, filter)
}

// UpdateWorkflow updates an existing workflow
func (s *workflowServiceImpl) UpdateWorkflow(ctx context.Context, workflow *types.Workflow) error {
	workflow.UpdatedAt = time.Now()

	if err := s.appStore.UpdateWorkflow(ctx, workflow); err != nil {
		return fmt.Errorf("failed to update workflow: %w", err)
	}

	s.logger.Info("Workflow updated",
		zap.String("id", workflow.ID),
		zap.String("name", workflow.Name))

	return nil
}

// DeleteWorkflow deletes a workflow
func (s *workflowServiceImpl) DeleteWorkflow(ctx context.Context, id string) error {
	if err := s.appStore.DeleteWorkflow(ctx, id); err != nil {
		return fmt.Errorf("failed to update workflow: %w", err)
	}

	s.logger.Info("Workflow deleted", zap.String("id", id))
	return nil
}

// ExecuteWorkflow executes a workflow and all its actions
func (s *workflowServiceImpl) ExecuteWorkflow(ctx context.Context, workflowID string, metadata map[string]string) (*types.WorkflowExecution, error) {
	// Get the workflow
	workflow, err := s.appStore.GetWorkflow(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to update workflow: %w", err)
	}
	if workflow == nil {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}

	// Check if workflow is active
	if workflow.Status != types.WorkflowStatusActive {
		return nil, fmt.Errorf("trigger is not active: %s", workflow.Status)
	}

	// Get workflow steps
	steps, err := s.appStore.ListWorkflowSteps(ctx, workflow.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow steps: %w", err)
	}

	// Extract conditions from workflow steps
	conditions := s.extractConditionsFromSteps(steps)
	if len(conditions) > 0 {
		if !s.evaluateConditions(conditions, metadata) {
			s.logger.Info("Trigger conditions not met, skipping execution",
				zap.String("workflow_id", workflow.ID),
				zap.Int("condition_count", len(conditions)))
			return nil, fmt.Errorf("workflow conditions not met")
		}
	}

	// Create execution record
	execution := &types.WorkflowExecution{
		ID:             uuid.New().String(),
		WorkflowID:     workflow.ID,
		WorkflowName:   workflow.Name,
		Status:         types.ExecutionStatusRunning,
		StartedAt:      time.Now(),
		Metadata:       metadata,
		CreatedAt:      time.Now(),
		ConfigsCreated: make([]string, 0),
	}

	if err := s.appStore.CreateWorkflowExecution(ctx, execution); err != nil {
		s.logger.Error("Failed to create execution record", zap.Error(err))
		// Continue anyway, this is just for tracking
	}

	// Build step map and dependency graph
	stepMap := make(map[string]*types.WorkflowStep)
	for _, step := range steps {
		stepMap[step.ID] = step
	}

	// Build dependency graph and get execution order
	executionOrder := s.topologicalSort(steps, stepMap)

	s.logger.Info("Executing workflow",
		zap.String("workflow_id", workflow.ID),
		zap.String("workflow_name", workflow.Name),
		zap.Int("step_count", len(executionOrder)))

	// Execution context with variables
	execCtx := &executionContext{
		workflow:       workflow,
		execution:      execution,
		stepMap:        stepMap,
		variables:      make(map[string]interface{}),
		stepExecutions: make(map[string]*types.StepExecution),
		metadata:       metadata,
	}

	// Execute steps in topological order
	for _, stepID := range executionOrder {
		step := stepMap[stepID]
		if step == nil {
			continue
		}

		// Skip group nodes (visual only)
		if step.Type == types.FlowNodeTypeGroup {
			continue
		}

		// Execute step
		stepErr := s.executeStep(ctx, step, execCtx)

		if stepErr != nil {
			execution.ActionsFailed++
			s.logger.Error("Step execution failed",
				zap.String("step_id", step.ID),
				zap.String("step_name", step.Name),
				zap.String("step_type", string(step.Type)),
				zap.Error(stepErr))

			// If step failed and continue_on_error is false, stop execution
			if !step.ContinueOnError {
				execution.Error = fmt.Sprintf("Step %s failed: %v", step.Name, stepErr)
				break
			}
		} else {
			execution.ActionsSucceeded++
		}

		execution.ActionsExecuted++
	}

	// Update execution status
	now := time.Now()
	execution.CompletedAt = &now
	duration := now.Sub(execution.StartedAt).Milliseconds()
	execution.DurationMs = &duration

	if execution.ActionsFailed == 0 {
		execution.Status = types.ExecutionStatusSuccess
	} else if execution.ActionsSucceeded > 0 {
		execution.Status = types.ExecutionStatusPartial
	} else {
		execution.Status = types.ExecutionStatusFailed
	}

	if err := s.appStore.UpdateWorkflowExecution(ctx, execution); err != nil {
		s.logger.Error("Failed to update execution record", zap.Error(err))
	}

	// Update workflow statistics
	workflow.RunCount++
	if execution.Status == types.ExecutionStatusFailed || execution.Status == types.ExecutionStatusPartial {
		workflow.ErrorCount++
		workflow.LastError = execution.Error
		workflow.Status = types.WorkflowStatusError
	}
	workflow.LastRun = &now

	if err := s.appStore.UpdateWorkflow(ctx, workflow); err != nil {
		s.logger.Error("Failed to update trigger statistics", zap.Error(err))
	}

	s.logger.Info("Trigger execution completed",
		zap.String("execution_id", execution.ID),
		zap.String("status", string(execution.Status)),
		zap.Int("succeeded", execution.ActionsSucceeded),
		zap.Int("failed", execution.ActionsFailed))

	return execution, nil
}

// executionContext holds state during workflow execution
type executionContext struct {
	workflow       *types.Workflow
	execution      *types.WorkflowExecution
	stepMap        map[string]*types.WorkflowStep
	variables      map[string]interface{}
	stepExecutions map[string]*types.StepExecution
	metadata       map[string]string
}

// topologicalSort sorts steps by their dependencies
func (s *workflowServiceImpl) topologicalSort(steps []*types.WorkflowStep, stepMap map[string]*types.WorkflowStep) []string {
	// Build dependency map
	deps := make(map[string][]string) // stepID -> []dependsOn
	inDegree := make(map[string]int)  // stepID -> in-degree count

	// Initialize
	for _, step := range steps {
		deps[step.ID] = step.DependsOn
		inDegree[step.ID] = len(step.DependsOn)
	}

	// Kahn's algorithm for topological sort
	var queue []string
	var result []string

	// Find all nodes with no dependencies
	for stepID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, stepID)
		}
	}

	// Process queue
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// Find all steps that depend on current
		for _, step := range steps {
			for _, depID := range step.DependsOn {
				if depID == current {
					inDegree[step.ID]--
					if inDegree[step.ID] == 0 {
						queue = append(queue, step.ID)
					}
				}
			}
		}
	}

	// If we couldn't process all steps, there's a cycle - fall back to order
	if len(result) < len(steps) {
		s.logger.Warn("Cycle detected in workflow dependencies, using order field")
		result = make([]string, len(steps))
		for i, step := range steps {
			result[i] = step.ID
		}
	}

	return result
}

// executeStep executes a single workflow step based on its type
func (s *workflowServiceImpl) executeStep(ctx context.Context, step *types.WorkflowStep, execCtx *executionContext) error {
	// Create step execution record
	stepExecution := &types.StepExecution{
		ID:                  uuid.New().String(),
		StepID:              step.ID,
		WorkflowID:          execCtx.workflow.ID,
		WorkflowExecutionID: execCtx.execution.ID,
		StepName:            step.Name,
		StepType:            step.Type,
		Status:              types.StepStatusRunning,
		StartedAt:           time.Now(),
		RetryAttempt:        0,
		CreatedAt:           time.Now(),
	}

	if err := s.appStore.CreateStepExecution(ctx, stepExecution); err != nil {
		s.logger.Error("Failed to create step execution", zap.Error(err))
	}

	execCtx.stepExecutions[step.ID] = stepExecution
	s.createStepExecutionLog(ctx, stepExecution.ID, types.LogLevelInfo, fmt.Sprintf("Starting step execution: %s", step.Name), nil)

	var stepErr error

	// Execute based on step type
	switch step.Type {
	case types.FlowNodeTypeAction:
		stepErr = s.executeActionStep(ctx, step, execCtx, stepExecution)
	case types.FlowNodeTypeDelay:
		stepErr = s.executeDelayStep(ctx, step, execCtx, stepExecution)
	case types.FlowNodeTypeLoop:
		stepErr = s.executeLoopStep(ctx, step, execCtx, stepExecution)
	case types.FlowNodeTypeParallel:
		stepErr = s.executeParallelStep(ctx, step, execCtx, stepExecution)
	case types.FlowNodeTypeSequential:
		stepErr = s.executeSequentialStep(ctx, step, execCtx, stepExecution)
	case types.FlowNodeTypeCondition:
		stepErr = s.executeConditionStep(ctx, step, execCtx, stepExecution)
	case types.FlowNodeTypeBranch:
		stepErr = s.executeBranchStep(ctx, step, execCtx, stepExecution)
	case types.FlowNodeTypeVariable:
		stepErr = s.executeVariableStep(ctx, step, execCtx, stepExecution)
	case types.FlowNodeTypeNotification:
		stepErr = s.executeNotificationStep(ctx, step, execCtx, stepExecution)
	case types.FlowNodeTypeErrorHandler:
		// Error handlers are executed when errors occur, not as regular steps
		stepErr = nil
	default:
		stepErr = fmt.Errorf("unsupported step type: %s", step.Type)
	}

	// Update step execution status
	now := time.Now()
	stepExecution.CompletedAt = &now
	duration := now.Sub(stepExecution.StartedAt).Milliseconds()
	stepExecution.DurationMs = &duration

	if stepErr != nil {
		stepExecution.Status = types.StepStatusFailed
		stepExecution.Error = stepErr.Error()
		s.createStepExecutionLog(ctx, stepExecution.ID, types.LogLevelError, fmt.Sprintf("Step execution failed: %v", stepErr), map[string]interface{}{"error": stepErr.Error()})
	} else {
		stepExecution.Status = types.StepStatusSuccess
		s.createStepExecutionLog(ctx, stepExecution.ID, types.LogLevelInfo, "Step execution completed successfully", nil)
	}

	if err := s.appStore.UpdateStepExecution(ctx, stepExecution); err != nil {
		s.logger.Error("Failed to update step execution", zap.Error(err))
	}

	return stepErr
}

// executeActionStep executes an action step
func (s *workflowServiceImpl) executeActionStep(ctx context.Context, step *types.WorkflowStep, execCtx *executionContext, stepExecution *types.StepExecution) error {
	var actionData types.ActionNodeData
	if err := json.Unmarshal([]byte(step.ConfigJSON), &actionData); err != nil {
		return fmt.Errorf("failed to parse action step config: %w", err)
	}

	return s.executeAction(ctx, execCtx.workflow, &actionData.Action, execCtx.execution)
}

// executeDelayStep executes a delay step
func (s *workflowServiceImpl) executeDelayStep(ctx context.Context, step *types.WorkflowStep, execCtx *executionContext, stepExecution *types.StepExecution) error {
	var delayData types.DelayNodeData
	if err := json.Unmarshal([]byte(step.ConfigJSON), &delayData); err != nil {
		return fmt.Errorf("failed to parse delay step config: %w", err)
	}

	// Calculate duration
	var duration time.Duration
	switch delayData.Unit {
	case "seconds":
		duration = time.Duration(delayData.Duration) * time.Second
	case "minutes":
		duration = time.Duration(delayData.Duration) * time.Minute
	case "hours":
		duration = time.Duration(delayData.Duration) * time.Hour
	default:
		return fmt.Errorf("invalid delay unit: %s", delayData.Unit)
	}

	s.createStepExecutionLog(ctx, stepExecution.ID, types.LogLevelInfo, fmt.Sprintf("Waiting for %v", duration), nil)
	time.Sleep(duration)
	return nil
}

// executeLoopStep executes a loop step
func (s *workflowServiceImpl) executeLoopStep(ctx context.Context, step *types.WorkflowStep, execCtx *executionContext, stepExecution *types.StepExecution) error {
	var loopData types.LoopNodeData
	if err := json.Unmarshal([]byte(step.ConfigJSON), &loopData); err != nil {
		return fmt.Errorf("failed to parse loop step config: %w", err)
	}

	// Find child steps (steps that depend on this loop step)
	childSteps := s.getChildSteps(step.ID, execCtx.stepMap)

	if len(childSteps) == 0 {
		s.logger.Warn("Loop step has no child steps", zap.String("step_id", step.ID))
		return nil
	}

	// Get items to loop over
	var items []map[string]interface{}
	var err error

	switch loopData.LoopType {
	case "agents":
		items, err = s.getLoopAgents(ctx, loopData.Filter)
	case "groups":
		items, err = s.getLoopGroups(ctx, loopData.Filter)
	case "range":
		items, err = s.getLoopRange(loopData.Filter)
	default:
		return fmt.Errorf("unsupported loop type: %s", loopData.LoopType)
	}

	if err != nil {
		return fmt.Errorf("failed to get loop items: %w", err)
	}

	// Limit iterations if maxIterations is set
	if loopData.MaxIterations > 0 && len(items) > loopData.MaxIterations {
		items = items[:loopData.MaxIterations]
	}

	s.createStepExecutionLog(ctx, stepExecution.ID, types.LogLevelInfo, fmt.Sprintf("Looping over %d items", len(items)), map[string]interface{}{"item_count": len(items)})

	// Execute child steps for each item
	if loopData.ParallelExecution {
		return s.executeLoopParallel(ctx, items, childSteps, execCtx, stepExecution)
	} else {
		return s.executeLoopSequential(ctx, items, childSteps, execCtx, stepExecution)
	}
}

// executeParallelStep executes a parallel step
func (s *workflowServiceImpl) executeParallelStep(ctx context.Context, step *types.WorkflowStep, execCtx *executionContext, stepExecution *types.StepExecution) error {
	var parallelData types.ParallelNodeData
	if err := json.Unmarshal([]byte(step.ConfigJSON), &parallelData); err != nil {
		return fmt.Errorf("failed to parse parallel step config: %w", err)
	}

	// Find child steps
	childSteps := s.getChildSteps(step.ID, execCtx.stepMap)

	if len(childSteps) == 0 {
		s.logger.Warn("Parallel step has no child steps", zap.String("step_id", step.ID))
		return nil
	}

	// Execute child steps in parallel
	var wg sync.WaitGroup
	errChan := make(chan error, len(childSteps))
	ctxWithTimeout := ctx

	if parallelData.Timeout > 0 {
		var cancel context.CancelFunc
		ctxWithTimeout, cancel = context.WithTimeout(ctx, time.Duration(parallelData.Timeout)*time.Second)
		defer cancel()
	}

	for _, childStep := range childSteps {
		wg.Add(1)
		go func(step *types.WorkflowStep) {
			defer wg.Done()
			if err := s.executeStep(ctxWithTimeout, step, execCtx); err != nil {
				errChan <- err
			}
		}(childStep)
	}

	wg.Wait()
	close(errChan)

	// Collect errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 && parallelData.WaitForAll {
		return fmt.Errorf("parallel execution had %d errors", len(errors))
	}

	return nil
}

// executeSequentialStep executes a sequential step
func (s *workflowServiceImpl) executeSequentialStep(ctx context.Context, step *types.WorkflowStep, execCtx *executionContext, stepExecution *types.StepExecution) error {
	var sequentialData types.SequentialNodeData
	if err := json.Unmarshal([]byte(step.ConfigJSON), &sequentialData); err != nil {
		return fmt.Errorf("failed to parse sequential step config: %w", err)
	}

	// Find child steps
	childSteps := s.getChildSteps(step.ID, execCtx.stepMap)

	if len(childSteps) == 0 {
		s.logger.Warn("Sequential step has no child steps", zap.String("step_id", step.ID))
		return nil
	}

	// Execute child steps sequentially with delays
	for i, childStep := range childSteps {
		if i > 0 && sequentialData.DelayBetween > 0 {
			time.Sleep(time.Duration(sequentialData.DelayBetween) * time.Second)
		}

		if err := s.executeStep(ctx, childStep, execCtx); err != nil {
			return fmt.Errorf("sequential step %s failed: %w", childStep.Name, err)
		}
	}

	return nil
}

// executeConditionStep executes a condition step (evaluates conditions)
func (s *workflowServiceImpl) executeConditionStep(ctx context.Context, step *types.WorkflowStep, execCtx *executionContext, stepExecution *types.StepExecution) error {
	var conditionData types.ConditionNodeData
	if err := json.Unmarshal([]byte(step.ConfigJSON), &conditionData); err != nil {
		return fmt.Errorf("failed to parse condition step config: %w", err)
	}

	// Merge metadata and variables for condition evaluation
	evalContext := make(map[string]string)
	for k, v := range execCtx.metadata {
		evalContext[k] = v
	}
	for k, v := range execCtx.variables {
		evalContext[k] = fmt.Sprintf("%v", v)
	}

	// Evaluate conditions
	result := s.evaluateConditions(conditionData.Conditions, evalContext)

	// Store result in variables
	execCtx.variables[step.ID+"_result"] = result

	s.createStepExecutionLog(ctx, stepExecution.ID, types.LogLevelInfo, fmt.Sprintf("Condition evaluated: %v", result), map[string]interface{}{"result": result})

	return nil
}

// executeBranchStep executes a branch step (multi-way conditional branching)
func (s *workflowServiceImpl) executeBranchStep(ctx context.Context, step *types.WorkflowStep, execCtx *executionContext, stepExecution *types.StepExecution) error {
	var branchData types.BranchNodeData
	if err := json.Unmarshal([]byte(step.ConfigJSON), &branchData); err != nil {
		return fmt.Errorf("failed to parse branch step config: %w", err)
	}

	// Merge metadata and variables for condition evaluation
	evalContext := make(map[string]string)
	for k, v := range execCtx.metadata {
		evalContext[k] = v
	}
	for k, v := range execCtx.variables {
		evalContext[k] = fmt.Sprintf("%v", v)
	}

	// Find matching branch
	var selectedBranch *types.BranchOption
	for i := range branchData.Branches {
		branch := &branchData.Branches[i]
		if s.evaluateConditions(branch.Conditions, evalContext) {
			selectedBranch = branch
			break
		}
	}

	// Use default branch if no match
	if selectedBranch == nil {
		for i := range branchData.Branches {
			if branchData.Branches[i].IsDefault {
				selectedBranch = &branchData.Branches[i]
				break
			}
		}
	}

	if selectedBranch == nil {
		return fmt.Errorf("no matching branch and no default branch")
	}

	s.createStepExecutionLog(ctx, stepExecution.ID, types.LogLevelInfo, fmt.Sprintf("Selected branch: %s", selectedBranch.Name), nil)

	// Find child steps for selected branch (this would need edge labels in the graph)
	// For now, execute all child steps
	childSteps := s.getChildSteps(step.ID, execCtx.stepMap)
	for _, childStep := range childSteps {
		if err := s.executeStep(ctx, childStep, execCtx); err != nil {
			return err
		}
	}

	return nil
}

// executeVariableStep executes a variable step
func (s *workflowServiceImpl) executeVariableStep(ctx context.Context, step *types.WorkflowStep, execCtx *executionContext, stepExecution *types.StepExecution) error {
	var variableData types.VariableNodeData
	if err := json.Unmarshal([]byte(step.ConfigJSON), &variableData); err != nil {
		return fmt.Errorf("failed to parse variable step config: %w", err)
	}

	switch variableData.Operation {
	case "set":
		execCtx.variables[variableData.VariableName] = variableData.Value
		s.createStepExecutionLog(ctx, stepExecution.ID, types.LogLevelInfo, fmt.Sprintf("Set variable %s = %v", variableData.VariableName, variableData.Value), nil)
	case "get":
		value, exists := execCtx.variables[variableData.VariableName]
		if !exists {
			return fmt.Errorf("variable %s not found", variableData.VariableName)
		}
		s.createStepExecutionLog(ctx, stepExecution.ID, types.LogLevelInfo, fmt.Sprintf("Get variable %s = %v", variableData.VariableName, value), nil)
	case "increment":
		current, exists := execCtx.variables[variableData.VariableName]
		if !exists {
			current = 0
		}
		if num, ok := current.(float64); ok {
			execCtx.variables[variableData.VariableName] = num + 1
		} else if num, ok := current.(int); ok {
			execCtx.variables[variableData.VariableName] = num + 1
		} else {
			return fmt.Errorf("variable %s is not numeric", variableData.VariableName)
		}
	case "append":
		current, exists := execCtx.variables[variableData.VariableName]
		if !exists {
			current = ""
		}
		execCtx.variables[variableData.VariableName] = fmt.Sprintf("%v%v", current, variableData.Value)
	default:
		return fmt.Errorf("unsupported variable operation: %s", variableData.Operation)
	}

	return nil
}

// executeNotificationStep executes a notification step
func (s *workflowServiceImpl) executeNotificationStep(ctx context.Context, step *types.WorkflowStep, execCtx *executionContext, stepExecution *types.StepExecution) error {
	var notificationData types.NotificationNodeData
	if err := json.Unmarshal([]byte(step.ConfigJSON), &notificationData); err != nil {
		return fmt.Errorf("failed to parse notification step config: %w", err)
	}

	// For now, just log the notification
	// In the future, this could send emails, Slack messages, webhooks, etc.
	s.logger.Info("Notification",
		zap.String("channel", notificationData.Channel),
		zap.String("severity", notificationData.Severity),
		zap.String("message", notificationData.Message),
		zap.Strings("recipients", notificationData.Recipients))

	s.createStepExecutionLog(ctx, stepExecution.ID, types.LogLevelInfo, fmt.Sprintf("Notification sent via %s: %s", notificationData.Channel, notificationData.Message), nil)

	return nil
}

// Helper functions for loop execution

func (s *workflowServiceImpl) getChildSteps(parentID string, stepMap map[string]*types.WorkflowStep) []*types.WorkflowStep {
	var children []*types.WorkflowStep
	for _, step := range stepMap {
		for _, depID := range step.DependsOn {
			if depID == parentID {
				children = append(children, step)
				break
			}
		}
	}
	return children
}

func (s *workflowServiceImpl) getLoopAgents(ctx context.Context, filter string) ([]map[string]interface{}, error) {
	agents, err := s.agentService.ListAgents(ctx)
	if err != nil {
		return nil, err
	}

	var items []map[string]interface{}
	for _, agent := range agents {
		item := map[string]interface{}{
			"id":       agent.ID.String(),
			"name":     agent.Name,
			"status":   string(agent.Status),
			"labels":   agent.Labels,
			"group_id": agent.GroupID,
		}
		items = append(items, item)
	}

	// TODO: Apply filter if provided
	return items, nil
}

func (s *workflowServiceImpl) getLoopGroups(ctx context.Context, filter string) ([]map[string]interface{}, error) {
	groups, err := s.agentService.ListGroups(ctx)
	if err != nil {
		return nil, err
	}

	var items []map[string]interface{}
	for _, group := range groups {
		item := map[string]interface{}{
			"id":     group.ID,
			"name":   group.Name,
			"labels": group.Labels,
		}
		items = append(items, item)
	}

	// TODO: Apply filter if provided
	return items, nil
}

func (s *workflowServiceImpl) getLoopRange(filter string) ([]map[string]interface{}, error) {
	// Parse range filter (e.g., "1-10" or "start:end:step")
	parts := strings.Split(filter, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid range format, expected 'start-end'")
	}

	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, fmt.Errorf("invalid start value: %w", err)
	}

	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, fmt.Errorf("invalid end value: %w", err)
	}

	var items []map[string]interface{}
	for i := start; i <= end; i++ {
		item := map[string]interface{}{
			"index": i,
			"value": i,
		}
		items = append(items, item)
	}

	return items, nil
}

func (s *workflowServiceImpl) executeLoopSequential(ctx context.Context, items []map[string]interface{}, childSteps []*types.WorkflowStep, execCtx *executionContext, parentExecution *types.StepExecution) error {
	for i, item := range items {
		// Set loop variable
		execCtx.variables["loop_item"] = item
		execCtx.variables["loop_index"] = i

		s.createStepExecutionLog(ctx, parentExecution.ID, types.LogLevelInfo, fmt.Sprintf("Loop iteration %d/%d", i+1, len(items)), map[string]interface{}{"index": i, "item": item})

		// Execute child steps for this iteration
		for _, childStep := range childSteps {
			if err := s.executeStep(ctx, childStep, execCtx); err != nil {
				return fmt.Errorf("loop iteration %d failed: %w", i, err)
			}
		}
	}

	return nil
}

func (s *workflowServiceImpl) executeLoopParallel(ctx context.Context, items []map[string]interface{}, childSteps []*types.WorkflowStep, execCtx *executionContext, parentExecution *types.StepExecution) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(items))

	for i, item := range items {
		wg.Add(1)
		go func(index int, item map[string]interface{}) {
			defer wg.Done()

			// Create a new execution context for this iteration to avoid variable conflicts
			iterCtx := &executionContext{
				workflow:       execCtx.workflow,
				execution:      execCtx.execution,
				stepMap:        execCtx.stepMap,
				variables:      make(map[string]interface{}),
				stepExecutions: make(map[string]*types.StepExecution),
				metadata:       execCtx.metadata,
			}

			// Copy variables
			for k, v := range execCtx.variables {
				iterCtx.variables[k] = v
			}

			// Set loop variables
			iterCtx.variables["loop_item"] = item
			iterCtx.variables["loop_index"] = index

			s.createStepExecutionLog(ctx, parentExecution.ID, types.LogLevelInfo, fmt.Sprintf("Loop iteration %d/%d (parallel)", index+1, len(items)), map[string]interface{}{"index": index, "item": item})

			// Execute child steps for this iteration
			for _, childStep := range childSteps {
				if err := s.executeStep(ctx, childStep, iterCtx); err != nil {
					errChan <- fmt.Errorf("loop iteration %d failed: %w", index, err)
					return
				}
			}
		}(i, item)
	}

	wg.Wait()
	close(errChan)

	// Collect errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("parallel loop had %d errors: %v", len(errors), errors[0])
	}

	return nil
}

// executeAction executes a single action
func (s *workflowServiceImpl) executeAction(ctx context.Context, workflow *types.Workflow, action *types.WorkflowAction, execution *types.WorkflowExecution) error {
	switch action.Type {
	case types.ActionTypeConfigUpdate:
		return s.executeConfigUpdate(ctx, workflow, action, execution)
	case types.ActionTypeDelayedAction:
		return s.executeDelayedAction(ctx, workflow, action, execution)
	case types.ActionTypeTailSampling:
		return s.executeTailSamplingUpdate(ctx, workflow, action, execution)
	default:
		return fmt.Errorf("unsupported action type: %s", action.Type)
	}
}

// executeConfigUpdate executes a config update action
func (s *workflowServiceImpl) executeConfigUpdate(ctx context.Context, workflow *types.Workflow, action *types.WorkflowAction, execution *types.WorkflowExecution) error {
	if action.ConfigUpdate == nil {
		return fmt.Errorf("config update action missing configuration")
	}

	// Resolve target ID (support metadata variables like ${app_name})
	targetID := s.resolveMetadataVariable(action.TargetID, execution.Metadata)

	// Get current config
	var currentConfig *Config
	var err error

	if action.TargetType == "agent" {
		agentID, err := uuid.Parse(targetID)
		if err != nil {
			return fmt.Errorf("invalid agent ID: %w", err)
		}
		currentConfig, err = s.agentService.GetLatestConfigForAgent(ctx, agentID)
	} else if action.TargetType == "group" {
		// Support group resolution by name if targetID is not a UUID
		if !s.isUUID(targetID) {
			group, err := s.agentService.GetGroupByName(ctx, targetID)
			if err != nil {
				return fmt.Errorf("failed to find group by name '%s': %w", targetID, err)
			}
			if group == nil {
				return fmt.Errorf("group not found: %s", targetID)
			}
			targetID = group.ID
		}
		currentConfig, err = s.agentService.GetLatestConfigForGroup(ctx, targetID)
	} else {
		return fmt.Errorf("unsupported target type: %s", action.TargetType)
	}

	// For replace operations, if there's no existing config, we can create a new one
	var newContent string
	if err != nil || currentConfig == nil {
		if action.ConfigUpdate.Operation == "replace" {
			// No existing config, use template directly
			newContent = action.ConfigUpdate.Template
		} else {
			return fmt.Errorf("failed to get current config for %s operation: %w", action.ConfigUpdate.Operation, err)
		}
	} else {
		// Apply config transformation
		newContent, err = s.applyConfigUpdate(currentConfig.Content, action.ConfigUpdate)
		if err != nil {
			return fmt.Errorf("failed to apply config update: %w", err)
		}
	}

	// Create new config version
	configName := fmt.Sprintf("Trigger: %s", workflow.Name)
	version := 1
	if currentConfig != nil {
		configName = fmt.Sprintf("%s (Trigger: %s)", currentConfig.Name, workflow.Name)
		version = currentConfig.Version + 1
	}

	newConfig := &Config{
		ID:      uuid.New().String(),
		Name:    configName,
		Content: newContent,
		Version: version,
	}

	// Use resolved targetID
	if action.TargetType == "agent" {
		agentID, _ := uuid.Parse(targetID)
		newConfig.AgentID = &agentID
	} else {
		newConfig.GroupID = &targetID
	}

	// Store config
	if err := s.agentService.CreateConfig(ctx, newConfig); err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	execution.ConfigsCreated = append(execution.ConfigsCreated, newConfig.ID)

	s.logger.Info("Config updated by trigger",
		zap.String("workflow_id", workflow.ID),
		zap.String("config_id", newConfig.ID),
		zap.String("target_type", action.TargetType),
		zap.String("target_id", action.TargetID))

	return nil
}

// applyConfigUpdate applies a config update operation
func (s *workflowServiceImpl) applyConfigUpdate(currentYAML string, update *types.ConfigUpdateAction) (string, error) {
	switch update.Operation {
	case "replace":
		// Simply replace the entire config with the template
		if update.Template == "" {
			return "", fmt.Errorf("template is required for replace operation")
		}
		return update.Template, nil

	case "merge":
		// Merge the template into the current config
		if update.Template == "" {
			return "", fmt.Errorf("template is required for merge operation")
		}

		var current, template map[string]interface{}

		if err := yaml.Unmarshal([]byte(currentYAML), &current); err != nil {
			return "", fmt.Errorf("failed to unmarshal current config: %w", err)
		}

		if err := yaml.Unmarshal([]byte(update.Template), &template); err != nil {
			return "", fmt.Errorf("failed to unmarshal template: %w", err)
		}

		// Deep merge
		merged := s.mergeMaps(current, template)

		output, err := yaml.Marshal(merged)
		if err != nil {
			return "", fmt.Errorf("failed to marshal merged config: %w", err)
		}

		return string(output), nil

	case "patch":
		// Update a specific path in the YAML
		if update.YAMLPath == "" {
			return "", fmt.Errorf("yaml_path is required for patch operation")
		}

		var config map[string]interface{}
		if err := yaml.Unmarshal([]byte(currentYAML), &config); err != nil {
			return "", fmt.Errorf("failed to unmarshal config: %w", err)
		}

		// Apply the patch
		if err := s.setYAMLPath(config, update.YAMLPath, update.Value); err != nil {
			return "", fmt.Errorf("failed to set YAML path: %w", err)
		}

		output, err := yaml.Marshal(config)
		if err != nil {
			return "", fmt.Errorf("failed to marshal patched config: %w", err)
		}

		return string(output), nil

	default:
		return "", fmt.Errorf("unsupported operation: %s", update.Operation)
	}
}

// mergeMaps deep merges two maps
func (s *workflowServiceImpl) mergeMaps(dst, src map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy dst
	for k, v := range dst {
		result[k] = v
	}

	// Merge src
	for k, v := range src {
		if existingValue, exists := result[k]; exists {
			// If both values are maps, merge recursively
			if existingMap, ok := existingValue.(map[string]interface{}); ok {
				if newMap, ok := v.(map[string]interface{}); ok {
					result[k] = s.mergeMaps(existingMap, newMap)
					continue
				}
			}
		}
		// Otherwise, overwrite
		result[k] = v
	}

	return result
}

// evaluateConditions evaluates trigger conditions against metadata
// All conditions must pass (AND logic)
func (s *workflowServiceImpl) evaluateConditions(conditions []types.WorkflowCondition, metadata map[string]string) bool {
	for _, condition := range conditions {
		fieldValue, exists := metadata[condition.Field]
		if !exists {
			// Field doesn't exist in metadata, condition fails
			return false
		}

		switch condition.Operator {
		case "equals":
			if fieldValue != condition.Value {
				return false
			}
		case "contains":
			if !strings.Contains(fieldValue, condition.Value) {
				return false
			}
		case "matches":
			// Simple wildcard matching (supports * and ?)
			matched, err := s.matchPattern(fieldValue, condition.Value)
			if err != nil || !matched {
				return false
			}
		default:
			s.logger.Warn("Unknown condition operator", zap.String("operator", condition.Operator))
			return false
		}
	}

	return true
}

// extractConditionsFromSteps extracts conditions from condition steps
func (s *workflowServiceImpl) extractConditionsFromSteps(steps []*types.WorkflowStep) []types.WorkflowCondition {
	var conditions []types.WorkflowCondition
	for _, step := range steps {
		if step.Type == types.FlowNodeTypeCondition {
			// Try to unmarshal the step config as ConditionNodeData
			var conditionData types.ConditionNodeData
			if err := json.Unmarshal([]byte(step.ConfigJSON), &conditionData); err != nil {
				s.logger.Warn("Failed to unmarshal condition step config",
					zap.String("step_id", step.ID),
					zap.Error(err))
				continue
			}

			conditions = append(conditions, conditionData.Conditions...)
		}
	}

	return conditions
}

// extractActionsFromSteps extracts actions from action steps
func (s *workflowServiceImpl) extractActionsFromSteps(steps []*types.WorkflowStep) []types.WorkflowAction {
	var actions []types.WorkflowAction
	for _, step := range steps {
		if step.Type == types.FlowNodeTypeAction {
			// Try to unmarshal the step config as ActionNodeData
			var actionData types.ActionNodeData
			if err := json.Unmarshal([]byte(step.ConfigJSON), &actionData); err != nil {
				s.logger.Warn("Failed to unmarshal action step config",
					zap.String("step_id", step.ID),
					zap.Error(err))
				continue
			}

			actions = append(actions, actionData.Action)
		}
	}

	return actions
}

// createStepExecutionLog creates a log entry for a step execution
func (s *workflowServiceImpl) createStepExecutionLog(ctx context.Context, stepExecutionID string, level types.LogLevel, message string, data map[string]interface{}) {
	logEntry := &types.StepExecutionLog{
		ID:              uuid.New().String(),
		StepExecutionID: stepExecutionID,
		Level:           level,
		Message:         message,
		Timestamp:       time.Now(),
	}

	if data != nil {
		dataJSON, err := json.Marshal(data)
		if err == nil {
			logEntry.Data = string(dataJSON)
		}
	}

	if err := s.appStore.CreateStepExecutionLog(ctx, logEntry); err != nil {
		s.logger.Warn("Failed to create step execution log", zap.Error(err))
	}
}

// matchPattern performs simple pattern matching (supports * and ? wildcards)
func (s *workflowServiceImpl) matchPattern(text, pattern string) (bool, error) {
	// Convert simple wildcard pattern to regex
	regexPattern := "^" + strings.ReplaceAll(
		strings.ReplaceAll(regexp.QuoteMeta(pattern), "\\*", ".*"),
		"\\?", ".",
	) + "$"

	matched, err := regexp.MatchString(regexPattern, text)
	return matched, err
}

// resolveMetadataVariable resolves ${variable} syntax from metadata
func (s *workflowServiceImpl) resolveMetadataVariable(value string, metadata map[string]string) string {
	if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
		varName := strings.TrimPrefix(strings.TrimSuffix(value, "}"), "${")
		if resolved, exists := metadata[varName]; exists {
			return resolved
		}
		// If variable not found, return original (could be intentional)
		s.logger.Warn("Metadata variable not found", zap.String("variable", varName))
	}
	return value
}

// isUUID checks if a string is a valid UUID
func (s *workflowServiceImpl) isUUID(str string) bool {
	_, err := uuid.Parse(str)
	return err == nil
}

// setYAMLPath sets a value at a specific YAML path (e.g., "service.pipelines.traces.samplers.probability")
func (s *workflowServiceImpl) setYAMLPath(config map[string]interface{}, path string, value interface{}) error {
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return fmt.Errorf("empty path")
	}

	current := config
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			// Create nested map if it doesn't exist
			newMap := make(map[string]interface{})
			current[part] = newMap
			current = newMap
		}
	}

	// Set the final value
	current[parts[len(parts)-1]] = value
	return nil
}

// executeDelayedAction schedules an action to run after a delay
func (s *workflowServiceImpl) executeDelayedAction(ctx context.Context, workflow *types.Workflow, action *types.WorkflowAction, execution *types.WorkflowExecution) error {
	if action.DelayedAction == nil {
		return fmt.Errorf("delayed action config missing")
	}

	// Parse the delay duration
	delay, err := time.ParseDuration(action.DelayedAction.Delay)
	if err != nil {
		return fmt.Errorf("invalid delay duration '%s': %w", action.DelayedAction.Delay, err)
	}

	scheduledFor := time.Now().Add(delay)

	// Resolve metadata variables in the nested action
	if action.DelayedAction.Action.TargetID != "" {
		action.DelayedAction.Action.TargetID = s.resolveMetadataVariable(action.DelayedAction.Action.TargetID, execution.Metadata)
	}

	// Create the delayed action queue item
	queueItem := &types.DelayedActionQueue{
		ID:           uuid.New().String(),
		WorkflowID:   workflow.ID,
		ExecutionID:  execution.ID,
		Action:       *action.DelayedAction.Action,
		ScheduledFor: scheduledFor,
		Status:       "pending",
		Metadata:     execution.Metadata,
		CreatedAt:    time.Now(),
	}

	if err := s.appStore.CreateDelayedAction(ctx, queueItem); err != nil {
		return fmt.Errorf("failed to create delayed action: %w", err)
	}

	s.logger.Info("Delayed action scheduled",
		zap.String("workflow_id", workflow.ID),
		zap.String("queue_id", queueItem.ID),
		zap.Duration("delay", delay),
		zap.Time("scheduled_for", scheduledFor))

	return nil
}

// executeTailSamplingUpdate updates tail sampling configuration for a service
func (s *workflowServiceImpl) executeTailSamplingUpdate(ctx context.Context, workflow *types.Workflow, action *types.WorkflowAction, execution *types.WorkflowExecution) error {
	if action.TailSampling == nil {
		return fmt.Errorf("tail sampling action config missing")
	}

	ts := action.TailSampling

	// Resolve service name from metadata
	serviceName := s.resolveMetadataVariable(ts.ServiceName, execution.Metadata)

	// Get current config
	targetID := s.resolveMetadataVariable(action.TargetID, execution.Metadata)

	var currentConfig *Config
	var err error

	if action.TargetType == "agent" {
		agentID, err := uuid.Parse(targetID)
		if err != nil {
			return fmt.Errorf("invalid agent ID: %w", err)
		}
		currentConfig, err = s.agentService.GetLatestConfigForAgent(ctx, agentID)
	} else if action.TargetType == "group" {
		// Support group resolution by name if targetID is not a UUID
		if !s.isUUID(targetID) {
			group, err := s.agentService.GetGroupByName(ctx, targetID)
			if err != nil {
				return fmt.Errorf("failed to find group by name '%s': %w", targetID, err)
			}
			if group == nil {
				return fmt.Errorf("group not found: %s", targetID)
			}
			targetID = group.ID
		}
		currentConfig, err = s.agentService.GetLatestConfigForGroup(ctx, targetID)
	} else {
		return fmt.Errorf("unsupported target type: %s", action.TargetType)
	}

	if err != nil || currentConfig == nil {
		return fmt.Errorf("failed to get current config: %w", err)
	}

	// Parse the current YAML config
	var config map[string]interface{}
	if err := yaml.Unmarshal([]byte(currentConfig.Content), &config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Find and update the tail sampling policy for this service
	updated, err := s.updateTailSamplingPolicy(config, serviceName, ts.SamplingPercentage, ts.PolicyName)
	if err != nil {
		return fmt.Errorf("failed to update tail sampling policy: %w", err)
	}

	if !updated {
		return fmt.Errorf("no tail sampling policy found for service '%s'", serviceName)
	}

	// Marshal back to YAML
	output, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal updated config: %w", err)
	}

	// Create new config version
	configName := fmt.Sprintf("%s (Tail Sampling: %s -> %.0f%%)", currentConfig.Name, serviceName, ts.SamplingPercentage)

	newConfig := &Config{
		ID:      uuid.New().String(),
		Name:    configName,
		Content: string(output),
		Version: currentConfig.Version + 1,
	}

	if action.TargetType == "agent" {
		agentID, _ := uuid.Parse(targetID)
		newConfig.AgentID = &agentID
	} else {
		newConfig.GroupID = &targetID
	}

	// Store config
	if err := s.agentService.CreateConfig(ctx, newConfig); err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	execution.ConfigsCreated = append(execution.ConfigsCreated, newConfig.ID)

	s.logger.Info("Tail sampling updated by trigger",
		zap.String("workflow_id", workflow.ID),
		zap.String("config_id", newConfig.ID),
		zap.String("service", serviceName),
		zap.Float64("sampling_percentage", ts.SamplingPercentage))

	// If revert_after is specified, schedule a delayed revert action
	if ts.RevertAfter != "" {
		revertAction := &types.WorkflowAction{
			Type:       types.ActionTypeTailSampling,
			TargetType: action.TargetType,
			TargetID:   action.TargetID,
			TailSampling: &types.TailSamplingAction{
				ServiceName:        ts.ServiceName, // Keep the variable reference
				SamplingPercentage: ts.RevertTo,
				PolicyName:         ts.PolicyName,
			},
		}

		delayedAction := &types.WorkflowAction{
			Type:       types.ActionTypeDelayedAction,
			TargetType: action.TargetType,
			TargetID:   action.TargetID,
			DelayedAction: &types.DelayedAction{
				Delay:  ts.RevertAfter,
				Action: revertAction,
			},
		}

		if err := s.executeDelayedAction(ctx, workflow, delayedAction, execution); err != nil {
			return fmt.Errorf("failed to schedule revert action: %w", err)
		}

		s.logger.Info("Scheduled tail sampling revert",
			zap.String("service", serviceName),
			zap.String("revert_after", ts.RevertAfter),
			zap.Float64("revert_to", ts.RevertTo))
	}

	return nil
}

// updateTailSamplingPolicy finds and updates the tail sampling policy for a specific service
func (s *workflowServiceImpl) updateTailSamplingPolicy(config map[string]interface{}, serviceName string, samplingPercentage float64, policyName string) (bool, error) {
	// Navigate to processors.tail_sampling.policies
	processors, ok := config["processors"].(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("processors section not found")
	}

	tailSampling, ok := processors["tail_sampling"].(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("tail_sampling processor not found")
	}

	policies, ok := tailSampling["policies"].([]interface{})
	if !ok {
		return false, fmt.Errorf("tail_sampling policies not found")
	}

	// Find the policy that matches the service
	for _, policyInterface := range policies {
		policy, ok := policyInterface.(map[string]interface{})
		if !ok {
			continue
		}

		// If policyName is specified, match by name
		if policyName != "" {
			name, _ := policy["name"].(string)
			if name != policyName {
				continue
			}
		} else {
			// Otherwise, look for policies with string_attribute matching service.name
			if !s.policyMatchesService(policy, serviceName) {
				continue
			}
		}

		// Update the sampling percentage
		// Handle both composite and simple probabilistic policies
		if composite, ok := policy["composite"].(map[string]interface{}); ok {
			// Composite policy - update the probabilistic sampler
			if err := s.updateCompositeProbabilisticSampler(composite, samplingPercentage); err != nil {
				return false, err
			}
		} else if probabilistic, ok := policy["probabilistic"].(map[string]interface{}); ok {
			// Simple probabilistic policy
			probabilistic["sampling_percentage"] = samplingPercentage
		} else {
			continue
		}

		return true, nil
	}

	return false, nil
}

// policyMatchesService checks if a policy matches a service name
func (s *workflowServiceImpl) policyMatchesService(policy map[string]interface{}, serviceName string) bool {
	// Check for string_attribute matching service.name
	stringAttr, ok := policy["string_attribute"].(map[string]interface{})
	if !ok {
		return false
	}

	key, _ := stringAttr["key"].(string)
	values, ok := stringAttr["values"].([]interface{})
	if !ok {
		return false
	}

	if key == "service.name" {
		for _, v := range values {
			if val, ok := v.(string); ok && val == serviceName {
				return true
			}
		}
	}

	return false
}

// updateCompositeProbabilisticSampler updates the probabilistic sampler in a composite policy
func (s *workflowServiceImpl) updateCompositeProbabilisticSampler(composite map[string]interface{}, samplingPercentage float64) error {
	policyOrder, ok := composite["policy_order"].([]interface{})
	if !ok {
		return fmt.Errorf("policy_order not found in composite policy")
	}

	// Find the probabilistic sampler
	for _, policyInterface := range policyOrder {
		policyItem, ok := policyInterface.(map[string]interface{})
		if !ok {
			continue
		}

		if probabilistic, ok := policyItem["probabilistic"].(map[string]interface{}); ok {
			probabilistic["sampling_percentage"] = samplingPercentage
			return nil
		}
	}

	return fmt.Errorf("probabilistic sampler not found in composite policy")
}

// GetWorkflowExecution retrieves a trigger execution by ID
func (s *workflowServiceImpl) GetWorkflowExecution(ctx context.Context, id string) (*types.WorkflowExecution, error) {
	return s.appStore.GetWorkflowExecution(ctx, id)
}

// ListWorkflowExecutions lists trigger executions with optional filtering
func (s *workflowServiceImpl) ListWorkflowExecutions(ctx context.Context, filter types.WorkflowExecutionFilter) ([]*types.WorkflowExecution, error) {
	return s.appStore.ListWorkflowExecutions(ctx, filter)
}

// ValidateWebhookSecret validates a webhook secret for a trigger
func (s *workflowServiceImpl) ValidateWebhookSecret(ctx context.Context, workflowID, secret string) bool {
	workflow, err := s.appStore.GetWorkflow(ctx, workflowID)
	if err != nil || workflow == nil {
		return false
	}

	// Use constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare([]byte(workflow.WebhookSecret), []byte(secret)) == 1
}
