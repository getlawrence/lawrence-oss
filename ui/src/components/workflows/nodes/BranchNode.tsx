import { Handle, Position, type NodeProps } from "@xyflow/react";
import { GitBranch, AlertCircle } from "lucide-react";
import { memo } from "react";

import type { BranchNodeData } from "../types/flow-types";

import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";

interface BranchNodeProps extends NodeProps {
  data: BranchNodeData;
  onNodeClick?: (nodeId: string, data: BranchNodeData) => void;
}

export const BranchNode = memo(
  ({ id, data, selected, onNodeClick }: BranchNodeProps) => {
    const isConfigured = !!(data.branches && data.branches.length > 0);

    const handleClick = () => {
      if (onNodeClick && id) {
        onNodeClick(id, data);
      }
    };

    return (
      <Card
        className={`w-[180px] h-[150px] overflow-hidden p-2.5 cursor-pointer transition-all bg-white dark:bg-gray-900 ${
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
            <div className="p-1.5 rounded-md bg-purple-100 dark:bg-purple-900/30 text-purple-600 dark:text-purple-400">
              <GitBranch className="h-4 w-4" />
            </div>
            <span className="font-semibold text-sm">Branch</span>
          </div>
          {!isConfigured && (
            <AlertCircle className="h-3.5 w-3.5 text-destructive" />
          )}
        </div>

        <div className="space-y-1.5">
          <div className="text-xs text-muted-foreground line-clamp-2">
            {data.description || data.label}
          </div>

          {data.branches && data.branches.length > 0 && (
            <div className="flex flex-wrap gap-1">
              <Badge variant="secondary" className="text-[10px] px-1.5">
                {data.branches.length} branch
                {data.branches.length !== 1 ? "es" : ""}
              </Badge>
              {data.branches.some((b) => b.isDefault) && (
                <Badge variant="outline" className="text-[10px] px-1.5">
                  Default
                </Badge>
              )}
            </div>
          )}

          {data.branches && data.branches.length > 0 && (
            <div className="space-y-1 mt-2">
              {data.branches.slice(0, 3).map((branch, idx) => (
                <div
                  key={idx}
                  className="text-[10px] bg-muted px-2 py-1 rounded truncate"
                >
                  {branch.name}
                  {branch.isDefault && (
                    <span className="ml-1 text-muted-foreground">
                      (default)
                    </span>
                  )}
                </div>
              ))}
              {data.branches.length > 3 && (
                <div className="text-[10px] text-muted-foreground">
                  +{data.branches.length - 3} more
                </div>
              )}
            </div>
          )}
        </div>

        {/* Dynamic outputs based on branches */}
        {data.branches && data.branches.length > 0 ? (
          <>
            {data.branches.slice(0, 4).map((branch, idx) => (
              <Handle
                key={branch.name}
                type="source"
                position={Position.Right}
                id={`branch-${idx}`}
                className="w-3 h-3 !bg-primary"
                style={{ top: `${20 + idx * 20}%` }}
              />
            ))}
          </>
        ) : (
          <Handle
            type="source"
            position={Position.Right}
            className="w-3 h-3 !bg-primary"
          />
        )}
      </Card>
    );
  },
);

BranchNode.displayName = "BranchNode";
