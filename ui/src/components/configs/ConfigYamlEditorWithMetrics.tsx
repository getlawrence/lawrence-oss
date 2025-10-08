import Editor, { type OnMount } from "@monaco-editor/react";
import * as monaco from "monaco-editor";
import { useEffect, useRef, useState } from "react";

import { type ComponentMetrics } from "@/api/collector-pipeline";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  parseYamlComponents,
  formatThroughput,
  formatErrorRate,
  getStatusIcon,
  type YamlComponent,
} from "@/utils/yaml-parser";

interface ConfigYamlEditorWithMetricsProps {
  value: string;
  onChange: (value: string) => void;
  metrics?: ComponentMetrics[]; // Optional metrics data
  readonly?: boolean;
}

export function ConfigYamlEditorWithMetrics({
  value,
  onChange,
  metrics,
  readonly = false,
}: ConfigYamlEditorWithMetricsProps) {
  const editorRef = useRef<monaco.editor.IStandaloneCodeEditor | null>(null);
  const [decorations, setDecorations] = useState<string[]>([]);
  const [parsedComponents, setParsedComponents] = useState<YamlComponent[]>([]);

  // Parse YAML to find components whenever value changes
  useEffect(() => {
    if (value) {
      const components = parseYamlComponents(value);
      setParsedComponents(components);
    }
  }, [value]);

  // Update decorations when metrics or parsed components change
  useEffect(() => {
    if (!editorRef.current || !metrics || metrics.length === 0) {
      return;
    }

    const editor = editorRef.current;
    const newDecorations: monaco.editor.IModelDeltaDecoration[] = [];

    // Match parsed components with metrics
    parsedComponents.forEach((component) => {
      // Find matching metrics for this component
      const componentMetrics = metrics.filter(
        (m) =>
          m.component_name === component.name &&
          mapComponentType(m.component_type) === component.type,
      );

      if (componentMetrics.length === 0) {
        return;
      }

      // Aggregate metrics across all pipeline types (traces, metrics, logs)
      const aggregated = aggregateMetrics(componentMetrics);

      // Create inline decoration with metrics
      const metricsText = formatMetricsText(aggregated);

      newDecorations.push({
        range: new monaco.Range(
          component.lineNumber,
          1,
          component.lineNumber,
          1,
        ),
        options: {
          after: {
            content: `  ${metricsText}`,
            inlineClassName: "metrics-decoration",
            inlineClassNameAffectsLetterSpacing: true,
          },
          afterContentClassName: "metrics-decoration-after",
          isWholeLine: false,
        },
      });

      // Add gutter decoration for visual status
      newDecorations.push({
        range: new monaco.Range(
          component.lineNumber,
          1,
          component.lineNumber,
          1,
        ),
        options: {
          isWholeLine: false,
          linesDecorationsClassName: "metrics-gutter-decoration",
          glyphMarginClassName: "metrics-glyph-margin",
          glyphMarginHoverMessage: {
            value: formatHoverMessage(component.name, componentMetrics),
          },
        },
      });

      // Add line highlight for components with errors
      if (aggregated.error_rate > 0) {
        newDecorations.push({
          range: new monaco.Range(
            component.lineNumber,
            1,
            component.lineNumber,
            1,
          ),
          options: {
            isWholeLine: true,
            className:
              aggregated.error_rate >= 5 ? "line-error" : "line-warning",
          },
        });
      }
    });

    // Apply decorations
    const newDecorationIds = editor.deltaDecorations(
      decorations,
      newDecorations,
    );
    setDecorations(newDecorationIds);
  }, [metrics, parsedComponents]);

  // Handle editor mount
  const handleEditorMount: OnMount = (editor, monaco) => {
    editorRef.current = editor;

    // Add custom CSS for decorations
    const style = document.createElement("style");
    style.innerHTML = `
      .metrics-decoration {
        color: #6b7280 !important;
        font-size: 0.85em !important;
        font-style: italic !important;
        opacity: 0.8 !important;
      }
      .line-warning {
        background-color: rgba(234, 179, 8, 0.1) !important;
      }
      .line-error {
        background-color: rgba(239, 68, 68, 0.1) !important;
      }
      .metrics-gutter-decoration {
        background-color: rgba(34, 197, 94, 0.2) !important;
        width: 3px !important;
        margin-left: 3px !important;
      }
    `;
    document.head.appendChild(style);

    // Register hover provider for detailed metrics
    if (metrics && metrics.length > 0) {
      monaco.languages.registerHoverProvider("yaml", {
        provideHover: (model, position) => {
          const line = position.lineNumber;
          const component = parsedComponents.find((c) => c.lineNumber === line);

          if (!component) {
            return null;
          }

          const componentMetrics = metrics.filter(
            (m) =>
              m.component_name === component.name &&
              mapComponentType(m.component_type) === component.type,
          );

          if (componentMetrics.length === 0) {
            return null;
          }

          return {
            range: new monaco.Range(
              line,
              1,
              line,
              model.getLineMaxColumn(line),
            ),
            contents: [
              { value: `**${component.name}** (${component.type})` },
              { value: formatHoverMessage(component.name, componentMetrics) },
            ],
          };
        },
      });
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">
          YAML Configuration
          {metrics && metrics.length > 0 && (
            <span className="ml-2 text-sm font-normal text-muted-foreground">
              (with live metrics)
            </span>
          )}
        </CardTitle>
        <CardDescription>
          {readonly
            ? "View OpenTelemetry collector configuration"
            : "Edit your OpenTelemetry collector configuration in YAML format"}
          {metrics && metrics.length > 0 && (
            <span className="block mt-1 text-xs">
              Hover over components to see detailed metrics
            </span>
          )}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="border rounded-lg overflow-hidden">
          <Editor
            height="60vh"
            defaultLanguage="yaml"
            value={value}
            onChange={(value) => !readonly && onChange(value || "")}
            theme="vs-light"
            onMount={handleEditorMount}
            options={{
              minimap: { enabled: false },
              fontSize: 13,
              lineNumbers: "on",
              roundedSelection: false,
              scrollBeyondLastLine: false,
              readOnly: readonly,
              automaticLayout: true,
              glyphMargin: metrics && metrics.length > 0, // Show glyph margin if we have metrics
            }}
          />
        </div>
      </CardContent>
    </Card>
  );
}

/**
 * Map component type from API to parser type
 */
function mapComponentType(
  apiType: string,
): "receiver" | "processor" | "exporter" {
  // API returns full type like "receiver", "processor", "exporter"
  if (apiType === "receiver") return "receiver";
  if (apiType === "processor") return "processor";
  if (apiType === "exporter") return "exporter";
  return "receiver"; // fallback
}

/**
 * Aggregate metrics across multiple pipeline types
 */
function aggregateMetrics(metrics: ComponentMetrics[]): ComponentMetrics {
  const aggregated: ComponentMetrics = {
    component_type: metrics[0].component_type,
    component_name: metrics[0].component_name,
    pipeline_type: "all",
    throughput: 0,
    errors: 0,
    error_rate: 0,
    received: 0,
    accepted: 0,
    refused: 0,
    dropped: 0,
    sent: 0,
    send_failed: 0,
    last_updated: metrics[0].last_updated,
  };

  metrics.forEach((m) => {
    aggregated.throughput += m.throughput;
    aggregated.errors += m.errors;
    aggregated.received = (aggregated.received || 0) + (m.received || 0);
    aggregated.accepted = (aggregated.accepted || 0) + (m.accepted || 0);
    aggregated.refused = (aggregated.refused || 0) + (m.refused || 0);
    aggregated.dropped = (aggregated.dropped || 0) + (m.dropped || 0);
    aggregated.sent = (aggregated.sent || 0) + (m.sent || 0);
    aggregated.send_failed =
      (aggregated.send_failed || 0) + (m.send_failed || 0);
  });

  // Recalculate error rate
  if (aggregated.throughput > 0) {
    aggregated.error_rate = (aggregated.errors / aggregated.throughput) * 100;
  }

  return aggregated;
}

/**
 * Format metrics text for inline display
 */
function formatMetricsText(metrics: ComponentMetrics): string {
  const parts: string[] = [];

  if (metrics.throughput > 0) {
    parts.push(`📊 ${formatThroughput(metrics.throughput)}`);
  }

  if (metrics.error_rate > 0) {
    parts.push(
      `${getStatusIcon(metrics.error_rate)} ${formatErrorRate(metrics.error_rate)} errors`,
    );
  } else if (metrics.throughput > 0) {
    parts.push(`${getStatusIcon(0)} healthy`);
  }

  return parts.join(" | ");
}

/**
 * Format hover message with detailed metrics
 */
function formatHoverMessage(
  _componentName: string,
  metrics: ComponentMetrics[],
): string {
  const lines: string[] = [];

  metrics.forEach((m) => {
    lines.push(`\n**Pipeline: ${m.pipeline_type}**`);
    lines.push(`- Throughput: ${formatThroughput(m.throughput)}`);
    lines.push(`- Error Rate: ${formatErrorRate(m.error_rate)}`);

    if (m.received !== undefined && m.received > 0) {
      lines.push(`- Received: ${m.received.toLocaleString()}`);
    }
    if (m.accepted !== undefined && m.accepted > 0) {
      lines.push(`- Accepted: ${m.accepted.toLocaleString()}`);
    }
    if (m.sent !== undefined && m.sent > 0) {
      lines.push(`- Sent: ${m.sent.toLocaleString()}`);
    }
    if (m.dropped !== undefined && m.dropped > 0) {
      lines.push(`- Dropped: ${m.dropped.toLocaleString()}`);
    }
    if (m.send_failed !== undefined && m.send_failed > 0) {
      lines.push(`- Send Failed: ${m.send_failed.toLocaleString()}`);
    }
  });

  const aggregated = aggregateMetrics(metrics);
  lines.push(`\n**Total Across All Pipelines**`);
  lines.push(`- Total Throughput: ${formatThroughput(aggregated.throughput)}`);
  lines.push(`- Overall Error Rate: ${formatErrorRate(aggregated.error_rate)}`);

  return lines.join("\n");
}
