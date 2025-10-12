import { CheckCircle, RotateCw, XCircle } from "lucide-react";
import { useState } from "react";
import useSWR from "swr";

import { getGroup, restartGroup } from "@/api/groups";
import { getGroupTopology } from "@/api/topology";
import { AgentDetailsDrawer } from "@/components/AgentDetailsDrawer";
import { GroupAgents } from "@/components/group-details/GroupAgents";
import { GroupLogs } from "@/components/group-details/GroupLogs";
import { GroupMetrics } from "@/components/group-details/GroupMetrics";
import { GroupOverview } from "@/components/group-details/GroupOverview";
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

interface GroupDetailsDrawerProps {
  groupId: string | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function GroupDetailsDrawer({
  groupId,
  open,
  onOpenChange,
}: GroupDetailsDrawerProps) {
  const [selectedAgentId, setSelectedAgentId] = useState<string | null>(null);
  const [agentDrawerOpen, setAgentDrawerOpen] = useState(false);
  const [isRestarting, setIsRestarting] = useState(false);
  const [restartMessage, setRestartMessage] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  const { data: group, isLoading: groupLoading } = useSWR(
    groupId && open ? `group-${groupId}` : null,
    () => (groupId ? getGroup(groupId) : null),
  );

  const { data: groupTopology } = useSWR(
    groupId && open ? `group-topology-${groupId}` : null,
    () => (groupId ? getGroupTopology(groupId) : null),
  );

  const getStatusIcon = (agentCount?: number) => {
    if (!agentCount || agentCount === 0) {
      return <XCircle className="h-5 w-5 text-gray-500" />;
    }
    return <CheckCircle className="h-5 w-5 text-green-500" />;
  };

  const handleAgentClick = (agentId: string) => {
    setSelectedAgentId(agentId);
    setAgentDrawerOpen(true);
  };

  const handleRestart = async () => {
    if (!groupId) return;
    setIsRestarting(true);
    setRestartMessage(null);
    try {
      const response = await restartGroup(groupId);
      if (response.success) {
        const summary = `${response.message} (${response.restarted_count} restarted, ${response.failed_count} failed)`;
        setRestartMessage({ type: "success", text: summary });
        // Clear message after 7 seconds
        setTimeout(() => setRestartMessage(null), 7000);
      } else {
        setRestartMessage({ type: "error", text: response.message });
      }
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : "Failed to restart group";
      setRestartMessage({ type: "error", text: errorMessage });
    } finally {
      setIsRestarting(false);
    }
  };

  const metrics = groupTopology
    ? {
        agent_count: groupTopology.agent_count || 0,
        metric_count: groupTopology.metric_count || 0,
        log_count: 0,
        trace_count: 0,
      }
    : undefined;

  return (
    <>
      <Sheet open={open} onOpenChange={onOpenChange}>
        <SheetContent className="w-full min-w[70vw] sm:max-w-2xl overflow-y-auto px-6">
          <SheetHeader>
            <div className="flex items-center justify-between">
              <div className="flex-1">
                <SheetTitle className="flex items-center gap-2">
                  {group && getStatusIcon(groupTopology?.agent_count)}
                  Group Details
                </SheetTitle>
                <SheetDescription>
                  {group?.name || "Loading..."}
                </SheetDescription>
              </div>
              {group && groupTopology && groupTopology.agent_count > 0 && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleRestart}
                  disabled={isRestarting}
                >
                  <RotateCw
                    className={`h-4 w-4 mr-2 ${isRestarting ? "animate-spin" : ""}`}
                  />
                  {isRestarting ? "Restarting..." : "Restart All"}
                </Button>
              )}
            </div>
          </SheetHeader>

          {groupLoading ? (
            <LoadingSpinner />
          ) : group ? (
            <Tabs defaultValue="overview" className="mt-6">
              <TabsList className="grid w-full grid-cols-4">
                <TabsTrigger value="overview">Overview</TabsTrigger>
                <TabsTrigger value="agents">Agents</TabsTrigger>
                <TabsTrigger value="metrics">Metrics</TabsTrigger>
                <TabsTrigger value="logs">Logs</TabsTrigger>
              </TabsList>

              <TabsContent value="overview" className="space-y-4">
                <GroupOverview
                  group={group}
                  metrics={metrics}
                  restartMessage={restartMessage}
                />
              </TabsContent>

              <TabsContent value="agents" className="space-y-4">
                {groupId && (
                  <GroupAgents
                    groupId={groupId}
                    onAgentClick={handleAgentClick}
                  />
                )}
              </TabsContent>

              <TabsContent value="metrics" className="space-y-4">
                {groupId && <GroupMetrics groupId={groupId} />}
              </TabsContent>

              <TabsContent value="logs" className="space-y-4">
                {groupId && <GroupLogs groupId={groupId} />}
              </TabsContent>
            </Tabs>
          ) : null}
        </SheetContent>
      </Sheet>

      <AgentDetailsDrawer
        agentId={selectedAgentId}
        open={agentDrawerOpen}
        onOpenChange={setAgentDrawerOpen}
      />
    </>
  );
}
