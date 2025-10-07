import { ChevronDown, ChevronRight } from 'lucide-react';
import { useState } from 'react';
import useSWR from 'swr';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { cn } from '@/lib/utils';
import { getAgents } from '@/api/agents';
import { getGroups } from '@/api/groups';
import type { TopologyNode } from './types';

interface DisplaySidebarProps {
  onNodeSelect?: (node: TopologyNode) => void;
  className?: string;
}

export function DisplaySidebar({ onNodeSelect, className }: DisplaySidebarProps) {
  const [groupsExpanded, setGroupsExpanded] = useState(true);
  const [agentsExpanded, setAgentsExpanded] = useState(true);

  return (
    <div className={cn('bg-background border-r border-border', className)}>
      <div className="h-full flex flex-col">
        {/* Header */}
        <div className="h-16 px-4 flex items-center border-b border-border bg-background flex-shrink-0">
          <h2 className="font-semibold">Resources</h2>
        </div>

        <ScrollArea className="flex-1">
          <div className="p-4 space-y-4">
            {/* Groups Section */}
            <div className="space-y-4 border-b border-border pb-4">
              <GroupsDisplaySection
                expanded={groupsExpanded}
                onToggle={() => setGroupsExpanded(!groupsExpanded)}
                onItemClick={onNodeSelect}
              />
            </div>

            {/* Agents Section */}
            <div className="space-y-4">
              <AgentsDisplaySection
                expanded={agentsExpanded}
                onToggle={() => setAgentsExpanded(!agentsExpanded)}
                onItemClick={onNodeSelect}
              />
            </div>
          </div>
        </ScrollArea>
      </div>
    </div>
  );
}

function GroupsDisplaySection({
  expanded,
  onToggle,
  onItemClick,
}: {
  expanded: boolean;
  onToggle: () => void;
  onItemClick?: (node: TopologyNode) => void;
}) {
  const { data } = useSWR('groups', getGroups);
  const groups = data?.groups || [];

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <div className="text-sm font-medium text-muted-foreground flex items-center gap-2">
          <Button
            variant="ghost"
            size="sm"
            onClick={onToggle}
            className="p-0 h-auto hover:bg-transparent"
          >
            {expanded ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
          </Button>
          Agent Groups
          <Badge variant="outline" className="text-xs">
            {groups.length}
          </Badge>
        </div>
      </div>
      {expanded && (
        <div className="space-y-2 max-h-64 overflow-y-auto">
          {groups.length > 0 ? (
            groups.map((group) => (
              <div
                key={group.id}
                className="flex items-center space-x-2 p-2 rounded-md hover:bg-muted cursor-pointer transition-colors"
                onClick={() =>
                  onItemClick?.({
                    id: group.id,
                    type: 'group',
                    name: group.name,
                    data: group,
                  })
                }
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    onItemClick?.({
                      id: group.id,
                      type: 'group',
                      name: group.name,
                      data: group,
                    });
                  }
                }}
                tabIndex={0}
                role="button"
              >
                <div className="flex-1 min-w-0">
                  <div className="text-sm font-medium truncate">{group.name}</div>
                  <div className="text-xs text-muted-foreground">
                    {Object.keys(group.labels || {}).length > 0 &&
                      `${Object.keys(group.labels).length} labels`
                    }
                  </div>
                </div>
              </div>
            ))
          ) : (
            <div className="text-xs text-muted-foreground text-center py-2">No groups found</div>
          )}
        </div>
      )}
    </div>
  );
}

function AgentsDisplaySection({
  expanded,
  onToggle,
  onItemClick,
}: {
  expanded: boolean;
  onToggle: () => void;
  onItemClick?: (node: TopologyNode) => void;
}) {
  const { data: agentsData } = useSWR('agents', getAgents);
  const agents = agentsData?.agents ? Object.values(agentsData.agents) : [];

  return (
    <div className="space-y-2">
      <div className="text-sm font-medium text-muted-foreground flex items-center gap-2">
        <Button
          variant="ghost"
          size="sm"
          onClick={onToggle}
          className="p-0 h-auto hover:bg-transparent"
        >
          {expanded ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
        </Button>
        Agents
        <Badge variant="outline" className="text-xs">
          {agents.length}
        </Badge>
      </div>
      {expanded && (
        <div className="space-y-2 max-h-64 overflow-y-auto">
          {agents.map((agent) => (
            <div
              key={agent.id}
              className="flex items-center space-x-2 p-2 rounded-md hover:bg-muted cursor-pointer transition-colors"
              onClick={() =>
                onItemClick?.({
                  id: agent.id,
                  type: 'agent',
                  name: agent.name || agent.id.slice(0, 8),
                  data: agent,
                })
              }
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  e.preventDefault();
                  onItemClick?.({
                    id: agent.id,
                    type: 'agent',
                    name: agent.name || agent.id.slice(0, 8),
                    data: agent,
                  });
                }
              }}
              tabIndex={0}
              role="button"
            >
              <div className="flex items-center gap-2 flex-1 min-w-0">
                <div
                  className={cn(
                    'w-2 h-2 rounded-full',
                    agent.status === 'online' ? 'bg-green-500' :
                    agent.status === 'error' ? 'bg-red-500' : 'bg-gray-400'
                  )}
                />
                <div className="flex-1 min-w-0">
                  <div className="text-sm font-medium truncate">
                    {agent.name || agent.id.slice(0, 8)}
                  </div>
                  <div className="text-xs text-muted-foreground">
                    {agent.status}
                  </div>
                </div>
              </div>
            </div>
          ))}
          {agents.length === 0 && (
            <div className="text-xs text-muted-foreground text-center py-2">No agents found</div>
          )}
        </div>
      )}
    </div>
  );
}
