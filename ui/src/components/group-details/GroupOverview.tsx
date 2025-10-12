import { RotateCw } from "lucide-react";
import { useState } from "react";

import { restartGroup } from "@/api/groups";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { InfoCard } from "@/components/ui/info-card";

interface Group {
  id: string;
  name: string;
  labels: Record<string, string>;
  created_at: string;
  updated_at: string;
}

interface GroupMetrics {
  agent_count: number;
  metric_count: number;
  log_count: number;
  trace_count: number;
  throughput_rps?: number;
}

interface GroupOverviewProps {
  group: Group;
  metrics?: GroupMetrics;
}

export function GroupOverview({ group, metrics }: GroupOverviewProps) {
  const [isRestarting, setIsRestarting] = useState(false);
  const [restartMessage, setRestartMessage] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  const handleRestart = async () => {
    setIsRestarting(true);
    setRestartMessage(null);
    try {
      const response = await restartGroup(group.id);
      if (response.success) {
        const summary = `${response.message} (${response.restarted_count} restarted, ${response.failed_count} failed)`;
        setRestartMessage({ type: "success", text: summary });
        // Clear message after 7 seconds
        setTimeout(() => setRestartMessage(null), 7000);
      } else {
        setRestartMessage({ type: "error", text: response.message });
      }
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : "Failed to restart group";
      setRestartMessage({ type: "error", text: errorMessage });
    } finally {
      setIsRestarting(false);
    }
  };
  const metricsItems = metrics
    ? [
        {
          label: "Agents",
          value: <span className="font-semibold">{metrics.agent_count}</span>,
        },
        {
          label: "Metrics",
          value: <span className="font-semibold">{metrics.metric_count}</span>,
        },
        {
          label: "Logs",
          value: <span className="font-semibold">{metrics.log_count}</span>,
        },
        {
          label: "Traces",
          value: <span className="font-semibold">{metrics.trace_count}</span>,
        },
        ...(metrics.throughput_rps !== undefined
          ? [
              {
                label: "Throughput",
                value: (
                  <span className="font-semibold">
                    {metrics.throughput_rps.toFixed(2)} rps
                  </span>
                ),
              },
            ]
          : []),
      ]
    : [];

  return (
    <div className="space-y-4">
      {restartMessage && (
        <Alert variant={restartMessage.type === "error" ? "destructive" : "default"}>
          <AlertDescription>{restartMessage.text}</AlertDescription>
        </Alert>
      )}

      <div className="flex items-center justify-between">
        <InfoCard
          title="Group Information"
          items={[
            { label: "ID", value: <span className="font-mono">{group.id}</span> },
            { label: "Name", value: group.name },
            {
              label: "Created",
              value: new Date(group.created_at).toLocaleString(),
            },
            {
              label: "Updated",
              value: new Date(group.updated_at).toLocaleString(),
            },
          ]}
        />
        {metrics && metrics.agent_count > 0 && (
          <Button
            variant="outline"
            size="sm"
            onClick={handleRestart}
            disabled={isRestarting}
            className="ml-4"
          >
            <RotateCw className={`h-4 w-4 mr-2 ${isRestarting ? "animate-spin" : ""}`} />
            {isRestarting ? "Restarting..." : "Restart All"}
          </Button>
        )}
      </div>

      {metrics && (
        <InfoCard title="Group Stats (Last 5 min)" items={metricsItems} />
      )}

      {group.labels && Object.keys(group.labels).length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Labels</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap gap-2">
              {Object.entries(group.labels).map(([key, value]) => (
                <Badge key={key} variant="outline">
                  {key}={String(value)}
                </Badge>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
