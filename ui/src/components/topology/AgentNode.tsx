import { CheckCircle, XCircle, AlertCircle, Server } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import type { AgentNodeData } from './types';

interface AgentNodeProps {
  data: AgentNodeData;
}

export function AgentNode({ data }: AgentNodeProps) {
  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'online':
        return <CheckCircle className="h-4 w-4 text-green-500" />;
      case 'offline':
        return <XCircle className="h-4 w-4 text-gray-500" />;
      case 'error':
        return <AlertCircle className="h-4 w-4 text-red-500" />;
      default:
        return <Server className="h-4 w-4 text-gray-400" />;
    }
  };

  return (
    <div className="bg-white border-2 border-blue-500 rounded-lg p-3 shadow-lg min-w-[200px]">
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2">
          <Server className="h-5 w-5 text-blue-500" />
          <span className="font-semibold text-sm">{data.name}</span>
        </div>
        {getStatusIcon(data.status)}
      </div>
      {data.metrics && (
        <div className="text-xs text-gray-600 space-y-1">
          <div>Metrics: {data.metrics.metric_count}</div>
          <div>Logs: {data.metrics.log_count}</div>
          <div>Throughput: {data.metrics.throughput_rps.toFixed(2)} rps</div>
        </div>
      )}
      {data.group_name && (
        <Badge variant="outline" className="mt-2 text-xs">
          {data.group_name}
        </Badge>
      )}
    </div>
  );
}
