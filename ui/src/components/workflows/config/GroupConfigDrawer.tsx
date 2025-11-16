import { useMemo, useState } from "react";

import type { GroupNodeData } from "../types/flow-types";

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
import { useFormState } from "@/hooks/useDrawerForm";

interface GroupConfigDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  nodeData: GroupNodeData | null;
  onSave: (data: GroupNodeData) => void;
}

export function GroupConfigDrawer({
  open,
  onOpenChange,
  nodeData,
  onSave,
}: GroupConfigDrawerProps) {
  const initialState = useMemo(
    () => ({
      label: nodeData?.label || "",
      description: nodeData?.description || "",
      collapsed: nodeData?.collapsed ?? false,
      color: nodeData?.color || "slate",
    }),
    [nodeData]
  );

  const { state, updateField } = useFormState(initialState);
  const [error, setError] = useState<string | null>(null);

  const colorOptions = [
    { value: "slate", label: "Slate" },
    { value: "blue", label: "Blue" },
    { value: "green", label: "Green" },
    { value: "purple", label: "Purple" },
    { value: "amber", label: "Amber" },
    { value: "red", label: "Red" },
    { value: "indigo", label: "Indigo" },
  ];

  const handleSave = () => {
    setError(null);

    if (!state.label.trim()) {
      setError("Label is required");
      return;
    }

    const updatedData: GroupNodeData = {
      label: state.label.trim(),
      description: state.description || undefined,
      collapsed: state.collapsed,
      color: state.color,
    };

    onSave(updatedData);
  };

  return (
    <ConfigDrawerLayout
      open={open}
      onOpenChange={onOpenChange}
      title="Configure Group"
      description="Configure visual grouping settings for workflow nodes"
      error={error}
      onSave={handleSave}
    >
      <div className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="label">Label *</Label>
          <Input
            id="label"
            placeholder="Group Name"
            value={state.label}
            onChange={(e) => updateField("label", e.target.value)}
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="description">Description (Optional)</Label>
          <Input
            id="description"
            placeholder="Brief description of this group"
            value={state.description}
            onChange={(e) => updateField("description", e.target.value)}
          />
        </div>
      </div>

      <div className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="color">Color</Label>
          <Select
            value={state.color}
            onValueChange={(value) => updateField("color", value)}
          >
            <SelectTrigger id="color">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {colorOptions.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="flex items-center justify-between">
          <div className="space-y-0.5">
            <Label htmlFor="collapsed">Collapsed by Default</Label>
            <p className="text-xs text-muted-foreground">
              Start with the group collapsed in the workflow view
            </p>
          </div>
          <Switch
            id="collapsed"
            checked={state.collapsed}
            onCheckedChange={(checked) => updateField("collapsed", checked)}
          />
        </div>
      </div>

      <div className="bg-muted p-4 rounded-md">
        <p className="text-sm font-medium mb-1">Group Preview</p>
        <p className="text-sm">
          <span className="font-semibold">{state.label || "Group Name"}</span>
          {state.description && (
            <span className="text-muted-foreground"> - {state.description}</span>
          )}
        </p>
        <p className="text-xs text-muted-foreground mt-2">
          Color:{" "}
          {colorOptions.find((c) => c.value === state.color)?.label || state.color}
        </p>
      </div>
    </ConfigDrawerLayout>
  );
}
