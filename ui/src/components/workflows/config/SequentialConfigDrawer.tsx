import { useMemo, useState } from "react";

import type { SequentialNodeData } from "../types/flow-types";

import { ConfigDrawerLayout } from "./ConfigDrawerLayout";

import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useFormState } from "@/hooks/useDrawerForm";

interface SequentialConfigDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  nodeData: SequentialNodeData | null;
  onSave: (data: SequentialNodeData) => void;
}

export function SequentialConfigDrawer({
  open,
  onOpenChange,
  nodeData,
  onSave,
}: SequentialConfigDrawerProps) {
  // Derive initial state from props - only computed when nodeData changes
  const initialState = useMemo(
    () => ({
      delayBetween: nodeData?.delayBetween || 0,
      label: nodeData?.label || "",
      description: nodeData?.description || "",
    }),
    [nodeData]
  );

  // Use single state object instead of multiple useState calls
  const { state, updateField } = useFormState(initialState);
  const [error, setError] = useState<string | null>(null);

  const handleSave = () => {
    setError(null);

    if (state.delayBetween < 0) {
      setError("Delay cannot be negative");
      return;
    }

    const updatedData: SequentialNodeData = {
      label: state.label || "Sequential Execution",
      description: state.description || undefined,
      delayBetween: state.delayBetween,
    };

    onSave(updatedData);
  };

  return (
    <ConfigDrawerLayout
      open={open}
      onOpenChange={onOpenChange}
      title="Configure Sequential Execution"
      description="Configure sequential execution with optional delays between actions"
      error={error}
      onSave={handleSave}
    >
      <div className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="label">Label</Label>
          <Input
            id="label"
            placeholder="Sequential Execution"
            value={state.label}
            onChange={(e) => updateField("label", e.target.value)}
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="description">Description (Optional)</Label>
          <Input
            id="description"
            placeholder="Brief description of sequential execution"
            value={state.description}
            onChange={(e) => updateField("description", e.target.value)}
          />
        </div>
      </div>

      <div className="space-y-2">
        <Label htmlFor="delayBetween">
          Delay Between Actions (seconds)
        </Label>
        <Input
          id="delayBetween"
          type="number"
          min="0"
          placeholder="0"
          value={state.delayBetween}
          onChange={(e) =>
            updateField("delayBetween", parseInt(e.target.value) || 0)
          }
        />
        <p className="text-xs text-muted-foreground">
          Delay in seconds between each action in the sequence. Set to 0 for
          no delay.
        </p>
      </div>

      <div className="bg-muted p-4 rounded-md">
        <p className="text-sm font-medium mb-1">Execution Mode</p>
        <p className="text-sm">
          Actions will execute sequentially
          {state.delayBetween > 0 && (
            <span className="font-semibold">
              {" "}
              with {state.delayBetween} second
              {state.delayBetween !== 1 ? "s" : ""} delay between each
            </span>
          )}
        </p>
      </div>
    </ConfigDrawerLayout>
  );
}
