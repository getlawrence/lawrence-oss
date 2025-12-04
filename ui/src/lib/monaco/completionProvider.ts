import type { languages, editor, Position } from "monaco-editor";

import {
  getPropInsertText,
  getPropertyDetail,
  SchemaCache,
} from "./schemaUtils";

import type { JSONSchema } from "@/api/schemas";
import { parseYamlContext } from "@/utils/yamlContext";

/**
 * Register YAML completion provider for OTel configuration
 */
export function registerYamlCompletionProvider(
  monaco: typeof import("monaco-editor"),
  schema: JSONSchema,
  schemaCache: SchemaCache,
): { dispose: () => void } {
  return monaco.languages.registerCompletionItemProvider("yaml", {
    triggerCharacters: [" ", ":", "\n", "-"],
    provideCompletionItems: async (
      model: editor.ITextModel,
      position: Position,
    ): Promise<languages.CompletionList> => {
      const textUntilPosition = model.getValueInRange({
        startLineNumber: position.lineNumber,
        startColumn: 1,
        endLineNumber: position.lineNumber,
        endColumn: position.column,
      });

      const suggestions: languages.CompletionItem[] = [];
      const word = model.getWordUntilPosition(position);
      const range = {
        startLineNumber: position.lineNumber,
        startColumn: word.startColumn,
        endLineNumber: position.lineNumber,
        endColumn: word.endColumn,
      };

      // Parse the context to understand where we are
      const lines = model.getValue().split("\n");
      const context = parseYamlContext(lines, position.lineNumber - 1);

      // Top-level property suggestions
      if (
        context.depth === 0 ||
        (textUntilPosition.trim() === "" && context.depth === 0)
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
              monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
            documentation:
              (schema.properties?.[prop]?.description as string) ||
              `${prop} configuration`,
            sortText: `0${index}`,
            range,
          });
        });
      }
      // Component-level suggestions (e.g., under receivers:)
      else if (context.section && !context.component && context.depth === 1) {
        const componentProps = schema.properties?.[context.section]
          ?.properties as Record<string, JSONSchema> | undefined;
        if (componentProps) {
          Object.keys(componentProps).forEach((componentName) => {
            const componentSchema = componentProps[componentName];
            suggestions.push({
              label: componentName,
              kind: monaco.languages.CompletionItemKind.Class,
              insertText: `${componentName}:\n  `,
              insertTextRules:
                monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
              documentation:
                (componentSchema?.description as string) ||
                `${componentName} component`,
              sortText: `1${componentName}`,
              range,
            });
          });
        }
      }
      // Component property suggestions (e.g., under receivers: otlp:)
      else if (context.section && context.component) {
        // Fetch the component schema on demand
        const componentSchema = await schemaCache.getComponentSchema(
          context.section,
          context.component,
        );

        if (componentSchema?.properties) {
          // Get properties for the current nesting level
          let currentSchema = componentSchema;

          // Navigate to the correct schema level based on the path
          for (const pathPart of context.path) {
            if (currentSchema.properties?.[pathPart]) {
              currentSchema = currentSchema.properties[pathPart] as JSONSchema;
            }
          }

          // Suggest properties from the current schema level
          if (currentSchema.properties) {
            Object.entries(currentSchema.properties).forEach(
              ([propName, propSchema]) => {
                const prop = propSchema as JSONSchema;
                const isRequired = currentSchema.required?.includes(propName);

                suggestions.push({
                  label: propName + (isRequired ? " *" : ""),
                  kind: monaco.languages.CompletionItemKind.Property,
                  insertText: getPropInsertText(propName, prop),
                  insertTextRules:
                    monaco.languages.CompletionItemInsertTextRule
                      .InsertAsSnippet,
                  documentation: (prop.description as string) || propName,
                  detail: getPropertyDetail(prop),
                  sortText: isRequired ? `0${propName}` : `1${propName}`,
                  range,
                });
              },
            );
          }

          // Handle enums
          if (currentSchema.enum) {
            currentSchema.enum.forEach((enumValue, index) => {
              suggestions.push({
                label: String(enumValue),
                kind: monaco.languages.CompletionItemKind.EnumMember,
                insertText: String(enumValue),
                documentation: `Allowed value`,
                sortText: `0${index}`,
                range,
              });
            });
          }
        }
      }

      return { suggestions };
    },
  });
}
