import { FileText } from "lucide-react";
import { useState } from "react";
import useSWR from "swr";

import { queryLogs, type LogData } from "@/api/telemetry";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { LoadingSpinner } from "@/components/ui/loading-spinner";
import { ScrollArea } from "@/components/ui/scroll-area";

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

  const entityType = agentId ? "agent" : "group";
  const entityId = agentId || groupId;
  const displayTitle =
    title || `${entityType === "agent" ? "Agent" : "Group"} Logs`;

  const { isLoading } = useSWR(
    `${entityType}-logs-${entityId}`,
    async () => {
      const endTime = new Date();
      const startTime = new Date(endTime.getTime() - 60 * 60 * 1000); // Last hour

      const result = await queryLogs({
        ...(agentId && { agent_id: agentId }),
        ...(groupId && { group_id: groupId }),
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

  if (isLoading) {
    return <LoadingSpinner />;
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <FileText className="h-4 w-4" />
          {displayTitle}
        </CardTitle>
        <CardDescription>{logsData.length} logs in last hour</CardDescription>
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
                  <div className="font-mono text-xs break-all">{log.body}</div>
                  <div className="text-xs text-gray-500 mt-1">
                    {showAgentId && `Agent: ${log.agent_id} • `}
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
  );
}
