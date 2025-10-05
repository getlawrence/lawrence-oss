import { apiGet, apiPost, apiDelete } from './base';

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

export interface GetGroupsResponse {
  groups: Group[];
  count: number;
}

// Get all groups
export const getGroups = (): Promise<GetGroupsResponse> => {
  return apiGet<GetGroupsResponse>('/groups');
};

// Get group by ID
export const getGroup = (id: string): Promise<Group> => {
  return apiGet<Group>(`/groups/${id}`);
};

// Create new group
export const createGroup = (data: CreateGroupRequest): Promise<Group> => {
  return apiPost<Group>('/groups', data);
};

// Delete group
export const deleteGroup = (id: string): Promise<void> => {
  return apiDelete<void>(`/groups/${id}`);
};
