import { Plus, Trash2, Save } from "lucide-react";
import { useEffect, useMemo, useState } from "react";

import type { ConditionNodeData } from "../types/flow-types";

import type { WorkflowCondition } from "@/api/workflows";
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
import { useArrayField, useFormState } from "@/hooks/useDrawerForm";

interface ConditionConfigDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  nodeData: ConditionNodeData | null;
  onSave: (data: ConditionNodeData) => void;
}

export function ConditionConfigDrawer({
  open,
  onOpenChange,
  nodeData,
  onSave,
}: ConditionConfigDrawerProps) {
  const initialState = useMemo(
    () => ({
      label: nodeData?.label || "",
    }),
    [nodeData],
  );

  const initialConditions = useMemo<WorkflowCondition[]>(
    () => nodeData?.conditions || [],
    [nodeData],
  );

  const { state, updateField, resetState } = useFormState(initialState);
  const { items: conditions, setItems: setConditions } =
    useArrayField<WorkflowCondition>(initialConditions);
  const [error, setError] = useState<string | null>(null);

  // Reset form state and conditions when drawer opens with new nodeData
  // This is appropriate useEffect usage - resetting state when a prop changes
  useEffect(() => {
    if (open && nodeData) {
      resetState(initialState);
      setConditions(initialConditions);
      setError(null);
    }
  }, [
    open,
    nodeData,
    initialState,
    initialConditions,
    resetState,
    setConditions,
  ]);

  const handleConditionChange = (
    index: number,
    field: keyof WorkflowCondition,
    value: string,
  ) => {
    const updated = [...conditions];
    updated[index] = {
      ...updated[index],
      [field]: value,
    };
    setConditions(updated);
  };

  const handleAddCondition = () => {
    setConditions([
      ...conditions,
      {
        field: "",
        operator: "equals",
        value: "",
      },
    ]);
  };

  const handleRemoveCondition = (index: number) => {
    setConditions(conditions.filter((_, i) => i !== index));
  };

  const handleSave = () => {
    if (!nodeData) return;

    // Validate conditions
    const invalidConditions = conditions.filter(
      (c) => !c.field.trim() || !c.value.trim(),
    );
    if (invalidConditions.length > 0) {
      setError("All conditions must have a field and value");
      return;
    }

    const updatedData: ConditionNodeData = {
      ...nodeData,
      label: state.label || "Condition",
      conditions,
    };

    onSave(updatedData);
    onOpenChange(false);
  };

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="w-full sm:max-w-xl overflow-y-auto p-6">
        <SheetHeader>
          <SheetTitle>Configure Condition</SheetTitle>
          <SheetDescription>
            Define conditions that must be met for the trigger to execute
          </SheetDescription>
        </SheetHeader>

        <div className="mt-4 space-y-4">
          {/* Label */}
          <div className="space-y-2">
            <Label className="text-sm">Label</Label>
            <Input
              value={state.label}
              onChange={(e) => updateField("label", e.target.value)}
              placeholder="Condition label"
              className="h-8 text-sm"
            />
          </div>

          {/* Conditions List */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <Label className="text-sm">Conditions</Label>
              <Button
                type="button"
                size="sm"
                variant="outline"
                onClick={handleAddCondition}
              >
                <Plus className="h-3 w-3 mr-1" />
                Add Condition
              </Button>
            </div>

            {conditions.length === 0 ? (
              <div className="border rounded-md p-4 text-center text-muted-foreground text-sm">
                No conditions defined. Click "Add Condition" to add one.
              </div>
            ) : (
              <div className="space-y-3 max-h-[40vh] overflow-y-auto pr-2">
                {conditions.map((condition, index) => (
                  <div
                    key={index}
                    className="border-l-2 border-muted pl-4 space-y-2 py-2"
                  >
                    <div className="flex items-center justify-between mb-2">
                      <span className="text-xs font-medium text-muted-foreground">
                        Condition {index + 1}
                      </span>
                      <Button
                        type="button"
                        size="sm"
                        variant="ghost"
                        onClick={() => handleRemoveCondition(index)}
                        className="h-6 w-6 p-0 text-destructive"
                      >
                        <Trash2 className="h-3 w-3" />
                      </Button>
                    </div>

                    <div className="grid grid-cols-3 gap-2">
                      <div className="space-y-1.5">
                        <Label className="text-xs">Field</Label>
                        <Input
                          value={condition.field}
                          onChange={(e) =>
                            handleConditionChange(
                              index,
                              "field",
                              e.target.value,
                            )
                          }
                          placeholder="field_name"
                          className="h-8 text-xs"
                        />
                      </div>

                      <div className="space-y-1.5">
                        <Label className="text-xs">Operator</Label>
                        <Select
                          value={condition.operator}
                          onValueChange={(value) =>
                            handleConditionChange(
                              index,
                              "operator",
                              value as WorkflowCondition["operator"],
                            )
                          }
                        >
                          <SelectTrigger className="h-8 text-xs">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="equals">Equals</SelectItem>
                            <SelectItem value="contains">Contains</SelectItem>
                            <SelectItem value="matches">Matches</SelectItem>
                          </SelectContent>
                        </Select>
                      </div>

                      <div className="space-y-1.5">
                        <Label className="text-xs">Value</Label>
                        <Input
                          value={condition.value}
                          onChange={(e) =>
                            handleConditionChange(
                              index,
                              "value",
                              e.target.value,
                            )
                          }
                          placeholder="value"
                          className="h-8 text-xs"
                        />
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

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
