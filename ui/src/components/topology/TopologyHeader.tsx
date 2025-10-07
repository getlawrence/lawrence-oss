import { RefreshCw } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';

interface TopologyHeaderProps {
  topologyLevel: 'instance' | 'group';
  onTopologyLevelChange: (level: 'instance' | 'group') => void;
  onRefresh: () => void;
  refreshing: boolean;
}

export function TopologyHeader({
  topologyLevel,
  onTopologyLevelChange,
  onRefresh,
  refreshing,
}: TopologyHeaderProps) {
  return (
    <div className="h-16 border-b bg-white px-4 flex items-center justify-between flex-shrink-0">
      <div>
        <h1 className="text-2xl font-bold">Topology</h1>
        <p className="text-sm text-gray-600">Visualize your agent infrastructure</p>
      </div>
      <div className="flex items-center gap-3">
        <Tabs value={topologyLevel} onValueChange={(value) => onTopologyLevelChange(value as 'instance' | 'group')}>
          <TabsList>
            <TabsTrigger value="instance">Instance</TabsTrigger>
            <TabsTrigger value="group">Group</TabsTrigger>
          </TabsList>
        </Tabs>
        <Button onClick={onRefresh} disabled={refreshing} size="sm">
          <RefreshCw className={`h-4 w-4 mr-2 ${refreshing ? 'animate-spin' : ''}`} />
          Refresh
        </Button>
      </div>
    </div>
  );
}
