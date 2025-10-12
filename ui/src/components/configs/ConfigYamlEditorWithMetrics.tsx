import Editor, { type OnMount } from "@monaco-editor/react";
import type * as Monaco from "monaco-editor";
import { useEffect, useRef, useState } from "react";

import { useTheme } from "../ThemeProvider";

import { type ComponentMetrics } from "@/api/collector-metrics";
import {
  parseYamlComponents,
  formatThroughput,
  formatErrorRate,
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
  const editorRef = useRef<Monaco.editor.IStandaloneCodeEditor | null>(null);
  const monacoRef = useRef<typeof Monaco | null>(null);
  const decorationsRef =
    useRef<Monaco.editor.IEditorDecorationsCollection | null>(null);
  const [parsedComponents, setParsedComponents] = useState<YamlComponent[]>([]);

  const { theme } = useTheme();

  // Parse YAML to find components whenever value changes
  useEffect(() => {
    if (value) {
      const components = parseYamlComponents(value);
      setParsedComponents(components);
    }
  }, [value]);

  // Update decorations when metrics or parsed components change
  useEffect(() => {
    if (!editorRef.current || !monacoRef.current || !metrics || metrics.length === 0) {
      return;
    }

    const editor = editorRef.current;
    const monaco = monacoRef.current;
    const model = editor.getModel();
    if (!model) {
      return;
    }

    const newDecorations: Monaco.editor.IModelDeltaDecoration[] = [];

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

      // Create simple decoration with after content
      const metricsText = formatMetricsTextCompact(aggregated);

      // Find the end of the line to place the decoration
      const lineLength = model.getLineMaxColumn(component.lineNumber);

      // Try both before and after to see which works
      newDecorations.push({
        range: new monaco.Range(
          component.lineNumber,
          lineLength,
          component.lineNumber,
          lineLength,
        ),
        options: {
          after: {
            content: metricsText,
          },
        },
      });
    });

    // Clear old decorations and create new ones
    if (decorationsRef.current) {
      decorationsRef.current.clear();
    }

    if (newDecorations.length > 0) {
      decorationsRef.current =
        editor.createDecorationsCollection(newDecorations);
    }
  }, [metrics, parsedComponents]);

  const handleEditorMount: OnMount = (editor, monaco) => {
    editorRef.current = editor;
    monacoRef.current = monaco;
  };

  return (
    <div className="border rounded-lg overflow-hidden">
      <Editor
        height="60vh"
        defaultLanguage="yaml"
        value={value}
        onMount={handleEditorMount}
        onChange={(value) => !readonly && onChange(value || "")}
        theme={theme === "dark" ? "vs-dark" : "vs-light"}
        options={{
          minimap: { enabled: false },
          fontSize: 13,
          lineNumbers: "on",
          roundedSelection: false,
          scrollBeyondLastLine: false,
          readOnly: readonly,
          automaticLayout: true,
          glyphMargin: false, // No gutter decorations
        }}
      />
    </div>
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
 * Format metrics text for inline display (compact version for inline hints)
 */
function formatMetricsTextCompact(metrics: ComponentMetrics): string {
  const parts: string[] = [];

  // Always show throughput
  parts.push(`${formatThroughput(metrics.throughput)}`);

  // Show error rate if > 0
  if (metrics.error_rate > 0) {
    parts.push(`err: ${formatErrorRate(metrics.error_rate)}`);
  }

  return parts.join(" • ");
}
