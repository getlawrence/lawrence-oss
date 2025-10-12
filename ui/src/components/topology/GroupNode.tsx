import { Users } from "lucide-react";

import type { GroupNodeData } from "./types";

interface GroupNodeProps {
  data: GroupNodeData;
}

export function GroupNode({ data }: GroupNodeProps) {
  return (
    <div className="bg-white border-2 border-purple-500 rounded-lg p-3 shadow-md hover:shadow-lg transition-all duration-200 hover:scale-[1.02] min-w-[200px] dark:bg-gray-800 dark:border-purple-400">
      <div className="flex items-center gap-2 mb-2">
        <Users className="h-5 w-5 text-purple-500" />
        <span className="font-semibold">{data.name}</span>
      </div>
      <div className="text-xs text-gray-600">
        {data.agent_count || 0} agents
      </div>
    </div>
  );
}
