import { RotateCw } from "lucide-react";
import { useState } from "react";

import { restartAgent } from "@/api/agents";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { InfoCard } from "@/components/ui/info-card";

interface Agent {
  id: string;
  name: string;
  version: string;
  status: string;
  group_name?: string;
  last_seen: string;
  labels?: Record<string, string | number | boolean>;
  capabilities?: string[];
}

interface Metrics {
  metric_count: number;
  log_count: number;
  trace_count: number;
  throughput_rps: number;
}

interface AgentOverviewProps {
  agent: Agent;
  metrics?: Metrics;
}

export function AgentOverview({ agent, metrics }: AgentOverviewProps) {
  const [isRestarting, setIsRestarting] = useState(false);
  const [restartMessage, setRestartMessage] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  // Check if agent supports restart command
  const supportsRestart =
    agent.capabilities?.includes("accepts_restart_command") ?? false;

  const handleRestart = async () => {
    setIsRestarting(true);
    setRestartMessage(null);
    try {
      const response = await restartAgent(agent.id);
      if (response.success) {
        setRestartMessage({ type: "success", text: response.message });
        // Clear message after 5 seconds
        setTimeout(() => setRestartMessage(null), 5000);
      } else {
        setRestartMessage({ type: "error", text: response.message });
      }
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : "Failed to restart agent";
      setRestartMessage({ type: "error", text: errorMessage });
    } finally {
      setIsRestarting(false);
    }
  };

  return (
    <div className="space-y-4">
      {restartMessage && (
        <Alert
          variant={restartMessage.type === "error" ? "destructive" : "default"}
        >
          <AlertDescription>{restartMessage.text}</AlertDescription>
        </Alert>
      )}

      <div className="flex items-center justify-between">
        <InfoCard
          title="Agent Information"
          items={[
            {
              label: "ID",
              value: <span className="font-mono">{agent.id}</span>,
            },
            { label: "Version", value: agent.version },
            {
              label: "Status",
              value: (
                <Badge
                  variant={agent.status === "online" ? "default" : "secondary"}
                >
                  {agent.status}
                </Badge>
              ),
            },
            { label: "Group", value: agent.group_name || "No Group" },
            {
              label: "Last Seen",
              value: new Date(agent.last_seen).toLocaleString(),
            },
          ]}
        />
        {supportsRestart && (
          <Button
            variant="outline"
            size="sm"
            onClick={handleRestart}
            disabled={agent.status !== "online" || isRestarting}
            className="ml-4"
          >
            <RotateCw
              className={`h-4 w-4 mr-2 ${isRestarting ? "animate-spin" : ""}`}
            />
            {isRestarting ? "Restarting..." : "Restart"}
          </Button>
        )}
      </div>

      {metrics && (
        <InfoCard
          title="Telemetry Stats (Last 5 min)"
          items={[
            {
              label: "Metrics",
              value: (
                <span className="font-semibold">{metrics.metric_count}</span>
              ),
            },
            {
              label: "Logs",
              value: <span className="font-semibold">{metrics.log_count}</span>,
            },
            {
              label: "Traces",
              value: (
                <span className="font-semibold">{metrics.trace_count}</span>
              ),
            },
            {
              label: "Throughput",
              value: (
                <span className="font-semibold">
                  {metrics.throughput_rps.toFixed(2)} rps
                </span>
              ),
            },
          ]}
        />
      )}

      {agent.capabilities && agent.capabilities.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Capabilities</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap gap-2">
              {agent.capabilities.map((capability) => (
                <Badge key={capability} variant="secondary">
                  {capability}
                </Badge>
              ))}
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
    </div>
  );
}
