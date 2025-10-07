export interface TopologyNode {
  id: string;
  type: 'group' | 'agent';
  name: string;
  data: any;
}

export interface AgentNodeData {
  name: string;
  status: string;
  group_name?: string;
  metrics?: {
    metric_count: number;
    log_count: number;
    throughput_rps: number;
  };
  labels?: Record<string, string>;
}

export interface GroupNodeData {
  name: string;
  status: string;
  agent_count: number;
}

export interface TopologyData {
  nodes: Array<{
    id: string;
    type: 'agent' | 'group' | 'service';
    name: string;
    status: string;
    group_name?: string;
    metrics?: AgentNodeData['metrics'];
    labels?: Record<string, string>;
  }>;
  edges: Array<{
    source: string;
    target: string;
    label?: string;
  }>;
  groups?: Array<{
    id: string;
    name: string;
    agent_count: number;
    labels?: Record<string, string>;
  }>;
}
