import {
  Clock,
  Check,
  X,
  Loader2,
  ChevronRight,
  ChevronDown,
  PlayCircle,
  AlertCircle,
  Copy,
  CheckCircle2,
} from "lucide-react";
import { useState } from "react";
import useSWR from "swr";

import {
  getWorkflowExecutions,
  getStepExecutions,
  type StepExecution,
} from "@/api/workflows";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type ExecutionStatus = "running" | "success" | "failed" | "partial";

type ExecutionLog = {
  node_id: string;
  node_name: string;
  status: ExecutionStatus;
  output?: any;
  error?: string;
};

type ExecutionRunsPanelProps = {
  workflowId: string;
  onSelectExecution?: (executionId: string, logs: ExecutionLog[]) => void;
  selectedExecutionId?: string;
};

const StatusIcon = ({ status }: { status: ExecutionStatus }) => {
  switch (status) {
    case "success":
      return <Check className="h-4 w-4 text-green-500" />;
    case "failed":
      return <X className="h-4 w-4 text-red-500" />;
    case "partial":
      return <AlertCircle className="h-4 w-4 text-yellow-500" />;
    case "running":
      return <Loader2 className="h-4 w-4 animate-spin text-blue-500" />;
    default:
      return <Clock className="h-4 w-4 text-gray-500" />;
  }
};

const StatusBadge = ({ status }: { status: ExecutionStatus }) => {
  return (
    <span
      className={cn(
        "text-xs font-medium capitalize",
        status === "success" && "text-green-500",
        status === "failed" && "text-red-500",
        status === "partial" && "text-yellow-500",
        status === "running" && "text-blue-500",
      )}
    >
      {status}
    </span>
  );
};

export function ExecutionRunsPanel({
  workflowId,
  onSelectExecution,
  selectedExecutionId,
}: ExecutionRunsPanelProps) {
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [expandedSteps, setExpandedSteps] = useState<Set<string>>(new Set());
  const [copiedId, setCopiedId] = useState<string | null>(null);

  // Load executions
  const {
    data: executionsData,
    error,
    isLoading,
  } = useSWR(
    workflowId ? [`workflow-executions`, workflowId] : null,
    () => getWorkflowExecutions(workflowId),
    { refreshInterval: 5000 },
  );

  const executions = executionsData?.executions || [];

  // Load step executions for expanded execution
  const { data: stepExecutionsData } = useSWR(
    expandedId && workflowId
      ? [`step-executions`, workflowId, expandedId]
      : null,
    () => getStepExecutions(workflowId, expandedId!),
    {
      refreshInterval:
        executions.find((e) => e.id === expandedId)?.status === "running"
          ? 2000
          : 0,
    },
  );

  const stepExecutions = stepExecutionsData?.step_executions || [];

  // Handle expand/collapse execution
  const handleExpand = async (executionId: string) => {
    if (expandedId === executionId) {
      setExpandedId(null);
      setExpandedSteps(new Set());
    } else {
      setExpandedId(executionId);
      setExpandedSteps(new Set());

      // Notify parent to update node statuses
      if (onSelectExecution) {
        const logs: ExecutionLog[] = stepExecutions.map(
          (step: StepExecution) => ({
            node_id: step.step_id,
            node_name: step.step_name,
            status: step.status as ExecutionStatus,
            output: step.output_data,
            error: step.error,
          }),
        );
        onSelectExecution(executionId, logs);
      }
    }
  };

  // Toggle step expansion
  const toggleStep = (stepId: string) => {
    setExpandedSteps((prev) => {
      const next = new Set(prev);
      if (next.has(stepId)) {
        next.delete(stepId);
      } else {
        next.add(stepId);
      }
      return next;
    });
  };

  // Copy to clipboard
  const copyToClipboard = async (text: string, id: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedId(id);
      setTimeout(() => setCopiedId(null), 2000);
    } catch (err) {
      console.error("Failed to copy:", err);
    }
  };

  // Get step status icon
  const getStepStatusIcon = (status: string) => {
    switch (status) {
      case "success":
        return <Check className="h-4 w-4 text-green-500" />;
      case "failed":
        return <X className="h-4 w-4 text-red-500" />;
      case "running":
        return <Loader2 className="h-4 w-4 animate-spin text-blue-500" />;
      default:
        return <Clock className="h-4 w-4 text-gray-500" />;
    }
  };

  // Format duration
  const formatDuration = (ms?: number) => {
    if (!ms) return "—";
    if (ms < 1000) return `${ms}ms`;
    return `${(ms / 1000).toFixed(2)}s`;
  };

  if (isLoading && executions.length === 0) {
    return (
      <div className="flex items-center justify-center h-full">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center h-full text-center p-8">
        <X className="h-12 w-12 text-red-500 mb-4" />
        <p className="text-sm text-muted-foreground">{error}</p>
      </div>
    );
  }

  if (executions.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-full text-center p-8">
        <PlayCircle className="h-12 w-12 text-muted-foreground mb-4" />
        <p className="text-sm font-medium mb-2">No executions yet</p>
        <p className="text-xs text-muted-foreground">
          Run this workflow to see execution history
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-2">
      {executions.map((execution) => {
        const isExpanded = expandedId === execution.id;
        const isSelected = selectedExecutionId === execution.id;

        return (
          <div
            key={execution.id}
            className={cn(
              "border rounded-lg overflow-hidden transition-colors",
              isSelected && "border-primary",
            )}
          >
            {/* Execution header */}
            <button
              onClick={() => handleExpand(execution.id)}
              className={cn(
                "w-full flex items-center justify-between p-3 hover:bg-muted/50 transition-colors text-left",
                isSelected && "bg-muted/30",
              )}
              type="button"
            >
              <div className="flex items-center gap-2 flex-1 min-w-0">
                {isExpanded ? (
                  <ChevronDown className="h-4 w-4 flex-shrink-0" />
                ) : (
                  <ChevronRight className="h-4 w-4 flex-shrink-0" />
                )}

                <StatusIcon status={execution.status} />

                <div className="flex-1 min-w-0">
                  <div className="text-sm font-medium truncate">
                    {new Date(execution.started_at).toLocaleString()}
                  </div>
                  {execution.duration_ms && (
                    <div className="text-xs text-muted-foreground">
                      {(execution.duration_ms / 1000).toFixed(2)}s
                    </div>
                  )}
                </div>
              </div>

              <StatusBadge status={execution.status} />
            </button>

            {/* Execution details (expanded) - Step by step like Vercel */}
            {isExpanded && (
              <div className="border-t bg-muted/20">
                {/* Run summary */}
                <div className="p-3 space-y-1 text-xs text-muted-foreground border-b">
                  <div className="flex items-center gap-4">
                    <span>
                      {new Date(execution.started_at).toLocaleString()}
                    </span>
                    {execution.duration_ms && (
                      <span>{formatDuration(execution.duration_ms)}</span>
                    )}
                    <span>
                      {stepExecutions.length} step
                      {stepExecutions.length !== 1 ? "s" : ""}
                    </span>
                  </div>
                </div>

                {/* Step executions */}
                <div className="p-3 space-y-2">
                  {stepExecutions.length === 0 ? (
                    <div className="text-xs text-muted-foreground text-center py-4">
                      Loading steps...
                    </div>
                  ) : (
                    stepExecutions.map((step: StepExecution) => {
                      const isStepExpanded = expandedSteps.has(step.id);
                      const inputJson = step.input_data
                        ? JSON.stringify(step.input_data, null, 2)
                        : "";
                      const outputJson = step.output_data
                        ? JSON.stringify(step.output_data, null, 2)
                        : "";
                      const errorText = step.error || "";

                      return (
                        <div
                          key={step.id}
                          className="border rounded-md overflow-hidden"
                        >
                          {/* Step header */}
                          <button
                            onClick={() => toggleStep(step.id)}
                            className="w-full flex items-center justify-between p-2 hover:bg-muted/50 transition-colors text-left"
                            type="button"
                          >
                            <div className="flex items-center gap-2 flex-1 min-w-0">
                              {isStepExpanded ? (
                                <ChevronDown className="h-3.5 w-3.5 flex-shrink-0" />
                              ) : (
                                <ChevronRight className="h-3.5 w-3.5 flex-shrink-0" />
                              )}
                              {getStepStatusIcon(step.status)}
                              <span className="text-sm font-medium truncate">
                                {step.step_name}
                              </span>
                              {step.duration_ms !== undefined && (
                                <span className="text-xs text-muted-foreground">
                                  {formatDuration(step.duration_ms)}
                                </span>
                              )}
                            </div>
                          </button>

                          {/* Step details (expanded) */}
                          {isStepExpanded && (
                            <div className="border-t bg-background p-2 space-y-2">
                              {/* INPUT section */}
                              {step.input_data &&
                                Object.keys(step.input_data).length > 0 && (
                                  <div>
                                    <button
                                      onClick={() => {
                                        const sectionId = `${step.id}-input`;
                                        if (expandedSteps.has(sectionId)) {
                                          setExpandedSteps((prev) => {
                                            const next = new Set(prev);
                                            next.delete(sectionId);
                                            return next;
                                          });
                                        } else {
                                          setExpandedSteps((prev) =>
                                            new Set(prev).add(sectionId),
                                          );
                                        }
                                      }}
                                      className="w-full flex items-center justify-between text-xs font-medium text-muted-foreground hover:text-foreground transition-colors"
                                      type="button"
                                    >
                                      <span className="flex items-center gap-1.5">
                                        {expandedSteps.has(
                                          `${step.id}-input`,
                                        ) ? (
                                          <ChevronDown className="h-3 w-3" />
                                        ) : (
                                          <ChevronRight className="h-3 w-3" />
                                        )}
                                        &gt; INPUT
                                      </span>
                                      <Button
                                        variant="ghost"
                                        size="sm"
                                        className="h-6 w-6 p-0"
                                        onClick={(e) => {
                                          e.stopPropagation();
                                          copyToClipboard(
                                            inputJson,
                                            `${step.id}-input-copy`,
                                          );
                                        }}
                                      >
                                        {copiedId ===
                                        `${step.id}-input-copy` ? (
                                          <CheckCircle2 className="h-3 w-3 text-green-500" />
                                        ) : (
                                          <Copy className="h-3 w-3" />
                                        )}
                                      </Button>
                                    </button>
                                    {expandedSteps.has(`${step.id}-input`) && (
                                      <div className="mt-1 p-2 bg-muted rounded text-xs font-mono overflow-x-auto">
                                        <pre className="whitespace-pre-wrap">
                                          {inputJson}
                                        </pre>
                                      </div>
                                    )}
                                  </div>
                                )}

                              {/* OUTPUT section */}
                              {step.output_data &&
                                Object.keys(step.output_data).length > 0 && (
                                  <div>
                                    <button
                                      onClick={() => {
                                        const sectionId = `${step.id}-output`;
                                        if (expandedSteps.has(sectionId)) {
                                          setExpandedSteps((prev) => {
                                            const next = new Set(prev);
                                            next.delete(sectionId);
                                            return next;
                                          });
                                        } else {
                                          setExpandedSteps((prev) =>
                                            new Set(prev).add(sectionId),
                                          );
                                        }
                                      }}
                                      className="w-full flex items-center justify-between text-xs font-medium text-muted-foreground hover:text-foreground transition-colors"
                                      type="button"
                                    >
                                      <span className="flex items-center gap-1.5">
                                        {expandedSteps.has(
                                          `${step.id}-output`,
                                        ) ? (
                                          <ChevronDown className="h-3 w-3" />
                                        ) : (
                                          <ChevronRight className="h-3 w-3" />
                                        )}
                                        &gt; OUTPUT
                                      </span>
                                      <Button
                                        variant="ghost"
                                        size="sm"
                                        className="h-6 w-6 p-0"
                                        onClick={(e) => {
                                          e.stopPropagation();
                                          copyToClipboard(
                                            outputJson,
                                            `${step.id}-output-copy`,
                                          );
                                        }}
                                      >
                                        {copiedId ===
                                        `${step.id}-output-copy` ? (
                                          <CheckCircle2 className="h-3 w-3 text-green-500" />
                                        ) : (
                                          <Copy className="h-3 w-3" />
                                        )}
                                      </Button>
                                    </button>
                                    {expandedSteps.has(`${step.id}-output`) && (
                                      <div className="mt-1 p-2 bg-muted rounded text-xs font-mono overflow-x-auto">
                                        <pre className="whitespace-pre-wrap">
                                          {outputJson}
                                        </pre>
                                      </div>
                                    )}
                                  </div>
                                )}

                              {/* ERROR section */}
                              {step.error && (
                                <div>
                                  <button
                                    onClick={() => {
                                      const sectionId = `${step.id}-error`;
                                      if (expandedSteps.has(sectionId)) {
                                        setExpandedSteps((prev) => {
                                          const next = new Set(prev);
                                          next.delete(sectionId);
                                          return next;
                                        });
                                      } else {
                                        setExpandedSteps((prev) =>
                                          new Set(prev).add(sectionId),
                                        );
                                      }
                                    }}
                                    className="w-full flex items-center justify-between text-xs font-medium text-red-600 dark:text-red-400 hover:text-red-700 dark:hover:text-red-300 transition-colors"
                                    type="button"
                                  >
                                    <span className="flex items-center gap-1.5">
                                      {expandedSteps.has(`${step.id}-error`) ? (
                                        <ChevronDown className="h-3 w-3" />
                                      ) : (
                                        <ChevronRight className="h-3 w-3" />
                                      )}
                                      &gt; ERROR
                                    </span>
                                    <Button
                                      variant="ghost"
                                      size="sm"
                                      className="h-6 w-6 p-0"
                                      onClick={(e) => {
                                        e.stopPropagation();
                                        copyToClipboard(
                                          errorText,
                                          `${step.id}-error-copy`,
                                        );
                                      }}
                                    >
                                      {copiedId === `${step.id}-error-copy` ? (
                                        <CheckCircle2 className="h-3 w-3 text-green-500" />
                                      ) : (
                                        <Copy className="h-3 w-3" />
                                      )}
                                    </Button>
                                  </button>
                                  {expandedSteps.has(`${step.id}-error`) && (
                                    <div className="mt-1 p-2 bg-red-50 dark:bg-red-950/30 rounded border border-red-200 dark:border-red-900 text-xs font-mono text-red-700 dark:text-red-300 overflow-x-auto">
                                      <pre className="whitespace-pre-wrap">
                                        {errorText}
                                      </pre>
                                    </div>
                                  )}
                                </div>
                              )}
                            </div>
                          )}
                        </div>
                      );
                    })
                  )}
                </div>
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
}
