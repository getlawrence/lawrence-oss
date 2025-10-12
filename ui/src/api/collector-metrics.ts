import { apiPost } from "./base";

/**
 * Component metrics for pipeline visualization
 */
export interface ComponentMetrics {
  component_type: string; // receiver, processor, exporter
  component_name: string;
  pipeline_type: string; // traces, metrics, logs
  throughput: number;
  errors: number;
  error_rate: number;
  received?: number;
  accepted?: number;
  refused?: number;
  dropped?: number;
  sent?: number;
  send_failed?: number;
  last_updated: string;
  labels?: Record<string, string>;
}

/**
 * Raw metric from telemetry API
 */
interface RawMetric {
  timestamp: string;
  agent_id: string;
  service_name: string;
  metric_name: string;
  value: number;
  labels?: Record<string, string>;
  metric_attributes?: Record<string, unknown>;
  type?: string;
}

interface MetricsQueryResponse {
  metrics: RawMetric[];
  count: number;
}

/**
 * Fetch component metrics for an agent using the telemetry metrics endpoint
 */
export async function fetchAgentComponentMetrics(
  agentId: string,
  timeRangeMinutes: number = 5,
): Promise<ComponentMetrics[]> {
  const endTime = new Date();
  const startTime = new Date(endTime.getTime() - timeRangeMinutes * 60 * 1000);

  // Query all collector self-telemetry metrics
  const response = await apiPost<MetricsQueryResponse>(
    "/telemetry/metrics/query",
    {
      agent_id: agentId,
      start_time: startTime.toISOString(),
      end_time: endTime.toISOString(),
      limit: 10000,
    },
  );

  // Filter for OpenTelemetry Collector metrics
  const collectorMetrics = response.metrics.filter(
    (m) =>
      m.metric_name.startsWith("otelcol_receiver_") ||
      m.metric_name.startsWith("otelcol_processor_") ||
      m.metric_name.startsWith("otelcol_exporter_"),
  );

  // Group metrics by component
  const componentMap = new Map<string, Partial<ComponentMetrics>>();

  for (const metric of collectorMetrics) {
    const labels = metric.labels || {};
    const attributes = metric.metric_attributes || {};

    // Extract component information from labels or attributes
    // Typical labels: receiver, processor, exporter, service_name, transport
    const componentType = extractComponentType(metric.metric_name);
    const componentName =
      labels.receiver ||
      labels.processor ||
      labels.exporter ||
      (attributes.receiver as string) ||
      (attributes.processor as string) ||
      (attributes.exporter as string) ||
      "unknown";

    // Try to get pipeline type from labels/attributes first, then fall back to inferring from metric name
    let pipelineType =
      labels.service_name || (attributes.service_name as string);
    if (
      !pipelineType ||
      pipelineType === "unknown" ||
      !["traces", "metrics", "logs"].includes(pipelineType)
    ) {
      // Infer pipeline type from metric name
      const inferredType = extractPipelineType(metric.metric_name);
      if (inferredType) {
        pipelineType = inferredType;
      }
    }

    // Skip metrics where we can't determine the pipeline type
    if (
      !pipelineType ||
      !["traces", "metrics", "logs"].includes(pipelineType)
    ) {
      continue;
    }

    const key = `${componentType}-${componentName}-${pipelineType}`;

    if (!componentMap.has(key)) {
      componentMap.set(key, {
        component_type: componentType,
        component_name: componentName,
        pipeline_type: pipelineType,
        throughput: 0,
        errors: 0,
        error_rate: 0,
        labels,
        last_updated: metric.timestamp,
      });
    }

    const component = componentMap.get(key)!;

    // Map metric names to component properties
    if (metric.metric_name.includes("_received_")) {
      component.received = (component.received || 0) + metric.value;
    } else if (metric.metric_name.includes("_accepted_")) {
      component.accepted = (component.accepted || 0) + metric.value;
    } else if (metric.metric_name.includes("_refused_")) {
      component.refused = (component.refused || 0) + metric.value;
      component.errors = (component.errors || 0) + metric.value;
    } else if (metric.metric_name.includes("_dropped_")) {
      component.dropped = (component.dropped || 0) + metric.value;
      component.errors = (component.errors || 0) + metric.value;
    } else if (metric.metric_name.includes("_sent_")) {
      component.sent = (component.sent || 0) + metric.value;
    } else if (metric.metric_name.includes("_send_failed_")) {
      component.send_failed = (component.send_failed || 0) + metric.value;
      component.errors = (component.errors || 0) + metric.value;
    }

    // Update last_updated to the most recent timestamp
    if (
      new Date(metric.timestamp) > new Date(component.last_updated as string)
    ) {
      component.last_updated = metric.timestamp;
    }
  }

  // Calculate derived metrics
  const result: ComponentMetrics[] = [];
  for (const [, component] of componentMap) {
    const throughput = calculateThroughput(component, timeRangeMinutes);
    const errorRate = calculateErrorRate(component);

    result.push({
      component_type: component.component_type!,
      component_name: component.component_name!,
      pipeline_type: component.pipeline_type!,
      throughput,
      errors: component.errors || 0,
      error_rate: errorRate,
      received: component.received,
      accepted: component.accepted,
      refused: component.refused,
      dropped: component.dropped,
      sent: component.sent,
      send_failed: component.send_failed,
      last_updated: component.last_updated!,
      labels: component.labels,
    });
  }

  return result;
}

/**
 * Extract component type from metric name
 */
function extractComponentType(metricName: string): string {
  if (metricName.startsWith("otelcol_receiver_")) {
    return "receiver";
  } else if (metricName.startsWith("otelcol_processor_")) {
    return "processor";
  } else if (metricName.startsWith("otelcol_exporter_")) {
    return "exporter";
  }
  return "unknown";
}

/**
 * Extract pipeline type from metric name
 * OTEL Collector metrics include the data type in the metric name
 */
function extractPipelineType(metricName: string): string | null {
  // Check for spans (traces)
  if (metricName.includes("_spans") || metricName.includes("_span_")) {
    return "traces";
  }
  // Check for metric points (metrics)
  if (
    metricName.includes("_metric_points") ||
    metricName.includes("_metrics")
  ) {
    return "metrics";
  }
  // Check for log records (logs)
  if (metricName.includes("_log_records") || metricName.includes("_logs")) {
    return "logs";
  }
  return null;
}

/**
 * Calculate throughput (items per second)
 */
function calculateThroughput(
  component: Partial<ComponentMetrics>,
  timeRangeMinutes: number,
): number {
  const total =
    (component.accepted || 0) +
    (component.sent || 0) +
    (component.received || 0);
  const seconds = timeRangeMinutes * 60;
  return total / seconds;
}

/**
 * Calculate error rate (percentage)
 */
function calculateErrorRate(component: Partial<ComponentMetrics>): number {
  const errors = component.errors || 0;
  const total =
    (component.received || 0) +
    (component.accepted || 0) +
    (component.sent || 0) +
    errors;

  if (total === 0) {
    return 0;
  }

  return (errors / total) * 100;
}
