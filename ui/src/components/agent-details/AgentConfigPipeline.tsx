import { Code, Workflow, Settings, AlertTriangle } from "lucide-react";
import { useState } from "react";
import useSWR from "swr";

import { type Agent } from "@/api/agents";
import { sendConfigToAgent } from "@/api/agents";
import { fetchAgentComponentMetrics } from "@/api/collector-metrics";
import { getConfigs } from "@/api/configs";
import { CollectorPipelineView } from "@/components/collector-pipeline/CollectorPipelineView";
import { ConfigYamlEditorWithMetrics } from "@/components/configs/ConfigYamlEditorWithMetrics";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { LoadingSpinner } from "@/components/ui/loading-spinner";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { type TimeRange, DEFAULT_TIME_RANGE } from "@/types/timeRange";

interface AgentConfigPipelineProps {
  agentId: string;
  effectiveConfig?: string; // Pass effective config from agent object
  agent?: Agent; // Pass full agent object to access capabilities
  agentName?: string;
}

export function AgentConfigPipeline({
  agentId,
  effectiveConfig,
  agent,
  agentName,
}: AgentConfigPipelineProps) {
  const [timeRange] = useState<TimeRange>(DEFAULT_TIME_RANGE);
  const [editedConfig, setEditedConfig] = useState("");
  const [isSending, setIsSending] = useState(false);
  const [viewMode, setViewMode] = useState<"code" | "pipeline">("code");

  // Check if agent supports remote config
  const supportsRemoteConfig =
    agent?.capabilities?.includes("accepts_remote_config") ?? false;

  // Fetch agent's config from configs API as fallback if effective config not provided
  const { data: configsData, isLoading: configLoading } = useSWR(
    effectiveConfig ? null : `agent-config-${agentId}`,
    async () => {
      const result = await getConfigs({ agent_id: agentId, limit: 1 });
      return result;
    },
  );

  // Fetch component metrics using the telemetry metrics endpoint (only for code view)
  const { data: metricsData, isLoading: metricsLoading } = useSWR(
    viewMode === "code"
      ? `agent-component-metrics-${agentId}-${timeRange}`
      : null,
    async () => {
      try {
        // Convert timeRange to minutes
        const minutes = timeRangeToMinutes(timeRange);
        const result = await fetchAgentComponentMetrics(agentId, minutes);
        return result;
      } catch (error) {
        console.error("Failed to fetch component metrics:", error);
        return [];
      }
    },
    {
      refreshInterval: 5000, // Refresh every 5 seconds
    },
  );

  // Use effective config from agent if available, otherwise fallback to configs API
  const configContent = effectiveConfig || configsData?.configs?.[0]?.content;
  const metrics = metricsData || [];

  // Get the current config value (edited or original)
  const currentConfig = editedConfig || configContent || "";

  const handleSendConfig = async () => {
    if (!supportsRemoteConfig) {
      return;
    }

    setIsSending(true);
    try {
      const response = await sendConfigToAgent(agentId, currentConfig);
      if (response.success) {
        // TODO: Optionally refresh agent data to get updated effective config
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
      <Tabs
        value={viewMode}
        onValueChange={(value) => setViewMode(value as "code" | "pipeline")}
      >
        <div className="flex items-center justify-between gap-4">
          <TabsList>
            <TabsTrigger value="code" className="flex items-center gap-2">
              <Code className="h-4 w-4" />
              Code View
            </TabsTrigger>
            <TabsTrigger value="pipeline" className="flex items-center gap-2">
              <Workflow className="h-4 w-4" />
              Pipeline View
            </TabsTrigger>
          </TabsList>

          {/* Show controls only in Code View */}
          {viewMode === "code" && (
            <div className="flex items-center gap-2">
              {supportsRemoteConfig ? (
                <>
                  <Badge
                    variant="outline"
                    className="bg-green-50 text-green-700 border-green-200"
                  >
                    <Settings className="w-3 h-3 mr-1" />
                    Remote Config
                  </Badge>
                  <Button
                    size="sm"
                    onClick={handleSendConfig}
                    disabled={isSending}
                  >
                    {isSending ? "Sending..." : "Send to Agent"}
                  </Button>
                </>
              ) : (
                <Badge
                  variant="outline"
                  className="bg-amber-50 text-amber-700 border-amber-200"
                >
                  <AlertTriangle className="w-3 h-3 mr-1" />
                  Read-only
                </Badge>
              )}
            </div>
          )}
        </div>

        <TabsContent value="code" className="space-y-4 mt-4">
          {metricsLoading && metrics.length === 0 ? (
            <div className="text-sm text-gray-500 text-center py-2">
              Loading metrics...
            </div>
          ) : null}

          {/* Config Editor with Metrics */}
          <ConfigYamlEditorWithMetrics
            value={currentConfig}
            onChange={setEditedConfig}
            metrics={metrics}
            readonly={!supportsRemoteConfig}
          />
        </TabsContent>

        <TabsContent value="pipeline" className="mt-4">
          <div className="h-[600px]">
            <CollectorPipelineView
              agentId={agentId}
              agentName={agentName || agent?.name}
              effectiveConfig={configContent}
            />
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}

// Helper function to convert TimeRange to minutes
function timeRangeToMinutes(timeRange: TimeRange): number {
  switch (timeRange) {
    case "1m":
      return 1;
    case "5m":
      return 5;
    case "15m":
      return 15;
    case "1h":
      return 60;
    case "6h":
      return 360;
    case "24h":
      return 1440;
    default:
      return 5;
  }
}
