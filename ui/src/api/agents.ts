import { apiGet, apiPatch, apiPost } from "./base";

export interface Agent {
  id: string;
  name: string;
  status: "online" | "offline" | "error";
  last_seen: string;
  version: string;
  group_id?: string;
  group_name?: string;
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

export interface GetAgentsResponse {
  agents: Record<string, Agent>;
  totalCount: number;
  activeCount: number;
  inactiveCount: number;
}

// Get all agents
export const getAgents = (): Promise<GetAgentsResponse> => {
  return apiGet<GetAgentsResponse>("/agents");
};

// Get agent by ID
export const getAgent = (id: string): Promise<Agent> => {
  return apiGet<Agent>(`/agents/${id}`);
};

// Get agent statistics
export const getAgentStats = (): Promise<AgentStats> => {
  return apiGet<AgentStats>("/agents/stats");
};

// Update agent group
export const updateAgentGroup = (
  id: string,
  groupId: string,
): Promise<void> => {
  return apiPatch<void>(`/agents/${id}/group`, { group_id: groupId });
};

// Send configuration to agent request/response types
export interface SendConfigToAgentRequest {
  content: string;
}

export interface SendConfigToAgentResponse {
  success: boolean;
  message: string;
  config_id?: string;
}

// Send configuration to agent
export const sendConfigToAgent = (
  agentId: string,
  content: string,
): Promise<SendConfigToAgentResponse> => {
  return apiPost<SendConfigToAgentResponse>(`/agents/${agentId}/config`, {
    content,
  });
};

// Restart agent request/response types
export interface RestartAgentResponse {
  success: boolean;
  message: string;
}

// Restart agent
export const restartAgent = (
  agentId: string,
): Promise<RestartAgentResponse> => {
  return apiPost<RestartAgentResponse>(`/agents/${agentId}/restart`, {});
};
