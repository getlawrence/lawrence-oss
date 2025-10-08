# Config YAML Editor with Metrics

## Overview

The enhanced YAML editor for OpenTelemetry Collector configurations now supports **inline metrics overlay**. This feature displays real-time performance metrics directly on the YAML configuration, making it easy to monitor component health and performance without switching views.

## Components

### `ConfigYamlEditor`

The standard YAML editor without metrics (backward compatible).

**Usage:**
```tsx
import { ConfigYamlEditor } from "@/components/configs";

<ConfigYamlEditor
  value={yamlContent}
  onChange={setYamlContent}
/>
```

### `ConfigYamlEditorWithMetrics`

Enhanced editor with optional metrics overlay support.

**Usage:**
```tsx
import { ConfigYamlEditorWithMetrics } from "@/components/configs";
import { getPipelineMetrics } from "@/api/collector-pipeline";

// Fetch metrics for an agent
const { data: metricsData } = useSWR(
  `agent-metrics-${agentId}`,
  () => getPipelineMetrics(agentId, "5m")
);

// Use the editor with metrics
<ConfigYamlEditorWithMetrics
  value={yamlContent}
  onChange={setYamlContent}
  metrics={metricsData?.components}
  readonly={false}
/>
```

**Props:**
- `value: string` - YAML configuration content
- `onChange: (value: string) => void` - Callback when content changes
- `metrics?: ComponentMetrics[]` - Optional array of component metrics
- `readonly?: boolean` - Whether the editor is read-only (default: false)

## Features

### 1. Inline Metrics Display

Metrics are displayed inline next to component definitions:

```yaml
receivers:
  otlp:                              # 📊 1.2k/s | ✅ healthy
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317

processors:
  batch:                             # 📊 1.2k/s | ✅ healthy
    timeout: 10s

exporters:
  otlp:                              # 📊 1.1k/s | ⚠️ 2.3% errors
    endpoint: localhost:4317
```

### 2. Hover Tooltips

Hovering over a component shows detailed metrics:
- Throughput per pipeline (traces, metrics, logs)
- Error rates
- Received/Accepted/Sent/Dropped counts
- Send failures
- Aggregated totals across all pipelines

### 3. Visual Health Indicators

- **✅ Green**: No errors (healthy)
- **⚠️ Yellow**: < 1% error rate (minor issues)
- **❌ Red**: >= 5% error rate (critical issues)

### 4. Line Highlighting

Lines with components experiencing errors are highlighted with a subtle background color to draw attention.

### 5. Auto-refresh

When integrated with SWR, metrics automatically refresh at your configured interval to show real-time data.

## Implementation Example

See `AgentConfig.tsx` for a complete implementation example that:
1. Fetches an agent's configuration
2. Fetches pipeline metrics with configurable time range
3. Displays the config with live metrics
4. Auto-refreshes metrics every 5 seconds
5. Handles cases where metrics are not available

## YAML Parser Utilities

The `yaml-parser.ts` utility provides helper functions:

- `parseYamlComponents(yamlContent)` - Extract component definitions with line numbers
- `formatThroughput(value)` - Format throughput for display (e.g., "1.2K/s")
- `formatErrorRate(value)` - Format error rates (e.g., "2.3%")
- `getStatusIcon(errorRate)` - Get status emoji based on error rate

## Backward Compatibility

The original `ConfigYamlEditor` component remains unchanged and continues to work as before. The new metrics feature is completely optional and degrades gracefully when metrics are not provided.

## Technical Details

### Monaco Editor Decorations

The component uses Monaco Editor's decoration API to add:
- **After decorations** - Inline text showing metrics
- **Glyph margin decorations** - Visual indicators in the left margin
- **Line decorations** - Background highlighting for errors
- **Hover providers** - Detailed metrics on hover

### YAML Parsing

A simple regex-based parser extracts component definitions by:
1. Detecting section headers (receivers, processors, exporters)
2. Finding component names at the appropriate indentation level
3. Recording line numbers for each component
4. Matching with metrics data by component name and type

### Metrics Aggregation

When multiple pipeline types (traces, metrics, logs) are configured for a single component, the metrics are aggregated to show:
- Total throughput across all pipelines
- Overall error rate
- Individual pipeline metrics in hover tooltips

## Future Enhancements

Potential improvements:
- Interactive charts in hover tooltips
- Historical metrics visualization
- Clickable metrics to drill down into details
- Warning annotations for performance issues
- Suggested optimizations based on metrics

