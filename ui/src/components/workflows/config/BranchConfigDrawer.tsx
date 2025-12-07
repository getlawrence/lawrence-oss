import { Plus, Trash2 } from "lucide-react";
import { useEffect, useMemo, useState } from "react";

import type { BranchNodeData } from "../types/flow-types";

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
import { Switch } from "@/components/ui/switch";
import { useFormState } from "@/hooks/useDrawerForm";

interface BranchConfigDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  nodeData: BranchNodeData | null;
  onSave: (data: BranchNodeData) => void;
}

type Branch = {
  name: string;
  condition: WorkflowCondition[];
  isDefault?: boolean;
};

export function BranchConfigDrawer({
  open,
  onOpenChange,
  nodeData,
  onSave,
}: BranchConfigDrawerProps) {
  const initialState = useMemo(
    () => ({
      label: nodeData?.label || "",
      description: nodeData?.description || "",
    }),
    [nodeData],
  );

  const initialBranches = useMemo<Branch[]>(
    () => nodeData?.branches || [],
    [nodeData],
  );

  const { state, updateField, resetState } = useFormState(initialState);
  const [branches, setBranches] = useState<Branch[]>(initialBranches);
  const [error, setError] = useState<string | null>(null);

  // Reset form state and branches when drawer opens with new nodeData
  // This is appropriate useEffect usage - resetting state when a prop changes
  useEffect(() => {
    if (open && nodeData) {
      resetState(initialState);
      setBranches(initialBranches);
      setError(null);
    }
  }, [open, nodeData, initialState, initialBranches, resetState]);

  const handleAddBranch = () => {
    setBranches([
      ...branches,
      {
        name: `Branch ${branches.length + 1}`,
        condition: [],
        isDefault: false,
      },
    ]);
  };

  const handleRemoveBranch = (index: number) => {
    setBranches(branches.filter((_, i) => i !== index));
  };

  const handleBranchNameChange = (index: number, name: string) => {
    const updated = [...branches];
    updated[index] = { ...updated[index], name };
    setBranches(updated);
  };

  const handleBranchDefaultChange = (index: number, isDefault: boolean) => {
    const updated = branches.map((branch, i) => ({
      ...branch,
      isDefault: i === index ? isDefault : false, // Only one default branch
    }));
    setBranches(updated);
  };

  const handleAddCondition = (branchIndex: number) => {
    const updated = [...branches];
    updated[branchIndex].condition = [
      ...updated[branchIndex].condition,
      {
        field: "",
        operator: "equals",
        value: "",
      },
    ];
    setBranches(updated);
  };

  const handleRemoveCondition = (
    branchIndex: number,
    conditionIndex: number,
  ) => {
    const updated = [...branches];
    updated[branchIndex].condition = updated[branchIndex].condition.filter(
      (_, i) => i !== conditionIndex,
    );
    setBranches(updated);
  };

  const handleConditionChange = (
    branchIndex: number,
    conditionIndex: number,
    field: keyof WorkflowCondition,
    value: string,
  ) => {
    const updated = [...branches];
    updated[branchIndex].condition[conditionIndex] = {
      ...updated[branchIndex].condition[conditionIndex],
      [field]: value,
    };
    setBranches(updated);
  };

  const handleSave = () => {
    setError(null);

    if (branches.length === 0) {
      setError("At least one branch is required");
      return;
    }

    // Validate branch names
    const invalidBranches = branches.filter((b) => !b.name.trim());
    if (invalidBranches.length > 0) {
      setError("All branches must have a name");
      return;
    }

    // Validate conditions (except default branch)
    for (const branch of branches) {
      if (!branch.isDefault) {
        const invalidConditions = branch.condition.filter(
          (c) => !c.field.trim() || !c.value.trim(),
        );
        if (invalidConditions.length > 0) {
          setError(`Branch "${branch.name}" has invalid conditions`);
          return;
        }
        if (branch.condition.length === 0) {
          setError(
            `Branch "${branch.name}" must have at least one condition or be marked as default`,
          );
          return;
        }
      }
    }

    const updatedData: BranchNodeData = {
      label: state.label || "Branch",
      description: state.description || undefined,
      branches,
    };

    onSave(updatedData);
  };

  return (
    <ConfigDrawerLayout
      open={open}
      onOpenChange={onOpenChange}
      title="Configure Branch"
      description="Configure multi-way branching with conditions for workflow paths"
      error={error}
      onSave={handleSave}
      maxWidth="3xl"
    >
      <div className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="label">Label</Label>
          <Input
            id="label"
            placeholder="Branch"
            value={state.label}
            onChange={(e) => updateField("label", e.target.value)}
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="description">Description (Optional)</Label>
          <Input
            id="description"
            placeholder="Brief description of branching logic"
            value={state.description}
            onChange={(e) => updateField("description", e.target.value)}
          />
        </div>
      </div>

      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <Label>Branches</Label>
          <Button
            type="button"
            size="sm"
            variant="outline"
            onClick={handleAddBranch}
          >
            <Plus className="h-3 w-3 mr-1" />
            Add Branch
          </Button>
        </div>

        {branches.length === 0 ? (
          <div className="border rounded-md p-4 text-center text-muted-foreground text-sm">
            No branches defined. Click "Add Branch" to add one.
          </div>
        ) : (
          <div className="space-y-4 max-h-[50vh] overflow-y-auto pr-2">
            {branches.map((branch, branchIndex) => (
              <div
                key={branchIndex}
                className="border rounded-md p-4 space-y-3 bg-muted/30"
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2 flex-1">
                    <Input
                      value={branch.name}
                      onChange={(e) =>
                        handleBranchNameChange(branchIndex, e.target.value)
                      }
                      placeholder="Branch name"
                      className="font-semibold"
                    />
                    {branch.isDefault && (
                      <span className="text-xs bg-primary text-primary-foreground px-2 py-1 rounded">
                        Default
                      </span>
                    )}
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="flex items-center gap-2">
                      <Label
                        htmlFor={`default-${branchIndex}`}
                        className="text-xs"
                      >
                        Default
                      </Label>
                      <Switch
                        id={`default-${branchIndex}`}
                        checked={branch.isDefault || false}
                        onCheckedChange={(checked) =>
                          handleBranchDefaultChange(branchIndex, checked)
                        }
                      />
                    </div>
                    <Button
                      type="button"
                      size="sm"
                      variant="ghost"
                      onClick={() => handleRemoveBranch(branchIndex)}
                      className="text-destructive"
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </div>

                {branch.isDefault ? (
                  <p className="text-xs text-muted-foreground">
                    This branch will be taken if no other branch conditions
                    match.
                  </p>
                ) : (
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <Label className="text-xs">Conditions</Label>
                      <Button
                        type="button"
                        size="sm"
                        variant="outline"
                        onClick={() => handleAddCondition(branchIndex)}
                        className="h-7 text-xs"
                      >
                        <Plus className="h-3 w-3 mr-1" />
                        Add Condition
                      </Button>
                    </div>

                    {branch.condition.length === 0 ? (
                      <div className="border-l-2 border-muted pl-3 py-2 text-xs text-muted-foreground">
                        No conditions. Add conditions or mark as default branch.
                      </div>
                    ) : (
                      <div className="space-y-2">
                        {branch.condition.map((condition, conditionIndex) => (
                          <div
                            key={conditionIndex}
                            className="border-l-2 border-primary/30 pl-3 py-2 space-y-2 bg-background rounded"
                          >
                            <div className="flex items-center justify-between">
                              <span className="text-xs font-medium text-muted-foreground">
                                Condition {conditionIndex + 1}
                              </span>
                              <Button
                                type="button"
                                size="sm"
                                variant="ghost"
                                onClick={() =>
                                  handleRemoveCondition(
                                    branchIndex,
                                    conditionIndex,
                                  )
                                }
                                className="h-6 w-6 p-0 text-destructive"
                              >
                                <Trash2 className="h-3 w-3" />
                              </Button>
                            </div>

                            <div className="grid grid-cols-3 gap-2">
                              <div className="space-y-1">
                                <Label className="text-xs">Field</Label>
                                <Input
                                  value={condition.field}
                                  onChange={(e) =>
                                    handleConditionChange(
                                      branchIndex,
                                      conditionIndex,
                                      "field",
                                      e.target.value,
                                    )
                                  }
                                  placeholder="field_name"
                                  className="h-7 text-xs"
                                />
                              </div>

                              <div className="space-y-1">
                                <Label className="text-xs">Operator</Label>
                                <Select
                                  value={condition.operator}
                                  onValueChange={(value) =>
                                    handleConditionChange(
                                      branchIndex,
                                      conditionIndex,
                                      "operator",
                                      value,
                                    )
                                  }
                                >
                                  <SelectTrigger className="h-7 text-xs">
                                    <SelectValue />
                                  </SelectTrigger>
                                  <SelectContent>
                                    <SelectItem value="equals">
                                      Equals
                                    </SelectItem>
                                    <SelectItem value="contains">
                                      Contains
                                    </SelectItem>
                                    <SelectItem value="matches">
                                      Matches
                                    </SelectItem>
                                  </SelectContent>
                                </Select>
                              </div>

                              <div className="space-y-1">
                                <Label className="text-xs">Value</Label>
                                <Input
                                  value={condition.value}
                                  onChange={(e) =>
                                    handleConditionChange(
                                      branchIndex,
                                      conditionIndex,
                                      "value",
                                      e.target.value,
                                    )
                                  }
                                  placeholder="value"
                                  className="h-7 text-xs"
                                />
                              </div>
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </ConfigDrawerLayout>
  );
}
