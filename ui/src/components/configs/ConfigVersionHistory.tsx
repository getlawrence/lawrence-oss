import type { Config } from "@/api/configs";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

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
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Configuration Versions</DialogTitle>
          <DialogDescription>
            View and restore previous configuration versions
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-2 max-h-96 overflow-y-auto">
          {versions.map((version) => (
            <div
              key={version.id}
              className="flex items-center justify-between p-3 border rounded hover:bg-gray-50 cursor-pointer"
              onClick={() => onLoadVersion(version)}
            >
              <div>
                <div className="font-medium">Version {version.version}</div>
                <div className="text-sm text-gray-600">
                  {new Date(version.created_at).toLocaleString()}
                </div>
              </div>
              <Badge variant="outline">
                {version.config_hash.substring(0, 8)}
              </Badge>
            </div>
          ))}
          {versions.length === 0 && (
            <div className="text-center py-8 text-gray-500">
              No version history available. Select a group to view versions.
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
