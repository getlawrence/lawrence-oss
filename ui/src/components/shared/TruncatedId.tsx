import { Check, Copy } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";

interface TruncatedIdProps {
  id: string;
  maxLength?: number;
  className?: string;
  showCopyButton?: boolean;
}

export function TruncatedId({
  id,
  maxLength = 8,
  className,
  showCopyButton = true,
}: TruncatedIdProps) {
  const [copied, setCopied] = useState(false);

  const truncated = id.length > maxLength ? `${id.slice(0, maxLength)}...` : id;
  const shouldTruncate = id.length > maxLength;

  const handleCopy = async (e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await navigator.clipboard.writeText(id);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error("Failed to copy:", err);
    }
  };

  return (
    <div className={cn("flex items-center gap-2", className)}>
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <code
              className={cn(
                "font-mono text-xs text-muted-foreground",
                shouldTruncate && "cursor-help",
              )}
            >
              {truncated}
            </code>
          </TooltipTrigger>
          {shouldTruncate && (
            <TooltipContent>
              <code className="font-mono text-xs">{id}</code>
            </TooltipContent>
          )}
        </Tooltip>
      </TooltipProvider>

      {showCopyButton && (
        <Button
          variant="ghost"
          size="icon"
          className="h-6 w-6 shrink-0"
          onClick={handleCopy}
        >
          {copied ? (
            <Check className="h-3 w-3 text-green-600" />
          ) : (
            <Copy className="h-3 w-3" />
          )}
        </Button>
      )}
    </div>
  );
}
