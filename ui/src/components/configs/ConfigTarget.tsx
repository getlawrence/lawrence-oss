import { FileText } from "lucide-react";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

interface Group {
  id: string;
  name: string;
}

interface ConfigTargetProps {
  mode: "create" | "edit";
  selectedGroupId: string;
  groups: Group[];
  currentVersion?: number;
  onGroupChange: (groupId: string) => void;
}

export function ConfigTarget({
  mode,
  selectedGroupId,
  groups,
  currentVersion,
  onGroupChange,
}: ConfigTargetProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Configuration Target</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid grid-cols-1 gap-4">
          <div>
            <label className="text-sm font-medium mb-2 block">Group</label>
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
          </div>
        </div>

        {mode === "edit" && currentVersion !== undefined && (
          <Alert>
            <FileText className="h-4 w-4" />
            <AlertDescription>
              Editing config version {currentVersion} - Saving will create
              version {currentVersion + 1}
            </AlertDescription>
          </Alert>
        )}
      </CardContent>
    </Card>
  );
}
