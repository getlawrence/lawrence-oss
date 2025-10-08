import { type Node } from "@xyflow/react";
import { RefreshCw } from "lucide-react";
import { useState, useCallback } from "react";
import useSWR from "swr";

import { getTopology } from "@/api/topology";
import { AgentDetailsDrawer } from "@/components/AgentDetailsDrawer";
import { GroupDetailsDrawer } from "@/components/GroupDetailsDrawer";
import {
  TopologyHeader,
  TopologyCanvas,
  DisplaySidebar,
  useTopologyLayout,
  type TopologyNode,
} from "@/components/topology";
import { Button } from "@/components/ui/button";

export default function TopologyPage() {
  const [refreshing, setRefreshing] = useState(false);
  const [selectedAgentId, setSelectedAgentId] = useState<string | null>(null);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [selectedGroupId, setSelectedGroupId] = useState<string | null>(null);
  const [groupDrawerOpen, setGroupDrawerOpen] = useState(false);
  const [topologyLevel, setTopologyLevel] = useState<"instance" | "group">(
    "instance",
  );

  const {
    data: topologyData,
    error,
    mutate,
  } = useSWR("topology", getTopology, { refreshInterval: 30000 });

  const { nodes, edges } = useTopologyLayout(topologyData, topologyLevel);

  const handleRefresh = async () => {
    setRefreshing(true);
    await mutate();
    setRefreshing(false);
  };

  const onNodeClick = useCallback((_event: React.MouseEvent, node: Node) => {
    if (node.type === "agent") {
      const agentId = node.id.replace("agent-", "");
      setSelectedAgentId(agentId);
      setDrawerOpen(true);
    } else if (node.type === "group") {
      const groupId = node.id.replace("group-", "");
      setSelectedGroupId(groupId);
      setGroupDrawerOpen(true);
    }
  }, []);

  const handleNodeSelect = (node: TopologyNode) => {
    if (node.type === "agent") {
      setSelectedAgentId(node.id);
      setDrawerOpen(true);
    } else if (node.type === "group") {
      setSelectedGroupId(node.id);
      setGroupDrawerOpen(true);
    }
  };

  if (error) {
    return (
      <div className="container mx-auto p-6">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-red-600 mb-4">
            Error Loading Topology
          </h1>
          <p className="text-gray-600">{error.message}</p>
          <Button onClick={handleRefresh} className="mt-4">
            <RefreshCw className="h-4 w-4 mr-2" />
            Retry
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="h-full w-full flex flex-col -m-4">
      <TopologyHeader
        topologyLevel={topologyLevel}
        onTopologyLevelChange={setTopologyLevel}
        onRefresh={handleRefresh}
        refreshing={refreshing}
      />

      <div className="flex-1 flex min-h-0">
        <DisplaySidebar
          onNodeSelect={handleNodeSelect}
          className="w-80 flex-shrink-0"
        />

        {topologyData && (
          <TopologyCanvas
            nodes={nodes}
            edges={edges}
            onNodeClick={onNodeClick}
          />
        )}

        <AgentDetailsDrawer
          agentId={selectedAgentId}
          open={drawerOpen}
          onOpenChange={setDrawerOpen}
        />

        <GroupDetailsDrawer
          groupId={selectedGroupId}
          open={groupDrawerOpen}
          onOpenChange={setGroupDrawerOpen}
        />
      </div>
    </div>
  );
}
