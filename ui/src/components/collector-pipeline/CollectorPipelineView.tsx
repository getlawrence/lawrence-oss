import {
  ReactFlow,
  Background,
  useNodesState,
  useEdgesState,
  type Node,
  type Edge,
} from "@xyflow/react";
import { RefreshCw, AlertCircle } from "lucide-react";
import { useEffect } from "react";
import useSWR from "swr";

import "@xyflow/react/dist/style.css";

import { generatePipelineNodes } from "./PipelineGenerator";
import {
  ReceiverNode,
  ProcessorNode,
  ExporterNode,
  SectionNode,
} from "./PipelineNode";

import { fetchAgentComponentMetrics } from "@/api/collector-metrics";
import { getConfigs } from "@/api/configs";
import { Alert, AlertDescription } from "@/components/ui/alert";

const nodeTypes = {
  section: SectionNode,
  receiver: ReceiverNode,
  processor: ProcessorNode,
  exporter: ExporterNode,
};

interface CollectorPipelineViewProps {
  agentId: string;
  agentName?: string;
  effectiveConfig?: string; // Pass effective config from agent object
}

export function CollectorPipelineView({
  agentId,
  agentName: _agentName,
  effectiveConfig: propEffectiveConfig,
}: CollectorPipelineViewProps) {
  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);

  // Fetch component metrics
  const { data: metricsData } = useSWR(
    `agent-component-metrics-pipeline-${agentId}`,
    async () => {
      try {
        return await fetchAgentComponentMetrics(agentId, 5); // 5 minutes
      } catch (error) {
        console.error("Failed to fetch component metrics:", error);
        return [];
      }
    },
    {
      refreshInterval: 5000, // Refresh every 5 seconds
    },
  );

  // Fetch agent config only if not provided via props
  const {
    data: configsData,
    error: configError,
    isLoading: configLoading,
  } = useSWR(
    propEffectiveConfig ? null : `agent-config-pipeline-${agentId}`,
    async () => {
      const configsResponse = await getConfigs({ agent_id: agentId, limit: 1 });
      return configsResponse;
    },
  );

  // Determine the effective config to use
  const effectiveConfig =
    propEffectiveConfig ||
    (configsData?.configs && configsData.configs.length > 0
      ? configsData.configs[0].content
      : null);

  const loading = !propEffectiveConfig && configLoading;
  const error = configError ? "Failed to fetch agent configuration" : null;

  // Generate nodes and edges for React Flow
  useEffect(() => {
    if (!effectiveConfig) {
      // No config available yet
      setNodes([]);
      setEdges([]);
      return;
    }

    // Use the generator to create nodes from config with metrics
    const { nodes: generatedNodes, edges: generatedEdges } =
      generatePipelineNodes(effectiveConfig, metricsData || []);

    setNodes(generatedNodes);
    setEdges(generatedEdges);
  }, [effectiveConfig, metricsData, setNodes, setEdges]);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <RefreshCw className="h-8 w-8 animate-spin" />
        <span className="ml-2">Loading pipeline configuration...</span>
      </div>
    );
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    );
  }

  return (
    <div className="h-full flex flex-col">
      <div className="flex-1 relative">
        {nodes.length === 0 ? (
          <div className="flex items-center justify-center h-full">
            <div className="text-center">
              <AlertCircle className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <p className="text-gray-600">No pipeline components found</p>
              <p className="text-sm text-gray-500">
                No components available for this agent
              </p>
            </div>
          </div>
        ) : (
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            nodeTypes={nodeTypes}
            fitView
            minZoom={0.2}
            maxZoom={2}
            attributionPosition="bottom-left"
          >
            <Background />
          </ReactFlow>
        )}
      </div>
    </div>
  );
}
