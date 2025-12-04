// Copyright (c) 2024 Lawrence OSS Contributors
// SPDX-License-Identifier: Apache-2.0

package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/getlawrence/lawrence-oss/internal/storage/applicationstore/types"
	"github.com/google/uuid"
)

// Store is an in-memory implementation of ApplicationStore
type Store struct {
	mu                 sync.RWMutex
	agents             map[uuid.UUID]*types.Agent
	groups             map[string]*types.Group
	configs            map[string]*types.Config
	workflows          map[string]*types.Workflow
	workflowExecutions map[string]*types.WorkflowExecution
	workflowTriggers   map[string]*types.WorkflowTrigger
	workflowSteps      map[string]*types.WorkflowStep
	stepExecutions     map[string]*types.StepExecution
	stepExecutionLogs  map[string]*types.StepExecutionLog
	delayedActions     map[string]*types.DelayedActionQueue
}

// NewStore creates a new in-memory store
func NewStore() *Store {
	return &Store{
		agents:             make(map[uuid.UUID]*types.Agent),
		groups:             make(map[string]*types.Group),
		configs:            make(map[string]*types.Config),
		workflows:          make(map[string]*types.Workflow),
		workflowExecutions: make(map[string]*types.WorkflowExecution),
		workflowTriggers:   make(map[string]*types.WorkflowTrigger),
		workflowSteps:      make(map[string]*types.WorkflowStep),
		stepExecutions:     make(map[string]*types.StepExecution),
		stepExecutionLogs:  make(map[string]*types.StepExecutionLog),
		delayedActions:     make(map[string]*types.DelayedActionQueue),
	}
}

// Agent management

func (s *Store) CreateAgent(ctx context.Context, agent *types.Agent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.agents[agent.ID]; exists {
		return fmt.Errorf("agent already exists: %s", agent.ID)
	}

	// Deep copy to prevent external modifications
	agentCopy := *agent
	if agent.Labels != nil {
		agentCopy.Labels = make(map[string]string, len(agent.Labels))
		for k, v := range agent.Labels {
			agentCopy.Labels[k] = v
		}
	}
	if agent.Capabilities != nil {
		agentCopy.Capabilities = make([]string, len(agent.Capabilities))
		copy(agentCopy.Capabilities, agent.Capabilities)
	}

	s.agents[agent.ID] = &agentCopy
	return nil
}

func (s *Store) GetAgent(ctx context.Context, id uuid.UUID) (*types.Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	agent, exists := s.agents[id]
	if !exists {
		return nil, nil
	}

	// Deep copy to prevent external modifications
	agentCopy := *agent
	if agent.Labels != nil {
		agentCopy.Labels = make(map[string]string, len(agent.Labels))
		for k, v := range agent.Labels {
			agentCopy.Labels[k] = v
		}
	}
	if agent.Capabilities != nil {
		agentCopy.Capabilities = make([]string, len(agent.Capabilities))
		copy(agentCopy.Capabilities, agent.Capabilities)
	}

	return &agentCopy, nil
}

func (s *Store) ListAgents(ctx context.Context) ([]*types.Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	agents := make([]*types.Agent, 0, len(s.agents))
	for _, agent := range s.agents {
		// Deep copy
		agentCopy := *agent
		if agent.Labels != nil {
			agentCopy.Labels = make(map[string]string, len(agent.Labels))
			for k, v := range agent.Labels {
				agentCopy.Labels[k] = v
			}
		}
		if agent.Capabilities != nil {
			agentCopy.Capabilities = make([]string, len(agent.Capabilities))
			copy(agentCopy.Capabilities, agent.Capabilities)
		}
		agents = append(agents, &agentCopy)
	}

	return agents, nil
}

func (s *Store) UpdateAgentStatus(ctx context.Context, id uuid.UUID, status types.AgentStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	agent, exists := s.agents[id]
	if !exists {
		return fmt.Errorf("agent not found: %s", id)
	}

	agent.Status = status
	agent.UpdatedAt = time.Now()
	return nil
}

func (s *Store) UpdateAgentLastSeen(ctx context.Context, id uuid.UUID, lastSeen time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	agent, exists := s.agents[id]
	if !exists {
		return fmt.Errorf("agent not found: %s", id)
	}

	agent.LastSeen = lastSeen
	agent.UpdatedAt = time.Now()
	return nil
}

func (s *Store) UpdateAgentEffectiveConfig(ctx context.Context, id uuid.UUID, effectiveConfig string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	agent, exists := s.agents[id]
	if !exists {
		return fmt.Errorf("agent not found: %s", id)
	}

	agent.EffectiveConfig = effectiveConfig
	agent.UpdatedAt = time.Now()
	return nil
}

func (s *Store) DeleteAgent(ctx context.Context, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.agents[id]; !exists {
		return fmt.Errorf("agent not found: %s", id)
	}

	delete(s.agents, id)
	return nil
}

// Group management

func (s *Store) CreateGroup(ctx context.Context, group *types.Group) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.groups[group.ID]; exists {
		return fmt.Errorf("group already exists: %s", group.ID)
	}

	// Deep copy
	groupCopy := *group
	if group.Labels != nil {
		groupCopy.Labels = make(map[string]string, len(group.Labels))
		for k, v := range group.Labels {
			groupCopy.Labels[k] = v
		}
	}

	s.groups[group.ID] = &groupCopy
	return nil
}

func (s *Store) GetGroup(ctx context.Context, id string) (*types.Group, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	group, exists := s.groups[id]
	if !exists {
		return nil, nil
	}

	// Deep copy
	groupCopy := *group
	if group.Labels != nil {
		groupCopy.Labels = make(map[string]string, len(group.Labels))
		for k, v := range group.Labels {
			groupCopy.Labels[k] = v
		}
	}

	return &groupCopy, nil
}

func (s *Store) ListGroups(ctx context.Context) ([]*types.Group, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	groups := make([]*types.Group, 0, len(s.groups))
	for _, group := range s.groups {
		// Deep copy
		groupCopy := *group
		if group.Labels != nil {
			groupCopy.Labels = make(map[string]string, len(group.Labels))
			for k, v := range group.Labels {
				groupCopy.Labels[k] = v
			}
		}
		groups = append(groups, &groupCopy)
	}

	return groups, nil
}

func (s *Store) DeleteGroup(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.groups[id]; !exists {
		return fmt.Errorf("group not found: %s", id)
	}

	delete(s.groups, id)
	return nil
}

// Config management

func (s *Store) CreateConfig(ctx context.Context, config *types.Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.configs[config.ID]; exists {
		return fmt.Errorf("config already exists: %s", config.ID)
	}

	// Deep copy
	configCopy := *config
	s.configs[config.ID] = &configCopy
	return nil
}

func (s *Store) GetConfig(ctx context.Context, id string) (*types.Config, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	config, exists := s.configs[id]
	if !exists {
		return nil, nil
	}

	// Deep copy
	configCopy := *config
	return &configCopy, nil
}

func (s *Store) GetLatestConfigForAgent(ctx context.Context, agentID uuid.UUID) (*types.Config, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var latestConfig *types.Config
	for _, config := range s.configs {
		if config.AgentID != nil && *config.AgentID == agentID {
			if latestConfig == nil || config.Version > latestConfig.Version ||
				(config.Version == latestConfig.Version && config.CreatedAt.After(latestConfig.CreatedAt)) {
				latestConfig = config
			}
		}
	}

	if latestConfig == nil {
		return nil, nil
	}

	// Deep copy
	configCopy := *latestConfig
	return &configCopy, nil
}

func (s *Store) GetLatestConfigForGroup(ctx context.Context, groupID string) (*types.Config, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var latestConfig *types.Config
	for _, config := range s.configs {
		if config.GroupID != nil && *config.GroupID == groupID {
			if latestConfig == nil || config.Version > latestConfig.Version ||
				(config.Version == latestConfig.Version && config.CreatedAt.After(latestConfig.CreatedAt)) {
				latestConfig = config
			}
		}
	}

	if latestConfig == nil {
		return nil, nil
	}

	// Deep copy
	configCopy := *latestConfig
	return &configCopy, nil
}

func (s *Store) ListConfigs(ctx context.Context, filter types.ConfigFilter) ([]*types.Config, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	configs := make([]*types.Config, 0)
	for _, config := range s.configs {
		// Apply filters
		if filter.AgentID != nil && (config.AgentID == nil || *config.AgentID != *filter.AgentID) {
			continue
		}
		if filter.GroupID != nil && (config.GroupID == nil || *config.GroupID != *filter.GroupID) {
			continue
		}

		// Deep copy
		configCopy := *config
		configs = append(configs, &configCopy)
	}

	// Apply limit
	if filter.Limit > 0 && len(configs) > filter.Limit {
		configs = configs[:filter.Limit]
	}

	return configs, nil
}

// Workflow management

func (s *Store) CreateWorkflow(ctx context.Context, workflow *types.Workflow) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.workflows[workflow.ID]; exists {
		return fmt.Errorf("workflow already exists: %s", workflow.ID)
	}

	workflowCopy := *workflow
	s.workflows[workflow.ID] = &workflowCopy
	return nil
}

func (s *Store) GetWorkflow(ctx context.Context, id string) (*types.Workflow, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	workflow, exists := s.workflows[id]
	if !exists {
		return nil, nil
	}

	workflowCopy := *workflow
	return &workflowCopy, nil
}

func (s *Store) ListWorkflows(ctx context.Context, filter types.WorkflowFilter) ([]*types.Workflow, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var workflows []*types.Workflow
	for _, workflow := range s.workflows {
		if filter.Type != nil && workflow.Type != *filter.Type {
			continue
		}
		if filter.Status != nil && workflow.Status != *filter.Status {
			continue
		}

		workflowCopy := *workflow
		workflows = append(workflows, &workflowCopy)
	}

	if filter.Limit > 0 && len(workflows) > filter.Limit {
		workflows = workflows[:filter.Limit]
	}

	return workflows, nil
}

func (s *Store) UpdateWorkflow(ctx context.Context, workflow *types.Workflow) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.workflows[workflow.ID]; !exists {
		return fmt.Errorf("workflow not found: %s", workflow.ID)
	}

	workflowCopy := *workflow
	s.workflows[workflow.ID] = &workflowCopy
	return nil
}

func (s *Store) DeleteWorkflow(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.workflows, id)
	return nil
}

// Workflow execution tracking

func (s *Store) CreateWorkflowExecution(ctx context.Context, execution *types.WorkflowExecution) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	executionCopy := *execution
	s.workflowExecutions[execution.ID] = &executionCopy
	return nil
}

func (s *Store) GetWorkflowExecution(ctx context.Context, id string) (*types.WorkflowExecution, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	execution, exists := s.workflowExecutions[id]
	if !exists {
		return nil, nil
	}

	executionCopy := *execution
	return &executionCopy, nil
}

func (s *Store) ListWorkflowExecutions(ctx context.Context, filter types.WorkflowExecutionFilter) ([]*types.WorkflowExecution, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var executions []*types.WorkflowExecution
	for _, execution := range s.workflowExecutions {
		if filter.WorkflowID != nil && execution.WorkflowID != *filter.WorkflowID {
			continue
		}
		if filter.Status != nil && execution.Status != *filter.Status {
			continue
		}

		executionCopy := *execution
		executions = append(executions, &executionCopy)
	}

	if filter.Limit > 0 && len(executions) > filter.Limit {
		executions = executions[:filter.Limit]
	}

	return executions, nil
}

func (s *Store) UpdateWorkflowExecution(ctx context.Context, execution *types.WorkflowExecution) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.workflowExecutions[execution.ID]; !exists {
		return fmt.Errorf("workflow execution not found: %s", execution.ID)
	}

	executionCopy := *execution
	s.workflowExecutions[execution.ID] = &executionCopy
	return nil
}

// Delayed action queue management

func (s *Store) CreateDelayedAction(ctx context.Context, action *types.DelayedActionQueue) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.delayedActions[action.ID]; exists {
		return fmt.Errorf("delayed action already exists: %s", action.ID)
	}

	actionCopy := *action
	if action.Metadata != nil {
		actionCopy.Metadata = make(map[string]string, len(action.Metadata))
		for k, v := range action.Metadata {
			actionCopy.Metadata[k] = v
		}
	}

	s.delayedActions[action.ID] = &actionCopy
	return nil
}

func (s *Store) GetDelayedAction(ctx context.Context, id string) (*types.DelayedActionQueue, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	action, exists := s.delayedActions[id]
	if !exists {
		return nil, nil
	}

	actionCopy := *action
	if action.Metadata != nil {
		actionCopy.Metadata = make(map[string]string, len(action.Metadata))
		for k, v := range action.Metadata {
			actionCopy.Metadata[k] = v
		}
	}

	return &actionCopy, nil
}

func (s *Store) ListPendingDelayedActions(ctx context.Context) ([]*types.DelayedActionQueue, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	var actions []*types.DelayedActionQueue

	for _, action := range s.delayedActions {
		if action.Status == "pending" && (action.ScheduledFor.Before(now) || action.ScheduledFor.Equal(now)) {
			actionCopy := *action
			if action.Metadata != nil {
				actionCopy.Metadata = make(map[string]string, len(action.Metadata))
				for k, v := range action.Metadata {
					actionCopy.Metadata[k] = v
				}
			}
			actions = append(actions, &actionCopy)
		}
	}

	return actions, nil
}

func (s *Store) UpdateDelayedAction(ctx context.Context, action *types.DelayedActionQueue) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.delayedActions[action.ID]; !exists {
		return fmt.Errorf("delayed action not found: %s", action.ID)
	}

	actionCopy := *action
	if action.Metadata != nil {
		actionCopy.Metadata = make(map[string]string, len(action.Metadata))
		for k, v := range action.Metadata {
			actionCopy.Metadata[k] = v
		}
	}

	s.delayedActions[action.ID] = &actionCopy
	return nil
}

func (s *Store) DeleteDelayedAction(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.delayedActions, id)
	return nil
}

// Workflow trigger management

func (s *Store) CreateWorkflowTrigger(ctx context.Context, trigger *types.WorkflowTrigger) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.workflowTriggers[trigger.WorkflowID]; exists {
		return fmt.Errorf("workflow trigger already exists: %s", trigger.WorkflowID)
	}

	triggerCopy := *trigger
	if trigger.Schedule != nil {
		scheduleCopy := *trigger.Schedule
		triggerCopy.Schedule = &scheduleCopy
	}
	s.workflowTriggers[trigger.WorkflowID] = &triggerCopy
	return nil
}

func (s *Store) GetWorkflowTrigger(ctx context.Context, workflowID string) (*types.WorkflowTrigger, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	trigger, exists := s.workflowTriggers[workflowID]
	if !exists {
		return nil, nil
	}

	triggerCopy := *trigger
	if trigger.Schedule != nil {
		scheduleCopy := *trigger.Schedule
		triggerCopy.Schedule = &scheduleCopy
	}
	return &triggerCopy, nil
}

func (s *Store) UpdateWorkflowTrigger(ctx context.Context, trigger *types.WorkflowTrigger) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.workflowTriggers[trigger.WorkflowID]; !exists {
		return fmt.Errorf("workflow trigger not found: %s", trigger.WorkflowID)
	}

	triggerCopy := *trigger
	if trigger.Schedule != nil {
		scheduleCopy := *trigger.Schedule
		triggerCopy.Schedule = &scheduleCopy
	}
	s.workflowTriggers[trigger.WorkflowID] = &triggerCopy
	return nil
}

func (s *Store) DeleteWorkflowTrigger(ctx context.Context, workflowID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.workflowTriggers, workflowID)
	return nil
}

// Workflow step management

func (s *Store) CreateWorkflowStep(ctx context.Context, step *types.WorkflowStep) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.workflowSteps[step.ID]; exists {
		return fmt.Errorf("workflow step already exists: %s", step.ID)
	}

	stepCopy := *step
	if step.DependsOn != nil {
		dependsOnCopy := make([]string, len(step.DependsOn))
		copy(dependsOnCopy, step.DependsOn)
		stepCopy.DependsOn = dependsOnCopy
	}
	s.workflowSteps[step.ID] = &stepCopy
	return nil
}

func (s *Store) GetWorkflowStep(ctx context.Context, stepID string) (*types.WorkflowStep, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	step, exists := s.workflowSteps[stepID]
	if !exists {
		return nil, nil
	}

	stepCopy := *step
	if step.DependsOn != nil {
		dependsOnCopy := make([]string, len(step.DependsOn))
		copy(dependsOnCopy, step.DependsOn)
		stepCopy.DependsOn = dependsOnCopy
	}
	return &stepCopy, nil
}

func (s *Store) ListWorkflowSteps(ctx context.Context, workflowID string) ([]*types.WorkflowStep, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var steps []*types.WorkflowStep
	for _, step := range s.workflowSteps {
		if step.WorkflowID == workflowID {
			stepCopy := *step
			if step.DependsOn != nil {
				dependsOnCopy := make([]string, len(step.DependsOn))
				copy(dependsOnCopy, step.DependsOn)
				stepCopy.DependsOn = dependsOnCopy
			}
			steps = append(steps, &stepCopy)
		}
	}

	// Sort by order
	for i := 0; i < len(steps)-1; i++ {
		for j := i + 1; j < len(steps); j++ {
			if steps[i].Order > steps[j].Order {
				steps[i], steps[j] = steps[j], steps[i]
			}
		}
	}

	return steps, nil
}

func (s *Store) UpdateWorkflowStep(ctx context.Context, step *types.WorkflowStep) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.workflowSteps[step.ID]; !exists {
		return fmt.Errorf("workflow step not found: %s", step.ID)
	}

	stepCopy := *step
	if step.DependsOn != nil {
		dependsOnCopy := make([]string, len(step.DependsOn))
		copy(dependsOnCopy, step.DependsOn)
		stepCopy.DependsOn = dependsOnCopy
	}
	s.workflowSteps[step.ID] = &stepCopy
	return nil
}

func (s *Store) DeleteWorkflowStep(ctx context.Context, stepID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.workflowSteps, stepID)
	return nil
}

func (s *Store) DeleteWorkflowSteps(ctx context.Context, workflowID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, step := range s.workflowSteps {
		if step.WorkflowID == workflowID {
			delete(s.workflowSteps, id)
		}
	}
	return nil
}

// Step execution management

func (s *Store) CreateStepExecution(ctx context.Context, execution *types.StepExecution) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.stepExecutions[execution.ID]; exists {
		return fmt.Errorf("step execution already exists: %s", execution.ID)
	}

	executionCopy := *execution
	if execution.InputData != nil {
		inputDataCopy := make(map[string]interface{})
		for k, v := range execution.InputData {
			inputDataCopy[k] = v
		}
		executionCopy.InputData = inputDataCopy
	}
	if execution.OutputData != nil {
		outputDataCopy := make(map[string]interface{})
		for k, v := range execution.OutputData {
			outputDataCopy[k] = v
		}
		executionCopy.OutputData = outputDataCopy
	}
	if execution.Metadata != nil {
		metadataCopy := make(map[string]string)
		for k, v := range execution.Metadata {
			metadataCopy[k] = v
		}
		executionCopy.Metadata = metadataCopy
	}
	s.stepExecutions[execution.ID] = &executionCopy
	return nil
}

func (s *Store) GetStepExecution(ctx context.Context, executionID string) (*types.StepExecution, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	execution, exists := s.stepExecutions[executionID]
	if !exists {
		return nil, nil
	}

	executionCopy := *execution
	if execution.InputData != nil {
		inputDataCopy := make(map[string]interface{})
		for k, v := range execution.InputData {
			inputDataCopy[k] = v
		}
		executionCopy.InputData = inputDataCopy
	}
	if execution.OutputData != nil {
		outputDataCopy := make(map[string]interface{})
		for k, v := range execution.OutputData {
			outputDataCopy[k] = v
		}
		executionCopy.OutputData = outputDataCopy
	}
	if execution.Metadata != nil {
		metadataCopy := make(map[string]string)
		for k, v := range execution.Metadata {
			metadataCopy[k] = v
		}
		executionCopy.Metadata = metadataCopy
	}
	return &executionCopy, nil
}

func (s *Store) ListStepExecutions(ctx context.Context, filter types.StepExecutionFilter) ([]*types.StepExecution, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var executions []*types.StepExecution
	for _, execution := range s.stepExecutions {
		if filter.WorkflowID != nil && execution.WorkflowID != *filter.WorkflowID {
			continue
		}
		if filter.WorkflowExecutionID != nil && execution.WorkflowExecutionID != *filter.WorkflowExecutionID {
			continue
		}
		if filter.StepID != nil && execution.StepID != *filter.StepID {
			continue
		}
		if filter.Status != nil && execution.Status != *filter.Status {
			continue
		}

		executionCopy := *execution
		if execution.InputData != nil {
			inputDataCopy := make(map[string]interface{})
			for k, v := range execution.InputData {
				inputDataCopy[k] = v
			}
			executionCopy.InputData = inputDataCopy
		}
		if execution.OutputData != nil {
			outputDataCopy := make(map[string]interface{})
			for k, v := range execution.OutputData {
				outputDataCopy[k] = v
			}
			executionCopy.OutputData = outputDataCopy
		}
		if execution.Metadata != nil {
			metadataCopy := make(map[string]string)
			for k, v := range execution.Metadata {
				metadataCopy[k] = v
			}
			executionCopy.Metadata = metadataCopy
		}
		executions = append(executions, &executionCopy)
	}

	// Sort by started_at descending
	for i := 0; i < len(executions)-1; i++ {
		for j := i + 1; j < len(executions); j++ {
			if executions[i].StartedAt.Before(executions[j].StartedAt) {
				executions[i], executions[j] = executions[j], executions[i]
			}
		}
	}

	if filter.Limit > 0 && len(executions) > filter.Limit {
		executions = executions[:filter.Limit]
	}

	return executions, nil
}

func (s *Store) UpdateStepExecution(ctx context.Context, execution *types.StepExecution) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.stepExecutions[execution.ID]; !exists {
		return fmt.Errorf("step execution not found: %s", execution.ID)
	}

	executionCopy := *execution
	if execution.InputData != nil {
		inputDataCopy := make(map[string]interface{})
		for k, v := range execution.InputData {
			inputDataCopy[k] = v
		}
		executionCopy.InputData = inputDataCopy
	}
	if execution.OutputData != nil {
		outputDataCopy := make(map[string]interface{})
		for k, v := range execution.OutputData {
			outputDataCopy[k] = v
		}
		executionCopy.OutputData = outputDataCopy
	}
	if execution.Metadata != nil {
		metadataCopy := make(map[string]string)
		for k, v := range execution.Metadata {
			metadataCopy[k] = v
		}
		executionCopy.Metadata = metadataCopy
	}
	s.stepExecutions[execution.ID] = &executionCopy
	return nil
}

// Step execution log management

func (s *Store) CreateStepExecutionLog(ctx context.Context, log *types.StepExecutionLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	logCopy := *log
	s.stepExecutionLogs[log.ID] = &logCopy
	return nil
}

func (s *Store) ListStepExecutionLogs(ctx context.Context, stepExecutionID string, limit int) ([]*types.StepExecutionLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var logs []*types.StepExecutionLog
	for _, log := range s.stepExecutionLogs {
		if log.StepExecutionID == stepExecutionID {
			logCopy := *log
			logs = append(logs, &logCopy)
		}
	}

	// Sort by timestamp descending
	for i := 0; i < len(logs)-1; i++ {
		for j := i + 1; j < len(logs); j++ {
			if logs[i].Timestamp.Before(logs[j].Timestamp) {
				logs[i], logs[j] = logs[j], logs[i]
			}
		}
	}

	if limit > 0 && len(logs) > limit {
		logs = logs[:limit]
	}

	return logs, nil
}

// purge removes all data from the store (for testing)
func (s *Store) purge(context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.agents = make(map[uuid.UUID]*types.Agent)
	s.groups = make(map[string]*types.Group)
	s.configs = make(map[string]*types.Config)
	s.workflows = make(map[string]*types.Workflow)
	s.workflowExecutions = make(map[string]*types.WorkflowExecution)
	s.workflowTriggers = make(map[string]*types.WorkflowTrigger)
	s.workflowSteps = make(map[string]*types.WorkflowStep)
	s.stepExecutions = make(map[string]*types.StepExecution)
	s.stepExecutionLogs = make(map[string]*types.StepExecutionLog)
	s.delayedActions = make(map[string]*types.DelayedActionQueue)
}
