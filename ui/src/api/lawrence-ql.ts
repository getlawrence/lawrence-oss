import { apiPost, apiGet } from "./base";

export interface LawrenceQLRequest {
  query: string;
  start_time?: string;
  end_time?: string;
  limit?: number;
  agent_id?: string;
  group_id?: string;
}

export interface QueryResult {
  type: "metrics" | "logs" | "traces";
  timestamp: string;
  labels: Record<string, string>;
  value: any;
  data?: Record<string, any>;
}

export interface QueryMeta {
  execution_time: number;
  row_count: number;
  query_type: string;
  used_rollups: boolean;
}

export interface LawrenceQLResponse {
  results: QueryResult[];
  meta: QueryMeta;
}

export interface ValidateQueryResponse {
  valid: boolean;
  error?: string;
  message?: string;
}

export interface SuggestionsResponse {
  suggestions: string[];
}

export interface QueryTemplate {
  id: string;
  name: string;
  description: string;
  query: string;
  category: string;
}

export interface TemplatesResponse {
  templates: QueryTemplate[];
}

export interface FunctionInfo {
  name: string;
  description: string;
  example: string;
}

export interface FunctionsResponse {
  functions: FunctionInfo[];
}

/**
 * Execute a Lawrence QL query
 */
export async function executeLawrenceQL(
  request: LawrenceQLRequest
): Promise<LawrenceQLResponse> {
  return apiPost<LawrenceQLResponse>("/telemetry/query", request);
}

/**
 * Validate a Lawrence QL query
 */
export async function validateQuery(query: string): Promise<ValidateQueryResponse> {
  return apiPost<ValidateQueryResponse>("/telemetry/query/validate", { query });
}

/**
 * Get query suggestions for auto-completion
 */
export async function getQuerySuggestions(
  query: string,
  cursorPos: number
): Promise<SuggestionsResponse> {
  return apiPost<SuggestionsResponse>("/telemetry/query/suggestions", {
    query,
    cursor_pos: cursorPos,
  });
}

/**
 * Get query templates
 */
export async function getQueryTemplates(): Promise<TemplatesResponse> {
  return apiGet<TemplatesResponse>("/telemetry/query/templates");
}

/**
 * Get available functions
 */
export async function getQueryFunctions(): Promise<FunctionsResponse> {
  return apiGet<FunctionsResponse>("/telemetry/query/functions");
}
