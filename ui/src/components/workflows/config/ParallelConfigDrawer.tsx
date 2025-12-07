import { useMemo, useState } from "react";

import type { ParallelNodeData } from "../types/flow-types";

import { ConfigDrawerLayout } from "./ConfigDrawerLayout";

import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { useFormState } from "@/hooks/useDrawerForm";

interface ParallelConfigDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  nodeData: ParallelNodeData | null;
  onSave: (data: ParallelNodeData) => void;
}

export function ParallelConfigDrawer({
  open,
  onOpenChange,
  nodeData,
  onSave,
}: ParallelConfigDrawerProps) {
  const initialState = useMemo(
    () => ({
      waitForAll: nodeData?.waitForAll ?? true,
      timeout: nodeData?.timeout,
      label: nodeData?.label || "",
      description: nodeData?.description || "",
    }),
    [nodeData],
  );

  const { state, updateField } = useFormState(initialState);
  const [error, setError] = useState<string | null>(null);

  const handleSave = () => {
    setError(null);

    if (state.timeout !== undefined && state.timeout <= 0) {
      setError("Timeout must be greater than 0");
      return;
    }

    const updatedData: ParallelNodeData = {
      label: state.label || "Parallel Execution",
      description: state.description || undefined,
      waitForAll: state.waitForAll,
      timeout: state.timeout && state.timeout > 0 ? state.timeout : undefined,
    };

    onSave(updatedData);
  };

  return (
    <ConfigDrawerLayout
      open={open}
      onOpenChange={onOpenChange}
      title="Configure Parallel Execution"
      description="Configure parallel execution settings for multiple workflow paths"
      error={error}
      onSave={handleSave}
    >
      <div className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="label">Label</Label>
          <Input
            id="label"
            placeholder="Parallel Execution"
            value={state.label}
            onChange={(e) => updateField("label", e.target.value)}
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="description">Description (Optional)</Label>
          <Input
            id="description"
            placeholder="Brief description of parallel execution"
            value={state.description}
            onChange={(e) => updateField("description", e.target.value)}
          />
        </div>
      </div>

      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div className="space-y-0.5">
            <Label htmlFor="waitForAll">Wait for All Paths</Label>
            <p className="text-xs text-muted-foreground">
              Wait for all parallel paths to complete before continuing
            </p>
          </div>
          <Switch
            id="waitForAll"
            checked={state.waitForAll}
            onCheckedChange={(checked) => updateField("waitForAll", checked)}
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="timeout">Timeout (seconds, optional)</Label>
          <Input
            id="timeout"
            type="number"
            min="1"
            placeholder="No timeout"
            value={state.timeout || ""}
            onChange={(e) => {
              const value = e.target.value;
              updateField("timeout", value ? parseInt(value) : undefined);
            }}
          />
          <p className="text-xs text-muted-foreground">
            Maximum time to wait for parallel paths to complete. Leave empty for
            no timeout.
          </p>
        </div>
      </div>

      <div className="bg-muted p-4 rounded-md">
        <p className="text-sm font-medium mb-2">Execution Mode</p>
        <p className="text-sm">
          {state.waitForAll
            ? "All parallel paths must complete before continuing"
            : "Continue as soon as any path completes"}
        </p>
        {state.timeout && (
          <p className="text-sm mt-2">Timeout: {state.timeout} seconds</p>
        )}
      </div>
    </ConfigDrawerLayout>
  );
}
