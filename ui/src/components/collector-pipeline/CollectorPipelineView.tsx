import { useState, useEffect, useMemo } from "react";
import {
  ReactFlow,
  Background,
  useNodesState,
  useEdgesState,
  Controls,
  MiniMap,
  type Node,
  type Edge,
} from "@xyflow/react";
import { RefreshCw, Clock, AlertCircle } from "lucide-react";

import "@xyflow/react/dist/style.css";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Alert, AlertDescription } from "@/components/ui/alert";

import { PipelineNode, type PipelineNodeData } from "./PipelineNode";
import { getPipelineMetrics, type ComponentMetrics, type PipelineMetricsResponse } from "@/api/collector-pipeline";

const nodeTypes = {
  pipeline: PipelineNode,
};

interface CollectorPipelineViewProps {
  agentId: string;
  agentName?: string;
}

type TimeRange = "1m" | "5m" | "15m" | "1h" | "6h" | "24h";

export function CollectorPipelineView({ agentId, agentName }: CollectorPipelineViewProps) {
  const [metrics, setMetrics] = useState<PipelineMetricsResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [timeRange, setTimeRange] = useState<TimeRange>("5m");
  const [selectedPipeline, setSelectedPipeline] = useState<string>("all");

  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);

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
    if (!filteredComponents.length) {
      setNodes([]);
      setEdges([]);
      return;
    }

    const newNodes: Node[] = [];
    const newEdges: Edge[] = [];

    // Group components by pipeline type and component type
    const groupedComponents = filteredComponents.reduce((acc, component) => {
      const key = `${component.pipeline_type}-${component.component_type}`;
      if (!acc[key]) {
        acc[key] = [];
      }
      acc[key].push(component);
      return acc;
    }, {} as Record<string, ComponentMetrics[]>);

    let nodeId = 0;
    const pipelineRows: Record<string, Node[]> = {};

    // Create nodes for each pipeline type
    Object.entries(groupedComponents).forEach(([key, components]) => {
      const [pipelineType, componentType] = key.split('-');
      
      if (!pipelineRows[pipelineType]) {
        pipelineRows[pipelineType] = [];
      }

      components.forEach((component) => {
        const node: Node = {
          id: `node-${nodeId++}`,
          type: 'pipeline',
          position: { 
            x: componentType === 'receiver' ? 100 : componentType === 'processor' ? 300 : 500,
            y: (pipelineRows[pipelineType].length * 150) + (pipelineTypes.indexOf(pipelineType) * 200)
          },
          data: { component } as unknown as PipelineNodeData,
        };
        
        newNodes.push(node);
        pipelineRows[pipelineType].push(node);
      });
    });

    // Create edges between components in the same pipeline
    Object.values(pipelineRows).forEach(pipelineNodes => {
      const receivers = pipelineNodes.filter(n => (n.data as unknown as PipelineNodeData).component.component_type === 'receiver');
      const processors = pipelineNodes.filter(n => (n.data as unknown as PipelineNodeData).component.component_type === 'processor');
      const exporters = pipelineNodes.filter(n => (n.data as unknown as PipelineNodeData).component.component_type === 'exporter');

      // Connect receivers to processors
      receivers.forEach(receiver => {
        processors.forEach(processor => {
          newEdges.push({
            id: `edge-${receiver.id}-${processor.id}`,
            source: receiver.id,
            target: processor.id,
            type: 'smoothstep',
            animated: true,
          });
        });
      });

      // Connect processors to exporters
      processors.forEach(processor => {
        exporters.forEach(exporter => {
          newEdges.push({
            id: `edge-${processor.id}-${exporter.id}`,
            source: processor.id,
            target: exporter.id,
            type: 'smoothstep',
            animated: true,
          });
        });
      });

      // If no processors, connect receivers directly to exporters
      if (processors.length === 0) {
        receivers.forEach(receiver => {
          exporters.forEach(exporter => {
            newEdges.push({
              id: `edge-${receiver.id}-${exporter.id}`,
              source: receiver.id,
              target: exporter.id,
              type: 'smoothstep',
              animated: true,
            });
          });
        });
      }
    });

    setNodes(newNodes);
    setEdges(newEdges);
  }, [filteredComponents, pipelineTypes]);

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
          <Select value={timeRange} onValueChange={(value: TimeRange) => setTimeRange(value)}>
            <SelectTrigger className="w-32">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="1m">1 minute</SelectItem>
              <SelectItem value="5m">5 minutes</SelectItem>
              <SelectItem value="15m">15 minutes</SelectItem>
              <SelectItem value="1h">1 hour</SelectItem>
              <SelectItem value="6h">6 hours</SelectItem>
              <SelectItem value="24h">24 hours</SelectItem>
            </SelectContent>
          </Select>
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
            <Controls />
            <MiniMap />
          </ReactFlow>
        )}
      </div>
    </div>
  );
}
