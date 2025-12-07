import yaml from "js-yaml";
import type { editor } from "monaco-editor";
import { useState, useEffect, useCallback, useRef } from "react";

import {
  validateComponentConfig,
  getComponentSchema,
  type ComponentType,
  type JSONSchema,
} from "@/api/schemas";
import {
  validateYamlConfig,
  type ValidationResult,
  type ValidationError,
  type Validator,
  type OTelConfig,
  defaultValidators,
} from "@/lib/validation";
import { findYamlPosition } from "@/lib/validation/yaml-position";

interface UseYamlValidationOptions {
  debounceMs?: number;
  validators?: Validator[];
  enableSchemaValidation?: boolean; // Enable async schema validation
}

export function useYamlValidation(
  yamlContent: string,
  editorRef: React.MutableRefObject<editor.IStandaloneCodeEditor | null>,
  options: UseYamlValidationOptions = {},
) {
  const {
    debounceMs = 500,
    validators = defaultValidators,
    enableSchemaValidation = true,
  } = options;
  const [validationResult, setValidationResult] = useState<ValidationResult>({
    valid: true,
    errors: [],
  });
  const [isValidating, setIsValidating] = useState(false);
  const abortControllerRef = useRef<AbortController | null>(null);

  const updateEditorMarkers = useCallback(
    (errors: ValidationError[]) => {
      if (!editorRef.current) {
        return;
      }

      const monaco = editorRef.current;
      const model = monaco.getModel();

      if (model) {
        // First, clear any existing markers from this owner to prevent duplicates
        const monacoEditor = (
          window as unknown as {
            monaco?: {
              editor: {
                setModelMarkers: (
                  model: editor.ITextModel,
                  owner: string,
                  markers: editor.IMarkerData[],
                ) => void;
              };
            };
          }
        ).monaco;

        if (monacoEditor) {
          // Clear existing markers first
          monacoEditor.editor.setModelMarkers(model, "otel-validator", []);

          // Create markers and deduplicate them
          const markerMap = new Map<string, editor.IMarkerData>();

          for (const error of errors) {
            const marker: editor.IMarkerData = {
              severity:
                error.severity === "error"
                  ? 8 // MarkerSeverity.Error
                  : 4, // MarkerSeverity.Warning
              startLineNumber: error.line,
              startColumn: error.column,
              endLineNumber: error.endLine || error.line,
              endColumn: error.endColumn || error.column + 1,
              message: error.message,
              source: "otel-validator",
            };

            // Create a unique key for deduplication
            const markerKey = `${marker.message}|${marker.startLineNumber}|${marker.startColumn}|${marker.endLineNumber}|${marker.endColumn}`;

            // Only add if we haven't seen this exact marker before
            if (!markerMap.has(markerKey)) {
              markerMap.set(markerKey, marker);
            }
          }

          // Convert map values to array and set markers
          const uniqueMarkers = Array.from(markerMap.values());
          monacoEditor.editor.setModelMarkers(
            model,
            "otel-validator",
            uniqueMarkers,
          );
        }
      }
    },
    [editorRef],
  );

  /**
   * Find invalid keys in a component config by comparing against schema
   */
  const findInvalidKeys = useCallback(
    (config: Record<string, unknown>, schema: JSONSchema): string[] => {
      const invalidKeys: string[] = [];
      const validKeys = new Set<string>();

      // Extract valid keys from schema properties
      const extractValidKeys = (schemaObj: JSONSchema, prefix = ""): void => {
        if (schemaObj.properties) {
          for (const [key, value] of Object.entries(schemaObj.properties)) {
            const fullKey = prefix ? `${prefix}.${key}` : key;
            validKeys.add(key); // Add the direct key
            validKeys.add(fullKey); // Add the full path key

            // Recursively extract nested properties
            if (value && typeof value === "object") {
              extractValidKeys(value as JSONSchema, fullKey);
            }
          }
        }

        // Handle oneOf, anyOf, allOf
        if (schemaObj.oneOf) {
          schemaObj.oneOf.forEach((subSchema) =>
            extractValidKeys(subSchema, prefix),
          );
        }
        if (schemaObj.anyOf) {
          schemaObj.anyOf.forEach((subSchema) =>
            extractValidKeys(subSchema, prefix),
          );
        }
        if (schemaObj.allOf) {
          schemaObj.allOf.forEach((subSchema) =>
            extractValidKeys(subSchema, prefix),
          );
        }
      };

      extractValidKeys(schema);

      // Check each config key against valid keys
      for (const key of Object.keys(config)) {
        if (!validKeys.has(key)) {
          invalidKeys.push(key);
        }
      }

      return invalidKeys;
    },
    [],
  );

  const performSchemaValidation = useCallback(
    async (
      content: string,
      parsed: OTelConfig,
      syncErrors: ValidationError[],
    ): Promise<ValidationError[]> => {
      if (!enableSchemaValidation) {
        return syncErrors;
      }

      const schemaErrors: ValidationError[] = [];
      const componentTypes: Array<{
        type: ComponentType;
        plural: string;
        components: Record<string, unknown> | undefined;
      }> = [
        { type: "receiver", plural: "receivers", components: parsed.receivers },
        {
          type: "processor",
          plural: "processors",
          components: parsed.processors,
        },
        { type: "exporter", plural: "exporters", components: parsed.exporters },
        {
          type: "connector",
          plural: "connectors",
          components: parsed.connectors,
        },
        {
          type: "extension",
          plural: "extensions",
          components: parsed.extensions,
        },
      ];

      // Validate each component against its schema
      const validationPromises = componentTypes.flatMap(
        ({ type, plural, components }) => {
          if (!components) {
            return [];
          }

          return Object.entries(components).map(
            async ([componentName, componentConfig]) => {
              // Skip if config is not an object (already caught by sync validators)
              if (
                !componentConfig ||
                typeof componentConfig !== "object" ||
                Array.isArray(componentConfig)
              ) {
                return;
              }

              try {
                // First try API validation
                const response = await validateComponentConfig({
                  type,
                  name: componentName,
                  config: componentConfig,
                });

                // If API says valid but we want to double-check, fetch schema and validate keys
                if (response.valid) {
                  try {
                    const schema = await getComponentSchema(
                      type,
                      componentName,
                    );
                    const invalidKeys = findInvalidKeys(
                      componentConfig as Record<string, unknown>,
                      schema,
                    );

                    if (invalidKeys.length > 0) {
                      // Find the position of this component in the YAML
                      const path = [plural, componentName];
                      const basePosition = findYamlPosition(content, path);

                      for (const invalidKey of invalidKeys) {
                        let errorLine = basePosition?.line || 1;
                        let errorColumn = basePosition?.column || 1;

                        // Try to find the invalid key in the YAML
                        const fieldPath = [...path, invalidKey];
                        const fieldPosition = findYamlPosition(
                          content,
                          fieldPath,
                        );
                        if (fieldPosition) {
                          errorLine = fieldPosition.line;
                          errorColumn = fieldPosition.column;
                        } else {
                          // Fallback: search for the key in the component's section
                          const lines = content.split("\n");
                          const componentStartLine = basePosition?.line
                            ? basePosition.line - 1
                            : 0;
                          for (
                            let i = componentStartLine;
                            i < lines.length;
                            i++
                          ) {
                            const line = lines[i];
                            const fieldIndex = line.indexOf(invalidKey + ":");
                            if (fieldIndex !== -1) {
                              errorLine = i + 1;
                              errorColumn = fieldIndex + 1;
                              break;
                            }
                            // Stop if we've moved to a different section
                            if (
                              i > componentStartLine &&
                              line.trim() &&
                              !line.startsWith(" ")
                            ) {
                              break;
                            }
                          }
                        }

                        schemaErrors.push({
                          message: `Unknown field '${invalidKey}' in ${type} '${componentName}'`,
                          severity: "error",
                          line: errorLine,
                          column: errorColumn,
                          endLine: errorLine,
                          endColumn: errorColumn + invalidKey.length,
                          path: fieldPath,
                        });
                      }
                    }
                  } catch {
                    // If we can't fetch schema, that's okay - just use API validation result
                    // Silently continue without schema-based validation
                  }
                }

                if (
                  !response.valid &&
                  response.errors &&
                  response.errors.length > 0
                ) {
                  // Find the position of this component in the YAML
                  const path = [plural, componentName];
                  const basePosition = findYamlPosition(content, path);

                  // Create errors for each schema validation error
                  for (const errorMsg of response.errors) {
                    // Try to extract field name from error message
                    // Error messages can be like "dffg-error: unknown field" or "unknown field: dffg-error"
                    let errorLine = basePosition?.line || 1;
                    let errorColumn = basePosition?.column || 1;
                    let fieldName: string | undefined;

                    // Try different patterns to extract field name
                    const fieldMatch1 = errorMsg.match(/^([^:]+):/); // "field: error message"
                    const fieldMatch2 = errorMsg.match(
                      /unknown field[:\s]+([^\s,]+)/i,
                    ); // "unknown field: fieldname"
                    const fieldMatch3 = errorMsg.match(
                      /field[:\s]+['"]?([^'",\s]+)['"]?/i,
                    ); // "field 'fieldname'"

                    if (fieldMatch1) {
                      fieldName = fieldMatch1[1].trim();
                    } else if (fieldMatch2) {
                      fieldName = fieldMatch2[1].trim();
                    } else if (fieldMatch3) {
                      fieldName = fieldMatch3[1].trim();
                    }

                    // Try to find the specific invalid field in the YAML
                    if (fieldName && basePosition) {
                      const fieldPath = [...path, fieldName];
                      const fieldPosition = findYamlPosition(
                        content,
                        fieldPath,
                      );
                      if (fieldPosition) {
                        errorLine = fieldPosition.line;
                        errorColumn = fieldPosition.column;
                      } else {
                        // If we can't find the exact position, search for the field name in the component's section
                        const lines = content.split("\n");
                        const componentStartLine = basePosition.line - 1; // Convert to 0-based

                        // Search for the field name in the component's section
                        for (
                          let i = componentStartLine;
                          i < lines.length;
                          i++
                        ) {
                          const line = lines[i];
                          const fieldIndex = line.indexOf(fieldName + ":");
                          if (fieldIndex !== -1) {
                            errorLine = i + 1; // Convert back to 1-based
                            errorColumn = fieldIndex + 1;
                            break;
                          }
                          // Stop if we've moved to a different section (less indentation)
                          if (
                            i > componentStartLine &&
                            line.trim() &&
                            !line.startsWith(" ")
                          ) {
                            break;
                          }
                        }
                      }
                    }

                    schemaErrors.push({
                      message: errorMsg,
                      severity: "error",
                      line: errorLine,
                      column: errorColumn,
                      endLine: errorLine,
                      endColumn: errorColumn + (fieldName?.length || 20),
                      path: fieldName ? [...path, fieldName] : path,
                    });
                  }
                }
              } catch (error) {
                // Log errors but don't break validation if schema service is unavailable
                console.warn(
                  `Schema validation API call failed for ${type}/${componentName}:`,
                  error,
                );
                // Don't add errors for network failures - just log them
              }
            },
          );
        },
      );

      await Promise.all(validationPromises);

      // Combine sync and async errors, then deduplicate
      const allErrors = [...syncErrors, ...schemaErrors];

      // Deduplicate errors based on message, line, column, and path
      const uniqueErrors: ValidationError[] = [];
      const seen = new Set<string>();

      for (const error of allErrors) {
        // Create a more specific key that includes path to avoid duplicates
        const pathKey = error.path ? error.path.join(".") : "";
        const key = `${error.message}|${error.line}|${error.column}|${pathKey}`;
        if (!seen.has(key)) {
          seen.add(key);
          uniqueErrors.push(error);
        }
      }

      return uniqueErrors;
    },
    [enableSchemaValidation, findInvalidKeys],
  );

  const performValidation = useCallback(
    async (content: string) => {
      // Cancel any pending schema validation
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
      abortControllerRef.current = new AbortController();

      if (!content.trim()) {
        setValidationResult({ valid: true, errors: [] });
        setIsValidating(false);
        updateEditorMarkers([]);
        return;
      }

      try {
        // First parse the YAML
        const parsed = yaml.load(content) as OTelConfig;

        // Run synchronous validators first
        const syncResult = validateYamlConfig(content, parsed, validators);

        // Then perform async schema validation
        if (
          enableSchemaValidation &&
          !abortControllerRef.current.signal.aborted
        ) {
          const allErrors = await performSchemaValidation(
            content,
            parsed,
            syncResult.errors,
          );

          // Check if validation was cancelled
          if (abortControllerRef.current.signal.aborted) {
            return;
          }

          const finalResult: ValidationResult = {
            valid: allErrors.filter((e) => e.severity === "error").length === 0,
            errors: allErrors,
          };

          setValidationResult(finalResult);
          updateEditorMarkers(allErrors);
        } else {
          // If schema validation is disabled, just use sync results
          setValidationResult(syncResult);
          updateEditorMarkers(syncResult.errors);
        }
      } catch (error) {
        // YAML parse error - don't run validators
        const parseError: ValidationError = {
          message: error instanceof Error ? error.message : "YAML parse error",
          severity: "error",
          line: 1,
          column: 1,
        };
        setValidationResult({
          valid: false,
          errors: [parseError],
        });
        updateEditorMarkers([parseError]);
      } finally {
        setIsValidating(false);
      }
    },
    [
      validators,
      enableSchemaValidation,
      performSchemaValidation,
      updateEditorMarkers,
    ],
  );

  useEffect(() => {
    setIsValidating(true);
    const timer = setTimeout(() => {
      performValidation(yamlContent);
    }, debounceMs);

    return () => {
      clearTimeout(timer);
    };
  }, [yamlContent, debounceMs, performValidation]);

  return { validationResult, isValidating };
}
