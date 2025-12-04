// Validator for OpenTelemetry collector component schemas
// Note: Full schema validation against OpenTelemetry component schemas requires async API calls.
// This validator performs synchronous structural validation (configs must be objects, etc.).
// Full schema validation should be done via async API calls in the validation hook.

import type { OTelConfig, ValidationError, Validator } from "../types";
import { createValidationError } from "../yaml-position";

/**
 * Validates OpenTelemetry collector component configurations for structural issues.
 * Performs synchronous validation of component structure and format.
 * Full schema validation against OpenTelemetry schemas should be done via async API calls.
 */
export class OTelSchemaValidator implements Validator {
  name = "otel-schema";

  validate(yamlContent: string, parsedData: unknown): ValidationError[] {
    const errors: ValidationError[] = [];

    if (!parsedData || typeof parsedData !== "object") {
      return errors;
    }

    const config = parsedData as OTelConfig;

    // Validate all component types
    this.validateComponents(config.receivers, "receivers", yamlContent, errors);
    this.validateComponents(
      config.processors,
      "processors",
      yamlContent,
      errors,
    );
    this.validateComponents(config.exporters, "exporters", yamlContent, errors);
    this.validateComponents(
      config.connectors,
      "connectors",
      yamlContent,
      errors,
    );
    this.validateComponents(
      config.extensions,
      "extensions",
      yamlContent,
      errors,
    );

    return errors;
  }

  /**
   * Validate components of a specific type
   */
  private validateComponents(
    components: Record<string, unknown> | undefined,
    componentType: string,
    yamlContent: string,
    errors: ValidationError[],
  ): void {
    if (!components) {
      return;
    }

    for (const [componentName, componentConfig] of Object.entries(components)) {
      const path = [componentType, componentName];

      // Validate component name format (alphanumeric with underscores and hyphens)
      if (!this.isValidComponentName(componentName)) {
        errors.push(
          createValidationError(
            `Invalid ${componentType} name '${componentName}': component names must be alphanumeric and may contain underscores or hyphens`,
            yamlContent,
            path,
            undefined,
            "warning",
          ),
        );
      }

      // Validate component config structure
      if (componentConfig === null) {
        errors.push(
          createValidationError(
            `${componentType} '${componentName}' has null configuration (should be an object or omitted)`,
            yamlContent,
            path,
            undefined,
            "error",
          ),
        );
      } else if (Array.isArray(componentConfig)) {
        errors.push(
          createValidationError(
            `${componentType} '${componentName}' has array configuration (should be an object)`,
            yamlContent,
            path,
            undefined,
            "error",
          ),
        );
      } else if (
        componentConfig !== undefined &&
        typeof componentConfig !== "object"
      ) {
        errors.push(
          createValidationError(
            `${componentType} '${componentName}' has invalid configuration type '${typeof componentConfig}' (should be an object)`,
            yamlContent,
            path,
            undefined,
            "error",
          ),
        );
      } else if (
        componentConfig !== undefined &&
        Object.keys(componentConfig).length === 0
      ) {
        // Empty config is valid but we can warn about it
        errors.push(
          createValidationError(
            `${componentType} '${componentName}' has empty configuration`,
            yamlContent,
            path,
            undefined,
            "warning",
          ),
        );
      }
    }
  }

  /**
   * Validate component name format
   * Component names should be alphanumeric and may contain underscores or hyphens
   */
  private isValidComponentName(name: string): boolean {
    // Allow alphanumeric characters, underscores, and hyphens
    // Must start with a letter or underscore
    return /^[a-zA-Z_][a-zA-Z0-9_-]*$/.test(name);
  }
}
