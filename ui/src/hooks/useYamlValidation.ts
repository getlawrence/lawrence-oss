import yaml from "js-yaml";
import type { editor } from "monaco-editor";
import { useState, useEffect, useCallback } from "react";

import {
  validateYamlConfig,
  type ValidationResult,
  type Validator,
  defaultValidators,
} from "@/lib/validation";

interface UseYamlValidationOptions {
  debounceMs?: number;
  validators?: Validator[];
}

export function useYamlValidation(
  yamlContent: string,
  editorRef: React.MutableRefObject<editor.IStandaloneCodeEditor | null>,
  options: UseYamlValidationOptions = {},
) {
  const { debounceMs = 500, validators = defaultValidators } = options;
  const [validationResult, setValidationResult] = useState<ValidationResult>({
    valid: true,
    errors: [],
  });
  const [isValidating, setIsValidating] = useState(false);

  const performValidation = useCallback(
    (content: string) => {
      if (!content.trim()) {
        setValidationResult({ valid: true, errors: [] });
        setIsValidating(false);
        return;
      }

      try {
        // First parse the YAML
        const parsed = yaml.load(content);

        // Run validators
        const result = validateYamlConfig(content, parsed, validators);
        setValidationResult(result);

        // Update Monaco editor markers
        if (editorRef.current) {
          const monaco = editorRef.current;
          const model = monaco.getModel();

          if (model) {
            const markers: editor.IMarkerData[] = result.errors.map(
              (error) => ({
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
              }),
            );

            // Set markers on the model
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
              monacoEditor.editor.setModelMarkers(
                model,
                "otel-validator",
                markers,
              );
            }
          }
        }
      } catch (error) {
        // YAML parse error - don't run validators
        setValidationResult({
          valid: false,
          errors: [
            {
              message:
                error instanceof Error ? error.message : "YAML parse error",
              severity: "error",
              line: 1,
              column: 1,
            },
          ],
        });
      } finally {
        setIsValidating(false);
      }
    },
    [editorRef, validators],
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
