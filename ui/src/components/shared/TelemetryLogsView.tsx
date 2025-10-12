import { FileText, X } from "lucide-react";
import { useState } from "react";
import useSWR from "swr";

import { queryLogs, type LogData } from "@/api/telemetry";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { LoadingSpinner } from "@/components/ui/loading-spinner";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

interface TelemetryLogsViewProps {
  /**
   * Filter by agent ID
   */
  agentId?: string;
  /**
   * Filter by group ID
   */
  groupId?: string;
  /**
   * Title for the logs card
   */
  title?: string;
  /**
   * Whether to show agent ID in the logs list
   */
  showAgentId?: boolean;
}

/**
 * Reusable telemetry logs view component
 * Can display logs for either an agent or a group
 */
export function TelemetryLogsView({
  agentId,
  groupId,
  title,
  showAgentId = false,
}: TelemetryLogsViewProps) {
  const [logsData, setLogsData] = useState<LogData[]>([]);
  const [severityFilter, setSeverityFilter] = useState<string>("all");
  const [searchFilter, setSearchFilter] = useState<string>("");

  const entityType = agentId ? "agent" : "group";
  const entityId = agentId || groupId;
  const displayTitle =
    title || `${entityType === "agent" ? "Agent" : "Group"} Logs`;

  const { isLoading } = useSWR(
    `${entityType}-logs-${entityId}-${severityFilter}-${searchFilter}`,
    async () => {
      const endTime = new Date();
      const startTime = new Date(endTime.getTime() - 60 * 60 * 1000); // Last hour

      const result = await queryLogs({
        ...(agentId && { agent_id: agentId }),
        ...(groupId && { group_id: groupId }),
        ...(severityFilter !== "all" && { severity: severityFilter }),
        ...(searchFilter && { search: searchFilter }),
        start_time: startTime.toISOString(),
        end_time: endTime.toISOString(),
        limit: 100,
      });
      setLogsData(result.logs || []);
      return result;
    },
    {
      revalidateOnFocus: false,
      revalidateOnReconnect: false,
    },
  );

  const getSeverityColor = (severity?: string) => {
    if (!severity)
      return "text-gray-600 bg-gray-50 dark:text-gray-400 dark:bg-gray-800";
    switch (severity.toUpperCase()) {
      case "ERROR":
      case "FATAL":
        return "text-red-600 bg-red-50 dark:text-red-400 dark:bg-red-950";
      case "WARN":
        return "text-yellow-600 bg-yellow-50 dark:text-yellow-400 dark:bg-yellow-950";
      case "INFO":
        return "text-blue-600 bg-blue-50 dark:text-blue-400 dark:bg-blue-950";
      case "DEBUG":
        return "text-gray-600 bg-gray-50 dark:text-gray-400 dark:bg-gray-800";
      default:
        return "text-gray-600 bg-gray-50 dark:text-gray-400 dark:bg-gray-800";
    }
  };

  const getSeverityOption = (value: string, label: string) => {
    const isAll = value === "all";
    return (
      <div className="flex items-center gap-2">
        {!isAll && (
          <Badge className={`text-xs ${getSeverityColor(value)}`}>
            {label}
          </Badge>
        )}
        {isAll && <span>{label}</span>}
      </div>
    );
  };

  if (isLoading) {
    return <LoadingSpinner />;
  }

  const logCountText =
    logsData.length > 0
      ? `${logsData.length} logs in last hour`
      : "No logs available";

  const handleClearFilters = () => {
    setSeverityFilter("all");
    setSearchFilter("");
  };

  const hasActiveFilters = severityFilter !== "all" || searchFilter !== "";

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <FileText className="h-4 w-4" />
          {displayTitle}
        </CardTitle>
        <CardDescription>{logCountText}</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col gap-3 mb-4">
          <div className="flex flex-wrap gap-2">
            <Select value={severityFilter} onValueChange={setSeverityFilter}>
              <SelectTrigger className="w-[160px]">
                <SelectValue placeholder="Severity">
                  {severityFilter === "all" ? (
                    "All Severities"
                  ) : (
                    <Badge
                      className={`text-xs ${getSeverityColor(severityFilter)}`}
                    >
                      {severityFilter}
                    </Badge>
                  )}
                </SelectValue>
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">
                  {getSeverityOption("all", "All Severities")}
                </SelectItem>
                <SelectItem value="error">
                  {getSeverityOption("error", "error")}
                </SelectItem>
                <SelectItem value="warn">
                  {getSeverityOption("warn", "warn")}
                </SelectItem>
                <SelectItem value="info">
                  {getSeverityOption("info", "info")}
                </SelectItem>
                <SelectItem value="debug">
                  {getSeverityOption("debug", "debug")}
                </SelectItem>
              </SelectContent>
            </Select>
            <div className="flex-1 min-w-[200px]">
              <Input
                placeholder="Search logs..."
                value={searchFilter}
                onChange={(e) => setSearchFilter(e.target.value)}
                className="w-full"
              />
            </div>
            {hasActiveFilters && (
              <Button
                variant="outline"
                size="sm"
                onClick={handleClearFilters}
                className="gap-1"
              >
                <X className="h-3 w-3" />
                Clear
              </Button>
            )}
          </div>
        </div>
        <ScrollArea className="h-96">
          {logsData.map((log, idx) => {
            const logDate = new Date(log.timestamp);
            const timeString = logDate.toLocaleTimeString("en-US", {
              hour: "2-digit",
              minute: "2-digit",
              second: "2-digit",
              hour12: false,
            });
            const dateString = logDate.toLocaleDateString("en-US", {
              month: "short",
              day: "numeric",
            });

            return (
              <div key={idx} className="py-1 border-b last:border-0">
                <div className="flex items-start gap-1.5">
                  <Badge
                    className={`text-xs shrink-0 py-0 px-1.5 ${getSeverityColor(log.severity_text)}`}
                  >
                    {log.severity_text || "UNKNOWN"}
                  </Badge>
                  <span className="text-xs text-gray-500 dark:text-gray-400 shrink-0 font-mono">
                    {dateString} {timeString}
                  </span>
                  <div className="flex-1 text-sm min-w-0">
                    <div className="font-mono text-xs break-all">
                      {log.body}
                    </div>
                    {showAgentId && log.agent_id && (
                      <div className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
                        Agent: {log.agent_id}
                      </div>
                    )}
                    {log.log_attributes &&
                      Object.keys(log.log_attributes).length > 0 && (
                        <div className="text-xs text-gray-600 dark:text-gray-400 mt-0.5 font-mono">
                          {Object.entries(log.log_attributes).map(
                            ([key, value]) => (
                              <span key={key} className="mr-2">
                                {key}={value}
                              </span>
                            ),
                          )}
                        </div>
                      )}
                  </div>
                </div>
              </div>
            );
          })}
          {logsData.length === 0 && (
            <div className="text-center py-8 text-gray-500">
              No logs available
            </div>
          )}
        </ScrollArea>
      </CardContent>
    </Card>
  );
}
