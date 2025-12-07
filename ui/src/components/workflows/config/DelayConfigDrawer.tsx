import { useMemo, useState } from "react";

import type { DelayNodeData } from "../types/flow-types";

import { ConfigDrawerLayout } from "./ConfigDrawerLayout";

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

interface DelayConfigDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  nodeData: DelayNodeData | null;
  onSave: (data: DelayNodeData) => void;
}

export function DelayConfigDrawer({
  open,
  onOpenChange,
  nodeData,
  onSave,
}: DelayConfigDrawerProps) {
  const initialState = useMemo(
    () => ({
      duration: nodeData?.duration || 5,
      unit: (nodeData?.unit || "seconds") as "seconds" | "minutes" | "hours",
      label: nodeData?.label || "",
      description: nodeData?.description || "",
    }),
    [nodeData],
  );

  const { state, updateField } = useFormState(initialState);
  const [error, setError] = useState<string | null>(null);

  const handleSave = () => {
    setError(null);

    if (state.duration <= 0) {
      setError("Duration must be greater than 0");
      return;
    }

    const updatedData: DelayNodeData = {
      label: state.label || "Delay",
      description: state.description || undefined,
      duration: state.duration,
      unit: state.unit,
    };

    onSave(updatedData);
  };

  return (
    <ConfigDrawerLayout
      open={open}
      onOpenChange={onOpenChange}
      title="Configure Delay"
      description="Configure delay duration for this workflow step"
      error={error}
      onSave={handleSave}
    >
      <div className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="label">Label</Label>
          <Input
            id="label"
            placeholder="Delay"
            value={state.label}
            onChange={(e) => updateField("label", e.target.value)}
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="description">Description (Optional)</Label>
          <Input
            id="description"
            placeholder="Brief description of this delay"
            value={state.description}
            onChange={(e) => updateField("description", e.target.value)}
          />
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label htmlFor="duration">Duration *</Label>
          <Input
            id="duration"
            type="number"
            min="1"
            placeholder="5"
            value={state.duration}
            onChange={(e) =>
              updateField("duration", parseInt(e.target.value) || 0)
            }
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="unit">Unit *</Label>
          <Select
            value={state.unit}
            onValueChange={(value: any) => updateField("unit", value)}
          >
            <SelectTrigger id="unit">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="seconds">Seconds</SelectItem>
              <SelectItem value="minutes">Minutes</SelectItem>
              <SelectItem value="hours">Hours</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <div className="bg-muted p-4 rounded-md">
        <p className="text-sm font-medium mb-1">Delay Duration</p>
        <p className="text-2xl font-bold">
          {state.duration} {state.unit}
        </p>
        <p className="text-xs text-muted-foreground mt-2">
          The workflow will pause for this duration before continuing to the
          next step.
        </p>
      </div>
    </ConfigDrawerLayout>
  );
}
