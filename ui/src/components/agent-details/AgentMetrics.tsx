import {
  Activity,
  Cpu,
  Database,
  Download,
  HardDrive,
  Upload,
} from "lucide-react";
import { useMemo, useState } from "react";
import {
  Area,
  AreaChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import useSWR from "swr";

import { queryMetrics, type MetricData } from "@/api/telemetry";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { LoadingSpinner } from "@/components/ui/loading-spinner";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

interface AgentMetricsProps {
  agentId: string;
}

interface ComponentMetrics {
  name: string;
  metrics: MetricData[];
}

interface HealthMetric {
  label: string;
  value: string;
  icon: React.ReactNode;
  trend?: "up" | "down" | "stable";
}

interface TimeSeriesPoint {
  time: string;
  value: number;
  timestamp: number;
}

export function AgentMetrics({ agentId }: AgentMetricsProps) {
  const [timeRange, setTimeRange] = useState<"1h" | "6h" | "24h">("1h");

  const { data: metricsData, isLoading } = useSWR(
    `agent-metrics-${agentId}-${timeRange}`,
    async () => {
      const endTime = new Date();
      const hours = timeRange === "1h" ? 1 : timeRange === "6h" ? 6 : 24;
      const startTime = new Date(endTime.getTime() - hours * 60 * 60 * 1000);

      const result = await queryMetrics({
        agent_id: agentId,
        start_time: startTime.toISOString(),
        end_time: endTime.toISOString(),
        limit: 10000,
      });
      return result.metrics || [];
    },
    {
      refreshInterval: 30000, // Refresh every 30 seconds
    },
  );

  // Group metrics by component type
  const groupedMetrics = useMemo(() => {
    if (!metricsData) return { receivers: [], processors: [], exporters: [] };

    const receivers: ComponentMetrics[] = [];
    const processors: ComponentMetrics[] = [];
    const exporters: ComponentMetrics[] = [];

    // Group by component
    const componentMap = new Map<
      string,
      { type: "receiver" | "processor" | "exporter"; metrics: MetricData[] }
    >();

    metricsData.forEach((metric) => {
      // Extract component from metric name or attributes
      let componentType: "receiver" | "processor" | "exporter" | null = null;
      let componentName = "";

      if (metric.metric_name.includes("receiver")) {
        componentType = "receiver";
        componentName =
          metric.metric_attributes?.receiver || "unknown_receiver";
      } else if (metric.metric_name.includes("processor")) {
        componentType = "processor";
        componentName =
          metric.metric_attributes?.processor || "unknown_processor";
      } else if (metric.metric_name.includes("exporter")) {
        componentType = "exporter";
        componentName =
          metric.metric_attributes?.exporter || "unknown_exporter";
      }

      if (componentType && componentName) {
        const key = `${componentType}:${componentName}`;
        if (!componentMap.has(key)) {
          componentMap.set(key, { type: componentType, metrics: [] });
        }
        componentMap.get(key)!.metrics.push(metric);
      }
    });

    // Convert to arrays
    componentMap.forEach((value, key) => {
      const name = key.split(":")[1];
      const component = { name, metrics: value.metrics };

      if (value.type === "receiver") {
        receivers.push(component);
      } else if (value.type === "processor") {
        processors.push(component);
      } else if (value.type === "exporter") {
        exporters.push(component);
      }
    });

    return { receivers, processors, exporters };
  }, [metricsData]);

  // Calculate health metrics
  const healthMetrics = useMemo<HealthMetric[]>(() => {
    if (!metricsData) return [];

    const metrics: HealthMetric[] = [];

    // Group metrics by name and get latest values
    const metricsByName = new Map<string, MetricData[]>();
    metricsData.forEach((m) => {
      if (!metricsByName.has(m.metric_name)) {
        metricsByName.set(m.metric_name, []);
      }
      metricsByName.get(m.metric_name)!.push(m);
    });

    // CPU usage - use cpu_seconds rate over time
    const cpuSecondsMetrics = metricsByName.get("otelcol_process_cpu_seconds");
    if (cpuSecondsMetrics && cpuSecondsMetrics.length >= 2) {
      const sorted = [...cpuSecondsMetrics].sort(
        (a, b) =>
          new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime(),
      );
      const first = sorted[0];
      const last = sorted[sorted.length - 1];
      const cpuDelta = last.value - first.value;
      const timeDelta =
        (new Date(last.timestamp).getTime() -
          new Date(first.timestamp).getTime()) /
        1000; // seconds
      const cpuPercent = timeDelta > 0 ? (cpuDelta / timeDelta) * 100 : 0;
      metrics.push({
        label: "CPU Usage",
        value: `${cpuPercent.toFixed(1)}%`,
        icon: <Cpu className="h-4 w-4" />,
      });
    }

    // Memory usage - use latest RSS memory
    const memMetrics = metricsByName.get("otelcol_process_memory_rss");
    if (memMetrics && memMetrics.length > 0) {
      const latest = memMetrics[memMetrics.length - 1];
      const memMB = (latest.value / (1024 * 1024)).toFixed(0);
      metrics.push({
        label: "Memory Usage",
        value: `${memMB} MB`,
        icon: <HardDrive className="h-4 w-4" />,
      });
    }

    // Data Received - sum latest values of all accepted metrics
    let totalReceived = 0;
    const acceptedSpans = metricsByName.get(
      "otelcol_receiver_accepted_spans",
    )?.[0];
    const acceptedLogs = metricsByName.get(
      "otelcol_receiver_accepted_log_records",
    )?.[0];
    if (acceptedSpans) totalReceived += acceptedSpans.value;
    if (acceptedLogs) totalReceived += acceptedLogs.value;

    if (totalReceived > 0) {
      metrics.push({
        label: "Data Received",
        value: formatNumber(totalReceived),
        icon: <Download className="h-4 w-4" />,
      });
    }

    // Data Sent - sum latest values of all sent metrics
    let totalSent = 0;
    const sentSpans = metricsByName.get("otelcol_exporter_sent_spans")?.[0];
    const sentLogs = metricsByName.get(
      "otelcol_exporter_sent_log_records",
    )?.[0];
    if (sentSpans) totalSent += sentSpans.value;
    if (sentLogs) totalSent += sentLogs.value;

    if (totalSent > 0) {
      metrics.push({
        label: "Data Sent",
        value: formatNumber(totalSent),
        icon: <Upload className="h-4 w-4" />,
      });
    }

    return metrics;
  }, [metricsData]);

  if (isLoading) {
    return <LoadingSpinner />;
  }

  return (
    <div className="space-y-4">
      {/* Health Metrics Overview */}
      <div className="grid grid-cols-2 gap-3">
        {healthMetrics.map((metric, idx) => (
          <Card key={idx}>
            <CardContent className="pt-4">
              <div className="flex items-center gap-2">
                <div className="rounded-lg bg-primary/10 p-2 text-primary">
                  {metric.icon}
                </div>
                <div className="flex-1">
                  <p className="text-xs text-muted-foreground">
                    {metric.label}
                  </p>
                  <p className="text-lg font-semibold">{metric.value}</p>
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Time Range Selector */}
      <div className="flex justify-end gap-2">
        <button
          onClick={() => setTimeRange("1h")}
          className={`px-3 py-1 text-xs rounded ${
            timeRange === "1h"
              ? "bg-primary text-primary-foreground"
              : "bg-muted hover:bg-muted/80"
          }`}
        >
          1h
        </button>
        <button
          onClick={() => setTimeRange("6h")}
          className={`px-3 py-1 text-xs rounded ${
            timeRange === "6h"
              ? "bg-primary text-primary-foreground"
              : "bg-muted hover:bg-muted/80"
          }`}
        >
          6h
        </button>
        <button
          onClick={() => setTimeRange("24h")}
          className={`px-3 py-1 text-xs rounded ${
            timeRange === "24h"
              ? "bg-primary text-primary-foreground"
              : "bg-muted hover:bg-muted/80"
          }`}
        >
          24h
        </button>
      </div>

      {/* Component Metrics Tabs */}
      <Tabs defaultValue="receivers" className="w-full">
        <TabsList className="grid w-full grid-cols-3">
          <TabsTrigger value="receivers">
            Receivers ({groupedMetrics.receivers.length})
          </TabsTrigger>
          <TabsTrigger value="processors">
            Processors ({groupedMetrics.processors.length})
          </TabsTrigger>
          <TabsTrigger value="exporters">
            Exporters ({groupedMetrics.exporters.length})
          </TabsTrigger>
        </TabsList>

        <TabsContent value="receivers" className="space-y-4">
          {groupedMetrics.receivers.length > 0 ? (
            groupedMetrics.receivers.map((component) => (
              <ComponentMetricCard
                key={component.name}
                component={component}
                icon={<Database className="h-4 w-4" />}
              />
            ))
          ) : (
            <Card>
              <CardContent className="py-8 text-center text-muted-foreground">
                No receiver metrics available
              </CardContent>
            </Card>
          )}
        </TabsContent>

        <TabsContent value="processors" className="space-y-4">
          {groupedMetrics.processors.length > 0 ? (
            groupedMetrics.processors.map((component) => (
              <ComponentMetricCard
                key={component.name}
                component={component}
                icon={<Activity className="h-4 w-4" />}
              />
            ))
          ) : (
            <Card>
              <CardContent className="py-8 text-center text-muted-foreground">
                No processor metrics available
              </CardContent>
            </Card>
          )}
        </TabsContent>

        <TabsContent value="exporters" className="space-y-4">
          {groupedMetrics.exporters.length > 0 ? (
            groupedMetrics.exporters.map((component) => (
              <ComponentMetricCard
                key={component.name}
                component={component}
                icon={<Upload className="h-4 w-4" />}
              />
            ))
          ) : (
            <Card>
              <CardContent className="py-8 text-center text-muted-foreground">
                No exporter metrics available
              </CardContent>
            </Card>
          )}
        </TabsContent>
      </Tabs>
    </div>
  );
}

interface ComponentMetricCardProps {
  component: ComponentMetrics;
  icon: React.ReactNode;
}

function ComponentMetricCard({ component, icon }: ComponentMetricCardProps) {
  // Group metrics by metric name for this component
  const metricsByName = useMemo(() => {
    const map = new Map<string, MetricData[]>();
    component.metrics.forEach((metric) => {
      if (!map.has(metric.metric_name)) {
        map.set(metric.metric_name, []);
      }
      map.get(metric.metric_name)!.push(metric);
    });
    return map;
  }, [component.metrics]);

  // Prepare time series data for each metric
  const timeSeriesData = useMemo(() => {
    const result = new Map<string, TimeSeriesPoint[]>();

    metricsByName.forEach((metrics, metricName) => {
      const points: TimeSeriesPoint[] = metrics
        .map((m) => ({
          time: new Date(m.timestamp).toLocaleTimeString([], {
            hour: "2-digit",
            minute: "2-digit",
          }),
          value: m.value,
          timestamp: new Date(m.timestamp).getTime(),
        }))
        .sort((a, b) => a.timestamp - b.timestamp)
        .slice(-20); // Last 20 points

      result.set(metricName, points);
    });

    return result;
  }, [metricsByName]);

  // Calculate summary stats for the component - use latest values
  const stats = useMemo(() => {
    // Get latest value for each metric type
    const getLatestValue = (filter: (name: string) => boolean): number => {
      let total = 0;
      metricsByName.forEach((metrics, name) => {
        if (filter(name) && metrics.length > 0) {
          // Sort by timestamp and get the latest
          const sorted = [...metrics].sort(
            (a, b) =>
              new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime(),
          );
          total += sorted[0].value;
        }
      });
      return total;
    };

    return {
      accepted: getLatestValue(
        (name) => name.includes("accepted") || name.includes("received"),
      ),
      refused: getLatestValue((name) => name.includes("refused")),
      sent: getLatestValue(
        (name) => name.includes("sent") || name.includes("exported"),
      ),
    };
  }, [metricsByName]);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          {icon}
          {component.name}
        </CardTitle>
        <CardDescription>
          {component.metrics.length} data points • {metricsByName.size} unique
          metrics
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Summary Stats */}
        <div className="grid grid-cols-3 gap-2 pb-2">
          {stats.accepted > 0 && (
            <div className="text-center p-2 bg-green-500/10 rounded">
              <p className="text-xs text-muted-foreground">Accepted</p>
              <p className="text-sm font-semibold text-green-600 dark:text-green-400">
                {formatNumber(stats.accepted)}
              </p>
            </div>
          )}
          {stats.refused > 0 && (
            <div className="text-center p-2 bg-red-500/10 rounded">
              <p className="text-xs text-muted-foreground">Refused</p>
              <p className="text-sm font-semibold text-red-600 dark:text-red-400">
                {formatNumber(stats.refused)}
              </p>
            </div>
          )}
          {stats.sent > 0 && (
            <div className="text-center p-2 bg-blue-500/10 rounded">
              <p className="text-xs text-muted-foreground">Sent</p>
              <p className="text-sm font-semibold text-blue-600 dark:text-blue-400">
                {formatNumber(stats.sent)}
              </p>
            </div>
          )}
        </div>

        {/* Time Series Charts for Top Metrics */}
        {Array.from(timeSeriesData.entries())
          .slice(0, 3)
          .map(([metricName, data]) => (
            <div key={metricName} className="space-y-1">
              <p className="text-xs font-medium text-muted-foreground">
                {metricName.replace("otelcol_", "").replace(/_/g, " ")}
              </p>
              <ResponsiveContainer width="100%" height={80}>
                <AreaChart data={data}>
                  <defs>
                    <linearGradient id={metricName} x1="0" y1="0" x2="0" y2="1">
                      <stop
                        offset="5%"
                        stopColor="hsl(var(--primary))"
                        stopOpacity={0.3}
                      />
                      <stop
                        offset="95%"
                        stopColor="hsl(var(--primary))"
                        stopOpacity={0}
                      />
                    </linearGradient>
                  </defs>
                  <XAxis
                    dataKey="time"
                    style={{ fontSize: "9px" }}
                    stroke="hsl(var(--muted-foreground))"
                  />
                  <YAxis
                    style={{ fontSize: "9px" }}
                    stroke="hsl(var(--muted-foreground))"
                  />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: "hsl(var(--background))",
                      border: "1px solid hsl(var(--border))",
                      borderRadius: "6px",
                      fontSize: "11px",
                    }}
                  />
                  <Area
                    type="monotone"
                    dataKey="value"
                    stroke="hsl(var(--primary))"
                    strokeWidth={2}
                    fill={`url(#${metricName})`}
                  />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          ))}

        {/* Show all metric names if more than 3 */}
        {metricsByName.size > 3 && (
          <details className="text-xs">
            <summary className="cursor-pointer text-muted-foreground hover:text-foreground">
              Show all {metricsByName.size} metrics
            </summary>
            <div className="mt-2 space-y-1 pl-2">
              {Array.from(metricsByName.keys()).map((name) => (
                <div key={name} className="text-muted-foreground">
                  • {name}
                </div>
              ))}
            </div>
          </details>
        )}
      </CardContent>
    </Card>
  );
}

function formatNumber(num: number): string {
  if (num >= 1_000_000) {
    return `${(num / 1_000_000).toFixed(1)}M`;
  } else if (num >= 1_000) {
    return `${(num / 1_000).toFixed(1)}K`;
  }
  return num.toFixed(0);
}
