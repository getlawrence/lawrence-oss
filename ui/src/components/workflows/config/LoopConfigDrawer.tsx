import { useMemo, useState } from "react";

import type { LoopNodeData } from "../types/flow-types";

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
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import { useFormState } from "@/hooks/useDrawerForm";

interface LoopConfigDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  nodeData: LoopNodeData | null;
  onSave: (data: LoopNodeData) => void;
}

export function LoopConfigDrawer({
  open,
  onOpenChange,
  nodeData,
  onSave,
}: LoopConfigDrawerProps) {
  const initialState = useMemo(
    () => ({
      loopType: (nodeData?.loopType || "agents") as "agents" | "groups" | "range",
      filter: nodeData?.filter || "",
      maxIterations: nodeData?.maxIterations,
      parallelExecution: nodeData?.parallelExecution ?? false,
      label: nodeData?.label || "",
      description: nodeData?.description || "",
    }),
    [nodeData]
  );

  const { state, updateField } = useFormState(initialState);
  const [error, setError] = useState<string | null>(null);

  const handleSave = () => {
    setError(null);

    if (!state.loopType) {
      setError("Loop type is required");
      return;
    }

    if (
      state.maxIterations !== undefined &&
      state.maxIterations <= 0
    ) {
      setError("Max iterations must be greater than 0");
      return;
    }

    const updatedData: LoopNodeData = {
      label: state.label || `Loop ${state.loopType}`,
      description: state.description || undefined,
      loopType: state.loopType,
      filter: state.filter.trim() || undefined,
      maxIterations:
        state.maxIterations && state.maxIterations > 0
          ? state.maxIterations
          : undefined,
      parallelExecution: state.parallelExecution,
    };

    onSave(updatedData);
  };

  return (
    <ConfigDrawerLayout
      open={open}
      onOpenChange={onOpenChange}
      title="Configure Loop"
      description="Configure loop settings for iterating over agents, groups, or ranges"
      error={error}
      onSave={handleSave}
    >
      <div className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="label">Label</Label>
          <Input
            id="label"
            placeholder="Loop"
            value={state.label}
            onChange={(e) => updateField("label", e.target.value)}
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="description">Description (Optional)</Label>
          <Input
            id="description"
            placeholder="Brief description of this loop"
            value={state.description}
            onChange={(e) => updateField("description", e.target.value)}
          />
        </div>
      </div>

      <div className="space-y-2">
        <Label htmlFor="loopType">Loop Type *</Label>
        <Select
          value={state.loopType}
          onValueChange={(value: any) => updateField("loopType", value)}
        >
          <SelectTrigger id="loopType">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="agents">Loop Agents</SelectItem>
            <SelectItem value="groups">Loop Groups</SelectItem>
            <SelectItem value="range">Loop Range</SelectItem>
          </SelectContent>
        </Select>
        <p className="text-xs text-muted-foreground">
          {state.loopType === "agents" && "Iterate over all agents"}
          {state.loopType === "groups" && "Iterate over all groups"}
          {state.loopType === "range" && "Iterate over a numeric range"}
        </p>
      </div>

      {(state.loopType === "agents" || state.loopType === "groups") && (
        <div className="space-y-2">
          <Label htmlFor="filter">Filter Expression (Optional)</Label>
          <Textarea
            id="filter"
            placeholder="name=production OR tags=critical"
            value={state.filter}
            onChange={(e) => updateField("filter", e.target.value)}
            rows={3}
            className="font-mono text-sm"
          />
          <p className="text-xs text-muted-foreground">
            Filter expression to select specific {state.loopType}. Supports
            pattern matching with * and ?.
          </p>
        </div>
      )}

      <div className="space-y-2">
        <Label htmlFor="maxIterations">Max Iterations (Optional)</Label>
        <Input
          id="maxIterations"
          type="number"
          min="1"
          placeholder="No limit"
          value={state.maxIterations || ""}
          onChange={(e) => {
            const value = e.target.value;
            updateField("maxIterations", value ? parseInt(value) : undefined);
          }}
        />
        <p className="text-xs text-muted-foreground">
          Maximum number of iterations. Leave empty for no limit.
        </p>
      </div>

      <div className="flex items-center justify-between">
        <div className="space-y-0.5">
          <Label htmlFor="parallelExecution">Parallel Execution</Label>
          <p className="text-xs text-muted-foreground">
            Execute loop iterations in parallel instead of sequentially
          </p>
        </div>
        <Switch
          id="parallelExecution"
          checked={state.parallelExecution}
          onCheckedChange={(checked) =>
            updateField("parallelExecution", checked)
          }
        />
      </div>

      <div className="bg-muted p-4 rounded-md">
        <p className="text-sm font-medium mb-1">Loop Configuration</p>
        <p className="text-sm">
          Type: <span className="font-semibold">{state.loopType}</span>
        </p>
        {state.filter && (
          <p className="text-sm mt-1">
            Filter: <span className="font-mono text-xs">{state.filter}</span>
          </p>
        )}
        {state.maxIterations && (
          <p className="text-sm mt-1">
            Max iterations:{" "}
            <span className="font-semibold">{state.maxIterations}</span>
          </p>
        )}
        <p className="text-sm mt-1">
          Execution:{" "}
          <span className="font-semibold">
            {state.parallelExecution ? "Parallel" : "Sequential"}
          </span>
        </p>
      </div>
    </ConfigDrawerLayout>
  );
}
