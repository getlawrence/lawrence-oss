import { Search, Loader2 } from "lucide-react";
import { useMemo, useState } from "react";

import type {
  TriggerNodeData,
  TelemetryTriggerConfig,
} from "../../types/flow-types";

import {
  queryLogs,
  queryMetrics,
  type LogData,
  type MetricData,
} from "@/api/telemetry";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useFormState } from "@/hooks/useDrawerForm";
import { validateCronExpression } from "@/utils";

interface TriggerConfigFormProps {
  nodeData: TriggerNodeData | null;
  onSave: (data: TriggerNodeData) => void;
  onCancel?: () => void;
}

export function TriggerConfigForm({
  nodeData,
  onSave,
  onCancel,
}: TriggerConfigFormProps) {
  const initialState = useMemo(
    () => ({
      triggerType: (nodeData?.triggerType || "manual") as
        | "manual"
        | "schedule"
        | "webhook"
        | "telemetry",
      cronExpression: nodeData?.cronExpression || "",
      timezone: nodeData?.timezone || "UTC",
      label: nodeData?.label || "",
      telemetryType: nodeData?.telemetryConfig?.type || "log",
      severity: nodeData?.telemetryConfig?.severity || "",
      pattern: nodeData?.telemetryConfig?.pattern || "",
      agentId: nodeData?.telemetryConfig?.agentId || "",
      serviceName: nodeData?.telemetryConfig?.serviceName || "",
      metricName: nodeData?.telemetryConfig?.metricName || "",
      operator: nodeData?.telemetryConfig?.operator || ">",
      threshold: nodeData?.telemetryConfig?.threshold?.toString() || "",
      timeWindow: nodeData?.telemetryConfig?.timeWindow || "",
    }),
    [nodeData],
  );

  const { state, updateField } = useFormState(initialState);
  const [error, setError] = useState<string | null>(null);
  const [cronError, setCronError] = useState<string | null>(null);
  const [previewLogs, setPreviewLogs] = useState<LogData[]>([]);
  const [previewMetrics, setPreviewMetrics] = useState<MetricData[]>([]);
  const [isPreviewLoading, setIsPreviewLoading] = useState(false);
  const [previewError, setPreviewError] = useState<string | null>(null);

  const handleSave = () => {
    if (!nodeData) return;

    setError(null);
    setCronError(null);

    // Validate schedule trigger
    if (state.triggerType === "schedule") {
      if (!state.cronExpression.trim()) {
        setError("Cron expression is required for schedule triggers");
        return;
      }

      // Validate cron expression format
      const validation = validateCronExpression(state.cronExpression);
      if (!validation.valid) {
        setCronError(validation.error || "Invalid cron expression");
        setError(validation.error || "Invalid cron expression");
        return;
      }
    }

    // Validate telemetry trigger
    if (state.triggerType === "telemetry") {
      if (state.telemetryType === "metric") {
        if (!state.metricName?.trim()) {
          setError("Metric name is required for metric triggers");
          return;
        }
        if (!state.threshold?.trim()) {
          setError("Threshold is required for metric triggers");
          return;
        }
        if (!state.timeWindow?.trim()) {
          setError("Time window is required for metric triggers");
          return;
        }
      }
    }

    let telemetryConfig: TelemetryTriggerConfig | undefined;
    if (state.triggerType === "telemetry") {
      telemetryConfig = {
        type: state.telemetryType as "log" | "metric",
      };

      if (state.telemetryType === "log") {
        if (state.severity)
          telemetryConfig.severity = state.severity as
            | "error"
            | "warn"
            | "info";
        if (state.pattern) telemetryConfig.pattern = state.pattern;
        if (state.agentId) telemetryConfig.agentId = state.agentId;
        if (state.serviceName) telemetryConfig.serviceName = state.serviceName;
      } else {
        telemetryConfig.metricName = state.metricName;
        telemetryConfig.operator = state.operator as ">" | "<" | ">=" | "<=";
        telemetryConfig.threshold = parseFloat(state.threshold);
        telemetryConfig.timeWindow = state.timeWindow;
        if (state.agentId) telemetryConfig.agentId = state.agentId;
        if (state.serviceName) telemetryConfig.serviceName = state.serviceName;
      }
    }

    const updatedData: TriggerNodeData = {
      ...nodeData,
      label: state.label || "Trigger",
      triggerType: state.triggerType,
      cronExpression:
        state.triggerType === "schedule" ? state.cronExpression : undefined,
      timezone: state.triggerType === "schedule" ? state.timezone : undefined,
      telemetryConfig:
        state.triggerType === "telemetry" ? telemetryConfig : undefined,
    };

    onSave(updatedData);
  };

  const handlePreview = async () => {
    if (state.triggerType !== "telemetry") return;

    setIsPreviewLoading(true);
    setPreviewError(null);
    setPreviewLogs([]);
    setPreviewMetrics([]);

    try {
      const endTime = new Date();
      const startTime = new Date(endTime.getTime() - 24 * 60 * 60 * 1000); // Last 24 hours

      if (state.telemetryType === "log") {
        // Query logs
        const queryParams: any = {
          start_time: startTime.toISOString(),
          end_time: endTime.toISOString(),
          limit: 100,
        };

        if (state.agentId) queryParams.agent_id = state.agentId;
        if (state.serviceName) queryParams.service_name = state.serviceName;
        if (state.severity) {
          // Map severity to OTLP severity numbers
          const severityMap: Record<string, string> = {
            error: "error",
            warn: "warn",
            info: "info",
          };
          queryParams.severity = severityMap[state.severity] || state.severity;
        }
        if (state.pattern) queryParams.search = state.pattern;

        const result = await queryLogs(queryParams);
        let filteredLogs = result.logs || [];

        // Apply pattern matching if specified (client-side filtering for regex)
        if (state.pattern && filteredLogs.length > 0) {
          try {
            const regex = new RegExp(state.pattern);
            filteredLogs = filteredLogs.filter((log) => regex.test(log.body));
          } catch {
            // If regex fails, fall back to simple string contains
            filteredLogs = filteredLogs.filter((log) =>
              log.body.includes(state.pattern),
            );
          }
        }

        setPreviewLogs(filteredLogs);
      } else {
        // Query metrics
        const queryParams: any = {
          start_time: startTime.toISOString(),
          end_time: endTime.toISOString(),
          limit: 1000,
        };

        if (state.agentId) queryParams.agent_id = state.agentId;
        if (state.serviceName) queryParams.service_name = state.serviceName;
        if (state.metricName) queryParams.metric_name = state.metricName;

        const result = await queryMetrics(queryParams);
        let filteredMetrics = result.metrics || [];

        // Filter by threshold if specified
        if (state.threshold && filteredMetrics.length > 0) {
          const thresholdValue = parseFloat(state.threshold);
          filteredMetrics = filteredMetrics.filter((metric) => {
            switch (state.operator) {
              case ">":
                return metric.value > thresholdValue;
              case "<":
                return metric.value < thresholdValue;
              case ">=":
                return metric.value >= thresholdValue;
              case "<=":
                return metric.value <= thresholdValue;
              default:
                return true;
            }
          });
        }

        setPreviewMetrics(filteredMetrics);
      }
    } catch (err) {
      setPreviewError(
        err instanceof Error ? err.message : "Failed to preview telemetry",
      );
    } finally {
      setIsPreviewLoading(false);
    }
  };

  if (!nodeData) {
    return (
      <div className="p-4 text-sm text-muted-foreground">
        No trigger node selected
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {error && (
        <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
          {error}
        </div>
      )}

      {/* Label */}
      <div className="space-y-2">
        <Label className="text-sm">Label</Label>
        <Input
          value={state.label}
          onChange={(e) => updateField("label", e.target.value)}
          placeholder="Trigger label"
          className="h-8 text-sm"
        />
      </div>

      {/* Trigger Type */}
      <div className="space-y-2">
        <Label className="text-sm">Trigger Type</Label>
        <Select
          value={state.triggerType}
          onValueChange={(value) => {
            updateField(
              "triggerType",
              value as "manual" | "schedule" | "webhook",
            );
            setError(null);
          }}
        >
          <SelectTrigger className="h-9">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="manual">Manual</SelectItem>
            <SelectItem value="schedule">Schedule</SelectItem>
            <SelectItem value="webhook">Webhook</SelectItem>
            <SelectItem value="telemetry">Telemetry</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Telemetry Configuration */}
      {state.triggerType === "telemetry" && (
        <div className="space-y-3 border-l-2 border-muted pl-4">
          <div className="space-y-2">
            <Label className="text-sm">Telemetry Type</Label>
            <Select
              value={state.telemetryType}
              onValueChange={(value) => {
                updateField("telemetryType", value as "log" | "metric");
                setError(null);
              }}
            >
              <SelectTrigger className="h-9">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="log">Log</SelectItem>
                <SelectItem value="metric">Metric</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {state.telemetryType === "log" && (
            <>
              <div className="space-y-2">
                <Label className="text-sm">Severity (Optional)</Label>
                <Select
                  value={state.severity || undefined}
                  onValueChange={(value) =>
                    updateField("severity", value || "")
                  }
                >
                  <SelectTrigger className="h-9">
                    <SelectValue placeholder="Any severity" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="error">Error</SelectItem>
                    <SelectItem value="warn">Warning</SelectItem>
                    <SelectItem value="info">Info</SelectItem>
                  </SelectContent>
                </Select>
                {state.severity && (
                  <button
                    type="button"
                    onClick={() => updateField("severity", "")}
                    className="text-xs text-muted-foreground hover:text-foreground"
                  >
                    Clear severity filter
                  </button>
                )}
              </div>

              <div className="space-y-2">
                <Label className="text-sm">Pattern (Optional)</Label>
                <Input
                  value={state.pattern}
                  onChange={(e) => updateField("pattern", e.target.value)}
                  placeholder="e.g., 'error' or regex pattern"
                  className="h-8 text-sm"
                />
                <p className="text-xs text-muted-foreground">
                  Pattern to match in log body (supports regex)
                </p>
              </div>

              <div className="space-y-2">
                <Label className="text-sm">Agent ID (Optional)</Label>
                <Input
                  value={state.agentId}
                  onChange={(e) => updateField("agentId", e.target.value)}
                  placeholder="Filter by agent ID"
                  className="h-8 text-sm"
                />
              </div>

              <div className="space-y-2">
                <Label className="text-sm">Service Name (Optional)</Label>
                <Input
                  value={state.serviceName}
                  onChange={(e) => updateField("serviceName", e.target.value)}
                  placeholder="Filter by service name"
                  className="h-8 text-sm"
                />
              </div>
            </>
          )}

          {state.telemetryType === "metric" && (
            <>
              <div className="space-y-2">
                <Label className="text-sm">
                  Metric Name <span className="text-destructive">*</span>
                </Label>
                <Input
                  value={state.metricName}
                  onChange={(e) => updateField("metricName", e.target.value)}
                  placeholder="e.g., cpu_usage"
                  className="h-8 text-sm"
                />
              </div>

              <div className="space-y-2">
                <Label className="text-sm">Operator</Label>
                <Select
                  value={state.operator}
                  onValueChange={(value) =>
                    updateField("operator", value as ">" | "<" | ">=" | "<=")
                  }
                >
                  <SelectTrigger className="h-9">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value=">">&gt; (Greater than)</SelectItem>
                    <SelectItem value="<">&lt; (Less than)</SelectItem>
                    <SelectItem value=">=">
                      &gt;= (Greater than or equal)
                    </SelectItem>
                    <SelectItem value="<=">
                      &lt;= (Less than or equal)
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label className="text-sm">
                  Threshold <span className="text-destructive">*</span>
                </Label>
                <Input
                  type="number"
                  value={state.threshold}
                  onChange={(e) => updateField("threshold", e.target.value)}
                  placeholder="e.g., 100"
                  className="h-8 text-sm"
                />
              </div>

              <div className="space-y-2">
                <Label className="text-sm">
                  Time Window <span className="text-destructive">*</span>
                </Label>
                <Input
                  value={state.timeWindow}
                  onChange={(e) => updateField("timeWindow", e.target.value)}
                  placeholder="e.g., 5m, 1h"
                  className="h-8 text-sm"
                />
                <p className="text-xs text-muted-foreground">
                  Duration string like "5m" (5 minutes) or "1h" (1 hour)
                </p>
              </div>

              <div className="space-y-2">
                <Label className="text-sm">Agent ID (Optional)</Label>
                <Input
                  value={state.agentId}
                  onChange={(e) => updateField("agentId", e.target.value)}
                  placeholder="Filter by agent ID"
                  className="h-8 text-sm"
                />
              </div>

              <div className="space-y-2">
                <Label className="text-sm">Service Name (Optional)</Label>
                <Input
                  value={state.serviceName}
                  onChange={(e) => updateField("serviceName", e.target.value)}
                  placeholder="Filter by service name"
                  className="h-8 text-sm"
                />
              </div>
            </>
          )}
        </div>
      )}

      {/* Preview Section for Telemetry Triggers */}
      {state.triggerType === "telemetry" && (
        <div className="space-y-3 border-t pt-4">
          <div className="flex items-center justify-between">
            <Label className="text-sm font-semibold">Preview (Last 24h)</Label>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={handlePreview}
              disabled={isPreviewLoading}
            >
              {isPreviewLoading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Querying...
                </>
              ) : (
                <>
                  <Search className="mr-2 h-4 w-4" />
                  Preview
                </>
              )}
            </Button>
          </div>

          {previewError && (
            <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
              {previewError}
            </div>
          )}

          {state.telemetryType === "log" && previewLogs.length > 0 && (
            <div className="space-y-2 rounded-md border p-3">
              <div className="text-sm font-medium">
                Found {previewLogs.length} matching log
                {previewLogs.length !== 1 ? "s" : ""}
              </div>
              <div className="max-h-60 space-y-2 overflow-y-auto">
                {previewLogs.slice(0, 10).map((log, idx) => (
                  <div
                    key={idx}
                    className="rounded border bg-muted/50 p-2 text-xs"
                  >
                    <div className="flex items-center justify-between">
                      <span className="font-mono text-xs text-muted-foreground">
                        {new Date(log.timestamp).toLocaleString()}
                      </span>
                      <span
                        className={`rounded px-1.5 py-0.5 text-xs ${
                          log.severity_text?.toUpperCase() === "ERROR"
                            ? "bg-red-100 text-red-800 dark:bg-red-950 dark:text-red-400"
                            : log.severity_text?.toUpperCase() === "WARN"
                              ? "bg-yellow-100 text-yellow-800 dark:bg-yellow-950 dark:text-yellow-400"
                              : "bg-blue-100 text-blue-800 dark:bg-blue-950 dark:text-blue-400"
                        }`}
                      >
                        {log.severity_text || "INFO"}
                      </span>
                    </div>
                    <div className="mt-1 truncate font-mono text-xs">
                      {log.body}
                    </div>
                    {log.service_name && (
                      <div className="mt-1 text-xs text-muted-foreground">
                        Service: {log.service_name}
                      </div>
                    )}
                  </div>
                ))}
                {previewLogs.length > 10 && (
                  <div className="text-xs text-muted-foreground">
                    ... and {previewLogs.length - 10} more
                  </div>
                )}
              </div>
            </div>
          )}

          {state.telemetryType === "log" &&
            !isPreviewLoading &&
            previewLogs.length === 0 &&
            !previewError && (
              <div className="rounded-md border p-3 text-sm text-muted-foreground">
                No matching logs found in the last 24 hours
              </div>
            )}

          {state.telemetryType === "metric" && previewMetrics.length > 0 && (
            <div className="space-y-2 rounded-md border p-3">
              <div className="text-sm font-medium">
                Found {previewMetrics.length} matching metric
                {previewMetrics.length !== 1 ? "s" : ""} exceeding threshold
              </div>
              <div className="max-h-60 space-y-2 overflow-y-auto">
                {previewMetrics.slice(0, 10).map((metric, idx) => (
                  <div
                    key={idx}
                    className="rounded border bg-muted/50 p-2 text-xs"
                  >
                    <div className="flex items-center justify-between">
                      <span className="font-semibold">
                        {metric.metric_name}
                      </span>
                      <span className="font-mono font-bold text-green-600 dark:text-green-400">
                        {metric.value.toFixed(2)}
                      </span>
                    </div>
                    <div className="mt-1 flex items-center gap-2 text-xs text-muted-foreground">
                      <span>{new Date(metric.timestamp).toLocaleString()}</span>
                      {metric.service_name && (
                        <>
                          <span>•</span>
                          <span>Service: {metric.service_name}</span>
                        </>
                      )}
                    </div>
                  </div>
                ))}
                {previewMetrics.length > 10 && (
                  <div className="text-xs text-muted-foreground">
                    ... and {previewMetrics.length - 10} more
                  </div>
                )}
              </div>
            </div>
          )}

          {state.telemetryType === "metric" &&
            !isPreviewLoading &&
            previewMetrics.length === 0 &&
            !previewError && (
              <div className="rounded-md border p-3 text-sm text-muted-foreground">
                No matching metrics found exceeding threshold in the last 24
                hours
              </div>
            )}
        </div>
      )}

      {/* Schedule Configuration */}
      {state.triggerType === "schedule" && (
        <div className="space-y-3 border-l-2 border-muted pl-4">
          <div className="space-y-2">
            <Label className="text-sm">
              Cron Expression <span className="text-destructive">*</span>
            </Label>
            <Input
              value={state.cronExpression}
              onChange={(e) => {
                updateField("cronExpression", e.target.value);
                // Validate on change
                if (e.target.value.trim()) {
                  const validation = validateCronExpression(e.target.value);
                  if (validation.valid) {
                    setCronError(null);
                  } else {
                    setCronError(validation.error || "Invalid cron expression");
                  }
                } else {
                  setCronError(null);
                }
              }}
              placeholder="0 0 0 * * *"
              className={`h-8 text-sm font-mono ${
                cronError ? "border-destructive" : ""
              }`}
            />
            {cronError && (
              <p className="text-xs text-destructive">{cronError}</p>
            )}
            <p className="text-xs text-muted-foreground">
              Format: second minute hour day month weekday (6 fields required)
              <br />
              Example: 0 0 0 * * * (runs daily at midnight)
            </p>
          </div>

          <div className="space-y-2">
            <Label className="text-sm">Timezone</Label>
            <Select
              value={state.timezone}
              onValueChange={(value) => updateField("timezone", value)}
            >
              <SelectTrigger className="h-9">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="UTC">UTC</SelectItem>
                <SelectItem value="America/New_York">
                  America/New_York
                </SelectItem>
                <SelectItem value="America/Los_Angeles">
                  America/Los_Angeles
                </SelectItem>
                <SelectItem value="Europe/London">Europe/London</SelectItem>
                <SelectItem value="Asia/Tokyo">Asia/Tokyo</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
      )}

      {/* Footer Actions */}
      <div className="flex justify-end gap-2 pt-4 border-t">
        {onCancel && (
          <Button variant="outline" onClick={onCancel}>
            Cancel
          </Button>
        )}
        <Button onClick={handleSave}>Save</Button>
      </div>
    </div>
  );
}
