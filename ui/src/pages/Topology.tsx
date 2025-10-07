import { useState, useCallback, useMemo } from 'react';
import useSWR from 'swr';
import {
  ReactFlow,
  Background,
  type Node,
  type Edge,
  useNodesState,
  useEdgesState,
  MarkerType,
  Position,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { RefreshCw, Server, Users, CheckCircle, XCircle, AlertCircle } from 'lucide-react';
import { getTopology } from '@/api/topology';
import { AgentDetailsDrawer } from '@/components/AgentDetailsDrawer';
import { DisplaySidebar } from '@/components/topology/DisplaySidebar';

// Custom Node Component for Agents
const AgentNode = ({ data }: { data: any }) => {
  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'online':
        return <CheckCircle className="h-4 w-4 text-green-500" />;
      case 'offline':
        return <XCircle className="h-4 w-4 text-gray-500" />;
      case 'error':
        return <AlertCircle className="h-4 w-4 text-red-500" />;
      default:
        return <Server className="h-4 w-4 text-gray-400" />;
    }
  };

  return (
    <div className="bg-white border-2 border-blue-500 rounded-lg p-3 shadow-lg min-w-[200px]">
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2">
          <Server className="h-5 w-5 text-blue-500" />
          <span className="font-semibold text-sm">{data.name}</span>
        </div>
        {getStatusIcon(data.status)}
      </div>
      {data.metrics && (
        <div className="text-xs text-gray-600 space-y-1">
          <div>Metrics: {data.metrics.metric_count}</div>
          <div>Logs: {data.metrics.log_count}</div>
          <div>Throughput: {data.metrics.throughput_rps.toFixed(2)} rps</div>
        </div>
      )}
      {data.group_name && (
        <Badge variant="outline" className="mt-2 text-xs">
          {data.group_name}
        </Badge>
      )}
    </div>
  );
};

// Custom Node Component for Groups
const GroupNode = ({ data }: { data: any }) => {
  return (
    <div className="bg-purple-50 border-2 border-purple-500 rounded-lg p-4 shadow-lg min-w-[250px]">
      <div className="flex items-center gap-2 mb-2">
        <Users className="h-5 w-5 text-purple-500" />
        <span className="font-semibold">{data.name}</span>
      </div>
      <div className="text-xs text-gray-600">
        {data.agent_count || 0} agents
      </div>
    </div>
  );
};

const nodeTypes = {
  agent: AgentNode,
  group: GroupNode,
};

export default function TopologyPage() {
  const [refreshing, setRefreshing] = useState(false);
  const [selectedAgentId, setSelectedAgentId] = useState<string | null>(null);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [topologyLevel, setTopologyLevel] = useState<'instance' | 'group'>('instance');

  const { data: topologyData, error, mutate } = useSWR(
    'topology',
    getTopology,
    { refreshInterval: 30000 }
  );

  const handleRefresh = async () => {
    setRefreshing(true);
    await mutate();
    setRefreshing(false);
  };

  // Convert API data to React Flow nodes
  const initialNodes: Node[] = useMemo(() => {
    if (!topologyData?.nodes || !Array.isArray(topologyData.nodes) || topologyData.nodes.length === 0) return [];

    // Calculate layout positions
    const nodes = topologyData.nodes;
    const agentNodes = nodes.filter(n => n.type === 'agent');
    const groupNodes = nodes.filter(n => n.type === 'group');

    const flowNodes: Node[] = [];

    // Filter nodes based on topology level
    if (topologyLevel === 'group') {
      // Only show group nodes
      groupNodes.forEach((node, idx) => {
        flowNodes.push({
          id: node.id,
          type: 'group',
          data: {
            name: node.name,
            status: node.status,
            agent_count: topologyData.groups?.find(g => g.id === node.id.replace('group-', ''))?.agent_count || 0,
          },
          position: { x: idx * 300 + 50, y: 150 },
          sourcePosition: Position.Bottom,
          targetPosition: Position.Top,
        });
      });
      return flowNodes;
    }

    // Position groups in top row
    groupNodes.forEach((node, idx) => {
      flowNodes.push({
        id: node.id,
        type: 'group',
        data: {
          name: node.name,
          status: node.status,
          agent_count: topologyData.groups?.find(g => g.id === node.id.replace('group-', ''))?.agent_count || 0,
        },
        position: { x: idx * 300 + 50, y: 50 },
        sourcePosition: Position.Bottom,
        targetPosition: Position.Top,
      });
    });

    // Position agents in rows below groups
    const agentsPerRow = 4;
    agentNodes.forEach((node, idx) => {
      const row = Math.floor(idx / agentsPerRow);
      const col = idx % agentsPerRow;

      flowNodes.push({
        id: node.id,
        type: 'agent',
        data: {
          name: node.name,
          status: node.status,
          group_name: node.group_name,
          metrics: node.metrics,
          labels: node.labels,
        },
        position: { x: col * 250 + 50, y: row * 200 + 300 },
        sourcePosition: Position.Top,
        targetPosition: Position.Bottom,
      });
    });

    return flowNodes;
  }, [topologyData, topologyLevel]);

  const initialEdges: Edge[] = useMemo(() => {
    if (!topologyData?.edges || !Array.isArray(topologyData.edges)) return [];

    // Filter edges based on topology level
    const filteredEdges = topologyData.edges.filter(edge => {
      if (topologyLevel === 'group') {
        // Only show edges between groups
        return edge.source.startsWith('group-') && edge.target.startsWith('group-');
      }
      // Instance level: show all edges
      return true;
    });

    return filteredEdges.map((edge, idx) => ({
      id: `edge-${idx}`,
      source: edge.source,
      target: edge.target,
      type: 'smoothstep',
      animated: false,
      label: edge.label,
      markerEnd: {
        type: MarkerType.ArrowClosed,
        color: '#9333ea',
      },
      style: {
        stroke: '#9333ea',
        strokeWidth: 2,
      },
    }));
  }, [topologyData, topologyLevel]);

  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);

  // Update nodes and edges when data changes
  useMemo(() => {
    setNodes(initialNodes);
    setEdges(initialEdges);
  }, [initialNodes, initialEdges, setNodes, setEdges]);

  const onNodeClick = useCallback((_event: React.MouseEvent, node: Node) => {
    // Only open drawer for agent nodes
    if (node.type === 'agent') {
      const agentId = node.id.replace('agent-', '');
      setSelectedAgentId(agentId);
      setDrawerOpen(true);
    }
  }, []);

  if (error) {
    return (
      <div className="container mx-auto p-6">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-red-600 mb-4">Error Loading Topology</h1>
          <p className="text-gray-600">{error.message}</p>
          <Button onClick={handleRefresh} className="mt-4">
            <RefreshCw className="h-4 w-4 mr-2" />
            Retry
          </Button>
        </div>
      </div>
    );
  }

  const handleNodeSelect = (node: { id: string; type: 'group' | 'agent'; name: string; data: any }) => {
    if (node.type === 'agent') {
      setSelectedAgentId(node.id);
      setDrawerOpen(true);
    }
  };

  return (
    <div className="h-full w-full flex flex-col -m-4">
      {/* Header */}
      <div className="h-16 border-b bg-white px-4 flex items-center justify-between flex-shrink-0">
        <div>
          <h1 className="text-2xl font-bold">Topology</h1>
          <p className="text-sm text-gray-600">Visualize your agent infrastructure</p>
        </div>
        <div className="flex items-center gap-3">
          <Tabs value={topologyLevel} onValueChange={(value) => setTopologyLevel(value as 'instance' | 'group')}>
            <TabsList>
              <TabsTrigger value="instance">Instance</TabsTrigger>
              <TabsTrigger value="group">Group</TabsTrigger>
            </TabsList>
          </Tabs>
          <Button onClick={handleRefresh} disabled={refreshing} size="sm">
            <RefreshCw className={`h-4 w-4 mr-2 ${refreshing ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </div>
      </div>

      <div className="flex-1 flex min-h-0">
        {/* Display Sidebar */}
        <DisplaySidebar onNodeSelect={handleNodeSelect} className="w-80 flex-shrink-0" />

        {/* Main Content - Topology Canvas */}
        <div className="flex-1 relative bg-gray-100">
          {topologyData && (
            <ReactFlow
              nodes={nodes}
              edges={edges}
              onNodesChange={onNodesChange}
              onEdgesChange={onEdgesChange}
              onNodeClick={onNodeClick}
              nodeTypes={nodeTypes}
              fitView
              minZoom={0.2}
              maxZoom={2}
            >
              <Background />
            </ReactFlow>
          )}
        </div>

        {/* Agent Details Drawer */}
        <AgentDetailsDrawer
          agentId={selectedAgentId}
          open={drawerOpen}
          onOpenChange={setDrawerOpen}
        />
      </div>
    </div>
  );
}
