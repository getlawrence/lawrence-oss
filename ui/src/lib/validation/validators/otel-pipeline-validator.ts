// Validator for OpenTelemetry collector pipeline configuration

import type { OTelConfig, ValidationError, Validator } from "../types";
import { createValidationError } from "../yaml-position";

/**
 * Validates that all components referenced in pipelines are defined
 */
export class OTelPipelineValidator implements Validator {
  name = "otel-pipeline";

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

    // Get all defined components
    const definedReceivers = new Set(Object.keys(config.receivers || {}));
    const definedProcessors = new Set(Object.keys(config.processors || {}));
    const definedExporters = new Set(Object.keys(config.exporters || {}));
    const definedConnectors = new Set(Object.keys(config.connectors || {}));

    // Connectors can be used as both receivers and exporters
    for (const connector of definedConnectors) {
      definedReceivers.add(connector);
      definedExporters.add(connector);
    }

    // Validate each pipeline
    for (const [pipelineName, pipeline] of Object.entries(
      config.service.pipelines,
    )) {
      // Validate receivers
      if (pipeline.receivers) {
        for (const receiver of pipeline.receivers) {
          if (!definedReceivers.has(receiver)) {
            errors.push(
              createValidationError(
                `Receiver '${receiver}' is used in pipeline '${pipelineName}' but not defined in receivers section`,
                yamlContent,
                ["service", "pipelines", pipelineName, "receivers"],
                receiver,
                "error",
              ),
            );
          }
        }
      }

      // Validate processors
      if (pipeline.processors) {
        for (const processor of pipeline.processors) {
          if (!definedProcessors.has(processor)) {
            errors.push(
              createValidationError(
                `Processor '${processor}' is used in pipeline '${pipelineName}' but not defined in processors section`,
                yamlContent,
                ["service", "pipelines", pipelineName, "processors"],
                processor,
                "error",
              ),
            );
          }
        }
      }

      // Validate exporters
      if (pipeline.exporters) {
        for (const exporter of pipeline.exporters) {
          if (!definedExporters.has(exporter)) {
            errors.push(
              createValidationError(
                `Exporter '${exporter}' is used in pipeline '${pipelineName}' but not defined in exporters section`,
                yamlContent,
                ["service", "pipelines", pipelineName, "exporters"],
                exporter,
                "error",
              ),
            );
          }
        }
      }
    }

    return errors;
  }
}
