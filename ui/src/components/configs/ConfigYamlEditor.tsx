import Editor, { type Monaco } from "@monaco-editor/react";
import yaml from "js-yaml";
import * as monaco from "monaco-editor";
import { useEffect, useRef, useState, useCallback } from "react";

import {
  getComponentSchemas,
  getComponentSchema,
  type ComponentType,
} from "@/api/schemas";
import type { JSONSchema } from "@/api/schemas";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

interface ConfigYamlEditorProps {
  value: string;
  onChange: (value: string) => void;
}

export function ConfigYamlEditor({ value, onChange }: ConfigYamlEditorProps) {
  const [isSchemaLoaded, setIsSchemaLoaded] = useState(false);
  const [schema, setSchema] = useState<JSONSchema | null>(null);
  const editorRef = useRef<monaco.editor.IStandaloneCodeEditor | null>(null);
  const monacoRef = useRef<Monaco | null>(null);

  // Fetch and build the schema
  useEffect(() => {
    const loadSchema = async () => {
      try {
        const { components } = await getComponentSchemas();
        const otelConfigSchema = await buildOTelCollectorSchema(components);
        setSchema(otelConfigSchema);
        setIsSchemaLoaded(true);
      } catch (error) {
        console.error("Failed to load OTel collector schemas:", error);
      }
    };

    loadSchema();
  }, []);

  // Validate YAML and provide suggestions
  const validateYAML = useCallback(
    (yamlText: string) => {
      if (!monacoRef.current || !editorRef.current || !schema) return;

      try {
        // Parse YAML to JSON
        yaml.load(yamlText);

        // Clear previous markers
        const model = editorRef.current.getModel();
        if (model) {
          monacoRef.current.editor.setModelMarkers(model, "yaml-validator", []);
        }
      } catch (error: unknown) {
        // Show YAML syntax errors
        const model = editorRef.current.getModel();
        const yamlError = error as {
          mark?: { line: number; column: number };
          message?: string;
        };
        if (model && yamlError.mark) {
          const markers: monaco.editor.IMarkerData[] = [
            {
              severity: monaco.MarkerSeverity.Error,
              startLineNumber: yamlError.mark.line + 1,
              startColumn: yamlError.mark.column + 1,
              endLineNumber: yamlError.mark.line + 1,
              endColumn: yamlError.mark.column + 2,
              message: yamlError.message || "YAML syntax error",
            },
          ];
          monacoRef.current.editor.setModelMarkers(
            model,
            "yaml-validator",
            markers,
          );
        }
      }
    },
    [schema],
  );

  // Setup autocomplete provider
  useEffect(() => {
    if (!monacoRef.current || !schema) return;

    const completionProvider =
      monacoRef.current.languages.registerCompletionItemProvider("yaml", {
        triggerCharacters: [" ", ":", "\n"],
        provideCompletionItems: (model, position) => {
          const textUntilPosition = model.getValueInRange({
            startLineNumber: position.lineNumber,
            startColumn: 1,
            endLineNumber: position.lineNumber,
            endColumn: position.column,
          });

          const suggestions: monaco.languages.CompletionItem[] = [];

          // Parse current YAML to understand context
          try {
            const fullText = model.getValue();
            yaml.load(fullText);

            // Provide top-level suggestions
            if (
              textUntilPosition.trim() === "" ||
              !textUntilPosition.includes(":")
            ) {
              const topLevelProps = [
                "receivers",
                "processors",
                "exporters",
                "connectors",
                "extensions",
                "service",
              ];
              topLevelProps.forEach((prop, index) => {
                suggestions.push({
                  label: prop,
                  kind: monaco.languages.CompletionItemKind.Property,
                  insertText: `${prop}:\n  `,
                  insertTextRules:
                    monaco.languages.CompletionItemInsertTextRule
                      .InsertAsSnippet,
                  documentation:
                    (schema.properties?.[prop]?.description as string) ||
                    `${prop} configuration`,
                  sortText: `0${index}`,
                  range: {
                    startLineNumber: position.lineNumber,
                    startColumn: position.column - textUntilPosition.length,
                    endLineNumber: position.lineNumber,
                    endColumn: position.column,
                  },
                });
              });
            }
          } catch {
            // If YAML is invalid, still provide basic suggestions
            const topLevelProps = [
              "receivers",
              "processors",
              "exporters",
              "service",
            ];
            topLevelProps.forEach((prop, index) => {
              suggestions.push({
                label: prop,
                kind: monaco.languages.CompletionItemKind.Property,
                insertText: `${prop}:\n  `,
                insertTextRules:
                  monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
                documentation: `${prop} configuration`,
                sortText: `0${index}`,
                range: {
                  startLineNumber: position.lineNumber,
                  startColumn: 1,
                  endLineNumber: position.lineNumber,
                  endColumn: position.column,
                },
              });
            });
          }

          return { suggestions };
        },
      });

    return () => {
      completionProvider.dispose();
    };
  }, [schema]);

  // Validate on change
  useEffect(() => {
    validateYAML(value);
  }, [value, validateYAML]);

  const handleEditorDidMount = (
    editor: monaco.editor.IStandaloneCodeEditor,
    monacoInstance: Monaco,
  ) => {
    editorRef.current = editor;
    monacoRef.current = monacoInstance;
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">
          YAML Configuration
          {isSchemaLoaded && (
            <span className="ml-2 text-xs text-muted-foreground font-normal">
              (Schema validation enabled)
            </span>
          )}
        </CardTitle>
        <CardDescription>
          Edit your OpenTelemetry collector configuration in YAML format with
          autocomplete and validation
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="border rounded-lg overflow-hidden">
          <Editor
            height="60vh"
            defaultLanguage="yaml"
            value={value}
            onChange={(value) => onChange(value || "")}
            theme="vs-light"
            onMount={handleEditorDidMount}
            options={{
              minimap: { enabled: false },
              fontSize: 13,
              lineNumbers: "on",
              roundedSelection: false,
              scrollBeyondLastLine: false,
              readOnly: false,
              automaticLayout: true,
              quickSuggestions: {
                other: true,
                comments: false,
                strings: true,
              },
              suggestOnTriggerCharacters: true,
              acceptSuggestionOnEnter: "on",
              tabCompletion: "on",
              wordBasedSuggestions: "off",
              parameterHints: {
                enabled: true,
              },
            }}
          />
        </div>
      </CardContent>
    </Card>
  );
}

// Helper function to build a comprehensive OTel collector config schema
async function buildOTelCollectorSchema(
  components: Array<{ type: ComponentType; name: string }>,
): Promise<JSONSchema> {
  // Group components by type
  const componentsByType: Record<string, string[]> = {
    receivers: [],
    processors: [],
    exporters: [],
    connectors: [],
    extensions: [],
  };

  // Map component type to plural form
  const typeToPluralMap: Record<ComponentType, keyof typeof componentsByType> =
    {
      receiver: "receivers",
      processor: "processors",
      exporter: "exporters",
      connector: "connectors",
      extension: "extensions",
    };

  // Organize components
  for (const component of components) {
    const pluralType = typeToPluralMap[component.type];
    if (pluralType) {
      componentsByType[pluralType].push(component.name);
    }
  }

  // Fetch a few common component schemas for better autocomplete
  const commonComponents = [
    { type: "receiver" as ComponentType, name: "otlp" },
    { type: "processor" as ComponentType, name: "batch" },
    { type: "exporter" as ComponentType, name: "otlp" },
    { type: "exporter" as ComponentType, name: "logging" },
  ];

  const componentSchemas: Record<string, JSONSchema> = {};

  await Promise.all(
    commonComponents.map(async ({ type, name }) => {
      try {
        const schema = await getComponentSchema(type, name);
        const pluralType = typeToPluralMap[type];
        const key = `${pluralType}.${name}`;
        componentSchemas[key] = schema;
      } catch (error) {
        console.warn(`Failed to fetch schema for ${type}/${name}:`, error);
      }
    }),
  );

  // Build the root schema for OTel collector config
  const schema: JSONSchema = {
    $schema: "http://json-schema.org/draft-07/schema#",
    type: "object",
    description: "OpenTelemetry Collector Configuration",
    properties: {
      receivers: {
        type: "object",
        description: "Receivers collect telemetry data from sources",
        additionalProperties: true,
        properties: buildComponentProperties(
          componentsByType.receivers,
          componentSchemas,
          "receivers",
        ),
      },
      processors: {
        type: "object",
        description: "Processors transform telemetry data",
        additionalProperties: true,
        properties: buildComponentProperties(
          componentsByType.processors,
          componentSchemas,
          "processors",
        ),
      },
      exporters: {
        type: "object",
        description: "Exporters send telemetry data to backends",
        additionalProperties: true,
        properties: buildComponentProperties(
          componentsByType.exporters,
          componentSchemas,
          "exporters",
        ),
      },
      connectors: {
        type: "object",
        description: "Connectors connect pipelines",
        additionalProperties: true,
        properties: buildComponentProperties(
          componentsByType.connectors,
          componentSchemas,
          "connectors",
        ),
      },
      extensions: {
        type: "object",
        description: "Extensions provide additional capabilities",
        additionalProperties: true,
        properties: buildComponentProperties(
          componentsByType.extensions,
          componentSchemas,
          "extensions",
        ),
      },
      service: {
        type: "object",
        description:
          "Service configuration defines pipelines and telemetry settings",
        properties: {
          extensions: {
            type: "array",
            description: "Extensions to enable",
            items: {
              type: "string",
              enum: componentsByType.extensions,
            },
          },
          pipelines: {
            type: "object",
            description: "Telemetry pipelines configuration",
            additionalProperties: {
              type: "object",
              properties: {
                receivers: {
                  type: "array",
                  description: "List of receivers for this pipeline",
                  items: {
                    type: "string",
                  },
                },
                processors: {
                  type: "array",
                  description: "List of processors for this pipeline",
                  items: {
                    type: "string",
                  },
                },
                exporters: {
                  type: "array",
                  description: "List of exporters for this pipeline",
                  items: {
                    type: "string",
                  },
                },
              },
              required: ["receivers", "exporters"],
            },
          },
          telemetry: {
            type: "object",
            description: "Collector's own telemetry configuration",
            properties: {
              logs: {
                type: "object",
                properties: {
                  level: {
                    type: "string",
                    enum: ["debug", "info", "warn", "error"],
                    default: "info",
                  },
                  encoding: {
                    type: "string",
                    enum: ["console", "json"],
                    default: "console",
                  },
                },
              },
              metrics: {
                type: "object",
                properties: {
                  level: {
                    type: "string",
                    enum: ["none", "basic", "normal", "detailed"],
                    default: "basic",
                  },
                  address: {
                    type: "string",
                    description: "Address to expose metrics on",
                  },
                },
              },
            },
          },
        },
      },
    },
    required: ["receivers", "exporters", "service"],
  };

  return schema;
}

function buildComponentProperties(
  componentNames: string[],
  componentSchemas: Record<string, JSONSchema>,
  componentType: string,
): Record<string, JSONSchema> {
  const properties: Record<string, JSONSchema> = {};

  for (const name of componentNames) {
    const schemaKey = `${componentType}.${name}`;

    // If we have the actual schema, use it; otherwise provide a basic schema
    if (componentSchemas[schemaKey]) {
      properties[name] = componentSchemas[schemaKey];
    } else {
      properties[name] = {
        type: "object",
        description: `Configuration for ${name} ${componentType}`,
        additionalProperties: true,
      };
    }
  }

  return properties;
}
