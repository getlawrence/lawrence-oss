import Editor from "@monaco-editor/react";
import { AlertCircle, Workflow, Loader2 } from "lucide-react";
import { useState } from "react";

import { CollectorPipelineView } from "@/components/collector-pipeline/CollectorPipelineView";
import { useTheme } from "@/components/ThemeProvider";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from "@/components/ui/resizable";
import { useYamlParser } from "@/hooks/useYamlParser";

interface ConfigEditorSideBySideProps {
  value: string;
  onChange: (value: string) => void;
}

export function ConfigEditorSideBySide({
  value,
  onChange,
}: ConfigEditorSideBySideProps) {
  const [showPipeline] = useState(true);
  const { parseResult, isParsing } = useYamlParser(value, { debounceMs: 300 });
  const { theme } = useTheme();

  return (
    <div className="flex flex-col h-full">
      {/* Status Bar - Top */}
      {!parseResult.valid && parseResult.error && (
        <Alert variant="destructive" className="rounded-none border-x-0 border-t-0">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            <strong>YAML Parse Error:</strong> {parseResult.error}
          </AlertDescription>
        </Alert>
      )}

      {/* Main Editor Area */}
      {showPipeline ? (
        <ResizablePanelGroup
          direction="horizontal"
          className="flex-1 min-h-0"
        >
          {/* Left Panel - YAML Editor */}
          <ResizablePanel defaultSize={50} minSize={30}>
            <div className="h-full flex flex-col border-r">
              {/* Editor Header */}
              <div className="flex items-center justify-between px-4 py-2 border-b bg-muted/30">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium">Config</span>
                </div>
                <div className="flex items-center gap-2">
                  {isParsing && (
                    <Badge variant="outline" className="gap-1 text-xs">
                      <Loader2 className="h-3 w-3 animate-spin" />
                      Updating...
                    </Badge>
                  )}
                </div>
              </div>

              {/* Monaco Editor */}
              <div className="flex-1 overflow-hidden">
                <Editor
                  height="100%"
                  defaultLanguage="yaml"
                  value={value}
                  onChange={(value) => onChange(value || "")}
                  theme={theme === "dark" ? "vs-dark" : "vs-light"}
                  options={{
                    minimap: { enabled: false },
                    fontSize: 13,
                    lineNumbers: "on",
                    roundedSelection: false,
                    scrollBeyondLastLine: false,
                    readOnly: false,
                    automaticLayout: true,
                    wordWrap: "on",
                    scrollbar: {
                      verticalScrollbarSize: 8,
                      horizontalScrollbarSize: 8,
                    },
                    padding: { top: 16, bottom: 16 },
                  }}
                />
              </div>
            </div>
          </ResizablePanel>

          <ResizableHandle withHandle />

          {/* Right Panel - Pipeline Visualization */}
          <ResizablePanel defaultSize={50} minSize={30}>
            <div className="h-full flex flex-col bg-background">
              {/* Pipeline Header */}
              <div className="flex items-center justify-between px-4 py-2 border-b bg-muted/30">
                <div className="flex items-center gap-2">
                  <Workflow className="h-4 w-4 text-muted-foreground" />
                  <span className="text-sm font-medium">Pipeline</span>
                </div>
                <div className="flex items-center gap-2">
                  {parseResult.valid && (
                    <Badge
                      variant="outline"
                      className="text-xs bg-green-50 text-green-700 border-green-200 dark:bg-green-950 dark:text-green-400 dark:border-green-800"
                    >
                      Valid
                    </Badge>
                  )}
                  {!parseResult.valid && (
                    <Badge
                      variant="outline"
                      className="text-xs bg-red-50 text-red-700 border-red-200 dark:bg-red-950 dark:text-red-400 dark:border-red-800"
                    >
                      <AlertCircle className="h-3 w-3 mr-1" />
                      Invalid
                    </Badge>
                  )}
                </div>
              </div>

              {/* Pipeline Content */}
              <div className="flex-1 overflow-hidden bg-muted/10">
                {parseResult.valid ? (
                  <CollectorPipelineView
                    effectiveConfig={value}
                    previewMode={true}
                  />
                ) : (
                  <div className="flex items-center justify-center h-full">
                    <div className="text-center space-y-3">
                      <AlertCircle className="h-16 w-16 text-muted-foreground/40 mx-auto" />
                      <div>
                        <p className="text-sm font-medium text-muted-foreground">
                          Invalid Configuration
                        </p>
                        <p className="text-xs text-muted-foreground/70 mt-1">
                          Fix YAML errors to see pipeline visualization
                        </p>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            </div>
          </ResizablePanel>
        </ResizablePanelGroup>
      ) : (
        <div className="flex-1 overflow-hidden">
          <Editor
            height="100%"
            defaultLanguage="yaml"
            value={value}
            onChange={(value) => onChange(value || "")}
            theme={theme === "dark" ? "vs-dark" : "vs-light"}
            options={{
              minimap: { enabled: false },
              fontSize: 13,
              lineNumbers: "on",
              roundedSelection: false,
              scrollBeyondLastLine: false,
              readOnly: false,
              automaticLayout: true,
              wordWrap: "on",
            }}
          />
        </div>
      )}
    </div>
  );
}
