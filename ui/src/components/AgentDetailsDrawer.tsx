import { CheckCircle, XCircle, AlertCircle, Server } from "lucide-react";
import useSWR from "swr";

import { getAgentTopology } from "@/api/topology";
import { AgentConfig } from "@/components/agent-details/AgentConfig";
import { AgentLogs } from "@/components/agent-details/AgentLogs";
import { AgentMetrics } from "@/components/agent-details/AgentMetrics";
import { AgentOverview } from "@/components/agent-details/AgentOverview";
import { CollectorPipelineView } from "@/components/collector-pipeline";
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
  const { data: agentTopology, isLoading } = useSWR(
    agentId && open ? `agent-topology-${agentId}` : null,
    () => (agentId ? getAgentTopology(agentId) : null),
  );

  const agent = agentTopology?.agent;
  const metrics = agentTopology?.metrics;

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
          <SheetTitle className="flex items-center gap-2">
            {agent && getStatusIcon(agent.status)}
            Agent Details
          </SheetTitle>
          <SheetDescription>{agent?.name || "Loading..."}</SheetDescription>
        </SheetHeader>

        {isLoading ? (
          <div className="flex justify-center items-center py-8">
            <LoadingSpinner />
          </div>
        ) : agent ? (
          <Tabs defaultValue="overview" className="mt-6">
            <TabsList className="grid w-full grid-cols-5">
              <TabsTrigger value="overview">Overview</TabsTrigger>
              <TabsTrigger value="config">Config</TabsTrigger>
              <TabsTrigger value="metrics">Metrics</TabsTrigger>
              <TabsTrigger value="logs">Logs</TabsTrigger>
              <TabsTrigger value="pipeline">Pipeline</TabsTrigger>
            </TabsList>

            <TabsContent value="overview" className="space-y-4 py-4">
              <AgentOverview agent={agent} metrics={metrics} />
            </TabsContent>

            <TabsContent value="config" className="space-y-4 py-4">
              {agentId && (
                <AgentConfig
                  agentId={agentId}
                  effectiveConfig={agent?.effective_config}
                  agent={agent}
                />
              )}
            </TabsContent>

            <TabsContent value="metrics" className="space-y-4 py-4">
              {agentId && <AgentMetrics agentId={agentId} />}
            </TabsContent>

            <TabsContent value="logs" className="space-y-4 py-4">
              {agentId && <AgentLogs agentId={agentId} />}
            </TabsContent>

            <TabsContent value="pipeline" className="space-y-4 py-4">
              {agentId ? (
                <div className="h-[600px]">
                  <CollectorPipelineView
                    agentId={agentId}
                    agentName={agent?.name}
                    effectiveConfig={agent?.effective_config}
                  />
                </div>
              ) : (
                <div className="text-center py-8 text-gray-500">
                  No agent selected
                </div>
              )}
            </TabsContent>
          </Tabs>
        ) : null}
      </SheetContent>
    </Sheet>
  );
}
