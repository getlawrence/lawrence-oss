// Main validation module exports

export * from "./types";
export * from "./yaml-position";
export * from "./validators";

import type { ValidationError, ValidationResult, Validator } from "./types";
import {
  OTelPipelineValidator,
  OTelExtensionsValidator,
  OTelEmptyPipelineValidator,
} from "./validators";

// Default validators for OpenTelemetry collector configs
export const defaultValidators: Validator[] = [
  new OTelPipelineValidator(),
  new OTelExtensionsValidator(),
  new OTelEmptyPipelineValidator(),
];

/**
 * Run all validators on YAML content
 */
export function validateYamlConfig(
  yamlContent: string,
  parsedData: unknown,
  validators: Validator[] = defaultValidators,
): ValidationResult {
  const allErrors: ValidationError[] = [];

  for (const validator of validators) {
    try {
      const errors = validator.validate(yamlContent, parsedData);
      allErrors.push(...errors);
    } catch (error) {
      console.error(`Validator ${validator.name} failed:`, error);
    }
  }

  return {
    valid: allErrors.filter((e) => e.severity === "error").length === 0,
    errors: allErrors,
  };
}
