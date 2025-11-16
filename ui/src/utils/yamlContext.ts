/**
 * Represents the context of a YAML cursor position
 */
export interface YamlContext {
  section: string | null; // e.g., "receivers", "processors"
  component: string | null; // e.g., "otlp", "batch"
  path: string[]; // nested path within component, e.g., ["protocols", "grpc"]
  depth: number; // indentation depth
}

/**
 * Parse YAML context to determine the cursor's position in the hierarchy
 *
 * @param lines - Array of YAML lines
 * @param lineNumber - Zero-based line number of cursor position
 * @returns YamlContext object with section, component, path, and depth
 *
 * @example
 * ```yaml
 * processors:          # line 0
 *   batch:             # line 1
 *     timeout: 10s     # line 2 (cursor here)
 * ```
 * Returns: {section: "processors", component: "batch", path: [], depth: 2}
 */
export function parseYamlContext(
  lines: string[],
  lineNumber: number,
): YamlContext {
  const context: YamlContext = {
    section: null,
    component: null,
    path: [],
    depth: 0,
  };

  // Calculate current line's indentation
  const currentLine = lines[lineNumber] || "";
  const currentIndent = currentLine.match(/^\s*/)?.[0].length || 0;
  context.depth = Math.floor(currentIndent / 2);

  // Build the hierarchy by finding all parent keys
  const hierarchy: Array<{ indent: number; key: string }> = [];

  // Parse backwards to find all parent keys
  for (let i = lineNumber - 1; i >= 0; i--) {
    const line = lines[i];
    const indent = line.match(/^\s*/)?.[0].length || 0;
    const keyMatch = line.match(/^\s*([a-zA-Z_][a-zA-Z0-9_-]*):/);

    if (!keyMatch) continue;

    const key = keyMatch[1];

    // Only consider lines with less indentation (these are parents)
    if (indent < currentIndent) {
      // Add this parent to hierarchy
      hierarchy.push({ indent, key });

      // If we found the top-level section, we're done
      if (
        indent === 0 &&
        [
          "receivers",
          "processors",
          "exporters",
          "connectors",
          "extensions",
          "service",
        ].includes(key)
      ) {
        break;
      }
    }
  }

  // Reverse hierarchy so it goes from root to current position
  hierarchy.reverse();

  // Parse the hierarchy to extract section, component, and path
  for (let i = 0; i < hierarchy.length; i++) {
    const item = hierarchy[i];

    if (
      item.indent === 0 &&
      [
        "receivers",
        "processors",
        "exporters",
        "connectors",
        "extensions",
        "service",
      ].includes(item.key)
    ) {
      context.section = item.key;
    } else if (context.section && item.indent === 2 && !context.component) {
      // Component is at indent 2 (one level under section)
      context.component = item.key;
    } else if (context.component && item.indent > 2) {
      // Everything deeper is part of the path
      context.path.push(item.key);
    }
  }

  return context;
}
