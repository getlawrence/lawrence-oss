import {
  RefreshCw,
  Server,
  AlertCircle,
  CheckCircle,
  XCircle,
} from "lucide-react";
import { useState } from "react";
import useSWR from "swr";

import { getAgents } from "@/api/agents";
import { AgentDetailsDrawer } from "@/components/AgentDetailsDrawer";
import { GroupDetailsDrawer } from "@/components/GroupDetailsDrawer";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

export default function AgentsPage() {
  const [refreshing, setRefreshing] = useState(false);
  const [selectedAgentId, setSelectedAgentId] = useState<string | null>(null);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [selectedGroupId, setSelectedGroupId] = useState<string | null>(null);
  const [groupDrawerOpen, setGroupDrawerOpen] = useState(false);

  const {
    data: agentsData,
    error: agentsError,
    mutate: mutateAgents,
  } = useSWR("agents", getAgents, { refreshInterval: 30000 });

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

  const agents = agentsData?.agents ? Object.values(agentsData.agents) : [];

  return (
    <div className="container mx-auto p-6 space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">Agents</h1>
          <p className="text-gray-600">
            Manage and monitor your OpenTelemetry agents
          </p>
        </div>
        <Button onClick={handleRefresh} disabled={refreshing}>
          <RefreshCw
            className={`h-4 w-4 mr-2 ${refreshing ? "animate-spin" : ""}`}
          />
          Refresh
        </Button>
      </div>

      {/* Agents Table */}
      <Card>
        <CardHeader>
          <CardTitle>Agents ({agents.length})</CardTitle>
          <CardDescription>
            All registered OpenTelemetry agents and their current status
          </CardDescription>
        </CardHeader>
        <CardContent>
          {agents.length === 0 ? (
            <div className="text-center py-8">
              <Server className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-gray-900 mb-2">
                No Agents Found
              </h3>
              <p className="text-gray-600">
                No agents are currently registered. Agents will appear here once
                they connect.
              </p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Status</TableHead>
                  <TableHead>Name</TableHead>
                  <TableHead>Version</TableHead>
                  <TableHead>Group</TableHead>
                  <TableHead>Last Seen</TableHead>
                  <TableHead>Labels</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {agents.map((agent) => (
                  <TableRow
                    key={agent.id}
                    onClick={() => handleAgentClick(agent.id)}
                    className="cursor-pointer hover:bg-muted/50"
                  >
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
                            agent.group_id &&
                            handleGroupClick(agent.group_id, e)
                          }
                          className="text-blue-600 hover:text-blue-800 cursor-pointer underline"
                        >
                          {agent.group_id}
                        </span>
                      ) : (
                        "No Group"
                      )}
                    </TableCell>
                    <TableCell>
                      {new Date(agent.last_seen).toLocaleString()}
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-1">
                        {Object.entries(agent.labels)
                          .slice(0, 2)
                          .map(([key, value]) => (
                            <Badge
                              key={key}
                              variant="outline"
                              className="text-xs"
                            >
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
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

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
  );
}
