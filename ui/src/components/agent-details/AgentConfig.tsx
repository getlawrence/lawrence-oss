import { useState } from "react";
import useSWR from "swr";

import { getConfigs } from "@/api/configs";
import { getPipelineMetrics } from "@/api/collector-pipeline";
import { ConfigYamlEditorWithMetrics } from "@/components/configs/ConfigYamlEditorWithMetrics";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { TimeRangeSelect } from "@/components/ui/time-range-select";
import { LoadingSpinner } from "@/components/ui/loading-spinner";
import { type TimeRange, DEFAULT_TIME_RANGE } from "@/types/timeRange";

interface AgentConfigProps {
  agentId: string;
  effectiveConfig?: string; // Pass effective config from agent object
}

export function AgentConfig({ agentId, effectiveConfig }: AgentConfigProps) {
  const [timeRange, setTimeRange] = useState<TimeRange>(DEFAULT_TIME_RANGE);

  // Fetch agent's config from configs API as fallback if effective config not provided
  const { data: configsData, isLoading: configLoading } = useSWR(
    effectiveConfig ? null : `agent-config-${agentId}`,
    async () => {
      const result = await getConfigs({ agent_id: agentId, limit: 1 });
      return result;
    }
  );

  // Fetch pipeline metrics for the agent
  const { data: metricsData, isLoading: metricsLoading } = useSWR(
    `agent-pipeline-metrics-${agentId}-${timeRange}`,
    async () => {
      try {
        const result = await getPipelineMetrics(agentId, timeRange);
        return result;
      } catch (error) {
        console.error("Failed to fetch pipeline metrics:", error);
        return null;
      }
    },
    {
      refreshInterval: 5000, // Refresh every 5 seconds
    }
  );

  // Use effective config from agent if available, otherwise fallback to configs API
  const configContent = effectiveConfig || configsData?.configs?.[0]?.content;
  const metrics = metricsData?.components || [];

  if (!effectiveConfig && configLoading) {
    return <LoadingSpinner />;
  }

  if (!configContent) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="text-base">No Configuration</CardTitle>
          <CardDescription>
            This agent does not have a configuration assigned yet.
          </CardDescription>
        </CardHeader>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <div>
          <h3 className="text-sm font-medium text-gray-700">Time Range</h3>
          <p className="text-xs text-gray-500">
            Metrics shown for the selected time range
          </p>
        </div>
        <TimeRangeSelect
          value={timeRange}
          onValueChange={setTimeRange}
          useShortLabels
          maxRange="1h"
        />
      </div>

      {metricsLoading && metrics.length === 0 ? (
        <div className="text-sm text-gray-500 text-center py-2">
          Loading metrics...
        </div>
      ) : null}

      <ConfigYamlEditorWithMetrics
        value={configContent}
        onChange={() => {}} // Read-only in this view
        metrics={metrics}
        readonly={true}
      />

      {metrics.length === 0 && !metricsLoading && (
        <Card className="border-yellow-200 bg-yellow-50">
          <CardContent className="pt-4">
            <p className="text-sm text-yellow-800">
              No metrics available yet. Metrics will appear once the agent starts
              sending telemetry data.
            </p>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

