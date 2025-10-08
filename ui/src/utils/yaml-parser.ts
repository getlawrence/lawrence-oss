/**
 * Utility functions for parsing YAML configurations and extracting component information
 */

export interface YamlComponent {
  name: string;
  type: "receiver" | "processor" | "exporter";
  lineNumber: number;
}

/**
 * Parse YAML content to extract component definitions with their line numbers
 * This is a simple parser that looks for common OpenTelemetry Collector patterns
 */
export function parseYamlComponents(yamlContent: string): YamlComponent[] {
  const components: YamlComponent[] = [];
  const lines = yamlContent.split("\n");

  let currentSection: "receiver" | "processor" | "exporter" | null = null;
  let sectionIndent = -1;

  lines.forEach((line, index) => {
    const trimmed = line.trim();
    const indent = line.search(/\S/);

    // Detect section headers
    if (trimmed === "receivers:") {
      currentSection = "receiver";
      sectionIndent = indent;
      return;
    } else if (trimmed === "processors:") {
      currentSection = "processor";
      sectionIndent = indent;
      return;
    } else if (trimmed === "exporters:") {
      currentSection = "exporter";
      sectionIndent = indent;
      return;
    } else if (
      trimmed === "service:" ||
      trimmed === "extensions:" ||
      trimmed === "connectors:"
    ) {
      currentSection = null;
      return;
    }

    // Extract component names (first-level items under sections)
    if (currentSection && indent > sectionIndent && indent !== -1) {
      // Check if this is a component definition (ends with : and is not deeply nested)
      if (trimmed.endsWith(":") && !trimmed.startsWith("-")) {
        const componentName = trimmed.slice(0, -1).trim();

        // Skip if it looks like a sub-property (too indented or contains special chars)
        const expectedIndent = sectionIndent + 2;
        if (
          Math.abs(indent - expectedIndent) <= 2 &&
          componentName &&
          !componentName.includes(" ")
        ) {
          components.push({
            name: componentName,
            type: currentSection,
            lineNumber: index + 1, // Monaco uses 1-based line numbers
          });
        }
      }
    }
  });

  return components;
}

/**
 * Format throughput for display
 */
export function formatThroughput(throughput: number): string {
  if (throughput >= 1000000) {
    return `${(throughput / 1000000).toFixed(1)}M/s`;
  } else if (throughput >= 1000) {
    return `${(throughput / 1000).toFixed(1)}K/s`;
  } else if (throughput >= 1) {
    return `${throughput.toFixed(0)}/s`;
  } else {
    return `${throughput.toFixed(2)}/s`;
  }
}

/**
 * Format error rate for display
 */
export function formatErrorRate(errorRate: number): string {
  if (errorRate >= 10) {
    return `${errorRate.toFixed(1)}%`;
  } else if (errorRate >= 0.1) {
    return `${errorRate.toFixed(2)}%`;
  } else if (errorRate > 0) {
    return `${errorRate.toFixed(3)}%`;
  }
  return "0%";
}

/**
 * Get status icon based on error rate
 */
export function getStatusIcon(errorRate: number): string {
  if (errorRate === 0) {
    return "✅";
  } else if (errorRate < 1) {
    return "⚠️";
  } else if (errorRate < 5) {
    return "⚠️";
  }
  return "❌";
}

/**
 * Get status color based on error rate
 */
export function getStatusColor(errorRate: number): string {
  if (errorRate === 0) {
    return "#22c55e"; // green
  } else if (errorRate < 1) {
    return "#eab308"; // yellow
  } else if (errorRate < 5) {
    return "#f97316"; // orange
  }
  return "#ef4444"; // red
}
