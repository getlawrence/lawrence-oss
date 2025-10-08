import { BarChart3 } from "lucide-react";
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

import { queryMetrics, type MetricData } from "@/api/telemetry";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { LoadingSpinner } from "@/components/ui/loading-spinner";
import { ScrollArea } from "@/components/ui/scroll-area";

interface TelemetryMetricsViewProps {
  /**
   * Filter by agent ID
   */
  agentId?: string;
  /**
   * Filter by group ID
   */
  groupId?: string;
  /**
   * Whether to show agent ID in the metrics list
   */
  showAgentId?: boolean;
}

/**
 * Reusable telemetry metrics view component
 * Can display metrics for either an agent or a group
 */
export function TelemetryMetricsView({
  agentId,
  groupId,
  showAgentId = false,
}: TelemetryMetricsViewProps) {
  const [metricsData, setMetricsData] = useState<MetricData[]>([]);

  const entityType = agentId ? "agent" : "group";
  const entityId = agentId || groupId;

  const { isLoading } = useSWR(
    `${entityType}-metrics-${entityId}`,
    async () => {
      const endTime = new Date();
      const startTime = new Date(endTime.getTime() - 60 * 60 * 1000); // Last hour

      const result = await queryMetrics({
        ...(agentId && { agent_id: agentId }),
        ...(groupId && { group_id: groupId }),
        start_time: startTime.toISOString(),
        end_time: endTime.toISOString(),
        limit: 100,
      });
      setMetricsData(result.metrics || []);
      return result;
    },
    {
      revalidateOnFocus: false,
      revalidateOnReconnect: false,
    },
  );

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

  if (isLoading) {
    return <LoadingSpinner />;
  }

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <BarChart3 className="h-4 w-4" />
            Metrics Timeline
          </CardTitle>
          <CardDescription>Last hour of metric activity</CardDescription>
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
                  <span className="font-medium">{metric.metric_name}</span>
                  <span className="text-gray-600">{metric.value}</span>
                </div>
                <div className="text-xs text-gray-500 mt-1">
                  {showAgentId && `Agent: ${metric.agent_id} • `}
                  {new Date(metric.timestamp).toLocaleTimeString()}
                </div>
              </div>
            ))}
            {metricsData.length === 0 && (
              <div className="text-center py-8 text-gray-500">
                No metrics available
              </div>
            )}
          </ScrollArea>
        </CardContent>
      </Card>
    </div>
  );
}
