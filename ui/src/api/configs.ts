import { apiGet, apiPost } from './base';

export interface Config {
  id: string;
  agent_id?: string;
  group_id?: string;
  config_hash: string;
  content: string;
  version: number;
  created_at: string;
}

export interface CreateConfigRequest {
  agent_id?: string;
  group_id?: string;
  config_hash: string;
  content: string;
  version: number;
}

export interface GetConfigsResponse {
  configs: Config[];
  count: number;
}

// Get all configs with optional filtering
export const getConfigs = (params?: {
  agent_id?: string;
  group_id?: string;
  limit?: number;
}): Promise<GetConfigsResponse> => {
  const queryParams: Record<string, string> = {};
  if (params?.agent_id) queryParams.agent_id = params.agent_id;
  if (params?.group_id) queryParams.group_id = params.group_id;
  if (params?.limit) queryParams.limit = params.limit.toString();

  return apiGet<GetConfigsResponse>('/configs', queryParams);
};

// Get config by ID
export const getConfig = (id: string): Promise<Config> => {
  return apiGet<Config>(`/configs/${id}`);
};

// Create new config
export const createConfig = (data: CreateConfigRequest): Promise<Config> => {
  return apiPost<Config>('/configs', data);
};
