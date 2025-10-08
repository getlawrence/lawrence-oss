import {
  CheckCircle,
  XCircle,
  AlertCircle,
  Server,
} from "lucide-react";
import useSWR from "swr";

import { getGroupAgents } from "@/api/groups";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { LoadingSpinner } from "@/components/ui/loading-spinner";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

interface GroupAgentsProps {
  groupId: string;
  onAgentClick?: (agentId: string) => void;
}

export function GroupAgents({ groupId, onAgentClick }: GroupAgentsProps) {
  const { data: agentsData, isLoading } = useSWR(
    `group-agents-${groupId}`,
    () => getGroupAgents(groupId),
    {
      revalidateOnFocus: false,
      revalidateOnReconnect: false,
    }
  );

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

  if (isLoading) {
    return <LoadingSpinner />;
  }

  const agents = agentsData?.agents || [];

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Group Agents</CardTitle>
        <CardDescription>{agents.length} agents in this group</CardDescription>
      </CardHeader>
      <CardContent>
        {agents.length === 0 ? (
          <div className="text-center py-8 text-gray-500">
            No agents in this group
          </div>
        ) : (
          <ScrollArea className="h-96">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Status</TableHead>
                  <TableHead>Name</TableHead>
                  <TableHead>Version</TableHead>
                  <TableHead>Last Seen</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {agents.map((agent) => (
                  <TableRow
                    key={agent.id}
                    onClick={() => onAgentClick?.(agent.id)}
                    className={onAgentClick ? "cursor-pointer hover:bg-muted/50" : ""}
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
                      {new Date(agent.last_seen).toLocaleString()}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </ScrollArea>
        )}
      </CardContent>
    </Card>
  );
}

