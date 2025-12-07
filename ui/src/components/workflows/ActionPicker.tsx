import {
  Zap,
  Calendar,
  Webhook,
  Filter,
  Settings,
  Bell,
  Layers,
  ArrowRight,
  Repeat,
  Clock,
  Variable,
  AlertTriangle,
  GitBranch,
  Code,
  Search,
  Activity,
} from "lucide-react";
import { useState, useMemo } from "react";

import type { NodeTemplate } from "./types/flow-types";
import { nodeTemplates } from "./utils/workflow-templates";

import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";

const iconMap: Record<string, React.ElementType> = {
  Zap,
  Calendar,
  Webhook,
  Filter,
  Settings,
  Bell,
  Layers,
  ArrowRight,
  Repeat,
  Clock,
  Variable,
  AlertTriangle,
  GitBranch,
  Code,
  Activity,
};

const colorMap: Record<string, string> = {
  blue: "bg-blue-100 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400 border-blue-200 dark:border-blue-800",
  amber:
    "bg-amber-100 dark:bg-amber-900/30 text-amber-600 dark:text-amber-400 border-amber-200 dark:border-amber-800",
  green:
    "bg-green-100 dark:bg-green-900/30 text-green-600 dark:text-green-400 border-green-200 dark:border-green-800",
  purple:
    "bg-purple-100 dark:bg-purple-900/30 text-purple-600 dark:text-purple-400 border-purple-200 dark:border-purple-800",
  indigo:
    "bg-indigo-100 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 border-indigo-200 dark:border-indigo-800",
  slate:
    "bg-slate-100 dark:bg-slate-900/30 text-slate-600 dark:text-slate-400 border-slate-200 dark:border-slate-800",
  cyan: "bg-cyan-100 dark:bg-cyan-900/30 text-cyan-600 dark:text-cyan-400 border-cyan-200 dark:border-cyan-800",
  red: "bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400 border-red-200 dark:border-red-800",
};

const getColorClass = (color: string) => {
  return colorMap[color] || colorMap.blue;
};

interface ActionPickerProps {
  onNodeSelect: (template: NodeTemplate) => void;
  onCancel?: () => void;
}

export function ActionPicker({ onNodeSelect, onCancel }: ActionPickerProps) {
  const [searchTerm, setSearchTerm] = useState("");

  const filteredTemplates = useMemo(() => {
    if (!searchTerm.trim()) {
      return nodeTemplates;
    }
    const term = searchTerm.toLowerCase();
    return nodeTemplates.filter(
      (template) =>
        template.name.toLowerCase().includes(term) ||
        template.description.toLowerCase().includes(term) ||
        template.type.toLowerCase().includes(term),
    );
  }, [searchTerm]);

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="p-4 border-b space-y-3">
        <div>
          <h3 className="text-sm font-semibold mb-1">Add Step</h3>
          <p className="text-xs text-muted-foreground">
            Select a step to add to your workflow
          </p>
        </div>
        {/* Search */}
        <div className="relative">
          <Search className="absolute left-2.5 top-2 h-3.5 w-3.5 text-muted-foreground" />
          <Input
            placeholder="Search actions..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-8 h-8 text-xs"
            autoFocus
          />
        </div>
      </div>

      {/* Content */}
      <ScrollArea className="flex-1">
        <div className="p-2">
          {filteredTemplates.length === 0 ? (
            <div className="text-center py-8 text-sm text-muted-foreground">
              No actions found
            </div>
          ) : (
            <div className="grid grid-cols-2 gap-2">
              {filteredTemplates.map((template) => {
                const NodeIcon = iconMap[template.icon] || Settings;

                return (
                  <button
                    key={template.id}
                    onClick={() => onNodeSelect(template)}
                    className="flex flex-col items-center gap-2 p-3 hover:bg-muted rounded-md transition-colors text-left border border-transparent hover:border-border group"
                  >
                    <div
                      className={`p-2 rounded ${getColorClass(template.color)}`}
                    >
                      <NodeIcon className="h-4 w-4" />
                    </div>
                    <div className="flex-1 min-w-0 w-full">
                      <div className="text-xs font-medium group-hover:text-primary transition-colors">
                        {template.name}
                      </div>
                      <div className="text-[10px] text-muted-foreground line-clamp-2 mt-0.5">
                        {template.description}
                      </div>
                    </div>
                  </button>
                );
              })}
            </div>
          )}
        </div>
      </ScrollArea>

      {/* Footer */}
      {onCancel && (
        <div className="p-2 border-t">
          <button
            onClick={onCancel}
            className="w-full text-xs text-muted-foreground hover:text-foreground text-center py-1"
          >
            Cancel
          </button>
        </div>
      )}
    </div>
  );
}
