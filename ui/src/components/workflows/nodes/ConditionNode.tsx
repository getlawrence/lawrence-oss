import { Handle, Position, type NodeProps } from "@xyflow/react";
import { Filter, AlertCircle } from "lucide-react";
import { memo } from "react";

import type { ConditionNodeData } from "../types/flow-types";

import type { WorkflowCondition } from "@/api/workflows";
import { Card } from "@/components/ui/card";

interface ConditionNodeProps extends NodeProps {
  data: ConditionNodeData;
  onNodeClick?: (nodeId: string, data: ConditionNodeData) => void;
}

export const ConditionNode = memo(
  ({ id, data, selected, onNodeClick }: ConditionNodeProps) => {
    const isConfigured = data.conditions.length > 0;

    const handleClick = () => {
      if (onNodeClick && id) {
        onNodeClick(id, data);
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
            <div className="p-1.5 rounded-md bg-amber-100 dark:bg-amber-900/30 text-amber-600 dark:text-amber-400">
              <Filter className="h-4 w-4" />
            </div>
            <span className="font-semibold text-sm">Condition</span>
          </div>
          {!isConfigured && (
            <AlertCircle className="h-3.5 w-3.5 text-destructive" />
          )}
        </div>

        <div className="space-y-1">
          {data.conditions.length === 0 ? (
            <p className="text-xs text-destructive font-medium">
              No conditions
            </p>
          ) : (
            data.conditions.map((condition: WorkflowCondition, idx: number) => (
              <div
                key={idx}
                className="text-xs bg-muted px-2 py-1 rounded text-[11px]"
              >
                {condition.field} {condition.operator} {condition.value}
              </div>
            ))
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

ConditionNode.displayName = "ConditionNode";
