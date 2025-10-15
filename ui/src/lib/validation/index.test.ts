import yaml from "js-yaml";
import { describe, it, expect } from "vitest";

import type { Validator, ValidationError } from "./types";

import { validateYamlConfig } from "./index";

describe("validateYamlConfig", () => {
  describe("default validators", () => {
    it("should run all default validators", () => {
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
      const result = validateYamlConfig(yamlContent, parsed);

      expect(result.valid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    it("should detect errors from multiple validators", () => {
      const yamlContent = `receivers:
  otlp:
service:
  pipelines:
    traces:
      receivers: [otlp, undefined_receiver]`;

      const parsed = yaml.load(yamlContent);
      const result = validateYamlConfig(yamlContent, parsed);

      expect(result.valid).toBe(false);
      expect(result.errors.length).toBeGreaterThan(0);
      expect(
        result.errors.some((e) => e.message.includes("undefined_receiver")),
      ).toBe(true);
    });

    it("should mark config as valid when only warnings exist", () => {
      const yamlContent = `receivers:
  otlp:
extensions:
  health_check:
service:
  pipelines:
    traces:
      receivers: [otlp]`;

      const parsed = yaml.load(yamlContent);
      const result = validateYamlConfig(yamlContent, parsed);

      // Should be valid if only warnings (no errors with severity: "error")
      const hasErrors = result.errors.some((e) => e.severity === "error");
      expect(result.valid).toBe(!hasErrors);
    });
  });

  describe("custom validators", () => {
    it("should run custom validators", () => {
      const customValidator: Validator = {
        name: "test-validator",
        validate: (_yamlContent, parsedData) => {
          const errors: ValidationError[] = [];
          if (parsedData && typeof parsedData === "object") {
            const config = parsedData as { test?: string };
            if (!config.test) {
              errors.push({
                message: "test field is required",
                severity: "error",
                line: 1,
                column: 1,
              });
            }
          }
          return errors;
        },
      };

      const yamlContent = `receivers:
  otlp:`;

      const parsed = yaml.load(yamlContent);
      const result = validateYamlConfig(yamlContent, parsed, [customValidator]);

      expect(result.valid).toBe(false);
      expect(result.errors).toHaveLength(1);
      expect(result.errors[0].message).toBe("test field is required");
    });
  });

  describe("error deduplication", () => {
    it("should deduplicate errors with same message, line, and column", () => {
      // Create a validator that produces duplicate errors
      const duplicatingValidator: Validator = {
        name: "duplicating-validator",
        validate: () => {
          return [
            {
              message: "Duplicate error",
              severity: "error" as const,
              line: 5,
              column: 10,
            },
            {
              message: "Duplicate error",
              severity: "error" as const,
              line: 5,
              column: 10,
            },
            {
              message: "Duplicate error",
              severity: "error" as const,
              line: 5,
              column: 10,
            },
          ];
        },
      };

      const yamlContent = `receivers:
  otlp:`;

      const parsed = yaml.load(yamlContent);
      const result = validateYamlConfig(yamlContent, parsed, [
        duplicatingValidator,
      ]);

      // Should only have 1 error after deduplication
      expect(result.errors).toHaveLength(1);
      expect(result.errors[0].message).toBe("Duplicate error");
    });

    it("should keep errors with same message but different positions", () => {
      const multiErrorValidator: Validator = {
        name: "multi-error-validator",
        validate: () => {
          return [
            {
              message: "Same error",
              severity: "error" as const,
              line: 5,
              column: 10,
            },
            {
              message: "Same error",
              severity: "error" as const,
              line: 6,
              column: 10,
            },
            {
              message: "Same error",
              severity: "error" as const,
              line: 5,
              column: 15,
            },
          ];
        },
      };

      const yamlContent = `receivers:
  otlp:`;

      const parsed = yaml.load(yamlContent);
      const result = validateYamlConfig(yamlContent, parsed, [
        multiErrorValidator,
      ]);

      // Should have all 3 errors since they're at different positions
      expect(result.errors).toHaveLength(3);
    });

    it("should deduplicate errors from multiple validators", () => {
      const validator1: Validator = {
        name: "validator-1",
        validate: () => [
          {
            message: "Error at position",
            severity: "error" as const,
            line: 3,
            column: 5,
          },
        ],
      };

      const validator2: Validator = {
        name: "validator-2",
        validate: () => [
          {
            message: "Error at position",
            severity: "error" as const,
            line: 3,
            column: 5,
          },
        ],
      };

      const yamlContent = `receivers:
  otlp:`;

      const parsed = yaml.load(yamlContent);
      const result = validateYamlConfig(yamlContent, parsed, [
        validator1,
        validator2,
      ]);

      // Should only have 1 error after deduplication
      expect(result.errors).toHaveLength(1);
    });
  });

  describe("error handling", () => {
    it("should handle validator that throws error", () => {
      const throwingValidator: Validator = {
        name: "throwing-validator",
        validate: () => {
          throw new Error("Validator crashed");
        },
      };

      const yamlContent = `receivers:
  otlp:`;

      const parsed = yaml.load(yamlContent);

      // Should not throw, should handle gracefully
      expect(() => {
        validateYamlConfig(yamlContent, parsed, [throwingValidator]);
      }).not.toThrow();
    });

    it("should continue with other validators when one fails", () => {
      const throwingValidator: Validator = {
        name: "throwing-validator",
        validate: () => {
          throw new Error("Validator crashed");
        },
      };

      const workingValidator: Validator = {
        name: "working-validator",
        validate: () => [
          {
            message: "Working error",
            severity: "error" as const,
            line: 1,
            column: 1,
          },
        ],
      };

      const yamlContent = `receivers:
  otlp:`;

      const parsed = yaml.load(yamlContent);
      const result = validateYamlConfig(yamlContent, parsed, [
        throwingValidator,
        workingValidator,
      ]);

      // Should have error from working validator
      expect(result.errors).toHaveLength(1);
      expect(result.errors[0].message).toBe("Working error");
    });
  });

  describe("validity determination", () => {
    it("should be valid when no errors", () => {
      const yamlContent = `receivers:
  otlp:
service:
  pipelines:
    traces:
      receivers: [otlp]`;

      const parsed = yaml.load(yamlContent);
      const result = validateYamlConfig(yamlContent, parsed);

      expect(result.valid).toBe(true);
    });

    it("should be invalid when errors exist", () => {
      const yamlContent = `service:
  pipelines:
    traces:
      receivers: [undefined]`;

      const parsed = yaml.load(yamlContent);
      const result = validateYamlConfig(yamlContent, parsed);

      expect(result.valid).toBe(false);
    });

    it("should be valid when only warnings exist", () => {
      const warningValidator: Validator = {
        name: "warning-validator",
        validate: () => [
          {
            message: "This is a warning",
            severity: "warning" as const,
            line: 1,
            column: 1,
          },
        ],
      };

      const yamlContent = `receivers:
  otlp:`;

      const parsed = yaml.load(yamlContent);
      const result = validateYamlConfig(yamlContent, parsed, [
        warningValidator,
      ]);

      expect(result.valid).toBe(true);
      expect(result.errors).toHaveLength(1);
      expect(result.errors[0].severity).toBe("warning");
    });

    it("should be invalid when both errors and warnings exist", () => {
      const mixedValidator: Validator = {
        name: "mixed-validator",
        validate: () => [
          {
            message: "This is an error",
            severity: "error" as const,
            line: 1,
            column: 1,
          },
          {
            message: "This is a warning",
            severity: "warning" as const,
            line: 2,
            column: 1,
          },
        ],
      };

      const yamlContent = `receivers:
  otlp:`;

      const parsed = yaml.load(yamlContent);
      const result = validateYamlConfig(yamlContent, parsed, [mixedValidator]);

      expect(result.valid).toBe(false);
      expect(result.errors).toHaveLength(2);
    });
  });
});
