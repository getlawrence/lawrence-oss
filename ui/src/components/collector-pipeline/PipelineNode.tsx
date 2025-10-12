import { Handle, Position, type NodeProps } from "@xyflow/react";
import { ArrowDownCircle, ArrowUpCircle, Settings } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { formatCount } from "@/utils";

interface SectionNodeData extends Record<string, unknown> {
  label: string;
  pipelineType: string;
  pipelineName: string;
  color: string;
  icon: React.ReactNode;
  width: number;
  height: number;
  metrics?: {
    received: number;
    errors: number;
  };
}

interface ComponentNodeData extends Record<string, unknown> {
  label: string;
  pipelineType: string;
  metrics?: {
    received?: number;
    processed?: number;
    batches?: number;
    exported?: number;
  };
}

// Component type configuration
const componentConfig = {
  receiver: {
    icon: <ArrowDownCircle className="h-4 w-4" />,
    color: "border-green-500 bg-green-50",
    iconColor: "text-green-600",
  },
  processor: {
    icon: <Settings className="h-4 w-4" />,
    color: "border-blue-500 bg-blue-50",
    iconColor: "text-blue-600",
  },
  exporter: {
    icon: <ArrowUpCircle className="h-4 w-4" />,
    color: "border-purple-500 bg-purple-50",
    iconColor: "text-purple-600",
  },
};

// Base component node used by Receiver, Processor, and Exporter
interface BaseNodeProps {
  componentType: "receiver" | "processor" | "exporter";
  label: string;
  pipelineType: string;
  selected?: boolean;
  metrics?: ComponentNodeData["metrics"];
  children?: React.ReactNode;
}

function BaseComponentNode({
  componentType,
  label,
  pipelineType,
  selected,
  metrics,
  children,
}: BaseNodeProps) {
  const config = componentConfig[componentType];

  return (
    <Card
      className={`min-w-[150px] transition-all duration-200 ${config.color} ${
        selected ? "ring-2 ring-blue-400 shadow-lg" : "shadow-md"
      }`}
    >
      <CardContent className="p-3">
        <Handle
          type="target"
          position={Position.Left}
          className="w-3 h-3 bg-gray-400"
        />
        <div className="flex items-center gap-2 mb-2">
          <span className={config.iconColor}>{config.icon}</span>
          <span className="font-semibold text-sm">{label}</span>
        </div>
        <Badge variant="outline" className="text-xs">
          {pipelineType}
        </Badge>
        {metrics && <div className="mt-2 text-xs">{children}</div>}
        <Handle
          type="source"
          position={Position.Right}
          className="w-3 h-3 bg-gray-400"
        />
      </CardContent>
    </Card>
  );
}

// Section Node - Container for pipeline
export function SectionNode({ data }: NodeProps) {
  const nodeData = data as unknown as SectionNodeData;
  return (
    <div
      className={`p-4 rounded-lg border-2 ${nodeData.color || "border-gray-400"} bg-white/50 backdrop-blur-sm`}
      style={{
        width: nodeData.width || 850,
        height: nodeData.height || 320,
      }}
    >
      <div className="flex items-center gap-2 mb-2">
        {nodeData.icon}
        <h3 className="text-lg font-semibold">{nodeData.label}</h3>
      </div>
      {nodeData.metrics && (
        <div className="text-xs text-gray-600">
          <span>Received: {formatCount(nodeData.metrics.received || 0)}</span>
          {nodeData.metrics.errors > 0 && (
            <span className="ml-4 text-red-600">
              Errors: {formatCount(nodeData.metrics.errors)}
            </span>
          )}
        </div>
      )}
    </div>
  );
}

// Receiver Node - Data input
export function ReceiverNode({ data, selected }: NodeProps) {
  const nodeData = data as unknown as ComponentNodeData;
  return (
    <BaseComponentNode
      componentType="receiver"
      label={nodeData.label}
      pipelineType={nodeData.pipelineType || "receiver"}
      selected={selected}
      metrics={nodeData.metrics}
    >
      {nodeData.metrics?.received !== undefined && (
        <div className="flex justify-between">
          <span className="text-gray-600">Received:</span>
          <span className="font-medium">
            {formatCount(nodeData.metrics.received)}
          </span>
        </div>
      )}
    </BaseComponentNode>
  );
}

// Processor Node - Data transformation
export function ProcessorNode({ data, selected }: NodeProps) {
  const nodeData = data as unknown as ComponentNodeData;
  return (
    <BaseComponentNode
      componentType="processor"
      label={nodeData.label}
      pipelineType={nodeData.pipelineType || "processor"}
      selected={selected}
      metrics={nodeData.metrics}
    >
      {nodeData.metrics?.processed !== undefined && (
        <div className="flex justify-between">
          <span className="text-gray-600">Processed:</span>
          <span className="font-medium">
            {formatCount(nodeData.metrics.processed)}
          </span>
        </div>
      )}
      {nodeData.metrics?.batches !== undefined && (
        <div className="flex justify-between">
          <span className="text-gray-600">Batches:</span>
          <span className="font-medium">
            {formatCount(nodeData.metrics.batches)}
          </span>
        </div>
      )}
    </BaseComponentNode>
  );
}

// Exporter Node - Data output
export function ExporterNode({ data, selected }: NodeProps) {
  const nodeData = data as unknown as ComponentNodeData;
  return (
    <BaseComponentNode
      componentType="exporter"
      label={nodeData.label}
      pipelineType={nodeData.pipelineType || "exporter"}
      selected={selected}
      metrics={nodeData.metrics}
    >
      {nodeData.metrics?.exported !== undefined && (
        <div className="flex justify-between">
          <span className="text-gray-600">Exported:</span>
          <span className="font-medium">
            {formatCount(nodeData.metrics.exported)}
          </span>
        </div>
      )}
    </BaseComponentNode>
  );
}
