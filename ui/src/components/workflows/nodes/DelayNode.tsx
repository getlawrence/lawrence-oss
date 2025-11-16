import { Handle, Position, type NodeProps } from "@xyflow/react";
import { Clock } from "lucide-react";
import { memo } from "react";

import type { DelayNodeData } from "../types/flow-types";

import { Card } from "@/components/ui/card";

interface DelayNodeProps extends NodeProps {
  data: DelayNodeData;
  onNodeClick?: (nodeId: string, data: DelayNodeData) => void;
}

export const DelayNode = memo(
  ({ id, data, selected, onNodeClick }: DelayNodeProps) => {
    const handleClick = () => {
      if (onNodeClick && id) {
        onNodeClick(id, data);
      }
    };

    const formatDuration = () => {
      const duration = data.duration || 0;
      const unit = data.unit || "seconds";
      return `${duration} ${unit}`;
    };

    return (
      <Card
        className={`min-w-[180px] p-3 cursor-pointer transition-all bg-white dark:bg-gray-900 ${
          selected ? "ring-2 ring-primary shadow-lg" : "shadow-sm"
        } hover:shadow-md border-border`}
        onClick={handleClick}
      >
        <Handle
          type="target"
          position={Position.Left}
          className="w-3 h-3 !bg-primary"
        />

        <div className="flex items-center justify-between mb-2">
          <div className="flex items-center gap-2">
            <div className="p-1.5 rounded-md bg-slate-100 dark:bg-slate-900/30 text-slate-600 dark:text-slate-400">
              <Clock className="h-4 w-4" />
            </div>
            <span className="font-semibold text-sm">Delay</span>
          </div>
        </div>

        <div className="space-y-1.5">
          <div className="text-sm font-medium text-center py-2 bg-muted rounded">
            {formatDuration()}
          </div>

          {data.description && (
            <div className="text-xs text-muted-foreground line-clamp-2">
              {data.description}
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

DelayNode.displayName = "DelayNode";
