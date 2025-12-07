import { AlertCircle, Info, XCircle } from "lucide-react";
import { useState, useEffect } from "react";
import useSWR from "swr";

import { getStepExecutionLogs, type StepExecutionLog } from "@/api/workflows";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { LoadingSpinner } from "@/components/ui/loading-spinner";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

interface StepExecutionLogsProps {
  stepExecutionId: string;
  autoRefresh?: boolean;
  limit?: number;
}

export function StepExecutionLogs({
  stepExecutionId,
  autoRefresh = false,
  limit = 1000,
}: StepExecutionLogsProps) {
  const [logLevelFilter, setLogLevelFilter] = useState<string>("all");

  const { data, error, mutate } = useSWR(
    stepExecutionId ? [`step-execution-logs`, stepExecutionId, limit] : null,
    () => getStepExecutionLogs(stepExecutionId, limit),
    {
      refreshInterval: autoRefresh ? 2000 : 0,
    },
  );

  useEffect(() => {
    if (autoRefresh) {
      const interval = setInterval(() => {
        mutate();
      }, 2000);
      return () => clearInterval(interval);
    }
  }, [autoRefresh, mutate]);

  const logs = data?.logs || [];
  const filteredLogs =
    logLevelFilter === "all"
      ? logs
      : logs.filter((log) => log.level === logLevelFilter);

  const getLogIcon = (level: string) => {
    switch (level) {
      case "error":
        return <XCircle className="h-4 w-4 text-red-500" />;
      case "warning":
        return <AlertCircle className="h-4 w-4 text-yellow-500" />;
      case "info":
        return <Info className="h-4 w-4 text-blue-500" />;
      case "debug":
        return <Info className="h-4 w-4 text-gray-500" />;
      default:
        return <Info className="h-4 w-4 text-gray-500" />;
    }
  };

  const getLogBadge = (level: string) => {
    switch (level) {
      case "error":
        return <Badge variant="destructive">Error</Badge>;
      case "warning":
        return <Badge className="bg-yellow-500">Warning</Badge>;
      case "info":
        return <Badge className="bg-blue-500">Info</Badge>;
      case "debug":
        return <Badge variant="outline">Debug</Badge>;
      default:
        return <Badge variant="outline">{level}</Badge>;
    }
  };

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString();
  };

  if (error) {
    return (
      <Card className="p-6">
        <div className="text-center text-red-600">
          <p>Failed to load logs</p>
          <p className="text-sm text-muted-foreground">{error.message}</p>
        </div>
      </Card>
    );
  }

  if (!data) {
    return (
      <Card className="p-6">
        <div className="flex justify-center items-center py-8">
          <LoadingSpinner />
        </div>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold">Execution Logs</h3>
        <Select value={logLevelFilter} onValueChange={setLogLevelFilter}>
          <SelectTrigger className="w-40">
            <SelectValue placeholder="Filter by level" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Levels</SelectItem>
            <SelectItem value="error">Error</SelectItem>
            <SelectItem value="warning">Warning</SelectItem>
            <SelectItem value="info">Info</SelectItem>
            <SelectItem value="debug">Debug</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {filteredLogs.length === 0 ? (
        <Card className="p-6 text-center">
          <Info className="h-8 w-8 mx-auto mb-2 text-muted-foreground" />
          <p className="text-muted-foreground">No logs found</p>
        </Card>
      ) : (
        <div className="space-y-2 max-h-96 overflow-y-auto">
          {filteredLogs.map((log: StepExecutionLog) => (
            <Card key={log.id} className="p-4">
              <div className="flex items-start gap-3">
                {getLogIcon(log.level)}
                <div className="flex-1 space-y-1">
                  <div className="flex items-center gap-2">
                    {getLogBadge(log.level)}
                    <span className="text-xs text-muted-foreground">
                      {formatTimestamp(log.timestamp)}
                    </span>
                  </div>
                  <p className="text-sm">{log.message}</p>
                  {log.data && (
                    <details className="text-xs">
                      <summary className="cursor-pointer text-muted-foreground hover:text-foreground">
                        View data
                      </summary>
                      <pre className="mt-2 p-2 bg-muted rounded overflow-x-auto">
                        {typeof log.data === "string"
                          ? log.data
                          : JSON.stringify(log.data, null, 2)}
                      </pre>
                    </details>
                  )}
                </div>
              </div>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
