import { Handle, Position, type NodeProps } from "@xyflow/react";
import { Repeat, AlertCircle } from "lucide-react";
import { memo } from "react";

import type { LoopNodeData } from "../types/flow-types";

import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";

interface LoopNodeProps extends NodeProps {
  data: LoopNodeData;
  onNodeClick?: (nodeId: string, data: LoopNodeData) => void;
}

export const LoopNode = memo(
  ({ id, data, selected, onNodeClick }: LoopNodeProps) => {
    const isConfigured = !!data.loopType;

    const handleClick = () => {
      if (onNodeClick && id) {
        onNodeClick(id, data);
      }
    };

    const getLoopTypeLabel = () => {
      switch (data.loopType) {
        case "agents":
          return "Loop Agents";
        case "groups":
          return "Loop Groups";
        case "range":
          return "Loop Range";
        default:
          return "Loop";
      }
    };

    return (
      <Card
        className={`w-[180px] h-[130px] overflow-hidden p-2.5 cursor-pointer transition-all bg-white dark:bg-gray-900 ${
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
            <div className="p-1.5 rounded-md bg-indigo-100 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400">
              <Repeat className="h-4 w-4" />
            </div>
            <span className="font-semibold text-sm">{getLoopTypeLabel()}</span>
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
            <Badge variant="outline" className="text-[10px] px-1.5">
              {data.loopType}
            </Badge>
            {data.maxIterations && (
              <Badge variant="secondary" className="text-[10px] px-1.5">
                max: {data.maxIterations}
              </Badge>
            )}
            {data.parallelExecution && (
              <Badge variant="secondary" className="text-[10px] px-1.5">
                parallel
              </Badge>
            )}
          </div>

          {data.filter && (
            <div className="text-xs bg-muted px-2 py-1 rounded text-[11px] font-mono truncate">
              {data.filter}
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

LoopNode.displayName = "LoopNode";
