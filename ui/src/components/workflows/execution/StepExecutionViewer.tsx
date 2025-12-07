import {
  AlertCircle,
  CheckCircle,
  Clock,
  RefreshCw,
  X,
  XCircle,
} from "lucide-react";
import { useState } from "react";
import useSWR from "swr";

import { StepExecutionLogs } from "./StepExecutionLogs";

import {
  getStepExecutions,
  retryStepExecution,
  type StepExecution,
} from "@/api/workflows";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { LoadingSpinner } from "@/components/ui/loading-spinner";

interface StepExecutionViewerProps {
  workflowId: string;
  executionId: string;
  autoRefresh?: boolean;
}

export function StepExecutionViewer({
  workflowId,
  executionId,
  autoRefresh = false,
}: StepExecutionViewerProps) {
  const [expandedStepId, setExpandedStepId] = useState<string | null>(null);
  const [retryingStepId, setRetryingStepId] = useState<string | null>(null);
  const [confirmRetryOpen, setConfirmRetryOpen] = useState(false);
  const [stepToRetry, setStepToRetry] = useState<StepExecution | null>(null);
  const [alertMessage, setAlertMessage] = useState<string>("");
  const [alertVariant, setAlertVariant] = useState<"default" | "destructive">(
    "default",
  );

  const { data, error, mutate } = useSWR(
    workflowId && executionId
      ? [`step-executions`, workflowId, executionId]
      : null,
    () => getStepExecutions(workflowId, executionId),
    {
      refreshInterval: autoRefresh ? 2000 : 0,
    },
  );

  const stepExecutions = data?.step_executions || [];

  const handleRetryClick = (stepExecution: StepExecution) => {
    setStepToRetry(stepExecution);
    setConfirmRetryOpen(true);
  };

  const handleRetryConfirm = async () => {
    if (!stepToRetry) return;

    setConfirmRetryOpen(false);
    setRetryingStepId(stepToRetry.id);
    try {
      await retryStepExecution(stepToRetry.id);
      await mutate();
      setAlertMessage("Step retry initiated");
      setAlertVariant("default");
    } catch (err) {
      console.error("Failed to retry step:", err);
      setAlertMessage("Failed to retry step. Please try again.");
      setAlertVariant("destructive");
    } finally {
      setRetryingStepId(null);
      setStepToRetry(null);
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "success":
        return <CheckCircle className="h-5 w-5 text-green-500" />;
      case "failed":
        return <XCircle className="h-5 w-5 text-red-500" />;
      case "running":
      case "retrying":
        return <Clock className="h-5 w-5 text-blue-500 animate-spin" />;
      case "pending":
        return <Clock className="h-5 w-5 text-gray-500" />;
      case "skipped":
        return <AlertCircle className="h-5 w-5 text-yellow-500" />;
      case "cancelled":
        return <XCircle className="h-5 w-5 text-gray-500" />;
      default:
        return <AlertCircle className="h-5 w-5 text-gray-500" />;
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
      case "retrying":
        return <Badge className="bg-blue-500">Retrying</Badge>;
      case "pending":
        return <Badge variant="outline">Pending</Badge>;
      case "skipped":
        return <Badge variant="outline">Skipped</Badge>;
      case "cancelled":
        return <Badge variant="outline">Cancelled</Badge>;
      default:
        return <Badge variant="outline">{status}</Badge>;
    }
  };

  const formatDuration = (ms?: number) => {
    if (!ms) return "—";
    if (ms < 1000) return `${ms}ms`;
    return `${(ms / 1000).toFixed(2)}s`;
  };

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString();
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

  if (stepExecutions.length === 0) {
    return (
      <Card className="p-6 text-center">
        <Clock className="h-8 w-8 mx-auto mb-2 text-muted-foreground" />
        <p className="text-muted-foreground">No step executions found</p>
      </Card>
    );
  }

  return (
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Step Executions</h3>

      {/* Alert Message */}
      {alertMessage && (
        <Alert variant={alertVariant} className="relative">
          {alertVariant === "destructive" ? (
            <AlertCircle className="h-4 w-4" />
          ) : (
            <CheckCircle className="h-4 w-4" />
          )}
          <AlertDescription className="pr-8">{alertMessage}</AlertDescription>
          <button
            onClick={() => setAlertMessage("")}
            className="absolute right-4 top-4 rounded-sm opacity-70 hover:opacity-100"
          >
            <X className="h-4 w-4" />
          </button>
        </Alert>
      )}
      <div className="space-y-2">
        {stepExecutions.map((stepExecution: StepExecution) => {
          const isExpanded = expandedStepId === stepExecution.id;
          const isRetrying = retryingStepId === stepExecution.id;

          return (
            <Card key={stepExecution.id} className="overflow-hidden">
              <div
                className="p-4 hover:bg-muted/50 transition-colors cursor-pointer"
                onClick={() =>
                  setExpandedStepId(isExpanded ? null : stepExecution.id)
                }
              >
                <div className="flex items-start justify-between">
                  <div className="flex items-start gap-4 flex-1">
                    {getStatusIcon(stepExecution.status)}
                    <div className="flex-1 space-y-2">
                      <div className="flex items-center gap-3">
                        <span className="font-medium">
                          {stepExecution.step_name}
                        </span>
                        {getStatusBadge(stepExecution.status)}
                        <span className="text-xs text-muted-foreground">
                          {stepExecution.step_type}
                        </span>
                      </div>

                      <div className="grid grid-cols-3 gap-4 text-sm">
                        <div>
                          <span className="text-muted-foreground">
                            Started:
                          </span>{" "}
                          <span>
                            {formatTimestamp(stepExecution.started_at)}
                          </span>
                        </div>
                        {stepExecution.completed_at && (
                          <div>
                            <span className="text-muted-foreground">
                              Completed:
                            </span>{" "}
                            <span>
                              {formatTimestamp(stepExecution.completed_at)}
                            </span>
                          </div>
                        )}
                        <div>
                          <span className="text-muted-foreground">
                            Duration:
                          </span>{" "}
                          <span>
                            {formatDuration(stepExecution.duration_ms)}
                          </span>
                        </div>
                      </div>

                      {stepExecution.retry_attempt > 0 && (
                        <div className="text-sm text-muted-foreground">
                          Retry attempt: {stepExecution.retry_attempt}
                        </div>
                      )}

                      {stepExecution.error && (
                        <div className="text-sm text-red-600 bg-red-50 dark:bg-red-950/20 p-2 rounded">
                          {stepExecution.error}
                        </div>
                      )}
                    </div>
                  </div>

                  <div className="flex gap-2">
                    {stepExecution.status === "failed" && (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleRetryClick(stepExecution);
                        }}
                        disabled={isRetrying}
                      >
                        {isRetrying ? (
                          <RefreshCw className="h-4 w-4 animate-spin" />
                        ) : (
                          <RefreshCw className="h-4 w-4" />
                        )}
                      </Button>
                    )}
                  </div>
                </div>
              </div>

              {isExpanded && (
                <div className="border-t p-4 bg-muted/30">
                  <StepExecutionLogs
                    stepExecutionId={stepExecution.id}
                    autoRefresh={
                      autoRefresh &&
                      (stepExecution.status === "running" ||
                        stepExecution.status === "retrying")
                    }
                  />
                </div>
              )}
            </Card>
          );
        })}
      </div>

      {/* Retry Confirmation Dialog */}
      <Dialog open={confirmRetryOpen} onOpenChange={setConfirmRetryOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Retry Step</DialogTitle>
            <DialogDescription>
              Are you sure you want to retry step "{stepToRetry?.step_name}"?
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => {
                setConfirmRetryOpen(false);
                setStepToRetry(null);
              }}
            >
              Cancel
            </Button>
            <Button onClick={handleRetryConfirm}>Retry</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
