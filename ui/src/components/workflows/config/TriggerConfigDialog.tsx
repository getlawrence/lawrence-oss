import { Save } from "lucide-react";
import { useEffect, useMemo, useState } from "react";

import type { TriggerNodeData } from "../types/flow-types";

import { Alert, AlertDescription } from "@/components/ui/alert";
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
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { useFormState } from "@/hooks/useDrawerForm";

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
      triggerType: (nodeData?.triggerType || "manual") as
        | "manual"
        | "schedule"
        | "webhook",
      cronExpression: nodeData?.cronExpression || "",
      timezone: nodeData?.timezone || "UTC",
      label: nodeData?.label || "",
    }),
    [nodeData],
  );

  const { state, updateField, resetState } = useFormState(initialState);
  const [error, setError] = useState<string | null>(null);

  // Reset form state when drawer opens with new nodeData
  // This is appropriate useEffect usage - resetting state when a prop changes
  useEffect(() => {
    if (open && nodeData) {
      resetState(initialState);
      setError(null);
    }
  }, [open, nodeData, initialState, resetState]);

  const handleSave = () => {
    if (!nodeData) return;

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
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="w-full sm:max-w-xl overflow-y-auto p-6">
        <SheetHeader>
          <SheetTitle>Configure Trigger</SheetTitle>
          <SheetDescription>
            Configure the trigger settings for this node
          </SheetDescription>
        </SheetHeader>

        <div className="mt-4 space-y-4">
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
                  onChange={(e) =>
                    updateField("cronExpression", e.target.value)
                  }
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

          {/* Error Alert */}
          {error && (
            <Alert variant="destructive">
              <AlertDescription className="text-sm">{error}</AlertDescription>
            </Alert>
          )}

          {/* Save Button */}
          <div className="flex justify-end gap-2 pt-3 border-t sticky bottom-0 bg-background pb-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button size="sm" onClick={handleSave}>
              <Save className="h-3 w-3 mr-2" />
              Save
            </Button>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  );
}
