import { useState } from "react";
import useSWR from "swr";

import { type Agent } from "@/api/agents";
import { getPipelineMetrics } from "@/api/collector-pipeline";
import { getConfigs } from "@/api/configs";
import { sendConfigToAgent } from "@/api/agents";
import { ConfigYamlEditorWithMetrics } from "@/components/configs/ConfigYamlEditorWithMetrics";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { LoadingSpinner } from "@/components/ui/loading-spinner";
import { TimeRangeSelect } from "@/components/ui/time-range-select";
import { type TimeRange, DEFAULT_TIME_RANGE } from "@/types/timeRange";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Settings, AlertTriangle } from "lucide-react";

interface AgentConfigProps {
  agentId: string;
  effectiveConfig?: string; // Pass effective config from agent object
  agent?: Agent; // Pass full agent object to access capabilities
}

export function AgentConfig({ agentId, effectiveConfig, agent }: AgentConfigProps) {
  const [timeRange, setTimeRange] = useState<TimeRange>(DEFAULT_TIME_RANGE);
  const [isEditing, setIsEditing] = useState(false);
  const [editedConfig, setEditedConfig] = useState("");
  const [isSending, setIsSending] = useState(false);

  // Check if agent supports remote config
  const supportsRemoteConfig = agent?.capabilities?.includes("accepts_remote_config") ?? false;

  // Fetch agent's config from configs API as fallback if effective config not provided
  const { data: configsData, isLoading: configLoading } = useSWR(
    effectiveConfig ? null : `agent-config-${agentId}`,
    async () => {
      const result = await getConfigs({ agent_id: agentId, limit: 1 });
      return result;
    },
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
    },
  );

  // Use effective config from agent if available, otherwise fallback to configs API
  const configContent = effectiveConfig || configsData?.configs?.[0]?.content;
  const metrics = metricsData?.components || [];

  // Initialize edited config when entering edit mode
  const handleStartEdit = () => {
    setEditedConfig(configContent || "");
    setIsEditing(true);
  };

  const handleCancelEdit = () => {
    setIsEditing(false);
    setEditedConfig("");
  };

  const handleSendConfig = async () => {
    if (!supportsRemoteConfig) {

      return;
    }

    setIsSending(true);
    try {
      const response = await sendConfigToAgent(agentId, editedConfig);
      if (response.success) {
        setIsEditing(false);
        // TODO: Optionally refresh agent data to get updated effective config
      } else {
      }
    } catch (error) {
      console.error("Failed to send config to agent:", error);
    } finally {
      setIsSending(false);
    }
  };

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
      {/* Header with capability status and action buttons */}
      <div className="flex justify-between items-start">
        <div className="flex-1">
          <div className="flex items-center gap-2 mb-1">
            <h3 className="text-sm font-medium text-gray-700">Configuration</h3>
            {supportsRemoteConfig ? (
              <Badge variant="outline" className="bg-green-50 text-green-700 border-green-200">
                <Settings className="w-3 h-3 mr-1" />
                Remote Config
              </Badge>
            ) : (
              <Badge variant="outline" className="bg-amber-50 text-amber-700 border-amber-200">
                <AlertTriangle className="w-3 h-3 mr-1" />
                Read-only
              </Badge>
            )}
          </div>
          {supportsRemoteConfig ? (
            <p className="text-xs text-gray-500">
              Edit and send configuration to agent (queued if offline)
            </p>
          ) : (
            <p className="text-xs text-gray-500">
              Agent does not support remote configuration
            </p>
          )}
        </div>

        <div className="flex gap-2">
          {!isEditing && supportsRemoteConfig && (
            <Button
              variant="outline"
              size="sm"
              onClick={handleStartEdit}
            >
              Edit Config
            </Button>
          )}

          {isEditing && (
            <>
              <Button
                variant="outline"
                size="sm"
                onClick={handleCancelEdit}
                disabled={isSending}
              >
                Cancel
              </Button>
              <Button
                size="sm"
                onClick={handleSendConfig}
                disabled={isSending}
              >
                {isSending ? "Sending..." : "Send to Agent"}
              </Button>
            </>
          )}
        </div>
      </div>

      {/* Time Range Selector */}
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

      {/* Config Editor */}
      <ConfigYamlEditorWithMetrics
        value={isEditing ? editedConfig : configContent}
        onChange={setEditedConfig}
        metrics={metrics}
        readonly={!isEditing}
      />
    </div>
  );
}
