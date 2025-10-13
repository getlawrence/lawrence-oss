import Editor, { type OnMount } from "@monaco-editor/react";
import { AlertCircle, Workflow, Loader2, AlertTriangle } from "lucide-react";
import type { editor } from "monaco-editor";
import { useState, useRef } from "react";

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
import { useYamlValidation } from "@/hooks/useYamlValidation";

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
  const editorRef = useRef<editor.IStandaloneCodeEditor | null>(null);
  const { validationResult, isValidating } = useYamlValidation(
    value,
    editorRef,
  );

  const handleEditorDidMount: OnMount = (editor) => {
    editorRef.current = editor;

    // Configure hover provider to show validation errors
    const monaco = (
      window as unknown as { monaco?: typeof import("monaco-editor") }
    ).monaco;
    if (monaco) {
      monaco.languages.registerHoverProvider("yaml", {
        provideHover: (model, position) => {
          const markers = monaco.editor.getModelMarkers({
            resource: model.uri,
          });

          const hoveredMarkers = markers.filter(
            (marker) =>
              marker.startLineNumber <= position.lineNumber &&
              marker.endLineNumber >= position.lineNumber &&
              marker.startColumn <= position.column &&
              marker.endColumn >= position.column,
          );

          if (hoveredMarkers.length > 0) {
            const contents = hoveredMarkers.map((marker) => ({
              value: `**${marker.severity === 8 ? "Error" : "Warning"}**: ${marker.message}`,
            }));

            return {
              contents,
            };
          }

          return null;
        },
      });
    }
  };

  return (
    <div className="flex flex-col h-full">
      {/* Status Bar - Top */}
      {!parseResult.valid && parseResult.error && (
        <Alert
          variant="destructive"
          className="rounded-none border-x-0 border-t-0"
        >
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            <strong>YAML Parse Error:</strong> {parseResult.error}
          </AlertDescription>
        </Alert>
      )}
      {parseResult.valid && validationResult.errors.length > 0 && (
        <Alert
          variant={
            validationResult.errors.some((e) => e.severity === "error")
              ? "destructive"
              : "default"
          }
          className="rounded-none border-x-0 border-t-0"
        >
          {validationResult.errors.some((e) => e.severity === "error") ? (
            <AlertCircle className="h-4 w-4" />
          ) : (
            <AlertTriangle className="h-4 w-4" />
          )}
          <AlertDescription>
            <strong>
              {validationResult.errors.filter((e) => e.severity === "error")
                .length > 0 && (
                <>
                  {
                    validationResult.errors.filter(
                      (e) => e.severity === "error",
                    ).length
                  }{" "}
                  error(s)
                </>
              )}
              {validationResult.errors.filter((e) => e.severity === "error")
                .length > 0 &&
                validationResult.errors.filter((e) => e.severity === "warning")
                  .length > 0 &&
                ", "}
              {validationResult.errors.filter((e) => e.severity === "warning")
                .length > 0 && (
                <>
                  {
                    validationResult.errors.filter(
                      (e) => e.severity === "warning",
                    ).length
                  }{" "}
                  warning(s)
                </>
              )}
            </strong>{" "}
            - Hover over highlighted text for details
          </AlertDescription>
        </Alert>
      )}

      {/* Main Editor Area */}
      {showPipeline ? (
        <ResizablePanelGroup direction="horizontal" className="flex-1 min-h-0">
          {/* Left Panel - YAML Editor */}
          <ResizablePanel defaultSize={50} minSize={30}>
            <div className="h-full flex flex-col border-r">
              {/* Editor Header */}
              <div className="flex items-center justify-between px-4 py-2 border-b bg-muted/30">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium">Config</span>
                </div>
                <div className="flex items-center gap-2">
                  {(isParsing || isValidating) && (
                    <Badge variant="outline" className="gap-1 text-xs">
                      <Loader2 className="h-3 w-3 animate-spin" />
                      {isValidating ? "Validating..." : "Updating..."}
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
                  onMount={handleEditorDidMount}
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
                  {parseResult.valid && validationResult.valid && (
                    <Badge
                      variant="outline"
                      className="text-xs bg-green-50 text-green-700 border-green-200 dark:bg-green-950 dark:text-green-400 dark:border-green-800"
                    >
                      Valid
                    </Badge>
                  )}
                  {parseResult.valid &&
                    !validationResult.valid &&
                    validationResult.errors.some(
                      (e) => e.severity === "warning",
                    ) &&
                    !validationResult.errors.some(
                      (e) => e.severity === "error",
                    ) && (
                      <Badge
                        variant="outline"
                        className="text-xs bg-yellow-50 text-yellow-700 border-yellow-200 dark:bg-yellow-950 dark:text-yellow-400 dark:border-yellow-800"
                      >
                        <AlertTriangle className="h-3 w-3 mr-1" />
                        Warnings
                      </Badge>
                    )}
                  {(!parseResult.valid ||
                    (parseResult.valid &&
                      validationResult.errors.some(
                        (e) => e.severity === "error",
                      ))) && (
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
