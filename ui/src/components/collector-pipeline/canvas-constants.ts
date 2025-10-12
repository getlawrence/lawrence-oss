import type { SectionType } from "./canvas-types";

export const LAYOUT_CONFIG = {
  // Layout strategy - vertical stacking for pipeline flow visualization
  LAYOUT_STRATEGY: "vertical" as const,

  // Section dimensions - optimized for vertical pipeline flow
  SECTION_HEIGHT: 500,
  SECTION_PADDING: 20,
  SECTION_MIN_WIDTH: 400,
  SECTION_MIN_HEIGHT: 400,
  SECTION_GAP: 16,

  // Pipeline-specific dimensions
  PIPELINE_MIN_WIDTH: 500,
  PIPELINE_COMPONENT_WIDTH: 180,
  PIPELINE_COMPONENT_SPACING: 100,

  // Node dimensions
  NODE_WIDTH: 150,
  NODE_HEIGHT: 40,
  NODE_SPACING: 150,

  // Layout spacing
  SIDEBAR_WIDTH: 80,
  HEADER_HEIGHT: 64,
  MARGIN: 16,

  // Z-index management
  MINIMAP_HEIGHT: 120,
  NODE_Z_INDEX: 20,
  SECTION_Z_INDEX: 0,
};

// Unified color scheme for all components
export const COLOR_SCHEME = {
  traces: {
    color: "blue",
    label: "Traces",
    background: "rgba(59, 130, 246, 0.05)",
  },
  metrics: {
    color: "green",
    label: "Metrics",
    background: "rgba(34, 197, 94, 0.05)",
  },
  logs: {
    color: "purple",
    label: "Logs",
    background: "rgba(168, 85, 247, 0.05)",
  },
  connectors: {
    color: "amber",
    label: "Connectors",
    background: "rgba(217, 119, 6, 0.05)",
  },
  exporters: {
    color: "purple",
    label: "Exporters",
    background: "rgba(168, 85, 247, 0.05)",
  },
  processors: {
    color: "green",
    label: "Processors",
    background: "rgba(34, 197, 94, 0.05)",
  },
  receivers: {
    color: "blue",
    label: "Receivers",
    background: "rgba(59, 130, 246, 0.05)",
  },
};

export const PIPELINE_SECTIONS: Record<
  SectionType,
  {
    label: string;
    background: string;
  }
> = {
  traces: COLOR_SCHEME.traces,
  metrics: COLOR_SCHEME.metrics,
  logs: COLOR_SCHEME.logs,
};

export const styles = {
  handleStyle: {
    width: "14px",
    height: "14px",
    background: "#555",
    border: "2px solid #fff",
    borderRadius: "7px",
    zIndex: 20,
  },

  validConnectionStyle: {
    stroke: "#ff9800",
    strokeWidth: 3,
    animated: true,
    zIndex: 1000,
    type: "smoothstep",
  },

  node: {
    base: {
      padding: "10px",
      borderRadius: "4px",
      border: "1px solid #ddd",
      background: "white",
      fontSize: "12px",
      transition: "all 0.2s ease",
    },
    receiver: {
      background: "#e3f2fd",
      borderColor: "#90caf9",
    },
    processor: {
      background: "#f1f8e9",
      borderColor: "#aed581",
    },
    exporter: {
      background: "#f3e5f5",
      borderColor: "#ce93d8",
    },
  },
};
