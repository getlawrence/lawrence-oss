import { apiGet, apiPost, apiPut } from "./base";

export interface Config {
  id: string;
  name: string;
  agent_id?: string;
  group_id?: string;
  config_hash: string;
  content: string;
  version: number;
  created_at: string;
}

export interface CreateConfigRequest {
  name?: string;
  agent_id?: string;
  group_id?: string;
  config_hash: string;
  content: string;
  version: number;
}

export interface UpdateConfigRequest {
  name?: string;
  content: string;
  version: number;
}

export interface ValidateConfigRequest {
  content: string;
}

export interface ValidateConfigResponse {
  valid: boolean;
  errors?: string[];
  warnings?: string[];
}

export interface PaginatedResponse<T> {
  data: T[];
  page: number;
  page_size: number;
  total_count: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
}

export type GetConfigsResponse = PaginatedResponse<Config>;

export interface GetVersionsResponse {
  versions: Config[];
  count: number;
}

// Get all configs with optional filtering and pagination
export const getConfigs = (params?: {
  agent_id?: string;
  group_id?: string;
  page?: number;
  page_size?: number;
}): Promise<GetConfigsResponse> => {
  const queryParams: Record<string, string> = {};
  if (params?.agent_id) queryParams.agent_id = params.agent_id;
  if (params?.group_id) queryParams.group_id = params.group_id;
  if (params?.page) queryParams.page = params.page.toString();
  if (params?.page_size) queryParams.page_size = params.page_size.toString();

  return apiGet<GetConfigsResponse>("/configs", queryParams);
};

// Get config by ID
export const getConfig = (id: string): Promise<Config> => {
  return apiGet<Config>(`/configs/${id}`);
};

// Create new config
export const createConfig = (data: CreateConfigRequest): Promise<Config> => {
  return apiPost<Config>("/configs", data);
};

// Update config (creates new version)
export const updateConfig = (
  id: string,
  data: UpdateConfigRequest,
): Promise<Config> => {
  return apiPut<Config>(`/configs/${id}`, data);
};

// Validate config
export const validateConfig = (
  data: ValidateConfigRequest,
): Promise<ValidateConfigResponse> => {
  return apiPost<ValidateConfigResponse>("/configs/validate", data);
};

// Get config versions
export const getConfigVersions = (params: {
  agent_id?: string;
  group_id?: string;
}): Promise<GetVersionsResponse> => {
  const queryParams: Record<string, string> = {};
  if (params.agent_id) queryParams.agent_id = params.agent_id;
  if (params.group_id) queryParams.group_id = params.group_id;

  return apiGet<GetVersionsResponse>("/configs/versions", queryParams);
};
