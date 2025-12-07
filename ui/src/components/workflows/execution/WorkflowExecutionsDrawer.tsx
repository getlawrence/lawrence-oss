import {
  Clock,
  CheckCircle,
  XCircle,
  AlertCircle,
  ChevronDown,
  ChevronUp,
  ChevronRight,
} from "lucide-react";
import { useState } from "react";
import useSWR from "swr";

import { StepDetailsPanel } from "./StepDetailsPanel";
import { WorkflowTimelineViewer } from "./WorkflowTimelineViewer";

import {
  getWorkflow,
  getWorkflowExecutions,
  getStepExecutions,
  type WorkflowExecution,
} from "@/api/workflows";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { LoadingSpinner } from "@/components/ui/loading-spinner";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";

interface WorkflowExecutionsDrawerProps {
  workflowId: string | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function WorkflowExecutionsDrawer({
  workflowId,
  open,
  onOpenChange,
}: WorkflowExecutionsDrawerProps) {
  const [expandedExecutionId, setExpandedExecutionId] = useState<string | null>(
    null,
  );
  const [selectedStepId, setSelectedStepId] = useState<string | null>(null);
  const [isDetailsPanelOpen, setIsDetailsPanelOpen] = useState(true);

  const { data: workflow } = useSWR(
    workflowId && open ? `/api/v1/workflows/${workflowId}` : null,
    () => getWorkflow(workflowId!),
  );

  const { data: executions } = useSWR(
    workflowId && open ? `/api/v1/workflows/${workflowId}/executions` : null,
    () => getWorkflowExecutions(workflowId!),
    { refreshInterval: 5000 },
  );

  // Get the expanded execution
  const expandedExecution = executions?.executions.find(
    (e) => e.id === expandedExecutionId,
  );

  // Fetch step executions for the expanded execution
  const { data: stepExecutionsData } = useSWR(
    workflowId && expandedExecutionId
      ? [`step-executions`, workflowId, expandedExecutionId]
      : null,
    () => getStepExecutions(workflowId!, expandedExecutionId!),
    {
      refreshInterval: expandedExecution?.status === "running" ? 2000 : 0,
    },
  );

  const handleToggleExecution = (execution: WorkflowExecution) => {
    if (expandedExecutionId === execution.id) {
      setExpandedExecutionId(null);
      setSelectedStepId(null);
    } else {
      setExpandedExecutionId(execution.id);
      setSelectedStepId(null);
      setIsDetailsPanelOpen(true);
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "success":
        return <CheckCircle className="h-5 w-5 text-green-500" />;
      case "failed":
        return <XCircle className="h-5 w-5 text-red-500" />;
      case "running":
        return <Clock className="h-5 w-5 text-blue-500 animate-spin" />;
      default:
        return <AlertCircle className="h-5 w-5 text-yellow-500" />;
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "success":
        return <Badge className="bg-green-500">Success</Badge>;
      case "failed":
        return <Badge variant="destructive">Failed</Badge>;
      case "running":
        return <Badge className="bg-blue-500">Running</Badge>;
      default:
        return <Badge variant="outline">{status}</Badge>;
    }
  };

  const formatDuration = (ms: number) => {
    if (ms < 1000) return `${ms}ms`;
    return `${(ms / 1000).toFixed(2)}s`;
  };

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString();
  };

  return (
    <>
      <Sheet open={open} onOpenChange={onOpenChange}>
        <SheetContent className="w-full min-w[70vw] sm:max-w-4xl overflow-hidden flex flex-col p-0">
          <SheetHeader className="px-6 pt-6 pb-4 border-b flex-shrink-0">
            <SheetTitle className="flex items-center gap-2">
              <Clock className="h-5 w-5" />
              Execution History
            </SheetTitle>
            <SheetDescription>
              {workflow?.name || "Loading..."}
            </SheetDescription>
          </SheetHeader>

          <div className="flex-1 overflow-y-auto px-6 py-6 space-y-4">
            {!executions ? (
              <div className="flex justify-center items-center py-8">
                <LoadingSpinner />
              </div>
            ) : executions.executions.length === 0 ? (
              <Card className="p-12 text-center">
                <Clock className="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
                <h3 className="text-lg font-semibold mb-2">
                  No Executions Yet
                </h3>
                <p className="text-muted-foreground">
                  This workflow hasn't been executed yet.
                </p>
              </Card>
            ) : (
              executions.executions.map((execution) => {
                const isExpanded = expandedExecutionId === execution.id;
                return (
                  <Card key={execution.id} className="overflow-hidden">
                    {/* Execution Summary Header */}
                    <div
                      className="p-6 hover:bg-muted/50 transition-colors cursor-pointer"
                      onClick={() => handleToggleExecution(execution)}
                    >
                      <div className="flex items-start justify-between">
                        <div className="flex items-start gap-4 flex-1">
                          {getStatusIcon(execution.status)}

                          <div className="flex-1 space-y-2">
                            <div className="flex items-center gap-3">
                              {getStatusBadge(execution.status)}
                              <span className="text-sm text-muted-foreground">
                                {formatTimestamp(execution.started_at)}
                              </span>
                              {execution.duration_ms !== undefined && (
                                <span className="text-sm text-muted-foreground">
                                  Duration:{" "}
                                  {formatDuration(execution.duration_ms)}
                                </span>
                              )}
                            </div>

                            <div className="grid grid-cols-3 gap-4 text-sm">
                              <div>
                                <span className="text-muted-foreground">
                                  Actions Executed:
                                </span>{" "}
                                <span className="font-medium">
                                  {execution.actions_executed}
                                </span>
                              </div>
                              <div>
                                <span className="text-muted-foreground">
                                  Succeeded:
                                </span>{" "}
                                <span className="font-medium text-green-600">
                                  {execution.actions_succeeded}
                                </span>
                              </div>
                              <div>
                                <span className="text-muted-foreground">
                                  Failed:
                                </span>{" "}
                                <span className="font-medium text-red-600">
                                  {execution.actions_failed}
                                </span>
                              </div>
                            </div>

                            {execution.configs_created &&
                              execution.configs_created.length > 0 && (
                                <div className="text-sm">
                                  <span className="text-muted-foreground">
                                    Configs Created:
                                  </span>{" "}
                                  <span className="font-mono text-xs">
                                    {execution.configs_created.length}
                                  </span>
                                </div>
                              )}

                            {execution.error && (
                              <div className="text-sm text-red-600 bg-red-50 dark:bg-red-950/20 p-2 rounded">
                                {execution.error}
                              </div>
                            )}

                            {execution.metadata && (
                              <div className="flex gap-2 text-xs text-muted-foreground">
                                {execution.metadata.source && (
                                  <span>
                                    Source: {execution.metadata.source}
                                  </span>
                                )}
                                {execution.metadata.user && (
                                  <span>User: {execution.metadata.user}</span>
                                )}
                              </div>
                            )}
                          </div>
                        </div>
                        <div className="ml-4">
                          {isExpanded ? (
                            <ChevronUp className="h-5 w-5 text-muted-foreground" />
                          ) : (
                            <ChevronDown className="h-5 w-5 text-muted-foreground" />
                          )}
                        </div>
                      </div>
                    </div>

                    {/* Expanded Details */}
                    {isExpanded && expandedExecution && workflowId && (
                      <div className="border-t bg-muted/30">
                        <div className="p-6 space-y-6">
                          {/* Execution Summary Details */}
                          <div className="grid grid-cols-2 gap-4 text-sm pb-4 border-b">
                            <div>
                              <span className="text-muted-foreground">
                                Execution ID:
                              </span>{" "}
                              <code className="text-xs">{execution.id}</code>
                            </div>
                            <div>
                              <span className="text-muted-foreground">
                                Status:
                              </span>{" "}
                              {getStatusBadge(execution.status)}
                            </div>
                            <div>
                              <span className="text-muted-foreground">
                                Started:
                              </span>{" "}
                              {formatTimestamp(execution.started_at)}
                            </div>
                            {execution.completed_at && (
                              <div>
                                <span className="text-muted-foreground">
                                  Completed:
                                </span>{" "}
                                {formatTimestamp(execution.completed_at)}
                              </div>
                            )}
                            <div>
                              <span className="text-muted-foreground">
                                Actions Executed:
                              </span>{" "}
                              <span className="font-medium">
                                {execution.actions_executed}
                              </span>
                            </div>
                            <div>
                              <span className="text-muted-foreground">
                                Duration:
                              </span>{" "}
                              {execution.duration_ms !== undefined &&
                                formatDuration(execution.duration_ms)}
                            </div>
                          </div>

                          {execution.error && (
                            <div className="text-sm text-red-600 bg-red-50 dark:bg-red-950/20 p-3 rounded">
                              <strong>Error:</strong> {execution.error}
                            </div>
                          )}

                          {/* Split View: Timeline on Left, Step Details on Right */}
                          <div
                            className={`grid gap-4 ${
                              isDetailsPanelOpen ? "grid-cols-2" : "grid-cols-1"
                            }`}
                            style={{ minHeight: "500px" }}
                          >
                            {/* Left: Timeline Viewer */}
                            <div
                              className={`overflow-y-auto ${isDetailsPanelOpen ? "pr-2" : ""}`}
                            >
                              <div className="flex items-center justify-between mb-2">
                                <h3 className="text-sm font-medium text-muted-foreground">
                                  Timeline
                                </h3>
                                {selectedStepId && (
                                  <button
                                    onClick={() =>
                                      setIsDetailsPanelOpen(!isDetailsPanelOpen)
                                    }
                                    className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground transition-colors"
                                    title={
                                      isDetailsPanelOpen
                                        ? "Hide details"
                                        : "Show details"
                                    }
                                  >
                                    {isDetailsPanelOpen ? (
                                      <>
                                        <ChevronRight className="h-4 w-4" />
                                        Hide Details
                                      </>
                                    ) : (
                                      <>
                                        <ChevronRight className="h-4 w-4" />
                                        Show Details
                                      </>
                                    )}
                                  </button>
                                )}
                              </div>
                              <WorkflowTimelineViewer
                                workflowId={workflowId}
                                execution={expandedExecution}
                                autoRefresh={
                                  expandedExecution.status === "running"
                                }
                                selectedStepId={selectedStepId}
                                onStepClick={(stepId) => {
                                  setSelectedStepId(stepId);
                                  if (stepId && !isDetailsPanelOpen) {
                                    setIsDetailsPanelOpen(true);
                                  }
                                }}
                              />
                            </div>

                            {/* Right: Step Details Panel */}
                            {isDetailsPanelOpen && (
                              <div className="overflow-y-auto pl-2 border-l">
                                <div className="flex items-center justify-between mb-2">
                                  <h3 className="text-sm font-medium text-muted-foreground">
                                    Step Details
                                  </h3>
                                  <button
                                    onClick={() => setIsDetailsPanelOpen(false)}
                                    className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground transition-colors"
                                    title="Hide details"
                                  >
                                    <ChevronRight className="h-4 w-4" />
                                    Hide
                                  </button>
                                </div>
                                <StepDetailsPanel
                                  step={
                                    selectedStepId
                                      ? stepExecutionsData?.step_executions?.find(
                                          (s: any) => s.id === selectedStepId,
                                        ) || null
                                      : null
                                  }
                                  execution={expandedExecution}
                                  autoRefresh={
                                    expandedExecution.status === "running"
                                  }
                                />
                              </div>
                            )}
                          </div>
                        </div>
                      </div>
                    )}
                  </Card>
                );
              })
            )}
          </div>
        </SheetContent>
      </Sheet>
    </>
  );
}
