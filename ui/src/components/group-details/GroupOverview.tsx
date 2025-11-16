import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { InfoCard } from "@/components/ui/info-card";

interface Group {
  id: string;
  name: string;
  labels?: Record<string, string> | null;
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
  restartMessage?: {
    type: "success" | "error";
    text: string;
  } | null;
}

export function GroupOverview({
  group,
  metrics,
  restartMessage,
}: GroupOverviewProps) {
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
        <Alert
          variant={restartMessage.type === "error" ? "destructive" : "default"}
        >
          <AlertDescription>{restartMessage.text}</AlertDescription>
        </Alert>
      )}

      <InfoCard
        title="Group Information"
        items={[
          {
            label: "ID",
            value: <span className="font-mono">{group.id}</span>,
          },
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
