import { Handle, Position, type NodeProps } from "@xyflow/react";
import { AlertTriangle } from "lucide-react";
import { memo } from "react";

import type { ErrorHandlerNodeData } from "../types/flow-types";

import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";

interface ErrorHandlerNodeProps extends NodeProps {
  data: ErrorHandlerNodeData;
  onNodeClick?: (nodeId: string, data: ErrorHandlerNodeData) => void;
}

export const ErrorHandlerNode = memo(
  ({ id, data, selected, onNodeClick }: ErrorHandlerNodeProps) => {
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
            <div className="p-1.5 rounded-md bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400">
              <AlertTriangle className="h-4 w-4" />
            </div>
            <span className="font-semibold text-sm">Error Handler</span>
          </div>
        </div>

        <div className="space-y-1.5">
          <div className="text-xs text-muted-foreground line-clamp-2">
            {data.description || data.label}
          </div>

          <div className="flex flex-wrap gap-1">
            {data.catchAll ? (
              <Badge variant="destructive" className="text-[10px] px-1.5">
                Catch All
              </Badge>
            ) : (
              <Badge variant="outline" className="text-[10px] px-1.5">
                Specific
              </Badge>
            )}
            {data.retryCount && data.retryCount > 0 && (
              <Badge variant="secondary" className="text-[10px] px-1.5">
                Retry: {data.retryCount}
              </Badge>
            )}
            {data.retryDelay && (
              <Badge variant="outline" className="text-[10px] px-1.5">
                Delay: {data.retryDelay}s
              </Badge>
            )}
          </div>

          {data.errorTypes && data.errorTypes.length > 0 && (
            <div className="flex flex-wrap gap-1">
              {data.errorTypes.slice(0, 2).map((errorType, idx) => (
                <Badge
                  key={idx}
                  variant="outline"
                  className="text-[10px] px-1.5"
                >
                  {errorType}
                </Badge>
              ))}
              {data.errorTypes.length > 2 && (
                <Badge variant="outline" className="text-[10px] px-1.5">
                  +{data.errorTypes.length - 2}
                </Badge>
              )}
            </div>
          )}
        </div>

        {/* Success and error outputs */}
        <Handle
          type="source"
          position={Position.Right}
          id="success"
          className="w-3 h-3 !bg-green-500 !top-[30%]"
        />
        <Handle
          type="source"
          position={Position.Right}
          id="error"
          className="w-3 h-3 !bg-red-500 !top-[70%]"
        />
      </Card>
    );
  },
);

ErrorHandlerNode.displayName = "ErrorHandlerNode";
