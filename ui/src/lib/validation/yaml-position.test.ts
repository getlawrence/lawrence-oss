import { describe, it, expect } from "vitest";

import {
  createValidationError,
  findYamlPosition,
} from "./yaml-position";

describe("findYamlPosition", () => {
  describe("basic path finding", () => {
    it("should find a top-level key", () => {
      const yaml = `receivers:
  otlp:
    protocols:
      grpc:`;

      const result = findYamlPosition(yaml, ["receivers"]);

      expect(result).toEqual({
        line: 1,
        column: 1,
        endLine: 1,
        endColumn: 10,
      });
    });

    it("should find a nested key", () => {
      const yaml = `receivers:
  otlp:
    protocols:
      grpc:`;

      const result = findYamlPosition(yaml, ["receivers", "otlp", "protocols"]);

      expect(result).toEqual({
        line: 3,
        column: 5,
        endLine: 3,
        endColumn: 14,
      });
    });

    it("should find deeply nested keys", () => {
      const yaml = `service:
  pipelines:
    traces:
      receivers:
        - otlp`;

      const result = findYamlPosition(yaml, [
        "service",
        "pipelines",
        "traces",
        "receivers",
      ]);

      expect(result).toEqual({
        line: 4,
        column: 7,
        endLine: 4,
        endColumn: 16,
      });
    });
  });

  describe("multi-line array component finding", () => {
    it("should find a component in a multi-line array", () => {
      const yaml = `service:
  pipelines:
    traces:
      receivers:
        - otlp
        - jaeger
        - zipkin`;

      const result = findYamlPosition(
        yaml,
        ["service", "pipelines", "traces", "receivers"],
        "jaeger",
      );

      expect(result).toEqual({
        line: 6,
        column: 11,
        endLine: 6,
        endColumn: 17,
      });
    });

    it("should find first component in multi-line array", () => {
      const yaml = `service:
  pipelines:
    traces:
      receivers:
        - otlp
        - jaeger`;

      const result = findYamlPosition(
        yaml,
        ["service", "pipelines", "traces", "receivers"],
        "otlp",
      );

      expect(result).toEqual({
        line: 5,
        column: 11,
        endLine: 5,
        endColumn: 15,
      });
    });

    it("should find last component in multi-line array", () => {
      const yaml = `service:
  pipelines:
    traces:
      receivers:
        - otlp
        - jaeger
        - zipkin`;

      const result = findYamlPosition(
        yaml,
        ["service", "pipelines", "traces", "receivers"],
        "zipkin",
      );

      expect(result).toEqual({
        line: 7,
        column: 11,
        endLine: 7,
        endColumn: 17,
      });
    });

    it("should handle array items with extra spaces", () => {
      const yaml = `service:
  pipelines:
    traces:
      receivers:
        -   otlp
        -  jaeger`;

      const result = findYamlPosition(
        yaml,
        ["service", "pipelines", "traces", "receivers"],
        "jaeger",
      );

      expect(result).toEqual({
        line: 6,
        column: 12,
        endLine: 6,
        endColumn: 18,
      });
    });
  });

  describe("inline array component finding", () => {
    it("should find a component in an inline array", () => {
      const yaml = `service:
  pipelines:
    traces:
      receivers: [otlp, jaeger, zipkin]`;

      const result = findYamlPosition(
        yaml,
        ["service", "pipelines", "traces", "receivers"],
        "jaeger",
      );

      expect(result).toEqual({
        line: 4,
        column: 25,
        endLine: 4,
        endColumn: 31,
      });
    });

    it("should find first component in inline array", () => {
      const yaml = `service:
  pipelines:
    traces:
      receivers: [otlp, jaeger]`;

      const result = findYamlPosition(
        yaml,
        ["service", "pipelines", "traces", "receivers"],
        "otlp",
      );

      expect(result).toEqual({
        line: 4,
        column: 19,
        endLine: 4,
        endColumn: 23,
      });
    });

    it("should find last component in inline array", () => {
      const yaml = `service:
  pipelines:
    traces:
      receivers: [otlp, jaeger, zipkin]`;

      const result = findYamlPosition(
        yaml,
        ["service", "pipelines", "traces", "receivers"],
        "zipkin",
      );

      expect(result).toEqual({
        line: 4,
        column: 33,
        endLine: 4,
        endColumn: 39,
      });
    });

    it("should handle inline arrays with spaces", () => {
      const yaml = `service:
  pipelines:
    traces:
      receivers: [ otlp , jaeger , zipkin ]`;

      const result = findYamlPosition(
        yaml,
        ["service", "pipelines", "traces", "receivers"],
        "jaeger",
      );

      expect(result).toEqual({
        line: 4,
        column: 27,
        endLine: 4,
        endColumn: 33,
      });
    });
  });

  describe("edge cases", () => {
    it("should return null for non-existent path", () => {
      const yaml = `receivers:
  otlp:`;

      const result = findYamlPosition(yaml, ["service", "pipelines"]);

      expect(result).toBeNull();
    });

    it("should fallback to key position when component not found in array", () => {
      const yaml = `service:
  pipelines:
    traces:
      receivers:
        - otlp
        - jaeger`;

      const result = findYamlPosition(
        yaml,
        ["service", "pipelines", "traces", "receivers"],
        "nonexistent",
      );

      // When component is not found, it returns the key position as fallback
      expect(result).toEqual({
        line: 4,
        column: 7,
        endLine: 4,
        endColumn: 16,
      });
    });

    it("should skip empty lines and comments", () => {
      const yaml = `service:
  pipelines:
    # This is a comment
    traces:
      receivers:
        # Another comment
        - otlp

        - jaeger`;

      const result = findYamlPosition(
        yaml,
        ["service", "pipelines", "traces", "receivers"],
        "jaeger",
      );

      expect(result).toEqual({
        line: 9,
        column: 11,
        endLine: 9,
        endColumn: 17,
      });
    });

    it("should handle multiple pipelines", () => {
      const yaml = `service:
  pipelines:
    traces:
      receivers:
        - otlp
      exporters:
        - otlp
    metrics:
      receivers:
        - prometheus`;

      const result = findYamlPosition(
        yaml,
        ["service", "pipelines", "metrics", "receivers"],
        "prometheus",
      );

      expect(result).toEqual({
        line: 10,
        column: 11,
        endLine: 10,
        endColumn: 21,
      });
    });

    it("should handle components with similar names", () => {
      const yaml = `service:
  pipelines:
    traces:
      receivers:
        - otlp
        - otlp/2`;

      const result = findYamlPosition(
        yaml,
        ["service", "pipelines", "traces", "receivers"],
        "otlp/2",
      );

      expect(result).toEqual({
        line: 6,
        column: 11,
        endLine: 6,
        endColumn: 17,
      });
    });
  });

  describe("processor and exporter arrays", () => {
    it("should find processor in array", () => {
      const yaml = `service:
  pipelines:
    traces:
      processors:
        - batch
        - memory_limiter`;

      const result = findYamlPosition(
        yaml,
        ["service", "pipelines", "traces", "processors"],
        "memory_limiter",
      );

      expect(result).toEqual({
        line: 6,
        column: 11,
        endLine: 6,
        endColumn: 25,
      });
    });

    it("should find exporter in array", () => {
      const yaml = `service:
  pipelines:
    traces:
      exporters:
        - otlp
        - logging`;

      const result = findYamlPosition(
        yaml,
        ["service", "pipelines", "traces", "exporters"],
        "logging",
      );

      expect(result).toEqual({
        line: 6,
        column: 11,
        endLine: 6,
        endColumn: 18,
      });
    });
  });
});

describe("createValidationError", () => {
  it("should create error with position information", () => {
    const yaml = `service:
  pipelines:
    traces:
      receivers:
        - otlp
        - data`;

    const error = createValidationError(
      "Receiver 'data' is not defined",
      yaml,
      ["service", "pipelines", "traces", "receivers"],
      "data",
      "error",
    );

    expect(error).toEqual({
      message: "Receiver 'data' is not defined",
      severity: "error",
      line: 6,
      column: 11,
      endLine: 6,
      endColumn: 15,
      path: ["service", "pipelines", "traces", "receivers"],
    });
  });

  it("should create error for inline array", () => {
    const yaml = `service:
  pipelines:
    traces:
      receivers: [otlp, data]`;

    const error = createValidationError(
      "Receiver 'data' is not defined",
      yaml,
      ["service", "pipelines", "traces", "receivers"],
      "data",
      "error",
    );

    expect(error).toEqual({
      message: "Receiver 'data' is not defined",
      severity: "error",
      line: 4,
      column: 25,
      endLine: 4,
      endColumn: 29,
      path: ["service", "pipelines", "traces", "receivers"],
    });
  });

  it("should fallback to line 1 when position not found", () => {
    const yaml = `receivers:
  otlp:`;

    const error = createValidationError(
      "Something is wrong",
      yaml,
      ["nonexistent", "path"],
      undefined,
      "warning",
    );

    expect(error).toEqual({
      message: "Something is wrong",
      severity: "warning",
      line: 1,
      column: 1,
      path: ["nonexistent", "path"],
    });
  });

  it("should create warning with correct severity", () => {
    const yaml = `service:
  pipelines:
    traces:
      receivers:
        - otlp`;

    const error = createValidationError(
      "This is a warning",
      yaml,
      ["service", "pipelines", "traces", "receivers"],
      "otlp",
      "warning",
    );

    expect(error.severity).toBe("warning");
  });

  it("should default to error severity", () => {
    const yaml = `service:
  pipelines:`;

    const error = createValidationError(
      "This is an error",
      yaml,
      ["service", "pipelines"],
    );

    expect(error.severity).toBe("error");
  });
});
