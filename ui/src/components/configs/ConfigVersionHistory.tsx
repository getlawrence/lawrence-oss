import type { Config } from "@/api/configs";
import { Badge } from "@/components/ui/badge";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";

interface ConfigVersionHistoryProps {
  open: boolean;
  versions: Config[];
  onOpenChange: (open: boolean) => void;
  onLoadVersion: (config: Config) => void;
}

export function ConfigVersionHistory({
  open,
  versions,
  onOpenChange,
  onLoadVersion,
}: ConfigVersionHistoryProps) {
  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="w-[400px] sm:w-[540px] p-6">
        <SheetHeader>
          <SheetTitle>Configuration Versions</SheetTitle>
          <SheetDescription>
            View and restore previous configuration versions
          </SheetDescription>
        </SheetHeader>
        <div className="space-y-2 mt-6 overflow-y-auto">
          {versions.map((version) => (
            <div
              key={version.id}
              className="flex items-center justify-between p-3 border rounded hover:bg-accent cursor-pointer transition-colors"
              onClick={() => onLoadVersion(version)}
            >
              <div>
                <div className="font-medium">Version {version.version}</div>
                <div className="text-sm text-muted-foreground">
                  {new Date(version.created_at).toLocaleString()}
                </div>
              </div>
              <Badge variant="outline">
                {version.config_hash.substring(0, 8)}
              </Badge>
            </div>
          ))}
          {versions.length === 0 && (
            <div className="text-center py-8 text-muted-foreground">
              No version history available. Select a group to view versions.
            </div>
          )}
        </div>
      </SheetContent>
    </Sheet>
  );
}
