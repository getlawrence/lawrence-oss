import yaml from "js-yaml";
import { describe, it, expect } from "vitest";

import { OTelPipelineValidator } from "./otel-pipeline-validator";

describe("OTelPipelineValidator", () => {
  const validator = new OTelPipelineValidator();

  describe("valid configurations", () => {
    it("should pass when all receivers are defined", () => {
      const yamlContent = `receivers:
  otlp:
    protocols:
      grpc:
service:
  pipelines:
    traces:
      receivers: [otlp]`;

      const parsed = yaml.load(yamlContent);
      const errors = validator.validate(yamlContent, parsed);

      expect(errors).toHaveLength(0);
    });

    it("should pass when all processors are defined", () => {
      const yamlContent = `receivers:
  otlp:
processors:
  batch:
service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]`;

      const parsed = yaml.load(yamlContent);
      const errors = validator.validate(yamlContent, parsed);

      expect(errors).toHaveLength(0);
    });

    it("should pass when all exporters are defined", () => {
      const yamlContent = `receivers:
  otlp:
exporters:
  logging:
service:
  pipelines:
    traces:
      receivers: [otlp]
      exporters: [logging]`;

      const parsed = yaml.load(yamlContent);
      const errors = validator.validate(yamlContent, parsed);

      expect(errors).toHaveLength(0);
    });

    it("should pass with multiple components", () => {
      const yamlContent = `receivers:
  otlp:
  jaeger:
processors:
  batch:
  memory_limiter:
exporters:
  otlp:
  logging:
service:
  pipelines:
    traces:
      receivers: [otlp, jaeger]
      processors: [batch, memory_limiter]
      exporters: [otlp, logging]`;

      const parsed = yaml.load(yamlContent);
      const errors = validator.validate(yamlContent, parsed);

      expect(errors).toHaveLength(0);
    });

    it("should pass with connectors used as receivers and exporters", () => {
      const yamlContent = `receivers:
  otlp:
connectors:
  spanmetrics:
exporters:
  logging:
service:
  pipelines:
    traces:
      receivers: [otlp]
      exporters: [spanmetrics]
    metrics:
      receivers: [spanmetrics]
      exporters: [logging]`;

      const parsed = yaml.load(yamlContent);
      const errors = validator.validate(yamlContent, parsed);

      expect(errors).toHaveLength(0);
    });
  });

  describe("invalid configurations", () => {
    it("should detect undefined receiver", () => {
      const yamlContent = `receivers:
  otlp:
service:
  pipelines:
    traces:
      receivers:
        - otlp
        - undefined_receiver`;

      const parsed = yaml.load(yamlContent);
      const errors = validator.validate(yamlContent, parsed);

      expect(errors).toHaveLength(1);
      expect(errors[0]).toMatchObject({
        message:
          "Receiver 'undefined_receiver' is used in pipeline 'traces' but not defined in receivers section",
        severity: "error",
        line: 8,
        column: 11,
      });
    });

    it("should detect undefined processor", () => {
      const yamlContent = `receivers:
  otlp:
processors:
  batch:
service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch, undefined_processor]`;

      const parsed = yaml.load(yamlContent);
      const errors = validator.validate(yamlContent, parsed);

      expect(errors).toHaveLength(1);
      expect(errors[0]).toMatchObject({
        message:
          "Processor 'undefined_processor' is used in pipeline 'traces' but not defined in processors section",
        severity: "error",
      });
    });

    it("should detect undefined exporter", () => {
      const yamlContent = `receivers:
  otlp:
exporters:
  logging:
service:
  pipelines:
    traces:
      receivers: [otlp]
      exporters: [logging, undefined_exporter]`;

      const parsed = yaml.load(yamlContent);
      const errors = validator.validate(yamlContent, parsed);

      expect(errors).toHaveLength(1);
      expect(errors[0]).toMatchObject({
        message:
          "Exporter 'undefined_exporter' is used in pipeline 'traces' but not defined in exporters section",
        severity: "error",
      });
    });

    it("should detect multiple undefined components", () => {
      const yamlContent = `receivers:
  otlp:
service:
  pipelines:
    traces:
      receivers: [otlp, data]
      processors: [batch]
      exporters: [logging]`;

      const parsed = yaml.load(yamlContent);
      const errors = validator.validate(yamlContent, parsed);

      expect(errors).toHaveLength(3);
      expect(errors[0].message).toContain("Receiver 'data'");
      expect(errors[1].message).toContain("Processor 'batch'");
      expect(errors[2].message).toContain("Exporter 'logging'");
    });

    it("should detect undefined components in multiple pipelines", () => {
      const yamlContent = `receivers:
  otlp:
exporters:
  logging:
service:
  pipelines:
    traces:
      receivers: [otlp, jaeger]
      exporters: [logging]
    metrics:
      receivers: [prometheus]
      exporters: [logging]`;

      const parsed = yaml.load(yamlContent);
      const errors = validator.validate(yamlContent, parsed);

      expect(errors).toHaveLength(2);
      expect(errors[0].message).toContain("Receiver 'jaeger'");
      expect(errors[0].message).toContain("pipeline 'traces'");
      expect(errors[1].message).toContain("Receiver 'prometheus'");
      expect(errors[1].message).toContain("pipeline 'metrics'");
    });
  });

  describe("edge cases", () => {
    it("should handle empty config", () => {
      const yamlContent = ``;
      const parsed = yaml.load(yamlContent);
      const errors = validator.validate(yamlContent, parsed);

      expect(errors).toHaveLength(0);
    });

    it("should handle config without service section", () => {
      const yamlContent = `receivers:
  otlp:`;

      const parsed = yaml.load(yamlContent);
      const errors = validator.validate(yamlContent, parsed);

      expect(errors).toHaveLength(0);
    });

    it("should handle config without pipelines section", () => {
      const yamlContent = `receivers:
  otlp:
service:
  extensions: [health_check]`;

      const parsed = yaml.load(yamlContent);
      const errors = validator.validate(yamlContent, parsed);

      expect(errors).toHaveLength(0);
    });

    it("should handle pipeline without receivers", () => {
      const yamlContent = `exporters:
  logging:
service:
  pipelines:
    traces:
      exporters: [logging]`;

      const parsed = yaml.load(yamlContent);
      const errors = validator.validate(yamlContent, parsed);

      expect(errors).toHaveLength(0);
    });

    it("should handle pipeline without processors", () => {
      const yamlContent = `receivers:
  otlp:
exporters:
  logging:
service:
  pipelines:
    traces:
      receivers: [otlp]
      exporters: [logging]`;

      const parsed = yaml.load(yamlContent);
      const errors = validator.validate(yamlContent, parsed);

      expect(errors).toHaveLength(0);
    });

    it("should handle pipeline without exporters", () => {
      const yamlContent = `receivers:
  otlp:
service:
  pipelines:
    traces:
      receivers: [otlp]`;

      const parsed = yaml.load(yamlContent);
      const errors = validator.validate(yamlContent, parsed);

      expect(errors).toHaveLength(0);
    });

    it("should handle components with named instances", () => {
      const yamlContent = `receivers:
  otlp:
  otlp/2:
  otlp/custom:
exporters:
  logging:
service:
  pipelines:
    traces:
      receivers: [otlp, otlp/2, otlp/custom]
      exporters: [logging]`;

      const parsed = yaml.load(yamlContent);
      const errors = validator.validate(yamlContent, parsed);

      expect(errors).toHaveLength(0);
    });

    it("should handle invalid YAML (non-object)", () => {
      const yamlContent = `just a string`;
      const parsed = yaml.load(yamlContent);
      const errors = validator.validate(yamlContent, parsed);

      expect(errors).toHaveLength(0);
    });
  });

  describe("position accuracy", () => {
    it("should highlight the specific invalid receiver in multi-line array", () => {
      const yamlContent = `receivers:
  otlp:
service:
  pipelines:
    traces:
      receivers:
        - otlp
        - data`;

      const parsed = yaml.load(yamlContent);
      const errors = validator.validate(yamlContent, parsed);

      expect(errors).toHaveLength(1);
      expect(errors[0].line).toBe(8);
      expect(errors[0].column).toBe(11);
      expect(errors[0].endColumn).toBe(15);
    });

    it("should highlight the specific invalid receiver in inline array", () => {
      const yamlContent = `receivers:
  otlp:
service:
  pipelines:
    traces:
      receivers: [otlp, data, jaeger]`;

      const parsed = yaml.load(yamlContent);
      const errors = validator.validate(yamlContent, parsed);

      expect(errors).toHaveLength(2);

      // Check that 'data' is highlighted
      const dataError = errors.find((e) => e.message.includes("'data'"));
      expect(dataError).toBeDefined();
      expect(dataError?.line).toBe(6);
      expect(dataError?.column).toBe(25);

      // Check that 'jaeger' is highlighted
      const jaegerError = errors.find((e) => e.message.includes("'jaeger'"));
      expect(jaegerError).toBeDefined();
      expect(jaegerError?.line).toBe(6);
      expect(jaegerError?.column).toBe(31);
    });
  });
});
