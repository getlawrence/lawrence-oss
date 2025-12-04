import { DiffEditor } from "@monaco-editor/react";
import { CheckCircle, Clock, XCircle, AlertCircle, Info } from "lucide-react";
import { useState, useEffect } from "react";

import { StepExecutionLogs } from "./StepExecutionLogs";

import { getConfig } from "@/api/configs";
import { type StepExecution } from "@/api/workflows";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";

interface StepDetailsPanelProps {
  step: StepExecution | null;
  execution: { configs_created?: string[] } | null;
  autoRefresh?: boolean;
}

export function StepDetailsPanel({
  step,
  execution,
  autoRefresh = false,
}: StepDetailsPanelProps) {
  const [configDiff, setConfigDiff] = useState<{
    old: string;
    new: string;
  } | null>(null);

  // Check if this step created a config and load it for diff view
  useEffect(() => {
    // Check if step output_data has config_id
    const configIdFromOutput = step?.output_data?.config_id as
      | string
      | undefined;

    // Or check if step type is action and execution has configs_created
    // For action steps that create configs, we'll try to match with execution's configs_created
    const configId =
      configIdFromOutput ||
      (step?.step_type === "action" &&
      execution?.configs_created &&
      execution.configs_created.length > 0
        ? execution.configs_created[0] // For now, use first config if multiple
        : undefined);

    if (configId) {
      getConfig(configId)
        .then((config) => {
          // Try to get previous version
          let oldContent = "";
          if (config.version > 1) {
            oldContent = "// Previous version not available";
          }
          setConfigDiff({
            old: oldContent,
            new: config.content || "",
          });
        })
        .catch((error) => {
          console.error("Failed to load config:", error);
          setConfigDiff(null);
        });
    } else {
      setConfigDiff(null);
    }
  }, [step?.output_data, step?.step_type, execution?.configs_created]);

  if (!step) {
    return (
      <Card className="p-6 h-full flex items-center justify-center">
        <div className="text-center text-muted-foreground">
          <Info className="h-12 w-12 mx-auto mb-4 opacity-50" />
          <p>Select a step to view details</p>
        </div>
      </Card>
    );
  }

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

  return (
    <div className="h-full overflow-y-auto space-y-6">
      {/* Step Header */}
      <div className="space-y-4">
        <div className="flex items-start gap-3">
          {getStatusIcon(step.status)}
          <div className="flex-1">
            <h2 className="text-xl font-semibold">{step.step_name}</h2>
            <div className="flex items-center gap-2 mt-1">
              {getStatusBadge(step.status)}
              <Badge variant="outline">{step.step_type}</Badge>
            </div>
          </div>
        </div>

        {/* Step Info */}
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <span className="text-muted-foreground">Step ID:</span>{" "}
            <code className="text-xs">{step.step_id}</code>
          </div>
          <div>
            <span className="text-muted-foreground">Status:</span>{" "}
            {getStatusBadge(step.status)}
          </div>
          <div>
            <span className="text-muted-foreground">Started:</span>{" "}
            {formatTimestamp(step.started_at)}
          </div>
          {step.completed_at && (
            <div>
              <span className="text-muted-foreground">Completed:</span>{" "}
              {formatTimestamp(step.completed_at)}
            </div>
          )}
          <div>
            <span className="text-muted-foreground">Duration:</span>{" "}
            {formatDuration(step.duration_ms)}
          </div>
          {step.retry_attempt > 0 && (
            <div>
              <span className="text-muted-foreground">Retry Attempt:</span>{" "}
              {step.retry_attempt}
            </div>
          )}
        </div>

        {step.error && (
          <div className="text-sm text-red-600 bg-red-50 dark:bg-red-950/20 p-3 rounded">
            <strong>Error:</strong> {step.error}
            {step.error_code && (
              <div className="mt-1 text-xs">
                <strong>Error Code:</strong> {step.error_code}
              </div>
            )}
          </div>
        )}
      </div>

      {/* Configuration Changes */}
      {configDiff && (
        <div className="space-y-2">
          <h3 className="font-semibold">Configuration Changes</h3>
          <div
            className="border rounded-lg overflow-hidden"
            style={{ height: "400px" }}
          >
            <DiffEditor
              original={configDiff.old}
              modified={configDiff.new}
              language="yaml"
              theme="vs-dark"
              options={{
                readOnly: true,
                renderSideBySide: true,
                minimap: { enabled: false },
                scrollBeyondLastLine: false,
              }}
            />
          </div>
        </div>
      )}

      {/* Input Data */}
      {step.input_data && Object.keys(step.input_data).length > 0 && (
        <div className="space-y-2">
          <h3 className="font-semibold">Input Data</h3>
          <Card className="p-4">
            <pre className="text-xs overflow-x-auto">
              {JSON.stringify(step.input_data, null, 2)}
            </pre>
          </Card>
        </div>
      )}

      {/* Output Data */}
      {step.output_data && Object.keys(step.output_data).length > 0 && (
        <div className="space-y-2">
          <h3 className="font-semibold">Output Data</h3>
          <Card className="p-4">
            <pre className="text-xs overflow-x-auto">
              {JSON.stringify(step.output_data, null, 2)}
            </pre>
          </Card>
        </div>
      )}

      {/* Metadata */}
      {step.metadata && Object.keys(step.metadata).length > 0 && (
        <div className="space-y-2">
          <h3 className="font-semibold">Metadata</h3>
          <Card className="p-4">
            <div className="space-y-1 text-sm">
              {Object.entries(step.metadata).map(([key, value]) => (
                <div key={key}>
                  <span className="text-muted-foreground">{key}:</span> {value}
                </div>
              ))}
            </div>
          </Card>
        </div>
      )}

      {/* Execution Logs */}
      <div className="space-y-2">
        <StepExecutionLogs
          stepExecutionId={step.id}
          autoRefresh={
            autoRefresh &&
            (step.status === "running" || step.status === "retrying")
          }
        />
      </div>
    </div>
  );
}
