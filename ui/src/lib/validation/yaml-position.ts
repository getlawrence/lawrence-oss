// Utility functions to find line/column positions in YAML content

import type { ValidationError } from "./types";

/**
 * Find the line and column number for a specific path in YAML content
 * @param yamlContent The raw YAML string
 * @param path Path array to the field (e.g., ["service", "pipelines", "traces", "receivers"])
 * @param componentName Optional component name to search for in arrays
 * @returns Line and column numbers (1-indexed) or null if not found
 */
export function findYamlPosition(
  yamlContent: string,
  path: string[],
  componentName?: string,
): {
  line: number;
  column: number;
  endLine?: number;
  endColumn?: number;
} | null {
  const lines = yamlContent.split("\n");
  let currentIndent = 0;
  let pathIndex = 0;

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const trimmedLine = line.trim();

    // Skip empty lines and comments
    if (!trimmedLine || trimmedLine.startsWith("#")) {
      continue;
    }

    // Calculate indentation
    const indent = line.length - line.trimStart().length;

    // Check if we're looking for a key at this level
    if (pathIndex < path.length) {
      const searchKey = path[pathIndex];

      // Match key: or key (for object keys)
      const keyMatch = trimmedLine.match(/^([^:#[\]]+):/);
      if (keyMatch && keyMatch[1].trim() === searchKey) {
        pathIndex++;
        currentIndent = indent;

        // If this is the last element in path, this is our target
        if (pathIndex === path.length) {
          if (componentName) {
            // Search for the component name in the following lines (in an array)
            for (let j = i + 1; j < lines.length; j++) {
              const nextLine = lines[j];
              const nextTrimmed = nextLine.trim();
              const nextIndent = nextLine.length - nextLine.trimStart().length;

              // Stop if we've gone back to a lower indent level
              if (
                nextIndent <= currentIndent &&
                nextTrimmed &&
                !nextTrimmed.startsWith("#")
              ) {
                break;
              }

              // Check for array item containing our component
              if (nextTrimmed.startsWith("-") || nextTrimmed.startsWith("[")) {
                const componentRegex = new RegExp(`\\b${componentName}\\b`);
                if (componentRegex.test(nextTrimmed)) {
                  const column = nextLine.indexOf(componentName) + 1;
                  return {
                    line: j + 1,
                    column,
                    endLine: j + 1,
                    endColumn: column + componentName.length,
                  };
                }
              }
            }
          }

          // Return the key's position
          const column = line.indexOf(searchKey) + 1;
          return {
            line: i + 1,
            column,
            endLine: i + 1,
            endColumn: column + searchKey.length,
          };
        }
        continue;
      }
    }

    // Reset if we've moved to a lower indentation level
    if (indent < currentIndent && trimmedLine) {
      // We've moved out of the current context, reset to find the path again
      pathIndex = 0;
      currentIndent = 0;
    }
  }

  return null;
}

/**
 * Create a validation error with position information
 */
export function createValidationError(
  message: string,
  yamlContent: string,
  path: string[],
  componentName?: string,
  severity: "error" | "warning" = "error",
): ValidationError {
  const position = findYamlPosition(yamlContent, path, componentName);

  if (position) {
    return {
      message,
      severity,
      ...position,
      path,
    };
  }

  // Fallback to line 1 if position not found
  return {
    message,
    severity,
    line: 1,
    column: 1,
    path,
  };
}
