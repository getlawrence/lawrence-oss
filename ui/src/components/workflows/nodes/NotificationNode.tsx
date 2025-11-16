import { Handle, Position, type NodeProps } from "@xyflow/react";
import {
  Bell,
  Mail,
  MessageSquare,
  Webhook,
  FileText,
  AlertCircle,
} from "lucide-react";
import { memo } from "react";

import type { NotificationNodeData } from "../types/flow-types";

import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";

interface NotificationNodeProps extends NodeProps {
  data: NotificationNodeData;
  onNodeClick?: (nodeId: string, data: NotificationNodeData) => void;
}

export const NotificationNode = memo(
  ({ id, data, selected, onNodeClick }: NotificationNodeProps) => {
    const isConfigured = !!data.message;

    const handleClick = () => {
      if (onNodeClick && id) {
        onNodeClick(id, data);
      }
    };

    const getChannelIcon = () => {
      switch (data.channel) {
        case "email":
          return <Mail className="h-4 w-4" />;
        case "slack":
          return <MessageSquare className="h-4 w-4" />;
        case "webhook":
          return <Webhook className="h-4 w-4" />;
        default:
          return <FileText className="h-4 w-4" />;
      }
    };

    const getSeverityColor = () => {
      switch (data.severity) {
        case "error":
          return "destructive";
        case "warning":
          return "secondary";
        case "success":
          return "default";
        default:
          return "outline";
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
            <div className="p-1.5 rounded-md bg-purple-100 dark:bg-purple-900/30 text-purple-600 dark:text-purple-400">
              <Bell className="h-4 w-4" />
            </div>
            <span className="font-semibold text-sm">Notification</span>
          </div>
          {!isConfigured && (
            <AlertCircle className="h-3.5 w-3.5 text-destructive" />
          )}
        </div>

        <div className="space-y-1.5">
          <div className="flex items-center gap-1 text-xs">
            <div className="p-1 rounded bg-muted">{getChannelIcon()}</div>
            <Badge variant="outline" className="text-[10px] px-1.5">
              {data.channel}
            </Badge>
            {data.severity && (
              <Badge
                variant={getSeverityColor() as any}
                className="text-[10px] px-1.5"
              >
                {data.severity}
              </Badge>
            )}
          </div>

          {data.message ? (
            <div className="text-xs bg-muted px-2 py-1.5 rounded line-clamp-2">
              {data.message}
            </div>
          ) : (
            <div className="text-xs text-destructive font-medium">
              Message not configured
            </div>
          )}

          {data.recipients && data.recipients.length > 0 && (
            <div className="text-[10px] text-muted-foreground">
              {data.recipients.length} recipient
              {data.recipients.length > 1 ? "s" : ""}
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

NotificationNode.displayName = "NotificationNode";
