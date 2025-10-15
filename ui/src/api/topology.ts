import { apiGet } from "./base";

export interface TopologyNode {
  id: string;
  type: "agent" | "group" | "service";
  name: string;
  status: string;
  group_id?: string;
  group_name?: string;
  labels: Record<string, string>;
  metrics?: NodeMetrics;
  last_seen?: string;
  metadata?: Record<string, unknown>;
}

export interface TopologyEdge {
  source: string;
  target: string;
  type: string;
  label?: string;
}

export interface NodeMetrics {
  metric_count: number;
  log_count: number;
  trace_count: number;
  error_rate: number;
  latency: number;
  throughput_rps: number;
}

export interface GroupSummary {
  id: string;
  name: string;
  agent_count: number;
  status: string;
}

export interface TopologyResponse {
  nodes: TopologyNode[];
  edges: TopologyEdge[];
  groups: GroupSummary[];
  services: string[];
  updated_at: string;
}

export interface AgentTopologyResponse {
  agent: unknown;
  metrics: NodeMetrics;
  pipeline: unknown;
}

export interface GroupTopologyResponse {
  group: unknown;
  agents: unknown[];
  agent_count: number;
  metric_count: number;
}

// Get full topology
export const getTopology = (): Promise<TopologyResponse> => {
  return apiGet<TopologyResponse>("/topology");
};

// Get agent topology
export const getAgentTopology = (
  agentId: string,
): Promise<AgentTopologyResponse> => {
  return apiGet<AgentTopologyResponse>(`/topology/agent/${agentId}`);
};

// Get group topology
export const getGroupTopology = (
  groupId: string,
): Promise<GroupTopologyResponse> => {
  return apiGet<GroupTopologyResponse>(`/topology/group/${groupId}`);
};
