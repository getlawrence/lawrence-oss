import {
  RefreshCw,
  Server,
  AlertCircle,
  CheckCircle,
  XCircle,
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
} from "lucide-react";
import { useState, useMemo } from "react";
import useSWR from "swr";

import { getAgents } from "@/api/agents";
import { getGroups } from "@/api/groups";
import { AgentDetailsDrawer } from "@/components/AgentDetailsDrawer";
import { GroupDetailsDrawer } from "@/components/GroupDetailsDrawer";
import { PageTable } from "@/components/shared/PageTable";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { TableCell } from "@/components/ui/table";
import type { Agent } from "@/types/agent";

type SortOrder = "asc" | "desc" | null;

export default function AgentsPage() {
  const [refreshing, setRefreshing] = useState(false);
  const [selectedAgentId, setSelectedAgentId] = useState<string | null>(null);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [selectedGroupId, setSelectedGroupId] = useState<string | null>(null);
  const [groupDrawerOpen, setGroupDrawerOpen] = useState(false);
  const [sortOrder, setSortOrder] = useState<SortOrder>("desc");
  const [showActiveOnly, setShowActiveOnly] = useState(false);

  const {
    data: agentsData,
    error: agentsError,
    mutate: mutateAgents,
  } = useSWR("agents", getAgents, { refreshInterval: 30000 });

  const { data: groupsData } = useSWR("groups", getGroups, {
    refreshInterval: 30000,
  });

  const handleRefresh = async () => {
    setRefreshing(true);
    await mutateAgents();
    setRefreshing(false);
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "online":
        return <CheckCircle className="h-4 w-4 text-green-500" />;
      case "offline":
        return <XCircle className="h-4 w-4 text-gray-500" />;
      case "error":
        return <AlertCircle className="h-4 w-4 text-red-500" />;
      default:
        return <Server className="h-4 w-4 text-gray-500" />;
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "online":
        return (
          <Badge variant="default" className="bg-green-100 text-green-800">
            Online
          </Badge>
        );
      case "offline":
        return <Badge variant="secondary">Offline</Badge>;
      case "error":
        return <Badge variant="destructive">Error</Badge>;
      default:
        return <Badge variant="outline">Unknown</Badge>;
    }
  };

  const handleAgentClick = (agentId: string) => {
    setSelectedAgentId(agentId);
    setDrawerOpen(true);
  };

  const handleGroupClick = (groupId: string, e: React.MouseEvent) => {
    e.stopPropagation();
    setSelectedGroupId(groupId);
    setGroupDrawerOpen(true);
  };

  const handleSortToggle = () => {
    setSortOrder((prev) => {
      if (prev === null) return "desc";
      if (prev === "desc") return "asc";
      return "desc";
    });
  };

  const getSortIcon = () => {
    if (sortOrder === "desc") return <ArrowDown className="h-4 w-4" />;
    if (sortOrder === "asc") return <ArrowUp className="h-4 w-4" />;
    return <ArrowUpDown className="h-4 w-4" />;
  };

  // Create a map of group IDs to group names
  const groupIdToName = useMemo(() => {
    const map: Record<string, string> = {};
    if (groupsData?.groups) {
      groupsData.groups.forEach((group) => {
        map[group.id] = group.name;
      });
    }
    return map;
  }, [groupsData]);

  // Filter and sort agents
  const filteredAndSortedAgents = useMemo(() => {
    const allAgents = agentsData?.agents
      ? Object.values(agentsData.agents)
      : [];

    // Apply status filter
    let filtered = showActiveOnly
      ? allAgents.filter((agent) => agent.status === "online")
      : allAgents;

    // Apply sorting by last_seen
    if (sortOrder !== null) {
      filtered = [...filtered].sort((a, b) => {
        const dateA = new Date(a.last_seen).getTime();
        const dateB = new Date(b.last_seen).getTime();
        return sortOrder === "asc" ? dateA - dateB : dateB - dateA;
      });
    }

    return filtered;
  }, [agentsData, showActiveOnly, sortOrder]);

  if (agentsError) {
    return (
      <div className="container mx-auto p-6">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-red-600 mb-4">
            Error Loading Data
          </h1>
          <p className="text-gray-600">
            {agentsError?.message || "Failed to load agent data"}
          </p>
          <Button onClick={handleRefresh} className="mt-4">
            <RefreshCw className="h-4 w-4 mr-2" />
            Retry
          </Button>
        </div>
      </div>
    );
  }

  const allAgents = agentsData?.agents ? Object.values(agentsData.agents) : [];
  const agents = filteredAndSortedAgents;

  return (
    <>
      <PageTable
        pageTitle="Agents"
        pageDescription="Manage and monitor your OpenTelemetry agents"
        pageActions={[
          {
            label: "Refresh",
            icon: RefreshCw,
            onClick: handleRefresh,
            disabled: refreshing,
          },
        ]}
        cardTitle={`Agents (${agents.length})`}
        cardDescription="All registered OpenTelemetry agents and their current status"
        cardHeaderExtra={
          <div className="flex items-center space-x-2">
            <Switch
              id="active-only"
              checked={showActiveOnly}
              onCheckedChange={setShowActiveOnly}
            />
            <Label htmlFor="active-only" className="cursor-pointer">
              Show active only
            </Label>
          </div>
        }
        columns={[
          { header: "Status", key: "status" },
          { header: "Name", key: "name" },
          { header: "Version", key: "version" },
          { header: "Group", key: "group" },
          {
            header: (
              <Button
                variant="ghost"
                size="sm"
                onClick={handleSortToggle}
                className="h-8 px-2 -ml-2"
              >
                Last Seen
                {getSortIcon()}
              </Button>
            ),
            key: "last_seen",
          },
          { header: "Labels", key: "labels" },
        ]}
        data={agents}
        getRowKey={(agent: Agent) => agent.id}
        onRowClick={(agent: Agent) => handleAgentClick(agent.id)}
        renderRow={(agent: Agent) => (
          <>
            <TableCell>
              <div className="flex items-center space-x-2">
                {getStatusIcon(agent.status)}
                {getStatusBadge(agent.status)}
              </div>
            </TableCell>
            <TableCell className="font-medium">{agent.name}</TableCell>
            <TableCell>{agent.version}</TableCell>
            <TableCell>
              {agent.group_id ? (
                <span
                  onClick={(e) =>
                    agent.group_id && handleGroupClick(agent.group_id, e)
                  }
                  className="text-blue-600 hover:text-blue-800 cursor-pointer underline"
                >
                  {groupIdToName[agent.group_id] || agent.group_id}
                </span>
              ) : (
                "No Group"
              )}
            </TableCell>
            <TableCell>{new Date(agent.last_seen).toLocaleString()}</TableCell>
            <TableCell>
              <div className="flex flex-wrap gap-1">
                {Object.entries(agent.labels)
                  .slice(0, 2)
                  .map(([key, value]) => (
                    <Badge key={key} variant="outline" className="text-xs">
                      {key}={value}
                    </Badge>
                  ))}
                {Object.keys(agent.labels).length > 2 && (
                  <Badge variant="outline" className="text-xs">
                    +{Object.keys(agent.labels).length - 2} more
                  </Badge>
                )}
              </div>
            </TableCell>
          </>
        )}
        emptyState={{
          icon: Server,
          title:
            allAgents.length === 0 ? "No Agents Found" : "No Matching Agents",
          description:
            allAgents.length === 0
              ? "No agents are currently registered. Agents will appear here once they connect."
              : "No agents match the current filter criteria.",
        }}
      />

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
    </>
  );
}
