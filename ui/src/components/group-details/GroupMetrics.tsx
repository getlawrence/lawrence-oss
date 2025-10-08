import { TelemetryMetricsView } from "@/components/shared/TelemetryMetricsView";

interface GroupMetricsProps {
  groupId: string;
}

export function GroupMetrics({ groupId }: GroupMetricsProps) {
  return <TelemetryMetricsView groupId={groupId} showAgentId />;
}
