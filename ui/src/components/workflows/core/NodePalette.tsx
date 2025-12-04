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
  ChevronDown,
  ChevronRight,
  Search,
} from "lucide-react";
import { useState, useMemo, useCallback } from "react";

import type { NodeTemplate } from "../types/flow-types";
import { nodeCategories
  
 } from "../utils/workflow-templates";

import { Badge } from "@/components/ui/badge";
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

interface NodePaletteProps {
  onNodeSelect: (template: NodeTemplate) => void;
}

export function NodePalette({ onNodeSelect }: NodePaletteProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const [expandedCategories, setExpandedCategories] = useState<Set<string>>(
    new Set(["triggers", "logic", "actions"]),
  );

  const toggleCategory = useCallback((categoryId: string) => {
    setExpandedCategories((prev) => {
      const newExpanded = new Set(prev);
      if (newExpanded.has(categoryId)) {
        newExpanded.delete(categoryId);
      } else {
        newExpanded.add(categoryId);
      }
      return newExpanded;
    });
  }, []);

  const filteredCategories = useMemo(() => {
    return nodeCategories
      .map((category) => ({
        ...category,
        nodes: category.nodes.filter(
          (node) =>
            node.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
            node.description.toLowerCase().includes(searchTerm.toLowerCase()),
        ),
      }))
      .filter((category) => category.nodes.length > 0);
  }, [searchTerm]);

  return (
    <div className="h-full flex flex-col bg-background border-r">
      {/* Header */}
      <div className="p-4 border-b space-y-3">
        {/* Search */}
        <div className="relative">
          <Search className="absolute left-2.5 top-2 h-3.5 w-3.5 text-muted-foreground" />
          <Input
            placeholder="Search nodes..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-8 h-8 text-xs"
          />
        </div>
      </div>

      {/* Content */}
      <ScrollArea className="flex-1">
        <div className="p-2 space-y-2">
          {filteredCategories.map((category) => {
            const CategoryIcon = iconMap[category.icon] || Code;
            const isExpanded = expandedCategories.has(category.id);

            return (
              <div key={category.id} className="space-y-1">
                {/* Category Header */}
                <button
                  onClick={() => toggleCategory(category.id)}
                  className="w-full flex items-center gap-2 px-2 py-1.5 text-xs font-medium hover:bg-muted rounded-md transition-colors"
                >
                  {isExpanded ? (
                    <ChevronDown className="h-3.5 w-3.5" />
                  ) : (
                    <ChevronRight className="h-3.5 w-3.5" />
                  )}
                  <CategoryIcon className="h-3.5 w-3.5" />
                  <span className="flex-1 text-left">{category.name}</span>
                  <Badge variant="secondary" className="h-4 text-[10px] px-1.5">
                    {category.nodes.length}
                  </Badge>
                </button>

                {/* Category Nodes */}
                {isExpanded && (
                  <div className="space-y-1 pl-6">
                    {category.nodes.map((node) => {
                      const NodeIcon = iconMap[node.icon] || Settings;

                      return (
                        <button
                          key={node.id}
                          onClick={() => onNodeSelect(node)}
                          className="w-full flex items-start gap-2 p-2 hover:bg-muted rounded-md transition-colors group text-left"
                        >
                          <div
                            className={`p-1.5 rounded ${getColorClass(node.color)}`}
                          >
                            <NodeIcon className="h-3 w-3" />
                          </div>
                          <div className="flex-1 min-w-0">
                            <div className="text-xs font-medium group-hover:text-primary transition-colors">
                              {node.name}
                            </div>
                            <div className="text-[10px] text-muted-foreground line-clamp-1">
                              {node.description}
                            </div>
                          </div>
                        </button>
                      );
                    })}
                  </div>
                )}
              </div>
            );
          })}

          {filteredCategories.length === 0 && (
            <div className="text-center py-8 text-sm text-muted-foreground">
              No nodes found
            </div>
          )}
        </div>
      </ScrollArea>

      {/* Footer */}
      <div className="p-2 border-t">
        <div className="text-[10px] text-muted-foreground text-center">
          Click a node to add it to the canvas
        </div>
      </div>
    </div>
  );
}
