import type { Node, Edge } from "@xyflow/react";

import type {
  Workflow,
  FlowNode,
  DelayNodeData,
  TriggerNodeData,
   ActionNodeData,
} from "../types/flow-types";

import type { Workflow as ApiWorkflow, WorkflowTrigger, WorkflowStep } from "@/api/workflows";

/**
 * Convert trigger + steps to flow_graph format
 */
function normalizeToFlowGraph(
  workflow: ApiWorkflow,
  trigger: WorkflowTrigger,
  steps: WorkflowStep[],
): Workflow {
  const nodes: FlowNode[] = [];
  const edges: Edge[] = [];

  // Create trigger node
  const triggerNodeId = `trigger-${workflow.id}`;
  const triggerNodeData: TriggerNodeData = {
    label: "Trigger",
    triggerType: trigger.type,
    description: "",
  };

  // Add schedule info if it's a schedule trigger
  if (trigger.type === "schedule" && trigger.schedule) {
    triggerNodeData.cronExpression = trigger.schedule.cron_expression;
    triggerNodeData.timezone = trigger.schedule.timezone;
  }

  nodes.push({
    id: triggerNodeId,
    type: "trigger",
    position: { x: 0, y: 0 },
    data: triggerNodeData,
  });

  // First, create all step nodes
  for (const step of steps) {
    try {
      // Parse config_json to get node data
      const stepData = JSON.parse(step.config_json);

      // Create node with position
      const node: FlowNode = {
        id: step.id,
        type: step.type as any,
        position: {
          x: step.position_x ?? 0,
          y: step.position_y ?? 0,
        },
        data: stepData,
      };

      nodes.push(node);
    } catch (error) {
      console.error(`Failed to parse config_json for step ${step.id}:`, error);
      // Create a basic node even if parsing fails
      nodes.push({
        id: step.id,
        type: step.type as any,
        position: {
          x: step.position_x ?? 0,
          y: step.position_y ?? 0,
        },
        data: {
          label: step.name,
          description: step.description || "",
        },
      });
    }
  }

  // Then, create edges from depends_on relationships
  for (const step of steps) {
    if (step.depends_on && step.depends_on.length > 0) {
      for (const depId of step.depends_on) {
        // Check if dependency is the trigger or another step
        const sourceId =
          depId === triggerNodeId || depId.startsWith("trigger-")
            ? triggerNodeId
            : depId;

        // Avoid duplicate edges
        const edgeId = `${sourceId}-${step.id}`;
        if (!edges.find((e) => e.id === edgeId)) {
          edges.push({
            id: edgeId,
            source: sourceId,
            target: step.id,
          });
        }
      }
    } else {
      // If no dependencies, connect to trigger (only if not already connected)
      const edgeId = `${triggerNodeId}-${step.id}`;
      if (!edges.find((e) => e.id === edgeId)) {
        edges.push({
          id: edgeId,
          source: triggerNodeId,
          target: step.id,
        });
      }
    }
  }

  return {
    nodes,
    edges,
  };
}

/**
 * Convert a Workflow object to an Enhanced Flow representation
 */
export function workflowToFlow(workflow: ApiWorkflow): Workflow {
  // If flow_graph is already present, use it
  if (workflow.flow_graph) {
    // Ensure all nodes have valid positions
    const normalizedNodes: FlowNode[] = workflow.flow_graph.nodes.map((node, index) => {
      // If node doesn't have position or position is invalid, assign a default
      const position = node.position;
      if (
        !position ||
        typeof position !== 'object' ||
        typeof position.x !== 'number' ||
        typeof position.y !== 'number' ||
        isNaN(position.x) ||
        isNaN(position.y)
      ) {
        return {
          ...node,
          position: {
            x: 100 + (index % 3) * 350,
            y: 100 + Math.floor(index / 3) * 150,
          },
        } as FlowNode;
      }
      return node as FlowNode;
    });
    
    return {
      nodes: normalizedNodes,
      edges: workflow.flow_graph.edges,
    };
  }

  // Otherwise, convert from trigger + steps format
  if (workflow.trigger && workflow.steps) {
    return normalizeToFlowGraph(workflow, workflow.trigger, workflow.steps);
  }

  throw new Error(
    "Workflow must have either flow_graph or trigger+steps",
  );
}

/**
 * Convert an Enhanced Flow representation back to a Workflow object
 * Uses the new flow_graph format for full feature support
 */
export function flowToWorkflow(
  flow: Workflow,
  existingWorkflow?: Partial<ApiWorkflow>,
): Partial<ApiWorkflow> {
  const triggerNode = flow.nodes.find((n) => n.type === "trigger") as
    | Node<TriggerNodeData, "trigger">
    | undefined;

  if (!triggerNode) {
    throw new Error("Flow must contain a trigger node");
  }

  const workflow: Partial<ApiWorkflow> = {
    ...existingWorkflow,
    type: triggerNode.data.triggerType,
  };

  // Add schedule if it's a schedule trigger
  if (
    triggerNode.data.triggerType === "schedule" &&
    triggerNode.data.cronExpression
  ) {
    workflow.schedule = {
      cron_expression: triggerNode.data.cronExpression,
      timezone: triggerNode.data.timezone || "UTC",
    };
  }

  // Always use flow_graph format
  workflow.flow_graph = {
    nodes: flow.nodes.map((node) => ({
      id: node.id,
      type: node.type,
      position: node.position,
      data: node.data,
    })),
    edges: flow.edges.map((edge) => ({
      id: edge.id,
      source: edge.source,
      target: edge.target,
      label: (edge as any).label,
    })),
  };

  return workflow;
}

/**
 * Validate that a flow is complete and can be converted to a workflow
 */
export function validateFlow(flow: Workflow): {
  valid: boolean;
  errors: string[];
} {
  const errors: string[] = [];

  // Check for trigger node
  const triggerNode = flow.nodes.find((n) => n.type === "trigger") as
    | Node<TriggerNodeData, "trigger">
    | undefined;
  if (!triggerNode) {
    errors.push("Flow must contain a trigger node");
  }

  // Check for at least one action
  const actionNodes = flow.nodes.filter((n) => n.type === "action") as Node<
    ActionNodeData,
    "action"
  >[];
  if (actionNodes.length === 0) {
    errors.push("Flow must contain at least one action node");
  }

  // Validate action nodes have required data
  actionNodes.forEach((node, index) => {
    if (!node.data.action.target_id) {
      errors.push(`Action ${index + 1} is missing a target`);
    }
    if (
      node.data.action.config_update?.operation === "replace" &&
      !node.data.action.config_update?.template
    ) {
      errors.push(`Action ${index + 1} is missing a config template`);
    }
  });

  // Validate delay nodes have positive duration
  const delayNodes = flow.nodes.filter((n) => n.type === "delay") as Node<
    DelayNodeData,
    "delay"
  >[];
  delayNodes.forEach((node, index) => {
    if (!node.data.duration || node.data.duration <= 0) {
      errors.push(`Delay ${index + 1} must have a positive duration`);
    }
  });

  // Validate schedule triggers have cron expression
  if (
    triggerNode &&
    triggerNode.data.triggerType === "schedule" &&
    !triggerNode.data.cronExpression
  ) {
    errors.push("Schedule trigger must have a cron expression");
  }

  return {
    valid: errors.length === 0,
    errors,
  };
}
