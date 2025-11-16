import { Handle, Position, type NodeProps } from "@xyflow/react";
import { Folder, ChevronDown, ChevronRight } from "lucide-react";
import { memo } from "react";

import type { GroupNodeData } from "../types/flow-types";

import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";

interface GroupNodeProps extends NodeProps {
  data: GroupNodeData;
  onNodeClick?: (nodeId: string, data: GroupNodeData) => void;
}

export const GroupNode = memo(
  ({ id, data, selected, onNodeClick }: GroupNodeProps) => {
    const handleClick = () => {
      if (onNodeClick && id) {
        onNodeClick(id, data);
      }
    };

    const colorClass =
      data.color ||
      "bg-slate-100 dark:bg-slate-900/30 text-slate-600 dark:text-slate-400 border-slate-200 dark:border-slate-800";

    return (
      <Card
        className={`w-[180px] h-[110px] overflow-hidden p-2.5 cursor-pointer transition-all bg-white dark:bg-gray-900 ${
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
            <div className={`p-1.5 rounded-md ${colorClass}`}>
              <Folder className="h-4 w-4" />
            </div>
            <span className="font-semibold text-sm">Group</span>
          </div>
          {data.collapsed ? (
            <ChevronRight className="h-3.5 w-3.5 text-muted-foreground" />
          ) : (
            <ChevronDown className="h-3.5 w-3.5 text-muted-foreground" />
          )}
        </div>

        <div className="space-y-1.5">
          <div className="text-xs text-muted-foreground line-clamp-2">
            {data.description || data.label}
          </div>

          {data.color && (
            <Badge variant="outline" className="text-[10px] px-1.5">
              {data.color}
            </Badge>
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

GroupNode.displayName = "GroupNode";
