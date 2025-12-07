import { Handle, Position, type NodeProps } from "@xyflow/react";
import { Settings, Users, Server, AlertCircle } from "lucide-react";
import { memo } from "react";

import type { ActionNodeData } from "../types/flow-types";

import { type WorkflowAction } from "@/api/workflows";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";

interface ActionNodeProps extends NodeProps {
  data: ActionNodeData;
  onNodeClick?: (nodeId: string, action: WorkflowAction) => void;
}

export const ActionNode = memo(
  ({ id, data, selected, onNodeClick }: ActionNodeProps) => {
    const getTargetIcon = () => {
      return data.action.target_type === "group" ? (
        <Users className="h-3 w-3" />
      ) : (
        <Server className="h-3 w-3" />
      );
    };

    const getActionLabel = () => {
      switch (data.action.type) {
        case "config_update":
          return "Config Update";
        case "tail_sampling":
          return "Tail Sampling";
        case "delayed_action":
          return "Delayed Action";
        default:
          return "Action";
      }
    };

    const isConfigured = () => {
      if (!data.action.target_id) return false;

      switch (data.action.type) {
        case "config_update":
          return !!data.action.config_update?.template;
        case "tail_sampling":
          return !!data.action.tail_sampling?.service_name;
        case "delayed_action":
          return !!(
            data.action.delayed_action?.delay &&
            data.action.delayed_action?.action
          );
        default:
          return false;
      }
    };

    const configured = isConfigured();

    const handleClick = () => {
      if (onNodeClick && id) {
        onNodeClick(id, data.action);
      }
    };

    return (
      <Card
        className={`w-[180px] h-[130px] overflow-hidden p-2.5 cursor-pointer transition-all bg-white dark:bg-gray-900 ${
          selected ? "ring-2 ring-primary shadow-lg" : "shadow-sm"
        } hover:shadow-md ${!configured ? "border-destructive/50" : "border-border"}`}
        onClick={handleClick}
      >
        <Handle
          type="target"
          position={Position.Left}
          className="w-3 h-3 !bg-primary"
        />

        <div className="flex items-center justify-between mb-2">
          <div className="flex items-center gap-2">
            <div className="p-1.5 rounded-md bg-green-100 dark:bg-green-900/30 text-green-600 dark:text-green-400">
              <Settings className="h-4 w-4" />
            </div>
            <span className="font-semibold text-sm">Action</span>
          </div>
          {!configured && (
            <AlertCircle className="h-3.5 w-3.5 text-destructive" />
          )}
        </div>

        <div className="space-y-1.5">
          <Badge variant="outline" className="text-xs">
            {getActionLabel()}
          </Badge>

          <div
            className={`flex items-center gap-1 text-xs ${!data.action.target_id ? "text-destructive font-medium" : "text-muted-foreground"}`}
          >
            {getTargetIcon()}
            <span className="truncate text-[11px]">
              {data.targetName || data.action.target_id || "No target selected"}
            </span>
          </div>

          {data.action.config_update && (
            <div className="text-xs">
              <Badge variant="secondary" className="text-xs">
                {data.action.config_update.operation}
              </Badge>
            </div>
          )}

          {data.action.tail_sampling && (
            <div className="space-y-1">
              <div className="text-xs text-muted-foreground">
                Service: {data.action.tail_sampling.service_name || "Not set"}
              </div>
              <div className="text-xs text-muted-foreground">
                Sample: {data.action.tail_sampling.sampling_percentage}%
              </div>
              {data.action.tail_sampling.revert_after && (
                <Badge variant="secondary" className="text-[10px]">
                  Revert in {data.action.tail_sampling.revert_after}
                </Badge>
              )}
            </div>
          )}

          {data.action.delayed_action && (
            <div className="text-xs">
              <Badge variant="secondary" className="text-xs">
                Delay: {data.action.delayed_action.delay}
              </Badge>
            </div>
          )}

          {data.action.type === "config_update" &&
            !data.action.config_update?.template && (
              <div className="text-xs text-destructive font-medium">
                Template not configured
              </div>
            )}
        </div>
      </Card>
    );
  },
);

ActionNode.displayName = "ActionNode";
