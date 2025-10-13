// Validation types and interfaces for YAML configuration validation

export interface ValidationError {
  message: string;
  severity: "error" | "warning";
  line: number;
  column: number;
  endLine?: number;
  endColumn?: number;
  path?: string[]; // Path to the problematic field in YAML (e.g., ["service", "pipelines", "traces"])
}

export interface ValidationResult {
  valid: boolean;
  errors: ValidationError[];
}

export interface Validator {
  name: string;
  validate: (yamlContent: string, parsedData: unknown) => ValidationError[];
}

export interface OTelConfig {
  receivers?: Record<string, unknown>;
  processors?: Record<string, unknown>;
  exporters?: Record<string, unknown>;
  extensions?: Record<string, unknown>;
  connectors?: Record<string, unknown>;
  service?: {
    pipelines?: Record<
      string,
      {
        receivers?: string[];
        processors?: string[];
        exporters?: string[];
      }
    >;
    extensions?: string[];
  };
}
