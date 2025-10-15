import { Handle, Position } from "@xyflow/react";
import { Bolt, Grip } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";

interface ProcessorNodeProps {
  data: any;
}

export const ProcessorNode = ({ data }: ProcessorNodeProps) => {
  const nodeStyle = {
    zIndex: 10,
  };

  return (
    <Card
      className="min-w-48 p-0 gap-0 shadow-md hover:shadow-lg transition-all duration-200 border-2 border-green-200 dark:border-green-800 bg-gradient-to-br from-green-50 to-white dark:from-green-950/30 dark:to-background"
      style={nodeStyle}
    >
      <div className="px-1 py-0.5 bg-green-100/50 dark:bg-green-900/20 flex items-center justify-between border-b border-green-200 dark:border-green-800">
        <div className="flex items-center gap-1 pl-1">
          <Grip size={12} className="text-green-400 dark:text-green-500" />
          <Badge
            variant="outline"
            className="text-[10px] py-0 h-4 bg-green-100 dark:bg-green-900/40 text-green-700 dark:text-green-300 border-green-300 dark:border-green-700"
          >
            Processor
          </Badge>
        </div>
      </div>
      <CardContent
        className={data.metrics?.processed !== undefined ? "p-2 pt-3" : "p-2"}
      >
        <div className="flex items-center gap-1.5">
          <div className="p-0.5 rounded-full bg-green-100 dark:bg-green-900/40 flex items-center justify-center">
            <Bolt size={12} className="text-green-600 dark:text-green-400" />
          </div>
          <div className="font-medium text-sm leading-none">{data.label}</div>
        </div>
        {data.metrics?.processed !== undefined && (
          <div className="text-xs text-green-600 dark:text-green-400 mt-2 pl-6">
            Processed: {data.metrics.processed.toLocaleString()}
          </div>
        )}
      </CardContent>
      <Handle
        type="target"
        position={Position.Left}
        style={{
          left: "-7px",
          backgroundColor: "#22c55e",
          border: "2px solid var(--background)",
          zIndex: 20,
          width: "14px",
          height: "14px",
        }}
        id="left"
      />
      <Handle
        type="source"
        position={Position.Right}
        style={{
          right: "-7px",
          backgroundColor: "#22c55e",
          border: "2px solid var(--background)",
          zIndex: 20,
          width: "14px",
          height: "14px",
        }}
        id="right"
      />
    </Card>
  );
};

export default ProcessorNode;
