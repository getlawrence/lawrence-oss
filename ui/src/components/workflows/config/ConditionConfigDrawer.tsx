import { Plus, Trash2 } from "lucide-react";
import { useMemo, useState } from "react";

import type { ConditionNodeData } from "../types/flow-types";

import { ConfigDrawerLayout } from "./ConfigDrawerLayout";

import type { WorkflowCondition } from "@/api/workflows";
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
import { useArrayField } from "@/hooks/useDrawerForm";

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
  // Derive initial conditions from nodeData
  const initialConditions = useMemo(
    () => nodeData?.conditions || [],
    [nodeData]
  );

  // Use array management hook
  const { items: conditions, addItem, removeItem, updateItem } = useArrayField<WorkflowCondition>(initialConditions);
  
  const [label, setLabel] = useState(nodeData?.label || "");
  const [error, setError] = useState<string | null>(null);

  const handleConditionChange = (
    index: number,
    field: keyof WorkflowCondition,
    value: string,
  ) => {
    const updated = {
      ...conditions[index],
      [field]: value,
    };
    updateItem(index, updated);
  };

  const handleAddCondition = () => {
    addItem({
      field: "",
      operator: "equals",
      value: "",
    });
  };

  const handleRemoveCondition = (index: number) => {
    removeItem(index);
  };

  const handleSave = () => {
    if (!nodeData) return;

    setError(null);

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
      label: label || "Condition",
      conditions,
    };

    onSave(updatedData);
    onOpenChange(false);
  };

  return (
    <ConfigDrawerLayout
      open={open}
      onOpenChange={onOpenChange}
      title="Configure Condition"
      description="Define conditions that must be met for the workflow to execute"
      error={error}
      onSave={handleSave}
      maxWidth="xl"
    >
      {/* Label */}
      <div className="space-y-2">
        <Label className="text-sm">Label</Label>
        <Input
          value={label}
          onChange={(e) => setLabel(e.target.value)}
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
    </ConfigDrawerLayout>
  );
}
