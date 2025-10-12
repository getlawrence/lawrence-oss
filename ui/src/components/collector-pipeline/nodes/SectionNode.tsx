import { COLOR_SCHEME } from "../canvas-constants";
import type { PipelineType } from "../canvas-types";

import { Badge } from "@/components/ui/badge";

interface SectionNodeProps {
  data: {
    type: PipelineType;
    width?: number;
    height?: number;
    label?: string;
    metrics?: {
      received: number;
      errors: number;
    };
  };
}

export const SectionNode = ({ data }: SectionNodeProps) => {
  const colorScheme = COLOR_SCHEME[data.type] || COLOR_SCHEME.traces;
  const baseColor = colorScheme.color;

  // Use explicit Tailwind classes based on the color
  const colors = (() => {
    switch (baseColor) {
      case "blue":
        return {
          bg: "bg-blue-50/50 dark:bg-blue-500/[0.02]",
          border: "border-blue-200 dark:border-blue-800/40",
          badge:
            "bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 border-blue-300 dark:border-blue-700",
        };
      case "green":
        return {
          bg: "bg-green-50/50 dark:bg-green-500/[0.02]",
          border: "border-green-200 dark:border-green-800/40",
          badge:
            "bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-300 border-green-300 dark:border-green-700",
        };
      case "purple":
        return {
          bg: "bg-purple-50/50 dark:bg-purple-500/[0.02]",
          border: "border-purple-200 dark:border-purple-800/40",
          badge:
            "bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-300 border-purple-300 dark:border-purple-700",
        };
      default:
        return {
          bg: "bg-gray-50/50 dark:bg-gray-500/[0.02]",
          border: "border-gray-200 dark:border-gray-800/40",
          badge:
            "bg-gray-100 dark:bg-gray-900/30 text-gray-700 dark:text-gray-300 border-gray-300 dark:border-gray-700",
        };
    }
  })();

  return (
    <div
      className={`${colors.bg} rounded-lg border-2 ${colors.border} shadow-sm transition-all duration-300 overflow-visible pointer-events-none`}
      style={{
        width: `${data.width || 900}px`,
        height: `${data.height || 400}px`,
        zIndex: 0,
      }}
    >
      {/* Section header */}
      <div
        className={`h-10 px-4 flex items-center justify-between border-b ${colors.border} bg-white dark:bg-background rounded-t-lg`}
      >
        <div className="flex items-center gap-2">
          <Badge variant="outline" className={`${colors.badge} shadow-sm`}>
            {data.label || colorScheme.label}
          </Badge>
          {data.metrics && (
            <div className="flex items-center gap-3 text-xs text-muted-foreground">
              <span>Received: {data.metrics.received.toLocaleString()}</span>
              {data.metrics.errors > 0 && (
                <span className="text-red-600 dark:text-red-400">
                  Errors: {data.metrics.errors.toLocaleString()}
                </span>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default SectionNode;
