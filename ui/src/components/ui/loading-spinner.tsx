import { cn } from "@/lib/utils";

interface LoadingSpinnerProps {
  /**
   * Size variant for the spinner
   */
  size?: "sm" | "md" | "lg";
  /**
   * Optional message to display below the spinner
   */
  message?: string;
  /**
   * Additional CSS classes
   */
  className?: string;
}

const sizeClasses = {
  sm: "h-4 w-4",
  md: "h-8 w-8",
  lg: "h-12 w-12",
};

/**
 * Reusable loading spinner component
 * 
 * @example
 * ```tsx
 * <LoadingSpinner size="md" message="Loading data..." />
 * ```
 */
export function LoadingSpinner({ 
  size = "md", 
  message,
  className 
}: LoadingSpinnerProps) {
  return (
    <div className={cn("flex items-center justify-center h-64", className)}>
      <div className="flex flex-col items-center gap-2">
        <div 
          className={cn(
            "animate-spin rounded-full border-b-2 border-gray-900",
            sizeClasses[size]
          )}
        />
        {message && <span className="text-sm text-gray-600">{message}</span>}
      </div>
    </div>
  );
}

