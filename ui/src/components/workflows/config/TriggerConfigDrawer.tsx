import { useMemo, useState } from "react";

import type { TriggerNodeData } from "../types/flow-types";

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

import { ConfigDrawerLayout } from "./ConfigDrawerLayout";

interface TriggerConfigDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  nodeData: TriggerNodeData | null;
  onSave: (data: TriggerNodeData) => void;
}

export function TriggerConfigDrawer({
  open,
  onOpenChange,
  nodeData,
  onSave,
}: TriggerConfigDrawerProps) {
  const initialState = useMemo(
    () => ({
      triggerType: (nodeData?.triggerType ||
        "manual") as "manual" | "schedule" | "webhook",
      cronExpression: nodeData?.cronExpression || "",
      timezone: nodeData?.timezone || "UTC",
      label: nodeData?.label || "",
    }),
    [nodeData]
  );

  const { state, updateField } = useFormState(initialState);
  const [error, setError] = useState<string | null>(null);

  const handleSave = () => {
    if (!nodeData) return;

    setError(null);

    // Validate schedule trigger
    if (state.triggerType === "schedule" && !state.cronExpression.trim()) {
      setError("Cron expression is required for schedule triggers");
      return;
    }

    const updatedData: TriggerNodeData = {
      ...nodeData,
      label: state.label || "Trigger",
      triggerType: state.triggerType,
      cronExpression:
        state.triggerType === "schedule" ? state.cronExpression : undefined,
      timezone: state.triggerType === "schedule" ? state.timezone : undefined,
    };

    onSave(updatedData);
    onOpenChange(false);
  };

  return (
    <ConfigDrawerLayout
      open={open}
      onOpenChange={onOpenChange}
      title="Configure Trigger"
      description="Configure the trigger settings for this node"
      error={error}
      onSave={handleSave}
      maxWidth="xl"
    >
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
              value as "manual" | "schedule" | "webhook"
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
          </SelectContent>
        </Select>
      </div>

      {/* Schedule Configuration */}
      {state.triggerType === "schedule" && (
        <div className="space-y-3 border-l-2 border-muted pl-4">
          <div className="space-y-2">
            <Label className="text-sm">
              Cron Expression <span className="text-destructive">*</span>
            </Label>
            <Input
              value={state.cronExpression}
              onChange={(e) => updateField("cronExpression", e.target.value)}
              placeholder="0 0 * * *"
              className="h-8 text-sm font-mono"
            />
            <p className="text-xs text-muted-foreground">
              Example: 0 0 * * * (runs daily at midnight)
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
    </ConfigDrawerLayout>
  );
}
