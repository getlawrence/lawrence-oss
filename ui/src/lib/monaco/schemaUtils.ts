import { getComponentSchema, type ComponentType } from "@/api/schemas";
import type { JSONSchema } from "@/api/schemas";

/**
 * Schema cache for component schemas
 */
export class SchemaCache {
  private cache: Record<string, JSONSchema> = {};

  /**
   * Fetch component schema on demand with caching
   */
  async getComponentSchema(
    section: string,
    componentName: string,
  ): Promise<JSONSchema | null> {
    const cacheKey = `${section}.${componentName}`;

    // Check cache first
    if (this.cache[cacheKey]) {
      return this.cache[cacheKey];
    }

    // Map section to component type
    const sectionToTypeMap: Record<string, ComponentType> = {
      receivers: "receiver",
      processors: "processor",
      exporters: "exporter",
      connectors: "connector",
      extensions: "extension",
    };

    const componentType = sectionToTypeMap[section];
    if (!componentType) return null;

    try {
      const componentSchema = await getComponentSchema(
        componentType,
        componentName,
      );
      this.cache[cacheKey] = componentSchema;
      return componentSchema;
    } catch (error) {
      console.warn(
        `Failed to fetch schema for ${section}/${componentName}:`,
        error,
      );
      return null;
    }
  }

  /**
   * Clear the cache
   */
  clear(): void {
    this.cache = {};
  }
}

/**
 * Generate appropriate insert text based on property type
 */
export function getPropInsertText(propName: string, prop: JSONSchema): string {
  const schemaType = Array.isArray(prop.type) ? prop.type[0] : prop.type;

  switch (schemaType) {
    case "object":
      return `${propName}:\n  `;
    case "array":
      return `${propName}:\n  - `;
    case "boolean":
      return `${propName}: false`;
    case "number":
    case "integer":
      return `${propName}: 0`;
    case "string":
      if (prop.enum && prop.enum.length > 0) {
        return `${propName}: ${prop.enum[0]}`;
      }
      return `${propName}: ""`;
    default:
      return `${propName}: `;
  }
}

/**
 * Get property detail string for autocomplete
 */
export function getPropertyDetail(prop: JSONSchema): string {
  const schemaType = Array.isArray(prop.type)
    ? prop.type.join(" | ")
    : prop.type;

  if (prop.enum) {
    return `${schemaType} (enum: ${prop.enum.slice(0, 3).join(", ")}${prop.enum.length > 3 ? "..." : ""})`;
  }

  if (prop.default !== undefined) {
    return `${schemaType} (default: ${JSON.stringify(prop.default)})`;
  }

  return schemaType || "any";
}
