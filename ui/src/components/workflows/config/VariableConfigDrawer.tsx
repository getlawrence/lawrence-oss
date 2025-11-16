import { useMemo, useState } from "react";

import type { VariableNodeData } from "../types/flow-types";

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
import { Textarea } from "@/components/ui/textarea";
import { useFormState } from "@/hooks/useDrawerForm";

interface VariableConfigDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  nodeData: VariableNodeData | null;
  onSave: (data: VariableNodeData) => void;
}

export function VariableConfigDrawer({
  open,
  onOpenChange,
  nodeData,
  onSave,
}: VariableConfigDrawerProps) {
  const initialState = useMemo(
    () => ({
      operation: (nodeData?.operation || "set") as "set" | "get" | "increment" | "append",
      variableName: nodeData?.variableName || "",
      value: nodeData?.value || "",
      label: nodeData?.label || "",
      description: nodeData?.description || "",
    }),
    [nodeData]
  );

  const { state, updateField } = useFormState(initialState);
  const [error, setError] = useState<string | null>(null);

  const handleSave = () => {
    setError(null);

    if (!state.variableName.trim()) {
      setError("Variable name is required");
      return;
    }

    if ((state.operation === "set" || state.operation === "append") && !state.value.trim()) {
      setError(`Value is required for ${state.operation} operation`);
      return;
    }

    const updatedData: VariableNodeData = {
      label: state.label || `Variable: ${state.operation}`,
      description: state.description || undefined,
      operation: state.operation,
      variableName: state.variableName.trim(),
      value:
        state.operation === "set" || state.operation === "append"
          ? state.value.trim()
          : undefined,
    };

    onSave(updatedData);
  };

  const requiresValue = state.operation === "set" || state.operation === "append";

  return (
    <ConfigDrawerLayout
      open={open}
      onOpenChange={onOpenChange}
      title="Configure Variable"
      description="Configure variable operations for workflow state management"
      error={error}
      onSave={handleSave}
    >
      <div className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="label">Label</Label>
          <Input
            id="label"
            placeholder="Variable Operation"
            value={state.label}
            onChange={(e) => updateField("label", e.target.value)}
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="description">Description (Optional)</Label>
          <Input
            id="description"
            placeholder="Brief description of this variable operation"
            value={state.description}
            onChange={(e) => updateField("description", e.target.value)}
          />
        </div>
      </div>

      <div className="space-y-2">
        <Label htmlFor="operation">Operation *</Label>
        <Select
          value={state.operation}
          onValueChange={(value: any) => updateField("operation", value)}
        >
          <SelectTrigger id="operation">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="set">Set - Assign a value</SelectItem>
            <SelectItem value="get">Get - Retrieve a value</SelectItem>
            <SelectItem value="increment">
              Increment - Increase by 1
            </SelectItem>
            <SelectItem value="append">
              Append - Add to existing value
            </SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-2">
        <Label htmlFor="variableName">Variable Name *</Label>
        <Input
          id="variableName"
          placeholder="myVariable"
          value={state.variableName}
          onChange={(e) => updateField("variableName", e.target.value)}
          className="font-mono"
        />
        <p className="text-xs text-muted-foreground">
          Variable name (alphanumeric and underscores only)
        </p>
      </div>

      {requiresValue && (
        <div className="space-y-2">
          <Label htmlFor="value">Value *</Label>
          <Textarea
            id="value"
            placeholder="Enter value..."
            value={state.value}
            onChange={(e) => updateField("value", e.target.value)}
            rows={3}
            className="font-mono text-sm"
          />
          <p className="text-xs text-muted-foreground">
            {state.operation === "set"
              ? "Value to assign to the variable"
              : "Value to append to the existing variable"}
          </p>
        </div>
      )}

      <div className="bg-muted p-4 rounded-md">
        <p className="text-sm font-medium mb-1">Operation Preview</p>
        <p className="text-sm font-mono">
          {state.operation === "set" &&
            `Set ${state.variableName || "variable"} = ${state.value || "value"}`}
          {state.operation === "get" && `Get ${state.variableName || "variable"}`}
          {state.operation === "increment" &&
            `Increment ${state.variableName || "variable"}++`}
          {state.operation === "append" &&
            `Append to ${state.variableName || "variable"}: ${state.value || "value"}`}
        </p>
      </div>
    </ConfigDrawerLayout>
  );
}
