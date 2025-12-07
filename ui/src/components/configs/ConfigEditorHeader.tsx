import {
  ArrowLeft,
  Save,
  RefreshCw,
  MoreVertical,
  History,
} from "lucide-react";

import { EditableConfigTitle } from "./EditableConfigTitle";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
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

interface ConfigEditorHeaderProps {
  isSaving: boolean;
  canSave: boolean;
  configName: string;
  selectedGroupId: string;
  groups: Group[];
  onBack: () => void;
  onShowVersions: () => void;
  onSave: () => void;
  onConfigNameChange: (name: string) => void;
  onGroupChange: (groupId: string) => void;
}

export function ConfigEditorHeader({
  isSaving,
  canSave,
  configName,
  selectedGroupId,
  groups,
  onBack,
  onShowVersions,
  onSave,
  onConfigNameChange,
  onGroupChange,
}: ConfigEditorHeaderProps) {
  return (
    <div className="flex justify-between items-center w-full">
      <div className="flex items-center gap-3 min-w-0 flex-1">
        <Button
          variant="ghost"
          size="icon"
          onClick={onBack}
          className="flex-shrink-0"
        >
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div className="min-w-0">
          <EditableConfigTitle
            value={configName}
            onChange={onConfigNameChange}
          />
        </div>
        <div className="w-[200px] flex-shrink-0">
          <Select value={selectedGroupId} onValueChange={onGroupChange}>
            <SelectTrigger className="h-9">
              <SelectValue placeholder="Select target group" />
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
      <div className="flex items-center gap-2">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon">
              <MoreVertical className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={onShowVersions}>
              <History className="h-4 w-4 mr-2" />
              Version History
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
        <Button onClick={onSave} disabled={isSaving || !canSave}>
          {isSaving ? (
            <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
          ) : (
            <Save className="h-4 w-4 mr-2" />
          )}
          Save
        </Button>
      </div>
    </div>
  );
}
