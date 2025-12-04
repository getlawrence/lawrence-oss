import { Clock } from "lucide-react";
import { useMemo } from "react";
import useSWR from "swr";

import {
  getStepExecutions,
  type StepExecution,
  type WorkflowExecution,
} from "@/api/workflows";
import { Card } from "@/components/ui/card";
import { LoadingSpinner } from "@/components/ui/loading-spinner";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

interface WorkflowTimelineViewerProps {
  workflowId: string;
  execution: WorkflowExecution;
  autoRefresh?: boolean;
  selectedStepId?: string | null;
  onStepClick?: (stepId: string | null) => void;
}

export function WorkflowTimelineViewer({
  workflowId,
  execution,
  autoRefresh = false,
  selectedStepId,
  onStepClick,
}: WorkflowTimelineViewerProps) {
  const { data, error } = useSWR(
    workflowId && execution.id
      ? [`step-executions`, workflowId, execution.id]
      : null,
    () => getStepExecutions(workflowId, execution.id),
    {
      refreshInterval: autoRefresh ? 2000 : 0,
    },
  );

  const stepExecutions = data?.step_executions || [];

  // Calculate timeline metrics
  const timelineData = useMemo(() => {
    if (stepExecutions.length === 0) {
      return null;
    }

    const workflowStartTime = new Date(execution.started_at).getTime();
    const workflowEndTime = execution.completed_at
      ? new Date(execution.completed_at).getTime()
      : Date.now();
    const totalDuration = Math.max(workflowEndTime - workflowStartTime, 1); // Ensure at least 1ms

    // Process step executions
    const processedSteps = stepExecutions.map((step: StepExecution) => {
      const startTime = new Date(step.started_at).getTime();
      const endTime = step.completed_at
        ? new Date(step.completed_at).getTime()
        : Date.now();
      const duration = Math.max(endTime - startTime, 1); // Ensure at least 1ms
      const offset = Math.max(startTime - workflowStartTime, 0);

      return {
        ...step,
        startTime,
        endTime,
        duration,
        offset,
        offsetPercent: totalDuration > 0 ? (offset / totalDuration) * 100 : 0,
        durationPercent:
          totalDuration > 0
            ? Math.max((duration / totalDuration) * 100, 0.5)
            : 0.5, // Minimum 0.5% width
      };
    });

    // Sort by start time
    processedSteps.sort((a, b) => a.startTime - b.startTime);

    return {
      workflowStartTime,
      workflowEndTime,
      totalDuration,
      steps: processedSteps,
    };
  }, [stepExecutions, execution]);

  const getStatusColor = (status: string) => {
    switch (status) {
      case "success":
        return "bg-green-500";
      case "failed":
        return "bg-red-500";
      case "running":
      case "retrying":
        return "bg-blue-500";
      case "pending":
        return "bg-gray-400";
      case "skipped":
        return "bg-yellow-500";
      case "cancelled":
        return "bg-gray-500";
      default:
        return "bg-gray-400";
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "success":
        return <div className="w-2 h-2 rounded-full bg-green-500" />;
      case "failed":
        return <div className="w-2 h-2 rounded-full bg-red-500" />;
      case "running":
      case "retrying":
        return (
          <div className="w-2 h-2 rounded-full bg-blue-500 animate-pulse" />
        );
      case "pending":
        return <div className="w-2 h-2 rounded-full bg-gray-400" />;
      case "skipped":
        return <div className="w-2 h-2 rounded-full bg-yellow-500" />;
      case "cancelled":
        return <div className="w-2 h-2 rounded-full bg-gray-500" />;
      default:
        return <div className="w-2 h-2 rounded-full bg-gray-400" />;
    }
  };

  const formatDuration = (ms: number) => {
    if (ms < 1000) return `${ms}ms`;
    return `${(ms / 1000).toFixed(2)}s`;
  };

  const formatTime = (timestamp: string) => {
    return new Date(timestamp).toLocaleTimeString();
  };

  if (error) {
    return (
      <Card className="p-6">
        <div className="text-center text-red-600">
          <p>Failed to load step executions</p>
          <p className="text-sm text-muted-foreground">{error.message}</p>
        </div>
      </Card>
    );
  }

  if (!data) {
    return (
      <Card className="p-6">
        <div className="flex justify-center items-center py-8">
          <LoadingSpinner />
        </div>
      </Card>
    );
  }

  if (!timelineData || timelineData.steps.length === 0) {
    return (
      <Card className="p-6 text-center">
        <Clock className="h-8 w-8 mx-auto mb-2 text-muted-foreground" />
        <p className="text-muted-foreground">No step executions found</p>
      </Card>
    );
  }

  const { steps, totalDuration } = timelineData;
  const maxDuration = Math.max(...steps.map((s) => s.duration), 1);
  const minBarHeight = 20;
  const maxBarHeight = 32;
  const barHeightRange = maxBarHeight - minBarHeight;

  return (
    <div>
      {/* Timeline header */}
      <div className="mb-2 pb-1 border-b border-border/50">
        <div className="flex justify-between text-[10px] text-muted-foreground mb-1">
          <span>0ms</span>
          <span>{formatDuration(totalDuration)}</span>
        </div>
        <div className="relative h-0.5 bg-border/50">
          {/* Time markers */}
          {[0, 25, 50, 75, 100].map((percent) => (
            <div
              key={percent}
              className="absolute top-0 h-full w-px bg-border/30"
              style={{ left: `${percent}%` }}
            />
          ))}
        </div>
      </div>

      {/* Timeline bars */}
      <div className="space-y-0.5">
        {steps.map((step) => {
          const barHeight =
            minBarHeight + (step.duration / maxDuration) * barHeightRange;
          const isRunning =
            step.status === "running" || step.status === "retrying";

          const isSelected = selectedStepId === step.id;

          return (
            <TooltipProvider key={step.id}>
              <Tooltip>
                <TooltipTrigger asChild>
                  <div
                    className={`relative group cursor-pointer py-0.5 ${
                      isSelected ? "ring-1 ring-primary rounded" : ""
                    }`}
                    onClick={() => onStepClick?.(isSelected ? null : step.id)}
                  >
                    <div className="flex items-center gap-2">
                      {/* Step label */}
                      <div className="flex items-center gap-1.5 min-w-[160px]">
                        {getStatusIcon(step.status)}
                        <span className="text-xs truncate text-foreground/80">
                          {step.step_name}
                        </span>
                      </div>

                      {/* Timeline bar container */}
                      <div
                        className="flex-1 relative"
                        style={{ height: `${maxBarHeight}px` }}
                      >
                        {/* Background track */}
                        <div className="absolute inset-0 bg-muted/30 rounded" />

                        {/* Step bar */}
                        <div
                          className={`absolute ${getStatusColor(
                            step.status,
                          )} rounded ${
                            isRunning ? "animate-pulse" : ""
                          } ${isSelected ? "ring-1 ring-primary/50" : ""}`}
                          style={{
                            left: `${Math.max(0, step.offsetPercent)}%`,
                            width: `${Math.max(0.5, step.durationPercent)}%`,
                            minWidth: "2px",
                            height: `${barHeight}px`,
                            top: "50%",
                            transform: "translateY(-50%)",
                          }}
                        >
                          {/* Duration label on bar */}
                          {step.durationPercent > 10 && (
                            <div className="absolute inset-0 flex items-center justify-center text-[10px] text-white font-medium whitespace-nowrap">
                              {formatDuration(step.duration)}
                            </div>
                          )}
                        </div>
                      </div>

                      {/* Duration display */}
                      <div className="min-w-[50px] text-right text-[10px] text-muted-foreground">
                        {formatDuration(step.duration)}
                      </div>
                    </div>
                  </div>
                </TooltipTrigger>
                <TooltipContent side="top" className="max-w-md">
                  <div className="space-y-1">
                    <div className="font-semibold">{step.step_name}</div>
                    <div className="text-xs">
                      <div>Type: {step.step_type}</div>
                      <div>Status: {step.status}</div>
                      <div>Started: {formatTime(step.started_at)}</div>
                      {step.completed_at && (
                        <div>Completed: {formatTime(step.completed_at)}</div>
                      )}
                      <div>Duration: {formatDuration(step.duration)}</div>
                      {step.error && (
                        <div className="text-red-500 mt-1">
                          Error: {step.error}
                        </div>
                      )}
                    </div>
                  </div>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          );
        })}
      </div>
    </div>
  );
}
