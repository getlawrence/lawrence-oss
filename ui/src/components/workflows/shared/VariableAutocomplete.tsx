import { useState, useRef } from "react";
import { Input } from "@/components/ui/input";

interface VariableAutocompleteProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  className?: string;
  availableVariables: string[];
  onBlur?: () => void;
}

export function VariableAutocomplete({
  value,
  onChange,
  placeholder,
  className,
  availableVariables,
  onBlur,
}: VariableAutocompleteProps) {
  const [open, setOpen] = useState(false);
  const [cursorPosition, setCursorPosition] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  // Check if we're typing a variable (after ${)
  const getVariableContext = (text: string, cursorPos: number) => {
    // Find the last ${ before cursor
    const beforeCursor = text.substring(0, cursorPos);
    const lastDollarBrace = beforeCursor.lastIndexOf("${");
    
    if (lastDollarBrace === -1) {
      return null;
    }

    // Check if we haven't closed it yet
    const afterDollarBrace = beforeCursor.substring(lastDollarBrace + 2);
    const closingBrace = afterDollarBrace.indexOf("}");
    
    if (closingBrace !== -1 && closingBrace < cursorPos - lastDollarBrace - 2) {
      // Already closed, no autocomplete
      return null;
    }

    // Extract the partial variable name
    const partialVar = afterDollarBrace.substring(0, cursorPos - lastDollarBrace - 2);
    
    return {
      start: lastDollarBrace,
      partial: partialVar,
    };
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.value;
    const newCursorPos = e.target.selectionStart || 0;
    
    onChange(newValue);
    setCursorPosition(newCursorPos);

    const context = getVariableContext(newValue, newCursorPos);
    if (context && availableVariables.length > 0) {
      setOpen(true);
    } else {
      setOpen(false);
    }
  };

  const handleSelectVariable = (variable: string) => {
    const context = getVariableContext(value, cursorPosition);
    if (!context) return;

    const before = value.substring(0, context.start);
    const after = value.substring(cursorPosition);
    const newValue = `${before}\${${variable}}${after}`;
    
    onChange(newValue);
    setOpen(false);
    
    // Set cursor position after the inserted variable
    setTimeout(() => {
      if (inputRef.current) {
        const newPos = context.start + variable.length + 3; // ${variable}
        inputRef.current.setSelectionRange(newPos, newPos);
        inputRef.current.focus();
      }
    }, 0);
  };

  const context = getVariableContext(value, cursorPosition);
  const filteredVariables = context
    ? availableVariables.filter((v) =>
        v.toLowerCase().startsWith(context.partial.toLowerCase())
      )
    : [];

  return (
    <div className="relative">
      <Input
        ref={inputRef}
        value={value}
        onChange={handleInputChange}
        onBlur={() => {
          // Delay closing to allow click on suggestion
          setTimeout(() => setOpen(false), 200);
          onBlur?.();
        }}
        onFocus={() => {
          const context = getVariableContext(value, cursorPosition);
          if (context && availableVariables.length > 0) {
            setOpen(true);
          }
        }}
        placeholder={placeholder}
        className={className}
        onKeyDown={(e) => {
          if (open && filteredVariables.length > 0) {
            if (e.key === "Escape") {
              setOpen(false);
            }
          }
        }}
      />
      {open && filteredVariables.length > 0 && (
        <div className="absolute z-50 w-full mt-1 bg-popover border rounded-md shadow-md max-h-60 overflow-auto">
          <div className="p-1">
            {filteredVariables.map((variable) => (
              <div
                key={variable}
                className="px-2 py-1.5 text-sm cursor-pointer hover:bg-accent rounded-sm font-mono"
                onMouseDown={(e) => {
                  e.preventDefault(); // Prevent input blur
                  handleSelectVariable(variable);
                }}
              >
                <span className="text-muted-foreground">{"${"}</span>
                <span className="font-semibold">{variable}</span>
                <span className="text-muted-foreground">{"}"}</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

