import { Save } from "lucide-react";
import { type ReactNode } from "react";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";

interface ConfigDrawerLayoutProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description: string;
  error?: string | null;
  onSave: () => void;
  onCancel?: () => void;
  children: ReactNode;
  maxWidth?: "sm" | "md" | "lg" | "xl" | "2xl" | "3xl";
  saveLabel?: string;
  saveIcon?: ReactNode;
  footerContent?: ReactNode;
}

const maxWidthClasses = {
  sm: "sm:max-w-sm",
  md: "sm:max-w-md",
  lg: "sm:max-w-lg",
  xl: "sm:max-w-xl",
  "2xl": "sm:max-w-2xl",
  "3xl": "sm:max-w-3xl",
};

/**
 * Reusable layout component for configuration drawers.
 * 
 * Eliminates repetitive markup across drawer components by providing:
 * - Sheet wrapper with consistent styling
 * - Header with title and description
 * - Error alert display
 * - Footer with Cancel and Save buttons
 * - Consistent spacing and layout
 * 
 * @example
 * ```tsx
 * <ConfigDrawerLayout
 *   open={open}
 *   onOpenChange={setOpen}
 *   title="Configure Node"
 *   description="Configure node settings"
 *   error={error}
 *   onSave={handleSave}
 * >
 *   {/* Your form fields here *\/}
 * </ConfigDrawerLayout>
 * ```
 */
export function ConfigDrawerLayout({
  open,
  onOpenChange,
  title,
  description,
  error,
  onSave,
  onCancel,
  children,
  maxWidth = "2xl",
  saveLabel = "Save",
  saveIcon = <Save className="h-4 w-4 mr-2" />,
  footerContent,
}: ConfigDrawerLayoutProps) {
  const handleCancel = () => {
    if (onCancel) {
      onCancel();
    } else {
      onOpenChange(false);
    }
  };

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className={`w-full ${maxWidthClasses[maxWidth]} overflow-y-auto p-6`}>
        <SheetHeader>
          <SheetTitle>{title}</SheetTitle>
          <SheetDescription>{description}</SheetDescription>
        </SheetHeader>

        <div className="mt-6 space-y-6">
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {children}

          <div className="flex justify-end gap-2 pt-4 border-t">
            {footerContent || (
              <>
                <Button variant="outline" onClick={handleCancel}>
                  Cancel
                </Button>
                <Button onClick={onSave}>
                  {saveIcon}
                  {saveLabel}
                </Button>
              </>
            )}
          </div>
        </div>
      </SheetContent>
    </Sheet>
  );
}

