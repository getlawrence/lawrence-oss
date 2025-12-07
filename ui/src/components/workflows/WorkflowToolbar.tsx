import {
  Plus,
  Undo2,
  Redo2,
  Save,
  Download,
  Play,
  Loader2,
} from "lucide-react";
import { Circle } from "lucide-react";

import { Button } from "@/components/ui/button";

interface WorkflowToolbarProps {
  onAddStep: () => void;
  onUndo?: () => void;
  onRedo?: () => void;
  onSave: () => void;
  onRun?: () => void;
  onExport?: () => void;
  hasUnsavedChanges?: boolean;
  isExecuting?: boolean;
  canUndo?: boolean;
  canRedo?: boolean;
  isNew?: boolean;
}

export function WorkflowToolbar({
  onAddStep,
  onUndo,
  onRedo,
  onSave,
  onRun,
  onExport,
  hasUnsavedChanges = false,
  isExecuting = false,
  canUndo = false,
  canRedo = false,
  isNew = false,
}: WorkflowToolbarProps) {
  return (
    <div className="flex items-center gap-2 px-2">
      {/* Add Step Button */}
      <Button
        type="button"
        variant="outline"
        size="sm"
        onClick={onAddStep}
        className="h-8"
      >
        <Plus className="h-4 w-4 mr-2" />
        Add Step
      </Button>

      {/* Undo/Redo Group */}
      <div className="flex items-center gap-1 border-l pl-2">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={onUndo}
          disabled={!canUndo}
          className="h-8 w-8 p-0"
          title="Undo (Cmd+Z)"
        >
          <Undo2 className="h-4 w-4" />
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={onRedo}
          disabled={!canRedo}
          className="h-8 w-8 p-0"
          title="Redo (Cmd+Shift+Z)"
        >
          <Redo2 className="h-4 w-4" />
        </Button>
      </div>

      {/* Save/Export Group */}
      <div className="flex items-center gap-1 border-l pl-2">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={onSave}
          className="h-8 relative"
          title="Save workflow (Cmd+S)"
        >
          {hasUnsavedChanges && (
            <Circle className="absolute -top-1 -right-1 h-2 w-2 fill-orange-500 text-orange-500" />
          )}
          <Save className="h-4 w-4 mr-2" />
          {isNew ? "Create" : "Save"}
        </Button>
        {onExport && (
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={onExport}
            className="h-8 w-8 p-0"
            title="Export workflow as code"
          >
            <Download className="h-4 w-4" />
          </Button>
        )}
      </div>

      {/* Run Button */}
      {onRun && !isNew && (
        <div className="flex items-center gap-1 border-l pl-2">
          <Button
            type="button"
            variant="default"
            size="sm"
            onClick={onRun}
            disabled={isExecuting}
            className="h-8"
            title="Run workflow (Cmd+Enter)"
          >
            {isExecuting ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <Play className="h-4 w-4 mr-2" />
            )}
            Run
          </Button>
        </div>
      )}
    </div>
  );
}
