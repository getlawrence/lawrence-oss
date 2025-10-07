import { useMemo } from "react";
import { Card } from "../ui/card";
import { Badge } from "../ui/badge";
import {
  LineChart,
  Line,
  BarChart,
  Bar,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from "recharts";
import type { QueryResult } from "../../api/lawrence-ql";
import { TrendingUp, BarChart3, Activity } from "lucide-react";

interface QueryResultsChartProps {
  results: QueryResult[];
  chartType?: "line" | "bar" | "area" | "auto";
}

export function QueryResultsChart({
  results,
  chartType = "auto",
}: QueryResultsChartProps) {
  // Transform results into chart data
  const chartData = useMemo(() => {
    if (results.length === 0) return [];

    // Group by timestamp and create data points
    const dataMap = new Map<string, any>();

    results.forEach((result) => {
      const timestamp = new Date(result.timestamp).toLocaleTimeString();

      if (!dataMap.has(timestamp)) {
        dataMap.set(timestamp, { timestamp });
      }

      const dataPoint = dataMap.get(timestamp)!;

      // Create a unique key for this series (based on labels)
      const seriesKey = result.labels
        ? Object.entries(result.labels)
            .map(([k, v]) => `${k}=${v}`)
            .join(",") || "value"
        : "value";

      // Convert value to number
      const numValue = typeof result.value === "number"
        ? result.value
        : parseFloat(String(result.value)) || 0;

      dataPoint[seriesKey] = numValue;
    });

    return Array.from(dataMap.values());
  }, [results]);

  // Determine the best chart type
  const resolvedChartType = useMemo(() => {
    if (chartType !== "auto") return chartType;

    // If we have time series data with multiple points, use line chart
    if (chartData.length > 1) {
      return "line";
    }

    // For single data points or aggregated data, use bar chart
    return "bar";
  }, [chartType, chartData]);

  // Extract series keys (all keys except 'timestamp')
  const seriesKeys = useMemo(() => {
    if (chartData.length === 0) return [];
    return Object.keys(chartData[0]).filter((key) => key !== "timestamp");
  }, [chartData]);

  // Color palette for multiple series
  const colors = [
    "#3b82f6", // blue
    "#10b981", // green
    "#f59e0b", // amber
    "#ef4444", // red
    "#8b5cf6", // violet
    "#ec4899", // pink
    "#06b6d4", // cyan
    "#84cc16", // lime
  ];

  if (chartData.length === 0 || seriesKeys.length === 0) {
    return null;
  }

  const renderChart = () => {
    const commonProps = {
      data: chartData,
      margin: { top: 5, right: 30, left: 20, bottom: 5 },
    };

    switch (resolvedChartType) {
      case "area":
        return (
          <AreaChart {...commonProps}>
            <defs>
              {seriesKeys.map((key, idx) => (
                <linearGradient
                  key={key}
                  id={`color-${idx}`}
                  x1="0"
                  y1="0"
                  x2="0"
                  y2="1"
                >
                  <stop
                    offset="5%"
                    stopColor={colors[idx % colors.length]}
                    stopOpacity={0.8}
                  />
                  <stop
                    offset="95%"
                    stopColor={colors[idx % colors.length]}
                    stopOpacity={0}
                  />
                </linearGradient>
              ))}
            </defs>
            <CartesianGrid strokeDasharray="3 3" opacity={0.3} />
            <XAxis
              dataKey="timestamp"
              tick={{ fontSize: 12 }}
              tickLine={false}
            />
            <YAxis tick={{ fontSize: 12 }} tickLine={false} />
            <Tooltip
              contentStyle={{
                backgroundColor: "hsl(var(--background))",
                border: "1px solid hsl(var(--border))",
                borderRadius: "6px",
              }}
            />
            <Legend wrapperStyle={{ fontSize: "12px" }} />
            {seriesKeys.map((key, idx) => (
              <Area
                key={key}
                type="monotone"
                dataKey={key}
                stroke={colors[idx % colors.length]}
                fillOpacity={1}
                fill={`url(#color-${idx})`}
                strokeWidth={2}
              />
            ))}
          </AreaChart>
        );

      case "bar":
        return (
          <BarChart {...commonProps}>
            <CartesianGrid strokeDasharray="3 3" opacity={0.3} />
            <XAxis
              dataKey="timestamp"
              tick={{ fontSize: 12 }}
              tickLine={false}
            />
            <YAxis tick={{ fontSize: 12 }} tickLine={false} />
            <Tooltip
              contentStyle={{
                backgroundColor: "hsl(var(--background))",
                border: "1px solid hsl(var(--border))",
                borderRadius: "6px",
              }}
            />
            <Legend wrapperStyle={{ fontSize: "12px" }} />
            {seriesKeys.map((key, idx) => (
              <Bar
                key={key}
                dataKey={key}
                fill={colors[idx % colors.length]}
                radius={[4, 4, 0, 0]}
              />
            ))}
          </BarChart>
        );

      case "line":
      default:
        return (
          <LineChart {...commonProps}>
            <CartesianGrid strokeDasharray="3 3" opacity={0.3} />
            <XAxis
              dataKey="timestamp"
              tick={{ fontSize: 12 }}
              tickLine={false}
            />
            <YAxis tick={{ fontSize: 12 }} tickLine={false} />
            <Tooltip
              contentStyle={{
                backgroundColor: "hsl(var(--background))",
                border: "1px solid hsl(var(--border))",
                borderRadius: "6px",
              }}
            />
            <Legend wrapperStyle={{ fontSize: "12px" }} />
            {seriesKeys.map((key, idx) => (
              <Line
                key={key}
                type="monotone"
                dataKey={key}
                stroke={colors[idx % colors.length]}
                strokeWidth={2}
                dot={{ r: 3 }}
                activeDot={{ r: 5 }}
              />
            ))}
          </LineChart>
        );
    }
  };

  return (
    <Card className="p-4">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          {resolvedChartType === "line" && (
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          )}
          {resolvedChartType === "bar" && (
            <BarChart3 className="h-4 w-4 text-muted-foreground" />
          )}
          {resolvedChartType === "area" && (
            <Activity className="h-4 w-4 text-muted-foreground" />
          )}
          <h3 className="font-medium">Visualization</h3>
        </div>
        <div className="flex gap-2">
          <Badge variant="outline" className="text-xs">
            {seriesKeys.length} series
          </Badge>
          <Badge variant="outline" className="text-xs capitalize">
            {resolvedChartType} chart
          </Badge>
        </div>
      </div>
      <ResponsiveContainer width="100%" height={300}>
        {renderChart()}
      </ResponsiveContainer>
    </Card>
  );
}
