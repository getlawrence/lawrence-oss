import { apiGet } from "./base";

export interface ComponentMetrics {
  component_type: string; // receiver, processor, exporter
  component_name: string;
  pipeline_type: string; // traces, metrics, logs
  throughput: number;
  errors: number;
  error_rate: number;
  received?: number;
  accepted?: number;
  refused?: number;
  dropped?: number;
  sent?: number;
  send_failed?: number;
  last_updated: string;
  labels?: Record<string, string>;
}

export interface PipelineMetricsResponse {
  agent_id: string;
  timestamp: string;
  components: ComponentMetrics[];
}

// Get pipeline metrics for a specific agent
export const getPipelineMetrics = (
  agentId: string,
  timeRange: string = "5m",
): Promise<PipelineMetricsResponse> => {
  return apiGet<PipelineMetricsResponse>(
    `/agents/${agentId}/pipeline-metrics?timeRange=${timeRange}`,
  );
};
