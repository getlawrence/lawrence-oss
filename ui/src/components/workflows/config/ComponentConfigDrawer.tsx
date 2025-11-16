import yaml from "js-yaml";
import { Loader2, Save } from "lucide-react";
import { useState, useEffect } from "react";
import useSWR from "swr";

import {
  getComponentSchemas,
  getComponentSchema,
  type ComponentType,
} from "@/api/schemas";
import { type WorkflowAction } from "@/api/workflows";
import { SchemaForm } from "@/components/configs/SchemaForm";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";

interface ComponentConfigDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  action: WorkflowAction | null;
  onSave: (action: WorkflowAction) => void;
  targetConfig?: string; // Optional: current config YAML for auto-detection
}

export function ComponentConfigDrawer({
  open,
  onOpenChange,
  action,
  onSave,
  targetConfig,
}: ComponentConfigDrawerProps) {
  const [componentType, setComponentType] = useState<ComponentType | "">("");
  const [componentName, setComponentName] = useState<string>("");
  const [autoDetect, setAutoDetect] = useState(false);
  const [formValue, setFormValue] = useState<unknown>(null);
  const [yamlPreview, setYamlPreview] = useState<string>("");
  const [error, setError] = useState<string | null>(null);

  // Fetch available components - fetch all when drawer opens, then filter by type
  const { data: allComponentsData } = useSWR(
    open ? "all-components" : null,
    () => getComponentSchemas(),
  );

  // Filter components by selected type
  const filteredComponents = componentType
    ? (allComponentsData?.components || []).filter(
        (c) => c.type === componentType,
      )
    : [];

  // Fetch schema for selected component
  const { data: schema, isLoading: schemaLoading } = useSWR(
    componentType && componentName
      ? `schema-${componentType}-${componentName}`
      : null,
    () =>
      componentType && componentName
        ? getComponentSchema(componentType, componentName)
        : null,
  );

  // Auto-detect component from target config
  useEffect(() => {
    if (autoDetect && targetConfig && !componentType && !componentName) {
      try {
        const parsed = yaml.load(targetConfig) as Record<string, unknown>;
        // Try to detect a component type and name from the config
        // This is a simple heuristic - could be improved
        if (parsed.processors) {
          const processors = parsed.processors as Record<string, unknown>;
          const firstProcessor = Object.keys(processors)[0];
          if (firstProcessor) {
            const processorConfig = processors[firstProcessor] as Record<
              string,
              unknown
            >;
            const processorType = Object.keys(processorConfig)[0];
            if (processorType) {
              setComponentType("processor");
              setComponentName(processorType);
              setFormValue(processorConfig[processorType]);
            }
          }
        }
      } catch (err) {
        console.error("Failed to auto-detect component:", err);
      }
    }
  }, [autoDetect, targetConfig, componentType, componentName]);

  // Update YAML preview when form value changes
  useEffect(() => {
    if (formValue && componentType && componentName) {
      try {
        const config = {
          [componentType]: {
            [componentName]: formValue,
          },
        };
        const yamlString = yaml.dump(config, { indent: 2 });
        setYamlPreview(yamlString);
      } catch {
        setError("Failed to generate YAML preview");
      }
    }
  }, [formValue, componentType, componentName]);

  // Reset form when drawer opens/closes
  useEffect(() => {
    if (!open) {
      setComponentType("");
      setComponentName("");
      setFormValue(null);
      setYamlPreview("");
      setError(null);
      setAutoDetect(false);
    } else if (action?.config_update?.template) {
      // Try to parse existing template
      try {
        const parsed = yaml.load(action.config_update.template) as Record<
          string,
          unknown
        >;
        // Extract component type and name from parsed YAML
        for (const [type, components] of Object.entries(parsed)) {
          if (
            [
              "receivers",
              "processors",
              "exporters",
              "connectors",
              "extensions",
            ].includes(type)
          ) {
            const comps = components as Record<string, unknown>;
            const compName = Object.keys(comps)[0];
            if (compName) {
              setComponentType(type.replace("s", "") as ComponentType);
              setComponentName(compName);
              setFormValue(comps[compName]);
              break;
            }
          }
        }
      } catch (err) {
        // If parsing fails, start fresh
        console.debug("Failed to parse existing template:", err);
      }
    }
  }, [open, action]);

  const handleSave = () => {
    if (!action || !componentType || !componentName || !formValue) {
      setError("Please select a component and configure it");
      return;
    }

    try {
      // Generate YAML config
      const config = {
        [componentType === "processor" ? "processors" : `${componentType}s`]: {
          [componentName]: formValue,
        },
      };
      const yamlString = yaml.dump(config, { indent: 2 });

      // Update action
      const updatedAction: WorkflowAction = {
        ...action,
        config_update: {
          ...action.config_update,
          operation: action.config_update?.operation || "merge",
          template: yamlString,
        },
      };

      onSave(updatedAction);
      onOpenChange(false);
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Failed to save configuration",
      );
    }
  };

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="w-full min-w-[70vw] sm:max-w-2xl overflow-y-auto p-6">
        <SheetHeader>
          <SheetTitle>Configure Component</SheetTitle>
          <SheetDescription>
            Select and configure an OpenTelemetry collector component for this
            action
          </SheetDescription>
        </SheetHeader>

        <div className="mt-4 space-y-4">
          {/* Component Selection */}
          <div className="space-y-3">
            <div className="flex items-center gap-2 pb-2">
              <input
                type="checkbox"
                id="auto-detect"
                checked={autoDetect}
                onChange={(e) => setAutoDetect(e.target.checked)}
                className="h-4 w-4"
              />
              <Label htmlFor="auto-detect" className="cursor-pointer text-sm">
                Auto-detect from target config
              </Label>
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div>
                <Label className="text-sm">Component Type</Label>
                <Select
                  value={componentType}
                  onValueChange={(value) => {
                    setComponentType(value as ComponentType);
                    setComponentName("");
                    setFormValue(null);
                  }}
                  disabled={autoDetect}
                >
                  <SelectTrigger className="h-9">
                    <SelectValue placeholder="Select type" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="receiver">Receiver</SelectItem>
                    <SelectItem value="processor">Processor</SelectItem>
                    <SelectItem value="exporter">Exporter</SelectItem>
                    <SelectItem value="connector">Connector</SelectItem>
                    <SelectItem value="extension">Extension</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div>
                <Label className="text-sm">Component Name</Label>
                <Select
                  value={componentName}
                  onValueChange={(value) => {
                    setComponentName(value);
                    setFormValue(null);
                  }}
                  disabled={!componentType || autoDetect}
                >
                  <SelectTrigger className="h-9">
                    <SelectValue placeholder="Select component" />
                  </SelectTrigger>
                  <SelectContent>
                    {filteredComponents.map((comp) => (
                      <SelectItem key={comp.name} value={comp.name}>
                        {comp.name}
                        {comp.description && ` - ${comp.description}`}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
          </div>

          {/* Error Alert */}
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {/* Schema Form and Preview */}
          {schema && componentType && componentName ? (
            <Tabs defaultValue="form" className="w-full">
              <TabsList className="h-9">
                <TabsTrigger value="form" className="text-sm">
                  Form
                </TabsTrigger>
                <TabsTrigger value="preview" className="text-sm">
                  YAML Preview
                </TabsTrigger>
              </TabsList>

              <TabsContent value="form" className="space-y-3 mt-3">
                {schemaLoading ? (
                  <div className="flex items-center justify-center py-6">
                    <Loader2 className="h-5 w-5 animate-spin" />
                  </div>
                ) : (
                  <div className="max-h-[60vh] overflow-y-auto pr-2">
                    <SchemaForm
                      schema={schema}
                      value={formValue}
                      onChange={setFormValue}
                      path={[componentName]}
                    />
                  </div>
                )}
              </TabsContent>

              <TabsContent value="preview" className="space-y-2 mt-3">
                <Label className="text-sm">Generated YAML</Label>
                <Textarea
                  value={yamlPreview}
                  readOnly
                  className="font-mono text-xs h-64"
                  placeholder="YAML will be generated here..."
                />
              </TabsContent>
            </Tabs>
          ) : (
            <div className="border rounded-md p-6 text-center text-muted-foreground">
              <p className="text-sm">
                {!componentType
                  ? "Select a component type and name above to configure"
                  : !componentName
                    ? "Select a component name above to configure"
                    : "Loading schema..."}
              </p>
            </div>
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
            <Button
              size="sm"
              onClick={handleSave}
              disabled={!componentType || !componentName || !formValue}
            >
              <Save className="h-3 w-3 mr-2" />
              Save Configuration
            </Button>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  );
}
