import { Handle, Position, type NodeProps } from "@xyflow/react";
import { Zap, Calendar, Webhook, AlertCircle } from "lucide-react";
import { memo } from "react";

import type { TriggerNodeData } from "../types/flow-types";

import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";

interface TriggerNodeProps extends NodeProps {
  data: TriggerNodeData;
  onNodeClick?: (nodeId: string, data: TriggerNodeData) => void;
}

export const TriggerNode = memo(
  ({ id, data, selected, onNodeClick }: TriggerNodeProps) => {
    const getIcon = () => {
      switch (data.triggerType) {
        case "schedule":
          return <Calendar className="h-4 w-4" />;
        case "webhook":
          return <Webhook className="h-4 w-4" />;
        default:
          return <Zap className="h-4 w-4" />;
      }
    };

    const isConfigured =
      data.triggerType === "schedule"
        ? !!(data.cronExpression && data.timezone)
        : true; // manual and webhook are configured by default

    const handleClick = () => {
      if (onNodeClick && id) {
        onNodeClick(id, data);
      }
    };

    return (
      <Card
        className={`w-[180px] h-[110px] overflow-hidden p-2.5 cursor-pointer transition-all bg-white dark:bg-gray-900 ${
          selected ? "ring-2 ring-primary shadow-lg" : "shadow-sm"
        } hover:shadow-md ${!isConfigured ? "border-destructive/50" : "border-border"}`}
        onClick={handleClick}
      >
        <div className="flex items-center justify-between mb-2">
          <div className="flex items-center gap-2">
            <div className="p-1.5 rounded-md bg-blue-100 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400">
              {getIcon()}
            </div>
            <span className="font-semibold text-sm">Trigger</span>
          </div>
          {!isConfigured && (
            <AlertCircle className="h-3.5 w-3.5 text-destructive" />
          )}
        </div>

        <div className="space-y-1.5">
          <Badge variant="outline" className="text-xs">
            {data.triggerType}
          </Badge>

          {data.triggerType === "schedule" && data.cronExpression && (
            <div className="text-xs text-muted-foreground">
              <code className="bg-muted px-1.5 py-0.5 rounded text-[11px]">
                {data.cronExpression}
              </code>
            </div>
          )}

          {data.triggerType === "schedule" && !isConfigured && (
            <div className="text-xs text-destructive font-medium">
              Schedule not configured
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

TriggerNode.displayName = "TriggerNode";
