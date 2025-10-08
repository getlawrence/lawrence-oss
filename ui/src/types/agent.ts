export interface Agent {
  id: string;
  name: string;
  status: "online" | "offline" | "error";
  last_seen: string;
  version: string;
  group_id?: string;
  labels: Record<string, string>;
  capabilities?: string[];
  effective_config?: string;
}

export interface AgentStats {
  totalAgents: number;
  onlineAgents: number;
  offlineAgents: number;
  errorAgents: number;
  groupsCount: number;
}
