import { Play } from "lucide-react";

import type { QueryTemplate } from "../../api/lawrence-ql";
import { Badge } from "../ui/badge";
import { Button } from "../ui/button";
import { Card } from "../ui/card";

interface QueryTemplatesProps {
  templates: QueryTemplate[];
  onSelect: (query: string) => void;
}

export function QueryTemplates({ templates, onSelect }: QueryTemplatesProps) {
  // Group templates by category
  const groupedTemplates = templates.reduce(
    (acc, template) => {
      const category = template.category || "general";
      if (!acc[category]) {
        acc[category] = [];
      }
      acc[category].push(template);
      return acc;
    },
    {} as Record<string, QueryTemplate[]>,
  );

  return (
    <div className="space-y-6">
      {Object.entries(groupedTemplates).map(([category, categoryTemplates]) => (
        <div key={category}>
          <h3 className="text-lg font-semibold mb-3 capitalize">{category}</h3>
          <div className="grid gap-3">
            {categoryTemplates.map((template) => (
              <Card
                key={template.id}
                className="p-4 hover:bg-muted/50 transition-colors"
              >
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1 space-y-2">
                    <div className="flex items-center gap-2">
                      <h4 className="font-medium">{template.name}</h4>
                      <Badge variant="outline" className="text-xs">
                        {template.category}
                      </Badge>
                    </div>
                    <p className="text-sm text-muted-foreground">
                      {template.description}
                    </p>
                    <pre className="text-xs bg-muted p-2 rounded overflow-auto">
                      {template.query}
                    </pre>
                  </div>
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => onSelect(template.query)}
                  >
                    <Play className="h-4 w-4" />
                  </Button>
                </div>
              </Card>
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}
