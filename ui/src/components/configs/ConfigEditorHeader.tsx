import { ArrowLeft, Save, RefreshCw, Target, MoreVertical, History } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

interface ConfigEditorHeaderProps {
  isSaving: boolean;
  canSave: boolean;
  selectedGroupName?: string;
  onBack: () => void;
  onShowVersions: () => void;
  onShowTarget: () => void;
  onSave: () => void;
}

export function ConfigEditorHeader({
  isSaving,
  canSave,
  selectedGroupName,
  onBack,
  onShowVersions,
  onShowTarget,
  onSave,
}: ConfigEditorHeaderProps) {
  return (
    <div className="flex justify-between items-center w-full">
      <div className="flex items-center gap-3">
        <Button variant="ghost" size="icon" onClick={onBack}>
          <ArrowLeft className="h-4 w-4" />
        </Button>
        {selectedGroupName && (
          <Badge variant="outline" className="text-sm">
            <Target className="h-3 w-3 mr-1" />
            {selectedGroupName}
          </Badge>
        )}
      </div>
      <div className="flex items-center gap-2">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon">
              <MoreVertical className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={onShowTarget}>
              <Target className="h-4 w-4 mr-2" />
              Target Group
            </DropdownMenuItem>
            <DropdownMenuSeparator />
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
