import { ArrowLeft, History, CheckCircle, Save, RefreshCw } from "lucide-react";

import { Button } from "@/components/ui/button";

interface ConfigEditorHeaderProps {
  mode: "create" | "edit";
  isValidating: boolean;
  isSaving: boolean;
  canSave: boolean;
  onBack: () => void;
  onShowVersions: () => void;
  onValidate: () => void;
  onSave: () => void;
}

export function ConfigEditorHeader({
  mode,
  isValidating,
  isSaving,
  canSave,
  onBack,
  onShowVersions,
  onValidate,
  onSave,
}: ConfigEditorHeaderProps) {
  return (
    <div className="flex justify-between items-center">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="sm" onClick={onBack}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to List
        </Button>
        <div>
          <h1 className="text-3xl font-bold">
            {mode === "create" ? "Create Configuration" : "Edit Configuration"}
          </h1>
          <p className="text-gray-600">
            {mode === "create"
              ? "Create a new OpenTelemetry collector configuration"
              : "Edit OpenTelemetry collector configuration"}
          </p>
        </div>
      </div>
      <div className="flex gap-2">
        <Button variant="outline" onClick={onShowVersions}>
          <History className="h-4 w-4 mr-2" />
          Version History
        </Button>
        <Button onClick={onValidate} disabled={isValidating}>
          {isValidating ? (
            <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
          ) : (
            <CheckCircle className="h-4 w-4 mr-2" />
          )}
          Validate
        </Button>
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
