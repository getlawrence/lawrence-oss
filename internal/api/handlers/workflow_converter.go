package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/getlawrence/lawrence-oss/internal/storage/applicationstore/types"
)

// FlowGraphNode represents a node in the flow graph
type FlowGraphNode struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Position *FlowGraphPosition     `json:"position,omitempty"`
	Data     map[string]interface{} `json:"data"`
}

// FlowGraphPosition represents the position of a node
type FlowGraphPosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// FlowGraphEdge represents an edge in the flow graph
type FlowGraphEdge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
	Label  string `json:"label,omitempty"`
}

// FlowGraph represents the flow graph structure from frontend
type FlowGraph struct {
	Nodes []FlowGraphNode `json:"nodes"`
	Edges []FlowGraphEdge `json:"edges"`
}

// WorkflowRequest represents a workflow request with normalized structure from frontend
type WorkflowRequest struct {
	types.Workflow
	FlowGraph *FlowGraph             `json:"flow_graph,omitempty"`
	Trigger   *types.WorkflowTrigger `json:"trigger,omitempty"`
	Steps     []*types.WorkflowStep  `json:"steps,omitempty"`
}

// WorkflowResponse represents a workflow response with normalized structure for frontend
type WorkflowResponse struct {
	types.Workflow
	Trigger   *types.WorkflowTrigger `json:"trigger"`
	Steps     []*types.WorkflowStep  `json:"steps"`
	FlowGraph *FlowGraph             `json:"flow_graph,omitempty"`
}

// ConvertFlowGraphToNormalized converts a flow_graph to trigger and steps
func ConvertFlowGraphToNormalized(workflowID string, workflowType types.WorkflowTriggerType, schedule *types.ScheduleConfig, webhookURL, webhookSecret string, flowGraph *FlowGraph) (*types.WorkflowTrigger, []*types.WorkflowStep, error) {
	if flowGraph == nil {
		return nil, nil, fmt.Errorf("flow_graph is required")
	}

	// Find trigger node
	var triggerNode *FlowGraphNode
	for i := range flowGraph.Nodes {
		if flowGraph.Nodes[i].Type == "trigger" {
			triggerNode = &flowGraph.Nodes[i]
			break
		}
	}

	if triggerNode == nil {
		return nil, nil, fmt.Errorf("trigger node not found in flow_graph")
	}

	// Extract trigger type from node data
	triggerTypeStr, ok := triggerNode.Data["triggerType"].(string)
	if !ok {
		triggerTypeStr = string(workflowType)
	}

	triggerType := types.WorkflowTriggerType(triggerTypeStr)

	// Create trigger
	trigger := &types.WorkflowTrigger{
		WorkflowID:    workflowID,
		Type:          triggerType,
		Enabled:       true,
		Schedule:      schedule,
		WebhookURL:    webhookURL,
		WebhookSecret: webhookSecret,
	}

	// Extract schedule from trigger node if it's a schedule trigger
	if triggerType == types.WorkflowTriggerTypeSchedule {
		if cronExpr, ok := triggerNode.Data["cronExpression"].(string); ok && cronExpr != "" {
			timezone := "UTC"
			if tz, ok := triggerNode.Data["timezone"].(string); ok && tz != "" {
				timezone = tz
			}
			trigger.Schedule = &types.ScheduleConfig{
				CronExpression: cronExpr,
				Timezone:       timezone,
			}
		}
	}

	// Build dependency map from edges
	dependencyMap := make(map[string][]string) // target -> []sources
	for _, edge := range flowGraph.Edges {
		dependencyMap[edge.Target] = append(dependencyMap[edge.Target], edge.Source)
	}

	// Convert nodes to steps (excluding trigger node)
	var steps []*types.WorkflowStep
	order := 0
	for _, node := range flowGraph.Nodes {
		// Skip trigger node
		if node.Type == "trigger" {
			continue
		}

		// Get dependencies for this node
		dependencies := dependencyMap[node.ID]
		// Remove trigger from dependencies (it's implicit)
		var filteredDeps []string
		for _, dep := range dependencies {
			if dep != triggerNode.ID {
				filteredDeps = append(filteredDeps, dep)
			}
		}

		// Get label/name from node data
		name := node.ID
		if label, ok := node.Data["label"].(string); ok && label != "" {
			name = label
		}

		// Get description
		description := ""
		if desc, ok := node.Data["description"].(string); ok {
			description = desc
		}

		// Get continueOnError for action nodes
		continueOnError := false
		if co, ok := node.Data["continueOnError"].(bool); ok {
			continueOnError = co
		}

		// Serialize node data to JSON
		configJSON, err := json.Marshal(node.Data)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to serialize node data for node %s: %w", node.ID, err)
		}

		// Get position
		var positionX, positionY *float64
		if node.Position != nil {
			positionX = &node.Position.X
			positionY = &node.Position.Y
		}

		step := &types.WorkflowStep{
			ID:              node.ID,
			WorkflowID:      workflowID,
			Type:            types.FlowNodeType(node.Type),
			Name:            name,
			Description:     description,
			Order:           order,
			PositionX:       positionX,
			PositionY:       positionY,
			ConfigJSON:      string(configJSON),
			ContinueOnError: continueOnError,
			DependsOn:       filteredDeps,
		}

		steps = append(steps, step)
		order++
	}

	return trigger, steps, nil
}

// ConvertNormalizedToFlowGraph converts normalized workflow data back into a flow graph representation
func ConvertNormalizedToFlowGraph(workflowID string, trigger *types.WorkflowTrigger, steps []*types.WorkflowStep) (*FlowGraph, error) {
	if trigger == nil {
		return nil, fmt.Errorf("trigger is required to build flow_graph")
	}

	triggerNodeID := fmt.Sprintf("trigger-%s", workflowID)
	triggerData := map[string]interface{}{
		"label":       "Trigger",
		"triggerType": trigger.Type,
	}

	if trigger.Type == types.WorkflowTriggerTypeSchedule && trigger.Schedule != nil {
		triggerData["cronExpression"] = trigger.Schedule.CronExpression
		triggerData["timezone"] = trigger.Schedule.Timezone
	}

	nodes := []FlowGraphNode{
		{
			ID:   triggerNodeID,
			Type: string(types.FlowNodeTypeTrigger),
			Data: triggerData,
		},
	}

	var edges []FlowGraphEdge
	edgeSet := make(map[string]struct{})

	for _, step := range steps {
		if step == nil {
			continue
		}

		data := make(map[string]interface{})
		if step.ConfigJSON != "" {
			if err := json.Unmarshal([]byte(step.ConfigJSON), &data); err != nil || data == nil {
				data = make(map[string]interface{})
			}
		}

		if _, ok := data["label"]; !ok && step.Name != "" {
			data["label"] = step.Name
		}

		if step.Description != "" {
			if _, ok := data["description"]; !ok {
				data["description"] = step.Description
			}
		}

		if step.ContinueOnError {
			if _, ok := data["continueOnError"]; !ok {
				data["continueOnError"] = step.ContinueOnError
			}
		}

		node := FlowGraphNode{
			ID:   step.ID,
			Type: string(step.Type),
			Data: data,
		}

		if step.PositionX != nil || step.PositionY != nil {
			pos := FlowGraphPosition{}
			if step.PositionX != nil {
				pos.X = *step.PositionX
			}
			if step.PositionY != nil {
				pos.Y = *step.PositionY
			}
			node.Position = &pos
		}

		nodes = append(nodes, node)

		if len(step.DependsOn) == 0 {
			edgeID := fmt.Sprintf("%s-%s", triggerNodeID, step.ID)
			if _, exists := edgeSet[edgeID]; !exists {
				edges = append(edges, FlowGraphEdge{
					ID:     edgeID,
					Source: triggerNodeID,
					Target: step.ID,
				})
				edgeSet[edgeID] = struct{}{}
			}
			continue
		}

		for _, dep := range step.DependsOn {
			if dep == "" {
				continue
			}
			edgeID := fmt.Sprintf("%s-%s", dep, step.ID)
			if _, exists := edgeSet[edgeID]; exists {
				continue
			}
			edges = append(edges, FlowGraphEdge{
				ID:     edgeID,
				Source: dep,
				Target: step.ID,
			})
			edgeSet[edgeID] = struct{}{}
		}
	}

	return &FlowGraph{
		Nodes: nodes,
		Edges: edges,
	}, nil
}
