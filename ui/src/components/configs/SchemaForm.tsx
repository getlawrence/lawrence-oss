import { useState, useEffect } from "react";

import type { JSONSchema } from "@/api/schemas";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import { Textarea } from "@/components/ui/textarea";

interface SchemaFormProps {
  schema: JSONSchema;
  value?: unknown;
  onChange?: (value: unknown) => void;
  path?: string[];
  required?: boolean;
}

export function SchemaForm({
  schema,
  value,
  onChange,
  path = [],
  required = false,
}: SchemaFormProps) {
  const [formValue, setFormValue] = useState<unknown>(
    value ?? getDefaultValue(schema),
  );

  useEffect(() => {
    if (value !== undefined) {
      setFormValue(value);
    }
  }, [value]);

  const handleChange = (newValue: unknown) => {
    setFormValue(newValue);
    onChange?.(newValue);
  };

  const schemaType = Array.isArray(schema.type) ? schema.type[0] : schema.type;
  const fieldName = path[path.length - 1] || "root";
  const isRequired = required || schema.required?.includes(fieldName) || false;

  // Handle object type
  if (schemaType === "object" || schema.properties) {
    return (
      <ObjectForm
        schema={schema}
        value={formValue as Record<string, unknown>}
        onChange={handleChange}
        path={path}
        required={isRequired}
      />
    );
  }

  // Handle array type
  if (schemaType === "array" || schema.items) {
    return (
      <ArrayForm
        schema={schema}
        value={formValue as unknown[]}
        onChange={handleChange}
        path={path}
        required={isRequired}
      />
    );
  }

  // Handle primitive types
  return (
    <div className="space-y-1.5">
      <Label htmlFor={fieldName} className="text-sm">
        {fieldName}
        {isRequired && <span className="text-destructive ml-1">*</span>}
      </Label>
      {schema.description && (
        <p className="text-xs text-muted-foreground mb-1">
          {schema.description}
        </p>
      )}
      {renderPrimitiveField(schema, formValue, handleChange, fieldName)}
    </div>
  );
}

function ObjectForm({
  schema,
  value,
  onChange,
  path,
}: {
  schema: JSONSchema;
  value?: Record<string, unknown>;
  onChange?: (value: Record<string, unknown>) => void;
  path: string[];
  required: boolean;
}) {
  const properties = schema.properties || {};
  const formValue = value || {};

  const handlePropertyChange = (key: string, newValue: unknown) => {
    const updated = { ...formValue, [key]: newValue };
    onChange?.(updated);
  };

  const fieldName = path[path.length - 1] || "root";
  const isRoot = fieldName === "root";

  return (
    <div
      className={
        isRoot ? "space-y-3" : "space-y-3 border-l-2 border-muted pl-4 ml-2"
      }
    >
      {!isRoot && (
        <div className="mb-2">
          <Label className="text-sm font-medium">{fieldName}</Label>
          {schema.description && (
            <p className="text-xs text-muted-foreground mt-0.5">
              {schema.description}
            </p>
          )}
        </div>
      )}
      <div className="space-y-3">
        {Object.entries(properties).map(([key, propSchema], index) => {
          const isRequired = schema.required?.includes(key) || false;
          return (
            <div key={key}>
              {index > 0 && !isRoot && <Separator className="my-3" />}
              <SchemaForm
                schema={propSchema as JSONSchema}
                value={formValue[key]}
                onChange={(newValue) => handlePropertyChange(key, newValue)}
                path={[...path, key]}
                required={isRequired}
              />
            </div>
          );
        })}
      </div>
    </div>
  );
}

function ArrayForm({
  schema,
  value,
  onChange,
  path,
  required,
}: {
  schema: JSONSchema;
  value?: unknown[];
  onChange?: (value: unknown[]) => void;
  path: string[];
  required: boolean;
}) {
  const itemsSchema = schema.items as JSONSchema | undefined;
  const formValue = value || [];
  const fieldName = path[path.length - 1] || "items";

  const handleItemChange = (index: number, newValue: unknown) => {
    const updated = [...formValue];
    updated[index] = newValue;
    onChange?.(updated);
  };

  const handleAddItem = () => {
    const newItem = getDefaultValue(itemsSchema);
    onChange?.([...formValue, newItem]);
  };

  const handleRemoveItem = (index: number) => {
    const updated = formValue.filter((_, i) => i !== index);
    onChange?.(updated);
  };

  return (
    <div className="space-y-2">
      <Label className="text-sm">
        {fieldName}
        {required && <span className="text-destructive ml-1">*</span>}
      </Label>
      {schema.description && (
        <p className="text-xs text-muted-foreground mb-1">
          {schema.description}
        </p>
      )}
      <div className="space-y-2">
        {formValue.map((item, index) => (
          <div
            key={index}
            className="flex gap-2 items-start border-l-2 border-muted pl-3 py-1"
          >
            <div className="flex-1 min-w-0">
              {itemsSchema ? (
                <SchemaForm
                  schema={itemsSchema}
                  value={item}
                  onChange={(newValue) => handleItemChange(index, newValue)}
                  path={[...path, `${index}`]}
                />
              ) : (
                <Input
                  value={String(item)}
                  onChange={(e) => handleItemChange(index, e.target.value)}
                  className="h-8 text-sm"
                />
              )}
            </div>
            <button
              type="button"
              onClick={() => handleRemoveItem(index)}
              className="text-destructive hover:underline text-xs px-2 py-1 shrink-0"
            >
              Remove
            </button>
          </div>
        ))}
        <button
          type="button"
          onClick={handleAddItem}
          className="text-xs text-primary hover:underline"
        >
          + Add item
        </button>
      </div>
    </div>
  );
}

function renderPrimitiveField(
  schema: JSONSchema,
  value: unknown,
  onChange: (value: unknown) => void,
  fieldName: string,
) {
  const schemaType = Array.isArray(schema.type) ? schema.type[0] : schema.type;

  // Handle enum
  if (schema.enum && schema.enum.length > 0) {
    return (
      <Select
        value={String(value ?? schema.default ?? "")}
        onValueChange={(val) => onChange(val)}
      >
        <SelectTrigger className="h-8 text-sm">
          <SelectValue placeholder="Select..." />
        </SelectTrigger>
        <SelectContent>
          {schema.enum.map((option) => (
            <SelectItem key={String(option)} value={String(option)}>
              {String(option)}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    );
  }

  // Handle boolean
  if (schemaType === "boolean") {
    return (
      <div className="flex items-center space-x-2">
        <Checkbox
          checked={Boolean(value ?? schema.default ?? false)}
          onCheckedChange={(checked) => onChange(checked)}
        />
        <Label className="text-sm font-normal">
          {value ? "Enabled" : "Disabled"}
        </Label>
      </div>
    );
  }

  // Handle number/integer
  if (schemaType === "number" || schemaType === "integer") {
    return (
      <Input
        type="number"
        value={
          value !== undefined ? String(value) : String(schema.default ?? "")
        }
        onChange={(e) => {
          const numValue =
            schemaType === "integer"
              ? parseInt(e.target.value, 10)
              : parseFloat(e.target.value);
          onChange(isNaN(numValue) ? undefined : numValue);
        }}
        placeholder={schema.description}
        className="h-8 text-sm"
      />
    );
  }

  // Handle string (default to textarea for longer strings, input for short ones)
  if (schemaType === "string") {
    const stringValue =
      value !== undefined ? String(value) : String(schema.default ?? "");
    // Use textarea if description suggests it's a longer field or if it's a multiline string
    const useTextarea =
      schema.description?.toLowerCase().includes("yaml") ||
      schema.description?.toLowerCase().includes("config") ||
      fieldName.toLowerCase().includes("config") ||
      fieldName.toLowerCase().includes("template");

    if (useTextarea) {
      return (
        <Textarea
          value={stringValue}
          onChange={(e) => onChange(e.target.value)}
          placeholder={schema.description}
          rows={4}
          className="font-mono text-xs"
        />
      );
    }

    return (
      <Input
        type="text"
        value={stringValue}
        onChange={(e) => onChange(e.target.value)}
        placeholder={schema.description}
        className="h-8 text-sm"
      />
    );
  }

  // Fallback to text input
  return (
    <Input
      value={value !== undefined ? String(value) : ""}
      onChange={(e) => onChange(e.target.value)}
      placeholder={schema.description}
      className="h-8 text-sm"
    />
  );
}

function getDefaultValue(schema?: JSONSchema): unknown {
  if (!schema) {
    return undefined;
  }

  if (schema.default !== undefined) {
    return schema.default;
  }

  const schemaType = Array.isArray(schema.type) ? schema.type[0] : schema.type;

  switch (schemaType) {
    case "object":
      return {};
    case "array":
      return [];
    case "boolean":
      return false;
    case "number":
    case "integer":
      return 0;
    case "string":
      return "";
    default:
      return undefined;
  }
}
