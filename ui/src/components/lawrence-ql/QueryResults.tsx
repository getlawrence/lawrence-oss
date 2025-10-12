import {
  Clock,
  Database,
  LineChart,
  TableIcon,
  ChevronDown,
  ChevronRight,
} from "lucide-react";
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
  const [expandedRows, setExpandedRows] = useState<Set<number>>(new Set());

  const toggleRowExpansion = (idx: number) => {
    const newExpanded = new Set(expandedRows);
    if (newExpanded.has(idx)) {
      newExpanded.delete(idx);
    } else {
      newExpanded.add(idx);
    }
    setExpandedRows(newExpanded);
  };

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
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-10 py-2 px-2"></TableHead>
                <TableHead className="w-20 py-2 px-2">Type</TableHead>
                <TableHead className="min-w-[120px] max-w-[200px] py-2 px-2">
                  Name
                </TableHead>
                <TableHead className="min-w-[140px] py-2 px-2">
                  Timestamp
                </TableHead>
                <TableHead className="min-w-[150px] py-2 px-2">Value</TableHead>
                <TableHead className="min-w-[150px] py-2 px-2">
                  Labels
                </TableHead>
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

                const otherLabels = Object.entries(result.labels || {}).filter(
                  ([key]) => key !== "__name__" && key !== "name",
                );

                const isExpanded = expandedRows.has(idx);
                const valueStr = formatValue(result.value);
                const hasLongValue = valueStr.length > 50;
                const hasMultipleLabels = otherLabels.length > 3;
                const shouldShowExpand =
                  hasLongValue || hasMultipleLabels || result.data;

                return (
                  <>
                    <TableRow
                      key={idx}
                      className={isExpanded ? "border-b-0" : ""}
                    >
                      <TableCell className="w-10 py-1 px-2">
                        {shouldShowExpand && (
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-5 w-5 p-0"
                            onClick={() => toggleRowExpansion(idx)}
                          >
                            {isExpanded ? (
                              <ChevronDown className="h-3 w-3" />
                            ) : (
                              <ChevronRight className="h-3 w-3" />
                            )}
                          </Button>
                        )}
                      </TableCell>
                      <TableCell className="w-20 py-1 px-2">
                        <Badge
                          variant={getTypeBadgeVariant(result.type)}
                          className="text-xs py-0"
                        >
                          {result.type}
                        </Badge>
                      </TableCell>
                      <TableCell
                        className="min-w-[120px] max-w-[200px] font-mono text-xs font-medium truncate py-1 px-2"
                        title={String(metricName)}
                      >
                        {String(metricName)}
                      </TableCell>
                      <TableCell className="min-w-[140px] font-mono text-xs whitespace-nowrap py-1 px-2">
                        {new Date(result.timestamp).toLocaleString()}
                      </TableCell>
                      <TableCell className="min-w-[150px] font-mono text-xs py-1 px-2">
                        <div
                          className="truncate max-w-[300px]"
                          title={valueStr}
                        >
                          {valueStr}
                        </div>
                      </TableCell>
                      <TableCell className="min-w-[150px] py-1 px-2">
                        <div className="flex flex-wrap gap-1 max-w-[300px]">
                          {otherLabels
                            .slice(0, isExpanded ? undefined : 3)
                            .map(([key, value]) => (
                              <Badge
                                key={key}
                                variant="secondary"
                                className="text-xs py-0 px-1.5"
                              >
                                <span
                                  className="truncate max-w-[150px]"
                                  title={`${key}=${value}`}
                                >
                                  {key}=
                                  {String(value).length > 20
                                    ? `${String(value).substring(0, 20)}...`
                                    : value}
                                </span>
                              </Badge>
                            ))}
                          {!isExpanded && otherLabels.length > 3 && (
                            <Badge
                              variant="outline"
                              className="text-xs py-0 px-1.5"
                            >
                              +{otherLabels.length - 3} more
                            </Badge>
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                    {isExpanded && (
                      <TableRow key={`${idx}-expanded`} className="bg-muted/50">
                        <TableCell colSpan={6} className="py-2 px-3">
                          <div className="space-y-2">
                            {/* Full Value */}
                            {hasLongValue && (
                              <div>
                                <div className="text-xs font-semibold text-muted-foreground mb-0.5">
                                  Full Value
                                </div>
                                <div className="font-mono text-xs p-2 bg-background rounded border break-all">
                                  {valueStr}
                                </div>
                              </div>
                            )}

                            {/* All Labels */}
                            {otherLabels.length > 0 && (
                              <div>
                                <div className="text-xs font-semibold text-muted-foreground mb-0.5">
                                  All Labels ({otherLabels.length})
                                </div>
                                <div className="flex flex-wrap gap-1">
                                  {otherLabels.map(([key, value]) => (
                                    <Badge
                                      key={key}
                                      variant="secondary"
                                      className="text-xs font-mono py-0 px-1.5"
                                    >
                                      {key}={value}
                                    </Badge>
                                  ))}
                                </div>
                              </div>
                            )}

                            {/* Additional Data */}
                            {result.data && (
                              <div>
                                <div className="text-xs font-semibold text-muted-foreground mb-0.5">
                                  Additional Data
                                </div>
                                <pre className="p-2 bg-background rounded border text-xs overflow-auto max-h-60">
                                  {JSON.stringify(result.data, null, 2)}
                                </pre>
                              </div>
                            )}
                          </div>
                        </TableCell>
                      </TableRow>
                    )}
                  </>
                );
              })}
            </TableBody>
          </Table>
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
