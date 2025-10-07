import { apiGet, apiPost } from "./base";

export interface TelemetryOverview {
  totalMetrics: number;
  totalLogs: number;
  totalTraces: number;
  activeAgents: number;
  services: string[];
  lastUpdated: string;
}

export interface ServicesResponse {
  services: string[];
  count: number;
}

export interface MetricQueryRequest {
  agent_id?: string;
  group_id?: string;
  metric_name?: string;
  start_time: string;
  end_time: string;
  limit?: number;
  use_rollups?: boolean;
}

export interface LogQueryRequest {
  agent_id?: string;
  group_id?: string;
  severity?: string;
  search?: string;
  start_time: string;
  end_time: string;
  limit?: number;
}

export interface TraceQueryRequest {
  agent_id?: string;
  group_id?: string;
  trace_id?: string;
  service_name?: string;
  start_time: string;
  end_time: string;
  limit?: number;
}

export interface MetricData {
  timestamp: string;
  agent_id: string;
  group_id?: string;
  service_name: string;
  metric_name: string;
  value: number;
  metric_attributes: Record<string, any>;
}

export interface LogData {
  timestamp: string;
  agent_id: string;
  group_id?: string;
  service_name: string;
  severity_text: string;
  severity_number: number;
  body: string;
  trace_id?: string;
  span_id?: string;
  log_attributes: Record<string, any>;
}

export interface TraceData {
  timestamp: string;
  agent_id: string;
  group_id?: string;
  trace_id: string;
  span_id: string;
  parent_span_id?: string;
  service_name: string;
  span_name: string;
  duration: number;
  status_code: string;
  span_attributes: Record<string, any>;
}

export interface MetricsQueryResponse {
  metrics: MetricData[];
  count: number;
}

export interface LogsQueryResponse {
  logs: LogData[];
  count: number;
}

export interface TracesQueryResponse {
  traces: TraceData[];
  count: number;
}

// Get telemetry overview
export const getTelemetryOverview = (): Promise<TelemetryOverview> => {
  return apiGet<TelemetryOverview>("/telemetry/overview");
};

// Get services list
export const getServices = (): Promise<ServicesResponse> => {
  return apiGet<ServicesResponse>("/telemetry/services");
};

// Query metrics
export const queryMetrics = (
  request: MetricQueryRequest,
): Promise<MetricsQueryResponse> => {
  return apiPost<MetricsQueryResponse>("/telemetry/metrics/query", request);
};

// Query logs
export const queryLogs = (
  request: LogQueryRequest,
): Promise<LogsQueryResponse> => {
  return apiPost<LogsQueryResponse>("/telemetry/logs/query", request);
};

// Query traces
export const queryTraces = (
  request: TraceQueryRequest,
): Promise<TracesQueryResponse> => {
  return apiPost<TracesQueryResponse>("/telemetry/traces/query", request);
};
