import { useState, useEffect, useMemo } from "react";
import {
  ReactFlow,
  Background,
  useNodesState,
  useEdgesState,
  type Node,
  type Edge,
} from "@xyflow/react";
import { RefreshCw, Clock, AlertCircle } from "lucide-react";

import "@xyflow/react/dist/style.css";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { TimeRangeSelect } from "@/components/ui/time-range-select";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

import { ReceiverNode, ProcessorNode, ExporterNode, SectionNode } from "./PipelineNode";
import { getPipelineMetrics, type PipelineMetricsResponse } from "@/api/collector-pipeline";
import { generatePipelineNodes } from "./PipelineGenerator";
import { getConfigs } from "@/api/configs";
import { type TimeRange, DEFAULT_TIME_RANGE } from "@/types/timeRange";

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

export function CollectorPipelineView({ agentId, agentName, effectiveConfig: propEffectiveConfig }: CollectorPipelineViewProps) {
  const [metrics, setMetrics] = useState<PipelineMetricsResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [timeRange, setTimeRange] = useState<TimeRange>(DEFAULT_TIME_RANGE);
  const [selectedPipeline, setSelectedPipeline] = useState<string>("all");

  const [effectiveConfig, setEffectiveConfig] = useState<string | null>(propEffectiveConfig || null);

  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);

  // Fetch agent config only if not provided via props
  const fetchConfig = async () => {
    if (propEffectiveConfig) {
      setEffectiveConfig(propEffectiveConfig);
      return;
    }
    
    try {
      const configsResponse = await getConfigs({ agent_id: agentId, limit: 1 });
      if (configsResponse.configs.length > 0) {
        setEffectiveConfig(configsResponse.configs[0].content);
      }
    } catch (err) {
      console.error("Failed to fetch agent config:", err);
      // Don't show error to user, just continue without config
    }
  };

  // Fetch pipeline metrics
  const fetchMetrics = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await getPipelineMetrics(agentId, timeRange);
      setMetrics(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to fetch pipeline metrics");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchConfig();
  }, [agentId, propEffectiveConfig]);

  useEffect(() => {
    fetchMetrics();
  }, [agentId, timeRange]);

  // Get unique pipeline types
  const pipelineTypes = useMemo(() => {
    if (!metrics) return [];
    const types = new Set(metrics.components.map(c => c.pipeline_type));
    return Array.from(types).sort();
  }, [metrics]);

  // Filter components by selected pipeline
  const filteredComponents = useMemo(() => {
    if (!metrics) return [];
    if (selectedPipeline === "all") return metrics.components;
    return metrics.components.filter(c => c.pipeline_type === selectedPipeline);
  }, [metrics, selectedPipeline]);

  // Generate nodes and edges for React Flow
  useEffect(() => {
    if (!effectiveConfig) {
      // No config available yet
      setNodes([]);
      setEdges([]);
      return;
    }

    // Use the generator to create nodes from config
    const { nodes: generatedNodes, edges: generatedEdges } = generatePipelineNodes(effectiveConfig);

    setNodes(generatedNodes);
    setEdges(generatedEdges);
  }, [effectiveConfig]);

  // Calculate summary statistics
  const summaryStats = useMemo(() => {
    if (!filteredComponents.length) return null;

    const totalThroughput = filteredComponents.reduce((sum, c) => sum + c.throughput, 0);
    const totalErrors = filteredComponents.reduce((sum, c) => sum + c.errors, 0);
    const avgErrorRate = filteredComponents.reduce((sum, c) => sum + c.error_rate, 0) / filteredComponents.length;
    const componentCounts = filteredComponents.reduce((acc, c) => {
      acc[c.component_type] = (acc[c.component_type] || 0) + 1;
      return acc;
    }, {} as Record<string, number>);

    return {
      totalThroughput,
      totalErrors,
      avgErrorRate,
      componentCounts,
    };
  }, [filteredComponents]);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <RefreshCw className="h-8 w-8 animate-spin" />
        <span className="ml-2">Loading pipeline metrics...</span>
      </div>
    );
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          Failed to load pipeline metrics: {error}
        </AlertDescription>
      </Alert>
    );
  }

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b">
        <div>
          <h2 className="text-xl font-semibold">
            Collector Pipeline - {agentName || agentId}
          </h2>
          <p className="text-sm text-gray-600">
            Real-time pipeline component metrics and flow
          </p>
        </div>
        <div className="flex items-center gap-2">
          <TimeRangeSelect value={timeRange} onValueChange={setTimeRange} />
          <Button onClick={fetchMetrics} size="sm" variant="outline">
            <RefreshCw className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {/* Pipeline Filter */}
      {pipelineTypes.length > 1 && (
        <div className="p-4 border-b">
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium">Pipeline:</span>
            <Select value={selectedPipeline} onValueChange={setSelectedPipeline}>
              <SelectTrigger className="w-40">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Pipelines</SelectItem>
                {pipelineTypes.map(type => (
                  <SelectItem key={type} value={type}>
                    {type.charAt(0).toUpperCase() + type.slice(1)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>
      )}

      {/* Summary Stats */}
      {summaryStats && (
        <div className="p-4 border-b">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <Card>
              <CardContent className="p-3">
                <div className="text-sm text-gray-600">Total Throughput</div>
                <div className="text-lg font-semibold">
                  {summaryStats.totalThroughput.toFixed(1)}/s
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-3">
                <div className="text-sm text-gray-600">Avg Error Rate</div>
                <div className="text-lg font-semibold">
                  {summaryStats.avgErrorRate.toFixed(2)}%
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-3">
                <div className="text-sm text-gray-600">Components</div>
                <div className="text-lg font-semibold">
                  {Object.entries(summaryStats.componentCounts).map(([type, count]) => (
                    <Badge key={type} variant="outline" className="mr-1">
                      {type}: {count}
                    </Badge>
                  ))}
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-3">
                <div className="text-sm text-gray-600">Last Updated</div>
                <div className="text-lg font-semibold flex items-center">
                  <Clock className="h-4 w-4 mr-1" />
                  {metrics?.timestamp ? new Date(metrics.timestamp).toLocaleTimeString() : 'N/A'}
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      )}

      {/* Pipeline Flow */}
      <div className="flex-1 relative">
        {nodes.length === 0 ? (
          <div className="flex items-center justify-center h-full">
            <div className="text-center">
              <AlertCircle className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <p className="text-gray-600">No pipeline components found</p>
              <p className="text-sm text-gray-500">
                {selectedPipeline === "all" 
                  ? "No components available for this agent"
                  : `No ${selectedPipeline} components found`
                }
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
