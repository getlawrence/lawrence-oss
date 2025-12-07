import { apiGet, apiPost, apiPut, apiDelete } from "./base";

export interface WorkflowTrigger {
  workflow_id: string;
  type: "webhook" | "schedule" | "manual" | "telemetry";
  webhook_url?: string;
  webhook_secret?: string;
  schedule?: {
    cron_expression: string;
    timezone: string;
  };
  telemetry_config?: TelemetryTriggerConfig;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface TelemetryTriggerConfig {
  type: "log" | "metric";
  // Log config
  severity?: "error" | "warn" | "info";
  pattern?: string;
  agent_id?: string;
  service_name?: string;
  // Metric config
  metric_name?: string;
  operator?: ">" | "<" | ">=" | "<=";
  threshold?: number;
  time_window?: string; // e.g., "5m", "1h"
}

export interface WorkflowStep {
  id: string;
  workflow_id: string;
  type: string;
  name: string;
  description?: string;
  order: number;
  position_x?: number;
  position_y?: number;
  config_json: string; // JSON string containing node data
  retry_enabled: boolean;
  retry_count: number;
  retry_delay_ms: number;
  continue_on_error: boolean;
  depends_on?: string[];
  created_at: string;
  updated_at: string;
}

export interface Workflow {
  id: string;
  name: string;
  description: string;
  type: "webhook" | "schedule" | "manual" | "telemetry";
  status: "active" | "inactive" | "error";
  schedule?: {
    cron_expression: string;
    timezone: string;
  };
  webhook_url?: string;
  webhook_secret?: string;
  telemetry_config?: TelemetryTriggerConfig;

  // Normalized structure (from API)
  trigger?: WorkflowTrigger;
  steps?: WorkflowStep[];

  // Flow graph representation (reconstructed from normalized structure)
  flow_graph?: WorkflowFlowGraph;

  created_by?: string;
  last_run?: string;
  next_run?: string;
  run_count: number;
  error_count: number;
  last_error?: string;
  created_at: string;
  updated_at: string;
}

export interface WorkflowFlowGraph {
  nodes: FlowNode[];
  edges: FlowEdge[];
}

export interface FlowNode {
  id: string;
  type: string;
  position?: { x: number; y: number };
  data: any;
}

export interface FlowEdge {
  id: string;
  source: string;
  target: string;
  label?: string;
}

export interface WorkflowAction {
  type: "config_update" | "delayed_action" | "tail_sampling";
  target_type: "agent" | "group";
  target_id: string;
  config_update?: {
    operation: "merge" | "replace" | "patch";
    yaml_path?: string;
    value?: unknown;
    template?: string;
  };
  delayed_action?: {
    delay: string;
    action: WorkflowAction;
  };
  tail_sampling?: {
    service_name: string;
    sampling_percentage: number;
    policy_name?: string;
    revert_after?: string;
    revert_to?: number;
  };
}

export interface WorkflowCondition {
  field: string;
  operator: "equals" | "contains" | "matches";
  value: string;
}

export interface WorkflowExecution {
  id: string;
  workflow_id: string;
  workflow_name: string;
  status: "running" | "success" | "failed" | "partial";
  started_at: string;
  completed_at?: string;
  duration_ms?: number;
  actions_executed: number;
  actions_succeeded: number;
  actions_failed: number;
  configs_created: string[];
  error?: string;
  metadata?: Record<string, string>;
  created_at: string;
}

export interface ListWorkflowsResponse {
  workflows: Workflow[];
  count: number;
}

export interface ListWorkflowExecutionsResponse {
  executions: WorkflowExecution[];
  count: number;
}

export interface StepExecution {
  id: string;
  step_id: string;
  workflow_id: string;
  workflow_execution_id: string;
  step_name: string;
  step_type: string;
  status:
    | "pending"
    | "running"
    | "success"
    | "failed"
    | "skipped"
    | "retrying"
    | "cancelled";
  started_at: string;
  completed_at?: string;
  duration_ms?: number;
  retry_attempt: number;
  error?: string;
  error_code?: string;
  input_data?: Record<string, unknown>;
  output_data?: Record<string, unknown>;
  metadata?: Record<string, string>;
  created_at: string;
}

export interface StepExecutionLog {
  id: string;
  step_execution_id: string;
  level: "debug" | "info" | "warning" | "error";
  message: string;
  data?: string;
  timestamp: string;
}

export interface ListStepExecutionsResponse {
  step_executions: StepExecution[];
  count: number;
}

export interface ListStepExecutionLogsResponse {
  logs: StepExecutionLog[];
  count: number;
}

export const getWorkflows = (params?: {
  type?: string;
  status?: string;
}): Promise<ListWorkflowsResponse> => {
  return apiGet("/workflows", params);
};

export const getWorkflow = (id: string): Promise<Workflow> => {
  return apiGet(`/workflows/${id}`);
};

export const createWorkflow = (data: Partial<Workflow>): Promise<Workflow> => {
  return apiPost("/workflows", data);
};

export const updateWorkflow = (
  id: string,
  data: Partial<Workflow>,
): Promise<Workflow> => {
  return apiPut(`/workflows/${id}`, data);
};

export const deleteWorkflow = (id: string): Promise<void> => {
  return apiDelete(`/workflows/${id}`);
};

export const executeWorkflow = (id: string): Promise<WorkflowExecution> => {
  return apiPost(`/workflows/${id}/execute`, {});
};

export const pauseWorkflow = (id: string): Promise<Workflow> => {
  return updateWorkflow(id, { status: "inactive" });
};

export const resumeWorkflow = (id: string): Promise<Workflow> => {
  return updateWorkflow(id, { status: "active" });
};

export const getWorkflowExecutions = (
  id: string,
): Promise<ListWorkflowExecutionsResponse> => {
  return apiGet(`/workflows/${id}/executions`);
};

export const getStepExecutions = (
  workflowId: string,
  executionId: string,
): Promise<ListStepExecutionsResponse> => {
  return apiGet(`/workflows/${workflowId}/executions/${executionId}/steps`);
};

export const getStepExecution = (
  executionId: string,
): Promise<StepExecution> => {
  return apiGet(`/step-executions/${executionId}`);
};

export const getStepExecutionLogs = (
  executionId: string,
  limit?: number,
): Promise<ListStepExecutionLogsResponse> => {
  const params = limit ? { limit: limit.toString() } : undefined;
  return apiGet(`/step-executions/${executionId}/logs`, params);
};

export const retryStepExecution = (
  executionId: string,
): Promise<{ message: string; step_execution_id: string }> => {
  return apiPost(`/step-executions/${executionId}/retry`, {});
};
