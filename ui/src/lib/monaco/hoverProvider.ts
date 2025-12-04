import type {
  languages,
  editor,
  IMarkdownString,
  IRange,
  Position,
} from "monaco-editor";

import { SchemaCache } from "./schemaUtils";

import type { JSONSchema } from "@/api/schemas";
import { parseYamlContext } from "@/utils/yamlContext";

/**
 * Register YAML hover provider for OTel configuration schema documentation
 */
export function registerYamlHoverProvider(
  monaco: typeof import("monaco-editor"),
  schemaCache: SchemaCache,
): { dispose: () => void } {
  return monaco.languages.registerHoverProvider("yaml", {
    provideHover: async (
      model: editor.ITextModel,
      position: Position,
    ): Promise<languages.Hover | null> => {
      const lines = model.getValue().split("\n");
      const currentLine = lines[position.lineNumber - 1];

      // Extract the property name from the full line
      const propertyMatch = currentLine.match(/^\s*([a-zA-Z_][a-zA-Z0-9_-]*):/);

      if (!propertyMatch) return null;

      const propertyName = propertyMatch[1];
      const propertyStartCol =
        propertyMatch.index! + propertyMatch[0].indexOf(propertyName) + 1;
      const propertyEndCol = propertyStartCol + propertyName.length;

      // Check if cursor is actually over the property name
      if (
        position.column < propertyStartCol ||
        position.column > propertyEndCol
      ) {
        return null;
      }

      // Parse context to understand where we are
      const context = parseYamlContext(lines, position.lineNumber - 1);

      // Only provide hover for component properties
      if (!context.section || !context.component) {
        return null;
      }

      // Fetch the component schema
      const componentSchema = await schemaCache.getComponentSchema(
        context.section,
        context.component,
      );

      if (!componentSchema?.properties) {
        return null;
      }

      // Navigate to the correct schema level based on path
      let currentSchema = componentSchema;
      for (const pathPart of context.path) {
        if (currentSchema.properties?.[pathPart]) {
          currentSchema = currentSchema.properties[pathPart] as JSONSchema;
        }
      }

      // Get the property schema
      const propSchema = currentSchema.properties?.[propertyName] as
        | JSONSchema
        | undefined;

      if (!propSchema) {
        return null;
      }

      // Build hover content
      const contents: IMarkdownString[] = [];

      // Add property name and type
      const schemaType = Array.isArray(propSchema.type)
        ? propSchema.type.join(" | ")
        : propSchema.type || "any";

      let header = `**${propertyName}**`;
      if (currentSchema.required?.includes(propertyName)) {
        header += " *(required)*";
      }
      header += `\n\n*Type:* \`${schemaType}\``;

      contents.push({ value: header });

      // Add description
      if (propSchema.description) {
        contents.push({ value: propSchema.description });
      }

      // Add default value
      if (propSchema.default !== undefined) {
        contents.push({
          value: `*Default:* \`${JSON.stringify(propSchema.default)}\``,
        });
      }

      // Add enum values
      if (propSchema.enum && propSchema.enum.length > 0) {
        const enumValues = propSchema.enum.map((v) => `\`${v}\``).join(", ");
        contents.push({
          value: `*Allowed values:* ${enumValues}`,
        });
      }

      // Add pattern for strings
      if (propSchema.pattern) {
        contents.push({
          value: `*Pattern:* \`${propSchema.pattern}\``,
        });
      }

      // Add constraints
      const constraints: string[] = [];
      if (typeof propSchema.minimum === "number") {
        constraints.push(`minimum: ${propSchema.minimum}`);
      }
      if (typeof propSchema.maximum === "number") {
        constraints.push(`maximum: ${propSchema.maximum}`);
      }
      if (typeof propSchema.minLength === "number") {
        constraints.push(`min length: ${propSchema.minLength}`);
      }
      if (typeof propSchema.maxLength === "number") {
        constraints.push(`max length: ${propSchema.maxLength}`);
      }
      if (constraints.length > 0) {
        contents.push({
          value: `*Constraints:* ${constraints.join(", ")}`,
        });
      }

      const range: IRange = {
        startLineNumber: position.lineNumber,
        startColumn: propertyStartCol,
        endLineNumber: position.lineNumber,
        endColumn: propertyEndCol,
      };

      return {
        contents,
        range,
      };
    },
  });
}
