import { getComponentSchema, type ComponentType } from "@/api/schemas";
import type { JSONSchema } from "@/api/schemas";

/**
 * Build a comprehensive OTel collector config schema
 */
export async function buildOTelCollectorSchema(
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
