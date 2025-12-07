import { FileText } from "lucide-react";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
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

interface Group {
  id: string;
  name: string;
}

interface ConfigTargetDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  mode: "create" | "edit";
  selectedGroupId: string;
  groups: Group[];
  currentVersion?: number;
  onGroupChange: (groupId: string) => void;
}

export function ConfigTargetDrawer({
  open,
  onOpenChange,
  mode,
  selectedGroupId,
  groups,
  currentVersion,
  onGroupChange,
}: ConfigTargetDrawerProps) {
  const selectedGroup = groups.find((g) => g.id === selectedGroupId);

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="p-6">
        <SheetHeader>
          <SheetTitle>Configuration Target</SheetTitle>
          <SheetDescription>
            {mode === "create"
              ? "Select the group to deploy this configuration to"
              : "View and change the target group for this configuration"}
          </SheetDescription>
        </SheetHeader>

        <div className="space-y-6 mt-6">
          {/* Group Selection */}
          <div>
            <label className="text-sm font-medium mb-2 block">
              Target Group
            </label>
            <Select value={selectedGroupId} onValueChange={onGroupChange}>
              <SelectTrigger>
                <SelectValue placeholder="Select a group" />
              </SelectTrigger>
              <SelectContent>
                {groups.map((group) => (
                  <SelectItem key={group.id} value={group.id}>
                    {group.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {selectedGroup && (
              <p className="text-sm text-muted-foreground mt-2">
                Configuration will be deployed to all agents in the{" "}
                <strong>{selectedGroup.name}</strong> group
              </p>
            )}
          </div>

          {/* Version Info for Edit Mode */}
          {mode === "edit" && currentVersion !== undefined && (
            <Alert>
              <FileText className="h-4 w-4" />
              <AlertDescription>
                Editing version <strong>{currentVersion}</strong>. Saving will
                create version <strong>{currentVersion + 1}</strong>.
              </AlertDescription>
            </Alert>
          )}

          {/* Warning if no group selected */}
          {!selectedGroupId && (
            <Alert variant="destructive">
              <AlertDescription>
                You must select a target group before saving the configuration.
              </AlertDescription>
            </Alert>
          )}

          {/* Action Buttons */}
          <div className="flex justify-end gap-2 pt-4">
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              Close
            </Button>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  );
}
