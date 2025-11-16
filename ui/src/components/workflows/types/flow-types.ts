import type { Node, Edge } from "@xyflow/react";

import { type WorkflowAction, type WorkflowCondition } from "@/api/workflows";

// Extended node types for complex workflows
export type FlowNodeType =
  | "trigger"
  | "condition"
  | "action"
  | "parallel" // Execute multiple actions in parallel
  | "sequential" // Execute actions in sequence with delays
  | "loop" // Loop through agents/groups
  | "delay" // Add time delays between actions
  | "notification" // Send notifications
  | "group" // Group nodes together visually
  | "variable" // Set/get variables
  | "error-handler" // Handle errors in flow
  | "branch"; // Multi-way branching based on conditions

// Original node data types
export interface TriggerNodeData extends Record<string, unknown> {
  label: string;
  triggerType: "manual" | "schedule" | "webhook";
  cronExpression?: string;
  timezone?: string;
  description?: string;
}

export interface ConditionNodeData extends Record<string, unknown> {
  label: string;
  conditions: WorkflowCondition[];
  logic?: "and" | "or"; // How to combine multiple conditions
  description?: string;
}

export interface ActionNodeData extends Record<string, unknown> {
  label: string;
  action: WorkflowAction;
  targetName?: string;
  description?: string;
  continueOnError?: boolean; // Don't stop flow if this action fails
}

// New enhanced node data types
export interface ParallelNodeData extends Record<string, unknown> {
  label: string;
  description?: string;
  waitForAll?: boolean; // Wait for all parallel paths to complete
  timeout?: number; // Timeout in seconds
}

export interface SequentialNodeData extends Record<string, unknown> {
  label: string;
  description?: string;
  delayBetween?: number; // Delay in seconds between each action
}

export interface LoopNodeData extends Record<string, unknown> {
  label: string;
  description?: string;
  loopType: "agents" | "groups" | "range";
  filter?: string; // Filter expression for agents/groups
  maxIterations?: number;
  parallelExecution?: boolean;
}

export interface DelayNodeData extends Record<string, unknown> {
  label: string;
  description?: string;
  duration: number; // Duration in seconds
  unit: "seconds" | "minutes" | "hours";
}

export interface NotificationNodeData extends Record<string, unknown> {
  label: string;
  description?: string;
  channel: "email" | "slack" | "webhook" | "log";
  recipients?: string[];
  message: string;
  severity?: "info" | "warning" | "error" | "success";
}

export interface GroupNodeData extends Record<string, unknown> {
  label: string;
  description?: string;
  collapsed?: boolean;
  color?: string;
}

export interface VariableNodeData extends Record<string, unknown> {
  label: string;
  description?: string;
  operation: "set" | "get" | "increment" | "append";
  variableName: string;
  value?: string;
}

export interface ErrorHandlerNodeData extends Record<string, unknown> {
  label: string;
  description?: string;
  catchAll?: boolean; // Catch all errors or specific ones
  errorTypes?: string[];
  retryCount?: number;
  retryDelay?: number; // Delay in seconds between retries
}

export interface BranchNodeData extends Record<string, unknown> {
  label: string;
  description?: string;
  branches: Array<{
    name: string;
    condition: WorkflowCondition[];
    isDefault?: boolean; // Default branch if no conditions match
  }>;
}

// Union type for all node data
export type FlowNode =
  | Node<TriggerNodeData, "trigger">
  | Node<ConditionNodeData, "condition">
  | Node<ActionNodeData, "action">
  | Node<ParallelNodeData, "parallel">
  | Node<SequentialNodeData, "sequential">
  | Node<LoopNodeData, "loop">
  | Node<DelayNodeData, "delay">
  | Node<NotificationNodeData, "notification">
  | Node<GroupNodeData, "group">
  | Node<VariableNodeData, "variable">
  | Node<ErrorHandlerNodeData, "error-handler">
  | Node<BranchNodeData, "branch">;



export interface Workflow {
  nodes: FlowNode[];
  edges: Edge[];
  variables?: Record<string, unknown>; // Flow-level variables
  metadata?: {
    version?: string;
    author?: string;
    tags?: string[];
    lastModified?: string;
  };
}

// Node categories for the palette
export interface NodeCategory {
  id: string;
  name: string;
  description: string;
  icon: string;
  nodes: NodeTemplate[];
}

export interface NodeTemplate {
  id: string;
  type: FlowNodeType;
  name: string;
  description: string;
  icon: string;
  defaultData: Record<string, unknown>;
  category: string;
  color: string;
  ports?: {
    inputs: number;
    outputs: number;
  };
}

// Workflow templates
export interface WorkflowTemplate {
  id: string;
  name: string;
  description: string;
  category: string;
  tags: string[];
  difficulty: "beginner" | "intermediate" | "advanced";
  flow: Workflow;
  thumbnail?: string;
}

// Flow validation result
export interface FlowValidation {
  valid: boolean;
  errors: Array<{
    nodeId?: string;
    message: string;
    severity: "error" | "warning" | "info";
  }>;
  warnings: Array<{
    nodeId?: string;
    message: string;
  }>;
}

// Flow execution context (for preview/simulation)
export interface FlowExecutionContext {
  currentNode?: string;
  executedNodes: string[];
  variables: Record<string, unknown>;
  logs: Array<{
    timestamp: string;
    nodeId: string;
    message: string;
    level: "info" | "warning" | "error";
  }>;
  status: "idle" | "running" | "paused" | "completed" | "error";
}