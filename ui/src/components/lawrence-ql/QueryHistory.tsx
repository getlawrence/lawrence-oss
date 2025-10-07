import { Play, Clock, Database } from "lucide-react";

import type { QueryHistoryItem } from "../../hooks/useLawrenceQL";
import { Badge } from "../ui/badge";
import { Button } from "../ui/button";
import { Card } from "../ui/card";

interface QueryHistoryProps {
  history: QueryHistoryItem[];
  onSelect: (query: string) => void;
}

export function QueryHistory({ history, onSelect }: QueryHistoryProps) {
  if (history.length === 0) {
    return (
      <Card className="p-8 text-center">
        <p className="text-muted-foreground">No query history yet</p>
        <p className="text-sm text-muted-foreground mt-1">
          Execute queries to see them here
        </p>
      </Card>
    );
  }

  return (
    <div className="space-y-3">
      {history.map((item, idx) => (
        <Card key={idx} className="p-4 hover:bg-muted/50 transition-colors">
          <div className="flex items-start justify-between gap-4">
            <div className="flex-1 space-y-2">
              <div className="flex items-center gap-3">
                <span className="text-xs text-muted-foreground">
                  {item.timestamp.toLocaleString()}
                </span>
                {item.executionTime !== undefined && (
                  <Badge variant="outline" className="text-xs">
                    <Clock className="h-3 w-3 mr-1" />
                    {(item.executionTime / 1000000).toFixed(2)}ms
                  </Badge>
                )}
                {item.rowCount !== undefined && (
                  <Badge variant="outline" className="text-xs">
                    <Database className="h-3 w-3 mr-1" />
                    {item.rowCount} results
                  </Badge>
                )}
              </div>
              <pre className="text-xs bg-muted p-2 rounded overflow-auto">
                {item.query}
              </pre>
            </div>
            <Button
              size="sm"
              variant="ghost"
              onClick={() => onSelect(item.query)}
            >
              <Play className="h-4 w-4" />
            </Button>
          </div>
        </Card>
      ))}
    </div>
  );
}
