import { apiGet, apiPost } from "./base";

export type ComponentType =
  | "receiver"
  | "processor"
  | "exporter"
  | "connector"
  | "extension";

export interface ComponentInfo {
  type: ComponentType;
  name: string;
  description?: string;
}

export interface GetComponentSchemasResponse {
  components: ComponentInfo[];
}

export interface ValidateComponentConfigRequest {
  type: ComponentType;
  name: string;
  config: unknown;
}

export interface ValidateComponentConfigResponse {
  valid: boolean;
  errors?: string[];
  warnings?: string[];
}

// JSON Schema type (simplified)
export interface JSONSchema {
  $schema?: string;
  type?: string | string[];
  properties?: Record<string, JSONSchema>;
  required?: string[];
  enum?: unknown[];
  default?: unknown;
  description?: string;
  items?: JSONSchema;
  oneOf?: JSONSchema[];
  anyOf?: JSONSchema[];
  allOf?: JSONSchema[];
  [key: string]: unknown;
}

// Get list of available components, optionally filtered by type
export const getComponentSchemas = (
  type?: ComponentType,
): Promise<GetComponentSchemasResponse> => {
  const params: Record<string, string> = {};
  if (type) {
    params.type = type;
  }
  return apiGet<GetComponentSchemasResponse>("/schemas/components", params);
};

// Get JSON schema for a specific component
export const getComponentSchema = (
  type: ComponentType,
  name: string,
): Promise<JSONSchema> => {
  return apiGet<JSONSchema>(`/schemas/components/${type}/${name}`);
};

// Validate a component configuration against its schema
export const validateComponentConfig = (
  data: ValidateComponentConfigRequest,
): Promise<ValidateComponentConfigResponse> => {
  return apiPost<ValidateComponentConfigResponse>("/schemas/validate", data);
};
