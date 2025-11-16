import { describe, it, expect } from "vitest";

import { parseYamlContext } from "./yamlContext";

describe("parseYamlContext", () => {
  describe("top-level position", () => {
    it("should detect cursor at top level (depth 0)", () => {
      const lines = [
        "receivers:",
        "  otlp:",
        "", // cursor here
      ];

      const result = parseYamlContext(lines, 2);

      expect(result).toEqual({
        section: null,
        component: null,
        path: [],
        depth: 0,
      });
    });

    it("should detect section from top-level position", () => {
      const lines = [
        "receivers:", // cursor here
        "  otlp:",
      ];

      const result = parseYamlContext(lines, 0);

      expect(result).toEqual({
        section: null,
        component: null,
        path: [],
        depth: 0,
      });
    });
  });

  describe("section level", () => {
    it("should detect section for processors", () => {
      const lines = [
        "processors:",
        "  ", // cursor here (indent 2)
      ];

      const result = parseYamlContext(lines, 1);

      expect(result).toEqual({
        section: "processors",
        component: null,
        path: [],
        depth: 1,
      });
    });

    it("should detect section for receivers", () => {
      const lines = [
        "receivers:",
        "  ", // cursor here
      ];

      const result = parseYamlContext(lines, 1);

      expect(result.section).toBe("receivers");
      expect(result.component).toBeNull();
    });

    it("should detect section for exporters", () => {
      const lines = [
        "exporters:",
        "  logging:", // cursor here
      ];

      const result = parseYamlContext(lines, 1);

      expect(result.section).toBe("exporters");
    });
  });

  describe("component level", () => {
    it("should detect component in processors section", () => {
      const lines = [
        "processors:",
        "  batch:",
        "    ", // cursor here (indent 4)
      ];

      const result = parseYamlContext(lines, 2);

      expect(result).toEqual({
        section: "processors",
        component: "batch",
        path: [],
        depth: 2,
      });
    });

    it("should detect component in receivers section", () => {
      const lines = [
        "receivers:",
        "  otlp:",
        "    protocols:", // cursor here
      ];

      const result = parseYamlContext(lines, 2);

      expect(result.section).toBe("receivers");
      expect(result.component).toBe("otlp");
      expect(result.path).toEqual([]);
    });

    it("should detect component in exporters section", () => {
      const lines = [
        "exporters:",
        "  logging:",
        "    ", // cursor here
      ];

      const result = parseYamlContext(lines, 2);

      expect(result.section).toBe("exporters");
      expect(result.component).toBe("logging");
    });
  });

  describe("nested properties", () => {
    it("should detect nested path in component", () => {
      const lines = [
        "receivers:",
        "  otlp:",
        "    protocols:",
        "      grpc:",
        "        ", // cursor here (indent 8)
      ];

      const result = parseYamlContext(lines, 4);

      expect(result).toEqual({
        section: "receivers",
        component: "otlp",
        path: ["protocols", "grpc"],
        depth: 4,
      });
    });

    it("should detect single nested property", () => {
      const lines = [
        "processors:",
        "  batch:",
        "    timeout:",
        "      ", // cursor here
      ];

      const result = parseYamlContext(lines, 3);

      expect(result.section).toBe("processors");
      expect(result.component).toBe("batch");
      expect(result.path).toEqual(["timeout"]);
    });

    it("should detect deeply nested path", () => {
      const lines = [
        "receivers:",
        "  otlp:",
        "    protocols:",
        "      http:",
        "        endpoint:",
        "          ", // cursor here
      ];

      const result = parseYamlContext(lines, 5);

      expect(result.path).toEqual(["protocols", "http", "endpoint"]);
    });
  });

  describe("complex YAML structures", () => {
    it("should handle multiple sections and find correct context", () => {
      const lines = [
        "receivers:",
        "  otlp:",
        "    protocols:",
        "      grpc:",
        "",
        "processors:",
        "  batch:",
        "    timeout: 10s", // cursor here (on the property line)
      ];

      const result = parseYamlContext(lines, 7);

      expect(result.section).toBe("processors");
      expect(result.component).toBe("batch");
      // When cursor is ON a property line, that property is not in the path
      expect(result.path).toEqual([]);
    });

    it("should handle service section", () => {
      const lines = [
        "service:",
        "  pipelines:",
        "    traces:", // cursor here
      ];

      const result = parseYamlContext(lines, 2);

      expect(result.section).toBe("service");
      expect(result.component).toBe("pipelines");
    });

    it("should ignore comments and empty lines", () => {
      const lines = [
        "processors:",
        "  # This is a comment",
        "  batch:",
        "    ",
        "    # Another comment",
        "    timeout: 10s", // cursor here (on the property line)
      ];

      const result = parseYamlContext(lines, 5);

      expect(result.section).toBe("processors");
      expect(result.component).toBe("batch");
      // When cursor is ON a property line, that property is not in the path
      expect(result.path).toEqual([]);
    });
  });

  describe("edge cases", () => {
    it("should handle empty file", () => {
      const lines: string[] = [];

      const result = parseYamlContext(lines, 0);

      expect(result).toEqual({
        section: null,
        component: null,
        path: [],
        depth: 0,
      });
    });

    it("should handle single line", () => {
      const lines = ["receivers:"];

      const result = parseYamlContext(lines, 0);

      expect(result.section).toBeNull();
    });

    it("should handle cursor at component name line", () => {
      const lines = [
        "processors:",
        "  batch:", // cursor here
      ];

      const result = parseYamlContext(lines, 1);

      expect(result.section).toBe("processors");
      expect(result.component).toBeNull();
      expect(result.depth).toBe(1);
    });

    it("should handle malformed YAML gracefully", () => {
      const lines = [
        "processors:",
        "batch", // no colon
        "  timeout: 10s", // cursor here
      ];

      const result = parseYamlContext(lines, 2);

      // Should still detect processors section
      expect(result.section).toBe("processors");
    });

    it("should handle component names with hyphens", () => {
      const lines = [
        "processors:",
        "  memory-limiter:",
        "    ", // cursor here
      ];

      const result = parseYamlContext(lines, 2);

      expect(result.component).toBe("memory-limiter");
    });

    it("should handle component names with underscores", () => {
      const lines = [
        "processors:",
        "  resource_detection:",
        "    ", // cursor here
      ];

      const result = parseYamlContext(lines, 2);

      expect(result.component).toBe("resource_detection");
    });
  });

  describe("real-world examples", () => {
    it("should parse standard OTel config", () => {
      const lines = [
        "receivers:",
        "  otlp:",
        "    protocols:",
        "      grpc:",
        "        endpoint: 0.0.0.0:4317",
        "      http:",
        "        endpoint: 0.0.0.0:4318",
        "",
        "processors:",
        "  batch:",
        "    timeout: 10s", // cursor here (on the property line)
        "    send_batch_size: 1024",
        "",
        "exporters:",
        "  otlp:",
        "    endpoint: localhost:4317",
      ];

      const result = parseYamlContext(lines, 10);

      expect(result).toEqual({
        section: "processors",
        component: "batch",
        path: [], // cursor is ON timeout line, not nested within it
        depth: 2,
      });
    });

    it("should handle cursor in service pipelines", () => {
      const lines = [
        "service:",
        "  pipelines:",
        "    traces:",
        "      receivers: [otlp]",
        "      processors: [batch]", // cursor here (on the property line)
        "      exporters: [otlp]",
      ];

      const result = parseYamlContext(lines, 4);

      expect(result.section).toBe("service");
      expect(result.component).toBe("pipelines");
      // cursor is ON processors line, not nested within it
      expect(result.path).toEqual(["traces"]);
    });
  });
});
