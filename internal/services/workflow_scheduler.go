package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/getlawrence/lawrence-oss/internal/storage/applicationstore/types"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// WorkflowScheduler manages scheduled trigger execution
type WorkflowScheduler struct {
	cron           *cron.Cron
	triggerService WorkflowService
	logger         *zap.Logger
	running        bool
	mu             sync.RWMutex
	entries        map[string]cron.EntryID // trigger ID -> cron entry ID
}

// NewWorkflowScheduler creates a new trigger scheduler
func NewWorkflowScheduler(triggerService WorkflowService, logger *zap.Logger) *WorkflowScheduler {
	return &WorkflowScheduler{
		cron:           cron.New(cron.WithSeconds()),
		triggerService: triggerService,
		logger:         logger,
		entries:        make(map[string]cron.EntryID),
	}
}

// Start initializes the scheduler and loads all active schedule triggers
func (s *WorkflowScheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("scheduler is already running")
	}

	s.logger.Info("Starting trigger scheduler")

	// Load all active schedule triggers
	scheduleType := types.WorkflowTriggerTypeSchedule
	activeStatus := types.WorkflowStatusActive

	triggers, err := s.triggerService.ListWorkflows(ctx, types.WorkflowFilter{
		Type:   &scheduleType,
		Status: &activeStatus,
	})

	if err != nil {
		return fmt.Errorf("failed to load triggers: %w", err)
	}

	// Schedule each trigger
	for _, workflow := range triggers {
		if err := s.scheduleTrigger(workflow); err != nil {
			s.logger.Error("Failed to schedule trigger",
				zap.String("workflow_id", workflow.ID),
				zap.String("workflow_name", workflow.Name),
				zap.Error(err))
		}
	}

	s.cron.Start()
	s.running = true

	s.logger.Info("Trigger scheduler started", zap.Int("workflows_scheduled", len(triggers)))
	return nil
}

// Stop gracefully stops the scheduler
func (s *WorkflowScheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.logger.Info("Stopping trigger scheduler")
	s.cron.Stop()
	s.running = false
	s.entries = make(map[string]cron.EntryID)
	s.logger.Info("Trigger scheduler stopped")
}

// AddWorkflow adds or updates a scheduled trigger
func (s *WorkflowScheduler) AddWorkflow(workflow *types.Workflow) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("scheduler is not running")
	}

	// Remove existing entry if present
	if entryID, exists := s.entries[workflow.ID]; exists {
		s.cron.Remove(entryID)
		delete(s.entries, workflow.ID)
	}

	// Only schedule if active and is a schedule trigger
	if workflow.Type != types.WorkflowTriggerTypeSchedule || workflow.Status != types.WorkflowStatusActive {
		return nil
	}

	return s.scheduleTrigger(workflow)
}

// RemoveWorkflow removes a trigger from the schedule
func (s *WorkflowScheduler) RemoveWorkflow(workflowID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, exists := s.entries[workflowID]; exists {
		s.cron.Remove(entryID)
		delete(s.entries, workflowID)
		s.logger.Debug("Trigger removed from schedule", zap.String("workflow_id", workflowID))
	}
}

// scheduleTrigger schedules a single trigger (must be called with lock held)
func (s *WorkflowScheduler) scheduleTrigger(workflow *types.Workflow) error {
	if workflow.Schedule == nil {
		return fmt.Errorf("trigger has no schedule config")
	}

	// Parse timezone
	location := time.UTC
	if workflow.Schedule.Timezone != "" {
		loc, err := time.LoadLocation(workflow.Schedule.Timezone)
		if err != nil {
			s.logger.Warn("Invalid timezone, using UTC",
				zap.String("workflow_id", workflow.ID),
				zap.String("timezone", workflow.Schedule.Timezone),
				zap.Error(err))
		} else {
			location = loc
		}
	}

	// Schedule with timezone
	entryID, err := s.cron.AddFunc(workflow.Schedule.CronExpression, func() {
		ctx := context.Background()
		s.logger.Info("Executing scheduled trigger",
			zap.String("workflow_id", workflow.ID),
			zap.String("workflow_name", workflow.Name))

		_, err := s.triggerService.ExecuteWorkflow(ctx, workflow.ID, map[string]string{
			"source":    "scheduler",
			"scheduled": time.Now().In(location).Format(time.RFC3339),
		})

		if err != nil {
			s.logger.Error("Scheduled trigger execution failed",
				zap.String("workflow_id", workflow.ID),
				zap.String("workflow_name", workflow.Name),
				zap.Error(err))
		}
	})

	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	s.entries[workflow.ID] = entryID

	// Calculate next run time
	entry := s.cron.Entry(entryID)
	nextRun := entry.Next

	s.logger.Info("Trigger scheduled",
		zap.String("workflow_id", workflow.ID),
		zap.String("workflow_name", workflow.Name),
		zap.String("cron_expression", workflow.Schedule.CronExpression),
		zap.Time("next_run", nextRun))

	return nil
}

// IsRunning returns whether the scheduler is running
func (s *WorkflowScheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}
