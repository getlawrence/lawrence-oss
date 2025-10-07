import {
  BarChart3,
  FileText,
  GitBranch,
  CheckCircle,
  XCircle,
  AlertCircle,
  Server,
} from "lucide-react";
import { useState, useMemo } from "react";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import useSWR from "swr";

import {
  queryMetrics,
  queryLogs,
  type MetricData,
  type LogData,
} from "@/api/telemetry";
import { getAgentTopology } from "@/api/topology";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

interface AgentDetailsDrawerProps {
  agentId: string | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function AgentDetailsDrawer({
  agentId,
  open,
  onOpenChange,
}: AgentDetailsDrawerProps) {
  const [metricsData, setMetricsData] = useState<MetricData[]>([]);
  const [logsData, setLogsData] = useState<LogData[]>([]);

  const { data: agentTopology, isLoading } = useSWR(
    agentId && open ? `agent-topology-${agentId}` : null,
    () => (agentId ? getAgentTopology(agentId) : null),
  );

  // Fetch metrics when drawer opens
  useSWR(agentId && open ? `agent-metrics-${agentId}` : null, async () => {
    if (!agentId) return null;
    const endTime = new Date();
    const startTime = new Date(endTime.getTime() - 60 * 60 * 1000); // Last hour

    const result = await queryMetrics({
      agent_id: agentId,
      start_time: startTime.toISOString(),
      end_time: endTime.toISOString(),
      limit: 100,
    });
    setMetricsData(result.metrics || []);
    return result;
  });

  // Fetch logs when drawer opens
  useSWR(agentId && open ? `agent-logs-${agentId}` : null, async () => {
    if (!agentId) return null;
    const endTime = new Date();
    const startTime = new Date(endTime.getTime() - 60 * 60 * 1000); // Last hour

    const result = await queryLogs({
      agent_id: agentId,
      start_time: startTime.toISOString(),
      end_time: endTime.toISOString(),
      limit: 100,
    });
    setLogsData(result.logs || []);
    return result;
  });

  const agent = agentTopology?.agent;
  const metrics = agentTopology?.metrics;
  const pipeline = agentTopology?.pipeline;

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "online":
        return <CheckCircle className="h-5 w-5 text-green-500" />;
      case "offline":
        return <XCircle className="h-5 w-5 text-gray-500" />;
      case "error":
        return <AlertCircle className="h-5 w-5 text-red-500" />;
      default:
        return <Server className="h-5 w-5 text-gray-400" />;
    }
  };

  const getSeverityColor = (severity?: string) => {
    if (!severity) return "text-gray-600 bg-gray-50";
    switch (severity.toUpperCase()) {
      case "ERROR":
      case "FATAL":
        return "text-red-600 bg-red-50";
      case "WARN":
        return "text-yellow-600 bg-yellow-50";
      case "INFO":
        return "text-blue-600 bg-blue-50";
      case "DEBUG":
        return "text-gray-600 bg-gray-50";
      default:
        return "text-gray-600 bg-gray-50";
    }
  };

  // Prepare chart data for metrics
  const chartData = useMemo(() => {
    if (!metricsData || !metricsData.length) return [];

    // Group by timestamp and aggregate
    const grouped = metricsData.reduce(
      (acc, metric) => {
        const time = new Date(metric.timestamp).toLocaleTimeString();
        if (!acc[time]) {
          acc[time] = { time, count: 0, totalValue: 0 };
        }
        acc[time].count += 1;
        acc[time].totalValue += metric.value;
        return acc;
      },
      {} as Record<string, { time: string; count: number; totalValue: number }>,
    );

    return Object.values(grouped).slice(-20); // Last 20 data points
  }, [metricsData]);

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="w-full sm:max-w-2xl overflow-y-auto">
        <SheetHeader>
          <SheetTitle className="flex items-center gap-2">
            {agent && getStatusIcon(agent.status)}
            Agent Details
          </SheetTitle>
          <SheetDescription>{agent?.name || "Loading..."}</SheetDescription>
        </SheetHeader>

        {isLoading ? (
          <div className="flex items-center justify-center h-64">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
          </div>
        ) : agent ? (
          <Tabs defaultValue="overview" className="mt-6">
            <TabsList className="grid w-full grid-cols-4">
              <TabsTrigger value="overview">Overview</TabsTrigger>
              <TabsTrigger value="metrics">Metrics</TabsTrigger>
              <TabsTrigger value="logs">Logs</TabsTrigger>
              <TabsTrigger value="pipeline">Pipeline</TabsTrigger>
            </TabsList>

            <TabsContent value="overview" className="space-y-4">
              <Card>
                <CardHeader>
                  <CardTitle className="text-base">Agent Information</CardTitle>
                </CardHeader>
                <CardContent className="space-y-3">
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-600">ID:</span>
                    <span className="text-sm font-mono">{agent.id}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-600">Version:</span>
                    <span className="text-sm">{agent.version}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-600">Status:</span>
                    <Badge
                      variant={
                        agent.status === "online" ? "default" : "secondary"
                      }
                    >
                      {agent.status}
                    </Badge>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-600">Group:</span>
                    <span className="text-sm">
                      {agent.group_name || "No Group"}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-600">Last Seen:</span>
                    <span className="text-sm">
                      {new Date(agent.last_seen).toLocaleString()}
                    </span>
                  </div>
                </CardContent>
              </Card>

              {metrics && (
                <Card>
                  <CardHeader>
                    <CardTitle className="text-base">
                      Telemetry Stats (Last 5 min)
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-2">
                    <div className="flex justify-between">
                      <span className="text-sm text-gray-600">Metrics:</span>
                      <span className="text-sm font-semibold">
                        {metrics.metric_count}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-sm text-gray-600">Logs:</span>
                      <span className="text-sm font-semibold">
                        {metrics.log_count}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-sm text-gray-600">Traces:</span>
                      <span className="text-sm font-semibold">
                        {metrics.trace_count}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-sm text-gray-600">Throughput:</span>
                      <span className="text-sm font-semibold">
                        {metrics.throughput_rps.toFixed(2)} rps
                      </span>
                    </div>
                  </CardContent>
                </Card>
              )}

              {agent.labels && Object.keys(agent.labels).length > 0 && (
                <Card>
                  <CardHeader>
                    <CardTitle className="text-base">Labels</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="flex flex-wrap gap-2">
                      {Object.entries(agent.labels).map(([key, value]) => (
                        <Badge key={key} variant="outline">
                          {key}={String(value)}
                        </Badge>
                      ))}
                    </div>
                  </CardContent>
                </Card>
              )}
            </TabsContent>

            <TabsContent value="metrics" className="space-y-4">
              <Card>
                <CardHeader>
                  <CardTitle className="text-base flex items-center gap-2">
                    <BarChart3 className="h-4 w-4" />
                    Metrics Timeline
                  </CardTitle>
                  <CardDescription>
                    Last hour of metric activity
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  {chartData.length > 0 ? (
                    <ResponsiveContainer width="100%" height={200}>
                      <LineChart data={chartData}>
                        <XAxis dataKey="time" style={{ fontSize: "10px" }} />
                        <YAxis style={{ fontSize: "10px" }} />
                        <Tooltip />
                        <Line
                          type="monotone"
                          dataKey="count"
                          stroke="#3b82f6"
                          strokeWidth={2}
                        />
                      </LineChart>
                    </ResponsiveContainer>
                  ) : (
                    <div className="text-center py-8 text-gray-500">
                      No metrics data available
                    </div>
                  )}
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="text-base">Recent Metrics</CardTitle>
                  <CardDescription>
                    {metricsData.length} metrics in last hour
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <ScrollArea className="h-64">
                    {metricsData.slice(0, 20).map((metric, idx) => (
                      <div key={idx} className="py-2 border-b last:border-0">
                        <div className="flex justify-between text-sm">
                          <span className="font-medium">
                            {metric.metric_name}
                          </span>
                          <span className="text-gray-600">{metric.value}</span>
                        </div>
                        <div className="text-xs text-gray-500 mt-1">
                          {new Date(metric.timestamp).toLocaleTimeString()}
                        </div>
                      </div>
                    ))}
                  </ScrollArea>
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="logs" className="space-y-4">
              <Card>
                <CardHeader>
                  <CardTitle className="text-base flex items-center gap-2">
                    <FileText className="h-4 w-4" />
                    Agent Logs
                  </CardTitle>
                  <CardDescription>
                    {logsData.length} logs in last hour
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <ScrollArea className="h-96">
                    {logsData.map((log, idx) => (
                      <div key={idx} className="py-3 border-b last:border-0">
                        <div className="flex items-start gap-2">
                          <Badge
                            className={`text-xs ${getSeverityColor(log.severity_text)}`}
                          >
                            {log.severity_text || "UNKNOWN"}
                          </Badge>
                          <div className="flex-1 text-sm">
                            <div className="font-mono text-xs break-all">
                              {log.body}
                            </div>
                            <div className="text-xs text-gray-500 mt-1">
                              {new Date(log.timestamp).toLocaleString()}
                            </div>
                          </div>
                        </div>
                      </div>
                    ))}
                    {logsData.length === 0 && (
                      <div className="text-center py-8 text-gray-500">
                        No logs available
                      </div>
                    )}
                  </ScrollArea>
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="pipeline" className="space-y-4">
              <Card>
                <CardHeader>
                  <CardTitle className="text-base flex items-center gap-2">
                    <GitBranch className="h-4 w-4" />
                    Pipeline Configuration
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  {pipeline ? (
                    <div className="space-y-3">
                      <div className="flex justify-between">
                        <span className="text-sm text-gray-600">
                          Config ID:
                        </span>
                        <span className="text-sm font-mono">
                          {pipeline.config_id}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-sm text-gray-600">Version:</span>
                        <span className="text-sm">
                          {pipeline.config_version}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-sm text-gray-600">Hash:</span>
                        <span className="text-sm font-mono text-xs">
                          {pipeline.config_hash}
                        </span>
                      </div>
                    </div>
                  ) : (
                    <div className="text-center py-8 text-gray-500">
                      No pipeline configuration available
                    </div>
                  )}
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        ) : null}
      </SheetContent>
    </Sheet>
  );
}
