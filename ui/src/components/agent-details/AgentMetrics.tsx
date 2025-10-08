import { TelemetryMetricsView } from "@/components/shared/TelemetryMetricsView";

interface AgentMetricsProps {
  agentId: string;
}

export function AgentMetrics({ agentId }: AgentMetricsProps) {
  return <TelemetryMetricsView agentId={agentId} />;
}
