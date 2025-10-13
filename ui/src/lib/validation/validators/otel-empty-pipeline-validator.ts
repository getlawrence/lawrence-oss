// Validator for empty or incomplete pipelines

import type { OTelConfig, ValidationError, Validator } from "../types";
import { createValidationError } from "../yaml-position";

/**
 * Validates that pipelines have required components
 */
export class OTelEmptyPipelineValidator implements Validator {
  name = "otel-empty-pipeline";

  validate(yamlContent: string, parsedData: unknown): ValidationError[] {
    const errors: ValidationError[] = [];

    if (!parsedData || typeof parsedData !== "object") {
      return errors;
    }

    const config = parsedData as OTelConfig;

    // Skip if no service.pipelines section
    if (!config.service?.pipelines) {
      return errors;
    }

    // Validate each pipeline
    for (const [pipelineName, pipeline] of Object.entries(
      config.service.pipelines,
    )) {
      // Check if pipeline has at least one receiver
      if (!pipeline.receivers || pipeline.receivers.length === 0) {
        errors.push(
          createValidationError(
            `Pipeline '${pipelineName}' must have at least one receiver`,
            yamlContent,
            ["service", "pipelines", pipelineName],
            undefined,
            "warning",
          ),
        );
      }

      // Check if pipeline has at least one exporter
      if (!pipeline.exporters || pipeline.exporters.length === 0) {
        errors.push(
          createValidationError(
            `Pipeline '${pipelineName}' must have at least one exporter`,
            yamlContent,
            ["service", "pipelines", pipelineName],
            undefined,
            "warning",
          ),
        );
      }
    }

    return errors;
  }
}
