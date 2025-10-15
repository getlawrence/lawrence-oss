import { Handle, Position } from "@xyflow/react";
import { ArrowUp, Grip } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";

interface ExporterNodeProps {
  data: any;
}

export const ExporterNode = ({ data }: ExporterNodeProps) => {
  const nodeStyle = {
    zIndex: 10,
  };

  return (
    <Card
      className="min-w-48 p-0 gap-0 shadow-md hover:shadow-lg transition-all duration-200 border-2 border-purple-200 dark:border-purple-800 bg-gradient-to-br from-purple-50 to-white dark:from-purple-950/30 dark:to-background"
      style={nodeStyle}
    >
      <div className="px-1 py-0.5 bg-purple-100/50 dark:bg-purple-900/20 flex items-center justify-between border-b border-purple-200 dark:border-purple-800">
        <div className="flex items-center gap-1 pl-1">
          <Grip size={12} className="text-purple-400 dark:text-purple-500" />
          <Badge
            variant="outline"
            className="text-[10px] py-0 h-4 bg-purple-100 dark:bg-purple-900/40 text-purple-700 dark:text-purple-300 border-purple-300 dark:border-purple-700"
          >
            Exporter
          </Badge>
        </div>
      </div>
      <CardContent
        className={
          data.metrics?.exported !== undefined || data.config?.endpoint
            ? "p-2 pt-3"
            : "p-2"
        }
      >
        <div className="flex items-center gap-1.5">
          <div className="p-0.5 rounded-full bg-purple-100 dark:bg-purple-900/40 flex items-center justify-center">
            <ArrowUp
              size={12}
              className="text-purple-600 dark:text-purple-400"
            />
          </div>
          <div className="font-medium text-sm leading-none">{data.label}</div>
        </div>
        {data.config?.endpoint && (
          <div className="text-xs text-muted-foreground mt-2 pl-6 truncate max-w-40">
            {data.config.endpoint}
          </div>
        )}
        {data.metrics?.exported !== undefined && (
          <div className="text-xs text-purple-600 dark:text-purple-400 mt-2 pl-6">
            Exported: {data.metrics.exported.toLocaleString()}
          </div>
        )}
      </CardContent>
      <Handle
        type="target"
        position={Position.Left}
        style={{
          left: "-7px",
          backgroundColor: "#a855f7",
          border: "2px solid var(--background)",
          zIndex: 20,
          width: "14px",
          height: "14px",
        }}
        id="left"
      />
    </Card>
  );
};

export default ExporterNode;
