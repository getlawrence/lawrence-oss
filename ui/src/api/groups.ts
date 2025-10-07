import type { Agent } from "../types/agent";

import { apiGet, apiPost, apiDelete } from "./base";
import type { Config } from "./configs";

export interface Group {
  id: string;
  name: string;
  labels: Record<string, string>;
  created_at: string;
  updated_at: string;
}

export interface CreateGroupRequest {
  name: string;
  labels?: Record<string, string>;
}

export interface AssignConfigRequest {
  config_id: string;
}

export interface AssignConfigResponse {
  message: string;
  config: Config;
}

export interface GetGroupsResponse {
  groups: Group[];
  count: number;
}

export interface GetGroupAgentsResponse {
  agents: Agent[];
  count: number;
}

// Get all groups
export const getGroups = (): Promise<GetGroupsResponse> => {
  return apiGet<GetGroupsResponse>("/groups");
};

// Get group by ID
export const getGroup = (id: string): Promise<Group> => {
  return apiGet<Group>(`/groups/${id}`);
};

// Create new group
export const createGroup = (data: CreateGroupRequest): Promise<Group> => {
  return apiPost<Group>("/groups", data);
};

// Delete group
export const deleteGroup = (id: string): Promise<void> => {
  return apiDelete<void>(`/groups/${id}`);
};

// Assign config to group
export const assignConfigToGroup = (
  groupId: string,
  data: AssignConfigRequest,
): Promise<AssignConfigResponse> => {
  return apiPost<AssignConfigResponse>(`/groups/${groupId}/config`, data);
};

// Get group's active config
export const getGroupConfig = (groupId: string): Promise<Config> => {
  return apiGet<Config>(`/groups/${groupId}/config`);
};

// Get agents in group
export const getGroupAgents = (
  groupId: string,
): Promise<GetGroupAgentsResponse> => {
  return apiGet<GetGroupAgentsResponse>(`/groups/${groupId}/agents`);
};
