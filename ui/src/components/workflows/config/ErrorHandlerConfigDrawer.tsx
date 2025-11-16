import { Plus, Trash2 } from "lucide-react";
import { useMemo, useState } from "react";

import type { ErrorHandlerNodeData } from "../types/flow-types";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { useFormState, useArrayField } from "@/hooks/useDrawerForm";

import { ConfigDrawerLayout } from "./ConfigDrawerLayout";

interface ErrorHandlerConfigDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  nodeData: ErrorHandlerNodeData | null;
  onSave: (data: ErrorHandlerNodeData) => void;
}

export function ErrorHandlerConfigDrawer({
  open,
  onOpenChange,
  nodeData,
  onSave,
}: ErrorHandlerConfigDrawerProps) {
  const initialState = useMemo(
    () => ({
      catchAll: nodeData?.catchAll ?? true,
      retryCount: nodeData?.retryCount,
      retryDelay: nodeData?.retryDelay,
      label: nodeData?.label || "",
      description: nodeData?.description || "",
    }),
    [nodeData]
  );

  const { state, updateField } = useFormState(initialState);
  const {
    items: errorTypes,
    addItem: addErrorType,
    removeItem: removeErrorType,
  } = useArrayField<string>(nodeData?.errorTypes || []);
  const [newErrorType, setNewErrorType] = useState<string>("");
  const [error, setError] = useState<string | null>(null);

  const handleAddErrorType = () => {
    if (newErrorType.trim() && !errorTypes.includes(newErrorType.trim())) {
      addErrorType(newErrorType.trim());
      setNewErrorType("");
    }
  };

  const handleRemoveErrorType = (index: number) => {
    removeErrorType(index);
  };

  const handleSave = () => {
    setError(null);

    if (!state.catchAll && errorTypes.length === 0) {
      setError(
        "At least one error type is required when not catching all errors"
      );
      return;
    }

    if (state.retryCount !== undefined && state.retryCount < 0) {
      setError("Retry count cannot be negative");
      return;
    }

    if (state.retryDelay !== undefined && state.retryDelay < 0) {
      setError("Retry delay cannot be negative");
      return;
    }

    const updatedData: ErrorHandlerNodeData = {
      label: state.label || "Error Handler",
      description: state.description || undefined,
      catchAll: state.catchAll,
      errorTypes: state.catchAll
        ? undefined
        : errorTypes.length > 0
          ? errorTypes
          : undefined,
      retryCount:
        state.retryCount && state.retryCount > 0 ? state.retryCount : undefined,
      retryDelay:
        state.retryDelay && state.retryDelay > 0 ? state.retryDelay : undefined,
    };

    onSave(updatedData);
  };

  return (
    <ConfigDrawerLayout
      open={open}
      onOpenChange={onOpenChange}
      title="Configure Error Handler"
      description="Configure error handling and retry logic for workflow steps"
      error={error}
      onSave={handleSave}
    >
      <div className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="label">Label</Label>
          <Input
            id="label"
            placeholder="Error Handler"
            value={state.label}
            onChange={(e) => updateField("label", e.target.value)}
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="description">Description (Optional)</Label>
          <Input
            id="description"
            placeholder="Brief description of error handling"
            value={state.description}
            onChange={(e) => updateField("description", e.target.value)}
          />
        </div>
      </div>

      <div className="flex items-center justify-between">
        <div className="space-y-0.5">
          <Label htmlFor="catchAll">Catch All Errors</Label>
          <p className="text-xs text-muted-foreground">
            Handle all error types, or specify specific error types
          </p>
        </div>
        <Switch
          id="catchAll"
          checked={state.catchAll}
          onCheckedChange={(checked) => updateField("catchAll", checked)}
        />
      </div>

      {!state.catchAll && (
        <div className="space-y-2">
          <Label>Error Types *</Label>
          <div className="flex gap-2">
            <Input
              placeholder="Error type name (e.g., ValidationError)"
              value={newErrorType}
              onChange={(e) => setNewErrorType(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  e.preventDefault();
                  handleAddErrorType();
                }
              }}
            />
            <Button type="button" onClick={handleAddErrorType} size="icon">
              <Plus className="h-4 w-4" />
            </Button>
          </div>
          {errorTypes.length > 0 && (
            <div className="space-y-1">
              {errorTypes.map((errorType, index) => (
                <div
                  key={index}
                  className="flex items-center justify-between bg-muted px-3 py-2 rounded-md"
                >
                  <span className="text-sm font-mono">{errorType}</span>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    onClick={() => handleRemoveErrorType(index)}
                    className="h-6 w-6"
                  >
                    <Trash2 className="h-3 w-3" />
                  </Button>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label htmlFor="retryCount">Retry Count (Optional)</Label>
          <Input
            id="retryCount"
            type="number"
            min="0"
            placeholder="No retry"
            value={state.retryCount || ""}
            onChange={(e) => {
              const value = e.target.value;
              updateField("retryCount", value ? parseInt(value) : undefined);
            }}
          />
          <p className="text-xs text-muted-foreground">
            Number of retry attempts
          </p>
        </div>

        <div className="space-y-2">
          <Label htmlFor="retryDelay">
            Retry Delay (seconds, Optional)
          </Label>
          <Input
            id="retryDelay"
            type="number"
            min="0"
            placeholder="No delay"
            value={state.retryDelay || ""}
            onChange={(e) => {
              const value = e.target.value;
              updateField("retryDelay", value ? parseInt(value) : undefined);
            }}
          />
          <p className="text-xs text-muted-foreground">
            Delay between retries
          </p>
        </div>
      </div>

      <div className="bg-muted p-4 rounded-md">
        <p className="text-sm font-medium mb-1">
          Error Handler Configuration
        </p>
        <p className="text-sm">
          Mode:{" "}
          <span className="font-semibold">
            {state.catchAll ? "Catch All Errors" : "Specific Errors"}
          </span>
        </p>
        {!state.catchAll && errorTypes.length > 0 && (
          <p className="text-sm mt-1">
            Error types:{" "}
            <span className="font-mono text-xs">
              {errorTypes.join(", ")}
            </span>
          </p>
        )}
        {state.retryCount && state.retryCount > 0 && (
          <p className="text-sm mt-1">
            Retries: <span className="font-semibold">{state.retryCount}</span>
            {state.retryDelay && state.retryDelay > 0 && (
              <span> with {state.retryDelay}s delay</span>
            )}
          </p>
        )}
      </div>
    </ConfigDrawerLayout>
  );
}
