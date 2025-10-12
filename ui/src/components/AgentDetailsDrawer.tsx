import {
  CheckCircle,
  XCircle,
  AlertCircle,
  Server,
  RotateCw,
} from "lucide-react";
import { useState } from "react";
import useSWR from "swr";

import { restartAgent } from "@/api/agents";
import { getAgentTopology } from "@/api/topology";
import { AgentConfigPipeline } from "@/components/agent-details/AgentConfigPipeline";
import { AgentLogs } from "@/components/agent-details/AgentLogs";
import { AgentMetrics } from "@/components/agent-details/AgentMetrics";
import { AgentOverview } from "@/components/agent-details/AgentOverview";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { LoadingSpinner } from "@/components/ui/loading-spinner";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

interface AgentDetailsDrawerProps {
  agentId: string | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function AgentDetailsDrawer({
  agentId,
  open,
  onOpenChange,
}: AgentDetailsDrawerProps) {
  const [isRestarting, setIsRestarting] = useState(false);
  const [restartMessage, setRestartMessage] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  const { data: agentTopology, isLoading } = useSWR(
    agentId && open ? `agent-topology-${agentId}` : null,
    () => (agentId ? getAgentTopology(agentId) : null),
  );

  const agent = agentTopology?.agent;
  const metrics = agentTopology?.metrics;

  // Check if agent supports restart command
  const supportsRestart =
    agent?.capabilities?.includes("accepts_restart_command") ?? false;

  const handleRestart = async () => {
    if (!agentId) return;

    setIsRestarting(true);
    setRestartMessage(null);
    try {
      const response = await restartAgent(agentId);
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

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "online":
        return <CheckCircle className="h-5 w-5 text-green-500" />;
      case "offline":
        return <XCircle className="h-5 w-5 text-gray-500" />;
      case "error":
        return <AlertCircle className="h-5 w-5 text-red-500" />;
      default:
        return <Server className="h-5 w-5 text-gray-400" />;
    }
  };

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="w-full min-w[70vw] sm:max-w-2xl overflow-y-auto px-6">
        <SheetHeader>
          <div className="flex items-center justify-between">
            <div className="flex-1">
              <SheetTitle className="flex items-center gap-2">
                {agent && getStatusIcon(agent.status)}
                Agent Details
              </SheetTitle>
              <SheetDescription>{agent?.name || "Loading..."}</SheetDescription>
            </div>
            {agent && supportsRestart && (
              <Button
                variant="outline"
                size="sm"
                onClick={handleRestart}
                disabled={agent.status !== "online" || isRestarting}
              >
                <RotateCw
                  className={`h-4 w-4 mr-2 ${isRestarting ? "animate-spin" : ""}`}
                />
                {isRestarting ? "Restarting..." : "Restart"}
              </Button>
            )}
          </div>
        </SheetHeader>

        {restartMessage && (
          <Alert
            variant={
              restartMessage.type === "error" ? "destructive" : "default"
            }
            className="mt-4"
          >
            <AlertDescription>{restartMessage.text}</AlertDescription>
          </Alert>
        )}

        {isLoading ? (
          <div className="flex justify-center items-center py-8">
            <LoadingSpinner />
          </div>
        ) : agent ? (
          <Tabs defaultValue="overview" className="mt-6">
            <TabsList className="grid w-full grid-cols-4">
              <TabsTrigger value="overview">Overview</TabsTrigger>
              <TabsTrigger value="config">Config</TabsTrigger>
              <TabsTrigger value="metrics">Metrics</TabsTrigger>
              <TabsTrigger value="logs">Logs</TabsTrigger>
            </TabsList>

            <TabsContent value="overview" className="space-y-4 py-4">
              <AgentOverview agent={agent} metrics={metrics} />
            </TabsContent>

            <TabsContent value="config" className="space-y-4 py-4">
              {agentId && (
                <AgentConfigPipeline
                  agentId={agentId}
                  effectiveConfig={agent?.effective_config}
                  agent={agent}
                  agentName={agent?.name}
                />
              )}
            </TabsContent>

            <TabsContent value="metrics" className="space-y-4 py-4">
              {agentId && <AgentMetrics agentId={agentId} />}
            </TabsContent>

            <TabsContent value="logs" className="space-y-4 py-4">
              {agentId && <AgentLogs agentId={agentId} />}
            </TabsContent>
          </Tabs>
        ) : null}
      </SheetContent>
    </Sheet>
  );
}
