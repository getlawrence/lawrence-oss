// Validator for OpenTelemetry collector extensions configuration

import type { OTelConfig, ValidationError, Validator } from "../types";
import { createValidationError } from "../yaml-position";

/**
 * Validates that all extensions referenced in service are defined
 */
export class OTelExtensionsValidator implements Validator {
  name = "otel-extensions";

  validate(yamlContent: string, parsedData: unknown): ValidationError[] {
    const errors: ValidationError[] = [];

    if (!parsedData || typeof parsedData !== "object") {
      return errors;
    }

    const config = parsedData as OTelConfig;

    // Skip if no service.extensions section
    if (!config.service?.extensions) {
      return errors;
    }

    // Get all defined extensions
    const definedExtensions = new Set(Object.keys(config.extensions || {}));

    // Validate each extension reference
    for (const extension of config.service.extensions) {
      if (!definedExtensions.has(extension)) {
        errors.push(
          createValidationError(
            `Extension '${extension}' is used in service but not defined in extensions section`,
            yamlContent,
            ["service", "extensions"],
            extension,
            "error",
          ),
        );
      }
    }

    return errors;
  }
}
