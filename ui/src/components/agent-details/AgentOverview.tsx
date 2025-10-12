import { Badge } from "@/components/ui/badge";
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
  return (
    <div className="space-y-4">
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
            <CardTitle className="text-lg font-semibold">
              Capabilities
            </CardTitle>
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
            <CardTitle className="text-lg font-semibold">Labels</CardTitle>
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
