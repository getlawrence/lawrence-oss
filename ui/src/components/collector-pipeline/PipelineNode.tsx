import { Handle, Position, type NodeProps } from "@xyflow/react";
import { 
  ArrowDownCircle, 
  ArrowUpCircle, 
  Settings, 
  Activity,
  AlertTriangle,
  CheckCircle
} from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import type { ComponentMetrics } from "@/api/collector-pipeline";

export interface PipelineNodeData extends Record<string, unknown> {
  component: ComponentMetrics;
  isSelected?: boolean;
}

const getComponentIcon = (componentType: string) => {
  switch (componentType) {
    case "receiver":
      return <ArrowDownCircle className="h-4 w-4" />;
    case "processor":
      return <Settings className="h-4 w-4" />;
    case "exporter":
      return <ArrowUpCircle className="h-4 w-4" />;
    default:
      return <Activity className="h-4 w-4" />;
  }
};

const getComponentColor = (componentType: string) => {
  switch (componentType) {
    case "receiver":
      return "border-green-500 bg-green-50";
    case "processor":
      return "border-blue-500 bg-blue-50";
    case "exporter":
      return "border-purple-500 bg-purple-50";
    default:
      return "border-gray-500 bg-gray-50";
  }
};

const getStatusIcon = (errorRate: number) => {
  if (errorRate === 0) {
    return <CheckCircle className="h-3 w-3 text-green-500" />;
  } else if (errorRate < 5) {
    return <AlertTriangle className="h-3 w-3 text-yellow-500" />;
  } else {
    return <AlertTriangle className="h-3 w-3 text-red-500" />;
  }
};

export function PipelineNode({ data, selected }: NodeProps) {
  const nodeData = data as unknown as PipelineNodeData;
  const { component } = nodeData;
  const isSelected = selected || nodeData.isSelected;

  return (
    <Card 
      className={`min-w-[200px] transition-all duration-200 ${
        getComponentColor(component.component_type)
      } ${isSelected ? "ring-2 ring-blue-400 shadow-lg" : "shadow-md"}`}
    >
      <CardContent className="p-3">
        {/* Header */}
        <div className="flex items-center justify-between mb-2">
          <div className="flex items-center gap-2">
            {getComponentIcon(component.component_type)}
            <span className="font-semibold text-sm capitalize">
              {component.component_name}
            </span>
          </div>
          {getStatusIcon(component.error_rate)}
        </div>

        {/* Pipeline Type Badge */}
        <Badge variant="outline" className="mb-2 text-xs">
          {component.pipeline_type}
        </Badge>

        {/* Metrics */}
        <div className="space-y-1 text-xs">
          <div className="flex justify-between">
            <span className="text-gray-600">Throughput:</span>
            <span className="font-medium">
              {component.throughput.toFixed(1)}/s
            </span>
          </div>
          
          {component.error_rate > 0 && (
            <div className="flex justify-between">
              <span className="text-gray-600">Error Rate:</span>
              <span className="font-medium text-red-600">
                {component.error_rate.toFixed(2)}%
              </span>
            </div>
          )}

          {/* Component-specific metrics */}
          {component.component_type === "receiver" && (
            <>
              {component.received && (
                <div className="flex justify-between">
                  <span className="text-gray-600">Received:</span>
                  <span className="font-medium">
                    {component.received.toLocaleString()}
                  </span>
                </div>
              )}
              {component.accepted && (
                <div className="flex justify-between">
                  <span className="text-gray-600">Accepted:</span>
                  <span className="font-medium text-green-600">
                    {component.accepted.toLocaleString()}
                  </span>
                </div>
              )}
              {component.refused && component.refused > 0 && (
                <div className="flex justify-between">
                  <span className="text-gray-600">Refused:</span>
                  <span className="font-medium text-red-600">
                    {component.refused.toLocaleString()}
                  </span>
                </div>
              )}
            </>
          )}

          {component.component_type === "exporter" && (
            <>
              {component.sent && (
                <div className="flex justify-between">
                  <span className="text-gray-600">Sent:</span>
                  <span className="font-medium text-green-600">
                    {component.sent.toLocaleString()}
                  </span>
                </div>
              )}
              {component.send_failed && component.send_failed > 0 && (
                <div className="flex justify-between">
                  <span className="text-gray-600">Failed:</span>
                  <span className="font-medium text-red-600">
                    {component.send_failed.toLocaleString()}
                  </span>
                </div>
              )}
            </>
          )}

          {component.component_type === "processor" && component.sent && (
            <div className="flex justify-between">
              <span className="text-gray-600">Processed:</span>
              <span className="font-medium text-blue-600">
                {component.sent.toLocaleString()}
              </span>
            </div>
          )}
        </div>

        {/* Handles for connections */}
        <Handle
          type="target"
          position={Position.Left}
          className="w-3 h-3 bg-gray-400"
        />
        <Handle
          type="source"
          position={Position.Right}
          className="w-3 h-3 bg-gray-400"
        />
      </CardContent>
    </Card>
  );
}
