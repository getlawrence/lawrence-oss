import { Check, Pencil, X } from "lucide-react";
import { useState, useRef, useEffect } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

interface EditableConfigTitleProps {
  value: string;
  onChange: (value: string) => void;
  className?: string;
}

export function EditableConfigTitle({
  value,
  onChange,
  className = "",
}: EditableConfigTitleProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [editValue, setEditValue] = useState(value);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    setEditValue(value);
  }, [value]);

  useEffect(() => {
    if (isEditing && inputRef.current) {
      inputRef.current.focus();
      inputRef.current.select();
    }
  }, [isEditing]);

  const handleSave = () => {
    const trimmedValue = editValue.trim();
    if (trimmedValue) {
      onChange(trimmedValue);
      setIsEditing(false);
    }
  };

  const handleCancel = () => {
    setEditValue(value);
    setIsEditing(false);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      handleSave();
    } else if (e.key === "Escape") {
      handleCancel();
    }
  };

  if (isEditing) {
    return (
      <div className={`flex items-center gap-2 ${className}`}>
        <Input
          ref={inputRef}
          value={editValue}
          onChange={(e) => setEditValue(e.target.value)}
          onKeyDown={handleKeyDown}
          className="h-8 text-lg font-semibold"
          placeholder="Config name"
        />
        <Button
          variant="ghost"
          size="icon"
          className="h-8 w-8"
          onClick={handleSave}
          disabled={!editValue.trim()}
        >
          <Check className="h-4 w-4" />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          className="h-8 w-8"
          onClick={handleCancel}
        >
          <X className="h-4 w-4" />
        </Button>
      </div>
    );
  }

  return (
    <div className={`flex items-center gap-2 group ${className}`}>
      <h1
        className="text-lg font-semibold truncate max-w-[400px]"
        title={value}
      >
        {value}
      </h1>
      <Button
        variant="ghost"
        size="icon"
        className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0"
        onClick={() => setIsEditing(true)}
      >
        <Pencil className="h-3 w-3" />
      </Button>
    </div>
  );
}
