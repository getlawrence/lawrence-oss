import { Plus, RefreshCw, FileText, Hash, Edit } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type { Config } from "@/api/configs";

interface ConfigsListProps {
  configs: Config[];
  refreshing: boolean;
  onRefresh: () => void;
  onCreateNew: () => void;
  onEditConfig: (config: Config) => void;
}

export function ConfigsList({
  configs,
  refreshing,
  onRefresh,
  onCreateNew,
  onEditConfig,
}: ConfigsListProps) {
  return (
    <div className="container mx-auto p-6 space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">Configurations</h1>
          <p className="text-gray-600">Manage agent and group configurations</p>
        </div>
        <div className="flex space-x-2">
          <Button onClick={onRefresh} disabled={refreshing}>
            <RefreshCw
              className={`h-4 w-4 mr-2 ${refreshing ? "animate-spin" : ""}`}
            />
            Refresh
          </Button>
          <Button onClick={onCreateNew}>
            <Plus className="h-4 w-4 mr-2" />
            Create Config
          </Button>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Configurations ({configs.length})</CardTitle>
          <CardDescription>All agent and group configurations</CardDescription>
        </CardHeader>
        <CardContent>
          {configs.length === 0 ? (
            <div className="text-center py-8">
              <FileText className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-gray-900 mb-2">
                No Configurations Found
              </h3>
              <p className="text-gray-600 mb-4">
                Create your first configuration to manage your agents.
              </p>
              <Button onClick={onCreateNew}>
                <Plus className="h-4 w-4 mr-2" />
                Create Configuration
              </Button>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>ID</TableHead>
                  <TableHead>Agent/Group</TableHead>
                  <TableHead>Version</TableHead>
                  <TableHead>Hash</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead>Content Preview</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {configs.map((config: Config) => (
                  <TableRow key={config.id}>
                    <TableCell className="font-mono text-sm">
                      {config.id.slice(0, 8)}...
                    </TableCell>
                    <TableCell>
                      {config.agent_id ? (
                        <span className="text-blue-600">
                          Agent: {config.agent_id.slice(0, 8)}...
                        </span>
                      ) : config.group_id ? (
                        <span className="text-green-600">
                          Group: {config.group_id}
                        </span>
                      ) : (
                        <span className="text-gray-500">Global</span>
                      )}
                    </TableCell>
                    <TableCell>
                      <span className="bg-blue-100 text-blue-800 px-2 py-1 rounded text-xs">
                        v{config.version}
                      </span>
                    </TableCell>
                    <TableCell className="font-mono text-xs">
                      <div className="flex items-center">
                        <Hash className="h-3 w-3 mr-1" />
                        {config.config_hash.slice(0, 8)}...
                      </div>
                    </TableCell>
                    <TableCell>
                      {new Date(config.created_at).toLocaleDateString()}
                    </TableCell>
                    <TableCell className="max-w-xs">
                      <div className="truncate text-sm text-gray-600">
                        {config.content.slice(0, 100)}
                        {config.content.length > 100 && "..."}
                      </div>
                    </TableCell>
                    <TableCell>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => onEditConfig(config)}
                      >
                        <Edit className="h-4 w-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

