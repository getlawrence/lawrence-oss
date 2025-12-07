import yaml from "js-yaml";
import { Save, Info, Loader2 } from "lucide-react";
import { useState, useEffect } from "react";
import useSWR from "swr";

import { VariableAutocomplete } from "../../shared/VariableAutocomplete";

import { getAgents, type Agent } from "@/api/agents";
import { getGroups, type Group } from "@/api/groups";
import {
  getComponentSchemas,
  getComponentSchema,
  type ComponentType,
} from "@/api/schemas";
import { type WorkflowAction } from "@/api/workflows";
import { SchemaForm } from "@/components/configs/SchemaForm";
import { Alert, AlertDescription } from "@/components/ui/alert";
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";


interface ActionConfigFormProps {
  action: WorkflowAction | null;
  onSave: (action: WorkflowAction) => void;
  onCancel?: () => void;
  availableVariables?: string[];
}

export function ActionConfigForm({
  action,
  onSave,
  onCancel,
  availableVariables = [],
}: ActionConfigFormProps) {
  const [targetType, setTargetType] = useState<"agent" | "group">("group");
  const [targetId, setTargetId] = useState<string>("");
  const [useCustomTarget, setUseCustomTarget] = useState<boolean>(false);
  const [operation, setOperation] = useState<"merge" | "replace" | "patch">(
    "patch",
  );
  const [yamlPath, setYamlPath] = useState<string>("");
  const [value, setValue] = useState<string>("");
  const [template, setTemplate] = useState<string>("");
  const [error, setError] = useState<string | null>(null);

  // UI Form state for merge/replace operations
  const [configMode, setConfigMode] = useState<"yaml" | "form">("yaml");
  const [componentType, setComponentType] = useState<ComponentType | "">("");
  const [componentName, setComponentName] = useState<string>("");
  const [formValue, setFormValue] = useState<unknown>(null);
  const [yamlPreview, setYamlPreview] = useState<string>("");

  // Fetch available components
  const { data: allComponentsData } = useSWR(
    operation === "merge" || operation === "replace" ? "all-components" : null,
    () => getComponentSchemas(),
  );

  // Fetch agents and groups
  const { data: agentsData } = useSWR(
    targetType === "agent" ? "agents" : null,
    () => getAgents(),
  );

  const { data: groupsData } = useSWR(
    targetType === "group" ? "groups" : null,
    () => getGroups(),
  );

  // Convert agents to array and extract relevant info
  const agents = agentsData?.agents
    ? Object.values(agentsData.agents).map((agent: Agent) => ({
        id: agent.id,
        name: agent.name || agent.id,
        display: agent.name ? `${agent.name} (${agent.id})` : agent.id,
      }))
    : [];

  // Extract groups info
  const groups = groupsData?.groups
    ? groupsData.groups.map((group: Group) => ({
        id: group.id,
        name: group.name,
        display: group.name,
      }))
    : [];

  // Filter components by selected type
  const filteredComponents = componentType
    ? (allComponentsData?.components || []).filter(
        (c) => c.type === componentType,
      )
    : [];

  // Fetch schema for selected component
  const { data: schema, isLoading: schemaLoading } = useSWR(
    componentType && componentName && configMode === "form"
      ? `schema-${componentType}-${componentName}`
      : null,
    () =>
      componentType && componentName
        ? getComponentSchema(componentType, componentName)
        : null,
  );

  // Update YAML preview when form value changes
  useEffect(() => {
    if (formValue && componentType && componentName && configMode === "form") {
      try {
        const pluralType =
          componentType === "processor" ? "processors" : `${componentType}s`;
        const config = {
          [pluralType]: {
            [componentName]: formValue,
          },
        };
        const yamlString = yaml.dump(config, { indent: 2 });
        setYamlPreview(yamlString);
        // Auto-update template when form changes
        setTemplate(yamlString);
      } catch (err) {
        console.error("Failed to generate YAML preview:", err);
      }
    }
  }, [formValue, componentType, componentName, configMode]);

  // When switching to form mode, try to parse existing YAML template
  useEffect(() => {
    if (configMode === "form" && template && !componentType && !componentName) {
      try {
        const parsed = yaml.load(template) as Record<string, unknown>;
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
              const singularType = type.replace("s", "") as ComponentType;
              setComponentType(singularType);
              setComponentName(compName);
              setFormValue(comps[compName]);
              break;
            }
          }
        }
      } catch (err) {
        // If parsing fails, user needs to select component manually
        console.debug("Failed to parse template for form mode:", err);
      }
    }
  }, [configMode, template, componentType, componentName]);

  // Reset form state when operation changes to patch
  useEffect(() => {
    if (operation === "patch") {
      setConfigMode("yaml");
      setComponentType("");
      setComponentName("");
      setFormValue(null);
      setYamlPreview("");
    }
  }, [operation]);

  // Initialize form from action
  useEffect(() => {
    if (action) {
      setTargetType(action.target_type);
      const existingTargetId = action.target_id || "";
      setTargetId(existingTargetId);
      setUseCustomTarget(false);

      setOperation(action.config_update?.operation || "patch");
      setYamlPath(action.config_update?.yaml_path || "");
      setValue(
        action.config_update?.value !== undefined
          ? String(action.config_update.value)
          : "",
      );
      const existingTemplate = action.config_update?.template || "";
      setTemplate(existingTemplate);

      // Try to parse existing template to populate form if it's a component config
      if (
        existingTemplate &&
        (action.config_update?.operation === "merge" ||
          action.config_update?.operation === "replace")
      ) {
        try {
          const parsed = yaml.load(existingTemplate) as Record<string, unknown>;
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
                const singularType = type.replace("s", "") as ComponentType;
                setComponentType(singularType);
                setComponentName(compName);
                setFormValue(comps[compName]);
                setConfigMode("form");
                break;
              }
            }
          }
        } catch (err) {
          // If parsing fails, keep YAML mode
          console.debug("Failed to parse existing template:", err);
        }
      }

      setError(null);
    } else {
      // Reset form when action is null
      setConfigMode("yaml");
      setComponentType("");
      setComponentName("");
      setFormValue(null);
      setYamlPreview("");
      setUseCustomTarget(false);
    }
  }, [action]);

  // Check if current targetId matches any available option
  useEffect(() => {
    if (!targetId) {
      setUseCustomTarget(false);
      return;
    }

    // Wait for data to load before checking
    if (targetType === "agent") {
      if (agents.length === 0 && agentsData) {
        // Data loaded but no agents found, so it's custom
        setUseCustomTarget(true);
        return;
      }
      if (agents.length > 0) {
        const matches = agents.some((agent) => agent.id === targetId);
        setUseCustomTarget(!matches);
      }
    } else {
      if (groups.length === 0 && groupsData) {
        // Data loaded but no groups found, so it's custom
        setUseCustomTarget(true);
        return;
      }
      if (groups.length > 0) {
        // Groups can be referenced by name or ID
        const matches = groups.some(
          (group) => group.id === targetId || group.name === targetId,
        );
        setUseCustomTarget(!matches);
      }
    }
  }, [targetId, targetType, agents, groups, agentsData, groupsData]);

  // Get the display value for the Select when targetId is set
  const getSelectValue = () => {
    if (!targetId) return "";
    if (targetType === "agent") {
      const agent = agents.find((a) => a.id === targetId);
      return agent ? agent.id : "";
    } else {
      // For groups, check both name and ID
      const group = groups.find(
        (g) => g.id === targetId || g.name === targetId,
      );
      return group ? group.name : "";
    }
  };

  // Reset custom target mode when target type changes
  useEffect(() => {
    setUseCustomTarget(false);
  }, [targetType]);

  const handleSave = () => {
    if (!action) return;

    // Validate based on operation
    if (operation === "patch") {
      if (!yamlPath.trim()) {
        setError("YAML path is required for patch operation");
        return;
      }
      if (!value.trim()) {
        setError("Value is required for patch operation");
        return;
      }
    } else {
      // For merge/replace, validate based on mode
      if (configMode === "form") {
        if (!componentType || !componentName || !formValue) {
          setError("Please select a component and configure it");
          return;
        }
        // Generate YAML from form value
        try {
          const pluralType =
            componentType === "processor" ? "processors" : `${componentType}s`;
          const config = {
            [pluralType]: {
              [componentName]: formValue,
            },
          };
          const yamlString = yaml.dump(config, { indent: 2 });
          setTemplate(yamlString);
        } catch (err) {
          setError(
            err instanceof Error
              ? err.message
              : "Failed to generate YAML from form",
          );
          return;
        }
      } else {
        if (!template.trim()) {
          setError("Template is required for merge/replace operation");
          return;
        }
      }
    }

    if (!targetId.trim()) {
      setError("Target ID is required");
      return;
    }

    // Parse value based on type (try number first, then string)
    let parsedValue: unknown = value;
    if (value.trim() !== "") {
      const numValue = Number(value);
      if (!isNaN(numValue) && isFinite(numValue)) {
        parsedValue = numValue;
      } else if (value.toLowerCase() === "true") {
        parsedValue = true;
      } else if (value.toLowerCase() === "false") {
        parsedValue = false;
      } else {
        parsedValue = value;
      }
    }

    const updatedAction: WorkflowAction = {
      ...action,
      target_type: targetType,
      target_id: targetId,
      config_update: {
        operation,
        ...(operation === "patch"
          ? {
              yaml_path: yamlPath,
              value: parsedValue,
            }
          : {
              template: template,
            }),
      },
    };

    onSave(updatedAction);
  };

  if (!action) {
    return (
      <div className="p-4 text-sm text-muted-foreground">
        No action node selected
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {error && (
        <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
          {error}
        </div>
      )}

      {/* Target Configuration */}
      <div className="space-y-3 border-l-2 border-muted pl-4">
        <div className="space-y-2">
          <Label className="text-sm">Target Type</Label>
          <Select
            value={targetType}
            onValueChange={(v) => setTargetType(v as "agent" | "group")}
          >
            <SelectTrigger className="h-9">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="group">Group</SelectItem>
              <SelectItem value="agent">Agent</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div className="space-y-2">
          <Label className="text-sm">
            Target ID <span className="text-destructive">*</span>
          </Label>
          {!useCustomTarget ? (
            <Select
              value={getSelectValue()}
              onValueChange={(value) => {
                if (value === "__custom__") {
                  setUseCustomTarget(true);
                  setTargetId("");
                } else {
                  // For groups, we store the name (backend accepts both name and ID)
                  // For agents, we store the ID
                  setTargetId(value);
                }
              }}
            >
              <SelectTrigger className="h-9">
                <SelectValue placeholder={`Select ${targetType}`} />
              </SelectTrigger>
              <SelectContent>
                {targetType === "agent" ? (
                  <>
                    {agents.length === 0 ? (
                      <SelectItem value="__loading__" disabled>
                        Loading agents...
                      </SelectItem>
                    ) : (
                      <>
                        {agents.map((agent) => (
                          <SelectItem key={agent.id} value={agent.id}>
                            {agent.display}
                          </SelectItem>
                        ))}
                        <SelectItem value="__custom__">
                          Custom (use variable or manual entry)...
                        </SelectItem>
                      </>
                    )}
                  </>
                ) : (
                  <>
                    {groups.length === 0 ? (
                      <SelectItem value="__loading__" disabled>
                        Loading groups...
                      </SelectItem>
                    ) : (
                      <>
                        {groups.map((group) => (
                          <SelectItem key={group.id} value={group.name}>
                            {group.display}
                          </SelectItem>
                        ))}
                        <SelectItem value="__custom__">
                          Custom (use variable or manual entry)...
                        </SelectItem>
                      </>
                    )}
                  </>
                )}
              </SelectContent>
            </Select>
          ) : (
            <VariableAutocomplete
              value={targetId}
              onChange={setTargetId}
              placeholder={
                targetType === "group"
                  ? "group-name or ${app_name}"
                  : "agent-uuid"
              }
              className="h-9 font-mono text-sm"
              availableVariables={availableVariables}
            />
          )}
          {useCustomTarget && (
            <Button
              type="button"
              variant="ghost"
              size="sm"
              className="h-7 text-xs"
              onClick={() => {
                setUseCustomTarget(false);
                setTargetId("");
              }}
            >
              ← Back to dropdown
            </Button>
          )}
          <p className="text-xs text-muted-foreground">
            {targetType === "group" ? (
              <>
                {useCustomTarget
                  ? "Use group name (e.g., 'user-service') or metadata variable (e.g., "
                  : "Select a group or choose 'Custom' to use a variable (e.g., "}
                <code className="text-xs bg-muted px-1 rounded">
                  {"${app_name}"}
                </code>
                {useCustomTarget ? ")" : ")"}
              </>
            ) : useCustomTarget ? (
              "Enter agent UUID"
            ) : (
              "Select an agent or choose 'Custom' to enter a UUID manually"
            )}
          </p>
        </div>
      </div>

      {/* Operation Configuration */}
      <div className="space-y-3 border-l-2 border-muted pl-4">
        <div className="space-y-2">
          <Label className="text-sm">Operation</Label>
          <Select
            value={operation}
            onValueChange={(v) => setOperation(v as typeof operation)}
          >
            <SelectTrigger className="h-9">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="patch">
                Patch (Update specific YAML path)
              </SelectItem>
              <SelectItem value="merge">Merge (Merge YAML template)</SelectItem>
              <SelectItem value="replace">
                Replace (Replace entire config)
              </SelectItem>
            </SelectContent>
          </Select>
          <p className="text-xs text-muted-foreground">
            {operation === "patch" &&
              "Update a specific value in the config (e.g., sampling rate)"}
            {operation === "merge" &&
              "Merge a YAML template into the existing config"}
            {operation === "replace" &&
              "Replace the entire config with a new template"}
          </p>
        </div>

        {/* Patch Operation Fields */}
        {operation === "patch" && (
          <div className="space-y-3 pt-2">
            <div className="space-y-2">
              <Label className="text-sm">
                YAML Path <span className="text-destructive">*</span>
              </Label>
              <Input
                value={yamlPath}
                onChange={(e) => setYamlPath(e.target.value)}
                placeholder="service.pipelines.traces.processors.probabilistic_sampler.sampling_percentage"
                className="h-9 font-mono text-sm"
              />
              <p className="text-xs text-muted-foreground">
                Dot-separated path to the value (e.g.,{" "}
                <code className="text-xs bg-muted px-1 rounded">
                  service.pipelines.traces.processors.probabilistic_sampler.sampling_percentage
                </code>
                )
              </p>
            </div>

            <div className="space-y-2">
              <Label className="text-sm">
                Value <span className="text-destructive">*</span>
              </Label>
              <Input
                value={value}
                onChange={(e) => setValue(e.target.value)}
                placeholder="100"
                className="h-9 font-mono text-sm"
              />
              <p className="text-xs text-muted-foreground">
                Value to set (numbers, strings, or booleans are auto-detected)
              </p>
            </div>

            {/* Common Sampling Paths Helper */}
            <Alert>
              <Info className="h-4 w-4" />
              <AlertDescription className="text-xs">
                <strong>Common sampling paths:</strong>
                <ul className="list-disc list-inside mt-1 space-y-0.5">
                  <li>
                    <code className="text-xs bg-muted px-1 rounded">
                      service.pipelines.traces.processors.probabilistic_sampler.sampling_percentage
                    </code>{" "}
                    (traces)
                  </li>
                  <li>
                    <code className="text-xs bg-muted px-1 rounded">
                      service.pipelines.metrics.processors.probabilistic_sampler.sampling_percentage
                    </code>{" "}
                    (metrics)
                  </li>
                  <li>
                    <code className="text-xs bg-muted px-1 rounded">
                      service.pipelines.logs.processors.probabilistic_sampler.sampling_percentage
                    </code>{" "}
                    (logs)
                  </li>
                </ul>
              </AlertDescription>
            </Alert>
          </div>
        )}

        {/* Merge/Replace Operation Fields */}
        {(operation === "merge" || operation === "replace") && (
          <div className="space-y-2 pt-2">
            <Label className="text-sm">
              Configuration <span className="text-destructive">*</span>
            </Label>
            <Tabs
              value={configMode}
              onValueChange={(v) => setConfigMode(v as "yaml" | "form")}
              className="w-full"
            >
              <TabsList className="h-9">
                <TabsTrigger value="yaml" className="text-xs">
                  YAML Editor
                </TabsTrigger>
                <TabsTrigger value="form" className="text-xs">
                  UI Form
                </TabsTrigger>
              </TabsList>

              {/* YAML Editor Tab */}
              <TabsContent value="yaml" className="space-y-2 mt-3">
                <Textarea
                  value={template}
                  onChange={(e) => setTemplate(e.target.value)}
                  placeholder="processors:&#10;  probabilistic_sampler:&#10;    sampling_percentage: 100"
                  className="font-mono text-xs h-64"
                />
              </TabsContent>

              {/* UI Form Tab */}
              <TabsContent value="form" className="space-y-3 mt-3">
                {/* Component Selection */}
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
                      disabled={!componentType}
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

                {/* Schema Form */}
                {schema && componentType && componentName ? (
                  <div className="space-y-3">
                    <div className="max-h-[50vh] overflow-y-auto pr-2 border rounded-md p-4">
                      {schemaLoading ? (
                        <div className="flex items-center justify-center py-6">
                          <Loader2 className="h-5 w-5 animate-spin" />
                        </div>
                      ) : (
                        <SchemaForm
                          schema={schema}
                          value={formValue}
                          onChange={setFormValue}
                          path={[componentName]}
                        />
                      )}
                    </div>

                    {/* YAML Preview */}
                    {yamlPreview && (
                      <div className="space-y-2">
                        <Label className="text-sm">
                          Generated YAML Preview
                        </Label>
                        <Textarea
                          value={yamlPreview}
                          readOnly
                          className="font-mono text-xs h-32"
                        />
                      </div>
                    )}
                  </div>
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
              </TabsContent>
            </Tabs>
          </div>
        )}
      </div>

      {/* Footer Actions */}
      <div className="flex justify-end gap-2 pt-4 border-t">
        {onCancel && (
          <Button variant="outline" onClick={onCancel}>
            Cancel
          </Button>
        )}
        <Button onClick={handleSave}>
          <Save className="h-4 w-4 mr-2" />
          Save Action
        </Button>
      </div>
    </div>
  );
}
