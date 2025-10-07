import { Play, Loader2, History, BookOpen } from "lucide-react";
import { useState } from "react";

import { useLawrenceQL, useQueryTemplates } from "../../hooks/useLawrenceQL";
import { Alert, AlertDescription } from "../ui/alert";
import { Button } from "../ui/button";
import { Card } from "../ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../ui/tabs";
import { Textarea } from "../ui/textarea";

import { QueryHistory } from "./QueryHistory";
import { QueryResults } from "./QueryResults";
import { QueryTemplates } from "./QueryTemplates";

export function LawrenceQLInterface() {
  const {
    query,
    setQuery,
    isExecuting,
    results,
    error,
    history,
    execute,
    clearResults,
  } = useLawrenceQL();

  const { templates } = useQueryTemplates();
  const [activeTab, setActiveTab] = useState("query");

  const handleExecute = () => {
    execute();
  };

  const handleTemplateSelect = (templateQuery: string) => {
    setQuery(templateQuery);
    setActiveTab("query");
  };

  const handleHistorySelect = (historyQuery: string) => {
    setQuery(historyQuery);
    setActiveTab("query");
  };

  return (
    <div className="flex flex-col gap-4 h-full">
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="query" className="gap-2">
            <Play className="h-4 w-4" />
            Query
          </TabsTrigger>
          <TabsTrigger value="templates" className="gap-2">
            <BookOpen className="h-4 w-4" />
            Templates
          </TabsTrigger>
          <TabsTrigger value="history" className="gap-2">
            <History className="h-4 w-4" />
            History
          </TabsTrigger>
        </TabsList>

        <TabsContent value="query" className="space-y-4">
          {/* Query Editor */}
          <Card className="p-4">
            <div className="space-y-4">
              <div>
                <label className="text-sm font-medium mb-2 block">
                  Lawrence QL Query
                </label>
                <Textarea
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  placeholder='Example: metrics{service="api"} [5m]'
                  className="font-mono min-h-[120px]"
                  onKeyDown={(e) => {
                    if ((e.metaKey || e.ctrlKey) && e.key === "Enter") {
                      handleExecute();
                    }
                  }}
                />
                <div className="text-xs text-muted-foreground mt-2">
                  Press Cmd/Ctrl + Enter to execute
                </div>
              </div>

              <div className="flex gap-2">
                <Button
                  onClick={handleExecute}
                  disabled={isExecuting || !query.trim()}
                >
                  {isExecuting ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Executing...
                    </>
                  ) : (
                    <>
                      <Play className="mr-2 h-4 w-4" />
                      Execute Query
                    </>
                  )}
                </Button>
                <Button
                  variant="outline"
                  onClick={clearResults}
                  disabled={!results && !error}
                >
                  Clear Results
                </Button>
              </div>
            </div>
          </Card>

          {/* Error Display */}
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {/* Results Display */}
          {results && <QueryResults results={results} />}
        </TabsContent>

        <TabsContent value="templates">
          <QueryTemplates
            templates={templates}
            onSelect={handleTemplateSelect}
          />
        </TabsContent>

        <TabsContent value="history">
          <QueryHistory history={history} onSelect={handleHistorySelect} />
        </TabsContent>
      </Tabs>
    </div>
  );
}
