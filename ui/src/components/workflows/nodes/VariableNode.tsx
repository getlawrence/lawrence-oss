import { Handle, Position, type NodeProps } from "@xyflow/react";
import { Variable, AlertCircle } from "lucide-react";
import { memo } from "react";

import type { VariableNodeData } from "../types/flow-types";

import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";

interface VariableNodeProps extends NodeProps {
  data: VariableNodeData;
  onNodeClick?: (nodeId: string, data: VariableNodeData) => void;
}

export const VariableNode = memo(
  ({ id, data, selected, onNodeClick }: VariableNodeProps) => {
    const isConfigured = !!(data.variableName && data.operation);

    const handleClick = () => {
      if (onNodeClick && id) {
        onNodeClick(id, data);
      }
    };

    const getOperationLabel = () => {
      switch (data.operation) {
        case "set":
          return "Set Variable";
        case "get":
          return "Get Variable";
        case "increment":
          return "Increment";
        case "append":
          return "Append";
        default:
          return "Variable";
      }
    };

    return (
      <Card
        className={`min-w-[200px] p-3 cursor-pointer transition-all bg-white dark:bg-gray-900 ${
          selected ? "ring-2 ring-primary shadow-lg" : "shadow-sm"
        } hover:shadow-md ${!isConfigured ? "border-destructive/50" : "border-border"}`}
        onClick={handleClick}
      >
        <Handle
          type="target"
          position={Position.Left}
          className="w-3 h-3 !bg-primary"
        />

        <div className="flex items-center justify-between mb-2">
          <div className="flex items-center gap-2">
            <div className="p-1.5 rounded-md bg-cyan-100 dark:bg-cyan-900/30 text-cyan-600 dark:text-cyan-400">
              <Variable className="h-4 w-4" />
            </div>
            <span className="font-semibold text-sm">{getOperationLabel()}</span>
          </div>
          {!isConfigured && (
            <AlertCircle className="h-3.5 w-3.5 text-destructive" />
          )}
        </div>

        <div className="space-y-1.5">
          <div className="text-xs text-muted-foreground line-clamp-2">
            {data.description || data.label}
          </div>

          <div className="flex flex-wrap gap-1">
            {data.variableName && (
              <Badge variant="outline" className="text-[10px] px-1.5 font-mono">
                {data.variableName}
              </Badge>
            )}
            {data.operation && (
              <Badge variant="secondary" className="text-[10px] px-1.5">
                {data.operation}
              </Badge>
            )}
          </div>

          {data.value && (
            <div className="text-xs bg-muted px-2 py-1 rounded text-[11px] font-mono truncate">
              {data.value}
            </div>
          )}

        </div>

        <Handle
          type="source"
          position={Position.Right}
          className="w-3 h-3 !bg-primary"
        />
      </Card>
    );
  },
);

VariableNode.displayName = "VariableNode";
