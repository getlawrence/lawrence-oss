import { Handle, Position, type NodeProps } from "@xyflow/react";
import { Layers, Clock } from "lucide-react";
import { memo } from "react";

import type { ParallelNodeData } from "../types/flow-types";

import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";

interface ParallelNodeProps extends NodeProps {
  data: ParallelNodeData;
  onNodeClick?: (nodeId: string, data: ParallelNodeData) => void;
}

export const ParallelNode = memo(
  ({ id, data, selected, onNodeClick }: ParallelNodeProps) => {
    const handleClick = () => {
      if (onNodeClick && id) {
        onNodeClick(id, data);
      }
    };

    return (
      <Card
        className={`min-w-[220px] p-3 cursor-pointer transition-all bg-white dark:bg-gray-900 ${
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
            <div className="p-1.5 rounded-md bg-indigo-100 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400">
              <Layers className="h-4 w-4" />
            </div>
            <span className="font-semibold text-sm">Parallel</span>
          </div>
        </div>

        <div className="space-y-1.5">
          <div className="text-xs text-muted-foreground line-clamp-2">
            {data.description || data.label}
          </div>

          <div className="flex flex-wrap gap-1">
            {data.waitForAll && (
              <Badge variant="secondary" className="text-[10px] px-1.5">
                Wait for all
              </Badge>
            )}
            {data.timeout && (
              <Badge
                variant="outline"
                className="text-[10px] px-1.5 flex items-center gap-0.5"
              >
                <Clock className="h-2.5 w-2.5" />
                {data.timeout}s
              </Badge>
            )}
          </div>
        </div>

        {/* Multiple outputs for parallel paths */}
        <Handle
          type="source"
          position={Position.Right}
          id="out-1"
          className="w-3 h-3 !bg-primary !top-[25%]"
        />
        <Handle
          type="source"
          position={Position.Right}
          id="out-2"
          className="w-3 h-3 !bg-primary !top-[50%]"
        />
        <Handle
          type="source"
          position={Position.Right}
          id="out-3"
          className="w-3 h-3 !bg-primary !top-[75%]"
        />
      </Card>
    );
  },
);

ParallelNode.displayName = "ParallelNode";
