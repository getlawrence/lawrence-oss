package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/getlawrence/lawrence-oss/internal/storage/applicationstore/types"
	"go.uber.org/zap"
)

// DelayedActionExecutor manages the execution of delayed actions
type DelayedActionExecutor struct {
	appStore        types.ApplicationStore
	workflowService WorkflowService
	logger          *zap.Logger
	running         bool
	mu              sync.RWMutex
	stopChan        chan struct{}
	interval        time.Duration
}

// NewDelayedActionExecutor creates a new delayed action executor
func NewDelayedActionExecutor(
	appStore types.ApplicationStore,
	workflowService WorkflowService,
	logger *zap.Logger,
) *DelayedActionExecutor {
	return &DelayedActionExecutor{
		appStore:        appStore,
		workflowService: workflowService,
		logger:          logger,
		interval:        10 * time.Second, // Check every 10 seconds
		stopChan:        make(chan struct{}),
	}
}

// Start begins processing delayed actions
func (e *DelayedActionExecutor) Start(ctx context.Context) error {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return fmt.Errorf("delayed action executor is already running")
	}
	e.running = true
	e.mu.Unlock()

	e.logger.Info("Starting delayed action executor", zap.Duration("interval", e.interval))

	go e.run(ctx)

	return nil
}

// Stop gracefully stops the executor
func (e *DelayedActionExecutor) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return
	}

	e.logger.Info("Stopping delayed action executor")
	close(e.stopChan)
	e.running = false
}

// run is the main loop that processes delayed actions
func (e *DelayedActionExecutor) run(ctx context.Context) {
	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()

	for {
		select {
		case <-e.stopChan:
			e.logger.Info("Delayed action executor stopped")
			return
		case <-ticker.C:
			if err := e.processDelayedActions(ctx); err != nil {
				e.logger.Error("Failed to process delayed actions", zap.Error(err))
			}
		}
	}
}

// processDelayedActions fetches and executes all pending delayed actions
func (e *DelayedActionExecutor) processDelayedActions(ctx context.Context) error {
	actions, err := e.appStore.ListPendingDelayedActions(ctx)
	if err != nil {
		return fmt.Errorf("failed to list pending actions: %w", err)
	}

	if len(actions) == 0 {
		return nil
	}

	e.logger.Info("Processing delayed actions", zap.Int("count", len(actions)))

	for _, action := range actions {
		if err := e.executeDelayedAction(ctx, action); err != nil {
			e.logger.Error("Failed to execute delayed action",
				zap.String("id", action.ID),
				zap.String("trigger_id", action.WorkflowID),
				zap.Error(err))

			// Mark as failed
			now := time.Now()
			action.Status = "failed"
			action.Error = err.Error()
			action.CompletedAt = &now

			if updateErr := e.appStore.UpdateDelayedAction(ctx, action); updateErr != nil {
				e.logger.Error("Failed to update delayed action status", zap.Error(updateErr))
			}
		}
	}

	return nil
}

// executeDelayedAction executes a single delayed action
func (e *DelayedActionExecutor) executeDelayedAction(ctx context.Context, queueItem *types.DelayedActionQueue) error {
	e.logger.Info("Executing delayed action",
		zap.String("id", queueItem.ID),
		zap.String("trigger_id", queueItem.WorkflowID),
		zap.Time("scheduled_for", queueItem.ScheduledFor))

	// Mark as executing
	queueItem.Status = "executing"
	if err := e.appStore.UpdateDelayedAction(ctx, queueItem); err != nil {
		return fmt.Errorf("failed to update action status to executing: %w", err)
	}

	// Get the workflow service implementation
	workflowServiceImpl, ok := e.workflowService.(*workflowServiceImpl)
	if !ok {
		return fmt.Errorf("workflow service is not the expected implementation")
	}

	// Get the original workflow for context
	trigger, err := e.appStore.GetWorkflow(ctx, queueItem.WorkflowID)
	if err != nil {
		return fmt.Errorf("failed to get trigger: %w", err)
	}

	// Create a mock execution record for this delayed action
	// This allows us to track what happened
	execution := &types.WorkflowExecution{
		ID:           queueItem.ExecutionID,
		WorkflowID:   queueItem.WorkflowID,
		WorkflowName: trigger.Name + " (Delayed Action)",
	}

	// Execute the action
	if err := workflowServiceImpl.executeAction(ctx, trigger, &queueItem.Action, execution); err != nil {
		return fmt.Errorf("failed to execute action: %w", err)
	}

	// Mark as completed
	now := time.Now()
	queueItem.Status = "completed"
	queueItem.CompletedAt = &now

	if err := e.appStore.UpdateDelayedAction(ctx, queueItem); err != nil {
		return fmt.Errorf("failed to update action status to completed: %w", err)
	}

	e.logger.Info("Delayed action completed",
		zap.String("id", queueItem.ID),
		zap.String("trigger_id", queueItem.WorkflowID))

	return nil
}

// IsRunning returns whether the executor is running
func (e *DelayedActionExecutor) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.running
}
