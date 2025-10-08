import { Clock, Database, LineChart, TableIcon } from "lucide-react";
import { useState } from "react";

import type { LawrenceQLResponse } from "../../api/lawrence-ql";
import { Badge } from "../ui/badge";
import { Button } from "../ui/button";
import { Card } from "../ui/card";
import {
  Table,
  TableHeader,
  TableRow,
  TableHead,
  TableBody,
  TableCell,
} from "../ui/table";

import { QueryResultsChart } from "./QueryResultsChart";

interface QueryResultsProps {
  results: LawrenceQLResponse;
}

export function QueryResults({ results }: QueryResultsProps) {
  const { results: data, meta } = results;
  const [activeView, setActiveView] = useState<"chart" | "table">("chart");

  // Determine if chart view is applicable (only for metrics)
  const hasMetrics = data.length > 0 && data[0].type === "metrics";
  const canShowChart = hasMetrics && data.length > 0;

  return (
    <div className="space-y-4">
      {/* Meta Information */}
      <Card className="p-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-6">
            <div className="flex items-center gap-2 text-sm">
              <Database className="h-4 w-4 text-muted-foreground" />
              <span className="font-medium">{meta.row_count}</span>
              <span className="text-muted-foreground">results</span>
            </div>
            <div className="flex items-center gap-2 text-sm">
              <Clock className="h-4 w-4 text-muted-foreground" />
              <span className="font-medium">
                {(meta.execution_time / 1000000).toFixed(2)}ms
              </span>
              <span className="text-muted-foreground">execution time</span>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Badge variant="outline">{meta.query_type}</Badge>
            {canShowChart && (
              <div className="flex gap-1 ml-2">
                <Button
                  variant={activeView === "chart" ? "default" : "ghost"}
                  size="sm"
                  onClick={() => setActiveView("chart")}
                >
                  <LineChart className="h-4 w-4" />
                </Button>
                <Button
                  variant={activeView === "table" ? "default" : "ghost"}
                  size="sm"
                  onClick={() => setActiveView("table")}
                >
                  <TableIcon className="h-4 w-4" />
                </Button>
              </div>
            )}
          </div>
        </div>
      </Card>

      {/* Chart View */}
      {canShowChart && activeView === "chart" && (
        <QueryResultsChart results={data} />
      )}

      {/* Results Table */}
      {(!canShowChart || activeView === "table") && (
        <Card className="overflow-hidden">
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-20">Type</TableHead>
                  <TableHead className="w-48">Name</TableHead>
                  <TableHead className="w-40">Timestamp</TableHead>
                  <TableHead className="w-28">Value</TableHead>
                  <TableHead className="w-80">Labels</TableHead>
                  <TableHead className="w-32">Data</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data.map((result, idx) => {
                  // Get metric/log/trace name from labels or data
                  const metricName =
                    result.labels?.__name__ ||
                    result.labels?.name ||
                    result.data?.name ||
                    (result.type === "metrics"
                      ? "metric"
                      : result.type === "logs"
                        ? "log"
                        : "trace");

                  const otherLabels = Object.entries(
                    result.labels || {},
                  ).filter(([key]) => key !== "__name__" && key !== "name");

                  return (
                    <TableRow key={idx}>
                      <TableCell className="w-20">
                        <Badge variant={getTypeBadgeVariant(result.type)}>
                          {result.type}
                        </Badge>
                      </TableCell>
                      <TableCell
                        className="w-48 font-mono text-sm font-medium truncate"
                        title={String(metricName)}
                      >
                        {String(metricName)}
                      </TableCell>
                      <TableCell className="w-40 font-mono text-xs">
                        {new Date(result.timestamp).toLocaleString()}
                      </TableCell>
                      <TableCell className="w-28 font-mono">
                        {formatValue(result.value)}
                      </TableCell>
                      <TableCell className="w-80">
                        <div className="flex flex-wrap gap-1 max-h-20 overflow-y-auto">
                          {otherLabels.map(([key, value]) => (
                            <Badge
                              key={key}
                              variant="secondary"
                              className="text-xs whitespace-nowrap"
                            >
                              {key}={value}
                            </Badge>
                          ))}
                        </div>
                      </TableCell>
                      <TableCell className="w-32">
                        {result.data && (
                          <details className="text-xs">
                            <summary className="cursor-pointer text-muted-foreground hover:text-foreground">
                              View details
                            </summary>
                            <pre className="mt-2 p-2 bg-muted rounded text-xs overflow-auto max-h-40">
                              {JSON.stringify(result.data, null, 2)}
                            </pre>
                          </details>
                        )}
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          </div>
        </Card>
      )}
    </div>
  );
}

function getTypeBadgeVariant(
  type: string,
): "default" | "secondary" | "destructive" | "outline" {
  switch (type) {
    case "metrics":
      return "default";
    case "logs":
      return "secondary";
    case "traces":
      return "outline";
    default:
      return "outline";
  }
}

function formatValue(value: any): string {
  if (typeof value === "number") {
    return value.toLocaleString();
  }
  if (typeof value === "object") {
    return JSON.stringify(value);
  }
  return String(value);
}
