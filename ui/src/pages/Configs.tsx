import { Plus, RefreshCw, FileText, Hash } from "lucide-react";
import { useState } from "react";
import useSWR from "swr";

import { getConfigs, createConfig } from "@/api/configs";
import type { Config, CreateConfigRequest } from "@/api/configs";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Textarea } from "@/components/ui/textarea";

export default function ConfigsPage() {
  const [refreshing, setRefreshing] = useState(false);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [createForm, setCreateForm] = useState<CreateConfigRequest>({
    agent_id: "",
    group_id: "",
    config_hash: "",
    content: "",
    version: 1,
  });

  const {
    data: configsData,
    error: configsError,
    mutate: mutateConfigs,
  } = useSWR("configs", () => getConfigs({ limit: 100 }), {
    refreshInterval: 30000,
  });

  const handleRefresh = async () => {
    setRefreshing(true);
    await mutateConfigs();
    setRefreshing(false);
  };

  const handleCreateConfig = async () => {
    try {
      const configData = {
        ...createForm,
        agent_id: createForm.agent_id || undefined,
        group_id: createForm.group_id || undefined,
        config_hash: createForm.config_hash || `hash_${Date.now()}`,
      };
      await createConfig(configData);
      setCreateDialogOpen(false);
      setCreateForm({
        agent_id: "",
        group_id: "",
        config_hash: "",
        content: "",
        version: 1,
      });
      await mutateConfigs();
    } catch (error) {
      console.error("Failed to create config:", error);
    }
  };

  if (configsError) {
    return (
      <div className="container mx-auto p-6">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-red-600 mb-4">
            Error Loading Configs
          </h1>
          <p className="text-gray-600">{configsError.message}</p>
          <Button onClick={handleRefresh} className="mt-4">
            <RefreshCw className="h-4 w-4 mr-2" />
            Retry
          </Button>
        </div>
      </div>
    );
  }

  const configs = configsData?.configs || [];

  return (
    <div className="container mx-auto p-6 space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">Configurations</h1>
          <p className="text-gray-600">Manage agent and group configurations</p>
        </div>
        <div className="flex space-x-2">
          <Button onClick={handleRefresh} disabled={refreshing}>
            <RefreshCw
              className={`h-4 w-4 mr-2 ${refreshing ? "animate-spin" : ""}`}
            />
            Refresh
          </Button>
          <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
            <DialogTrigger asChild>
              <Button>
                <Plus className="h-4 w-4 mr-2" />
                Create Config
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-2xl">
              <DialogHeader>
                <DialogTitle>Create New Configuration</DialogTitle>
                <DialogDescription>
                  Create a new configuration for agents or groups.
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label htmlFor="agent_id">Agent ID (optional)</Label>
                    <Input
                      id="agent_id"
                      value={createForm.agent_id}
                      onChange={(e) =>
                        setCreateForm({
                          ...createForm,
                          agent_id: e.target.value,
                        })
                      }
                      placeholder="Agent UUID"
                    />
                  </div>
                  <div>
                    <Label htmlFor="group_id">Group ID (optional)</Label>
                    <Input
                      id="group_id"
                      value={createForm.group_id}
                      onChange={(e) =>
                        setCreateForm({
                          ...createForm,
                          group_id: e.target.value,
                        })
                      }
                      placeholder="Group ID"
                    />
                  </div>
                </div>
                <div>
                  <Label htmlFor="config_hash">Config Hash</Label>
                  <Input
                    id="config_hash"
                    value={createForm.config_hash}
                    onChange={(e) =>
                      setCreateForm({
                        ...createForm,
                        config_hash: e.target.value,
                      })
                    }
                    placeholder="Configuration hash"
                  />
                </div>
                <div>
                  <Label htmlFor="version">Version</Label>
                  <Input
                    id="version"
                    type="number"
                    value={createForm.version}
                    onChange={(e) =>
                      setCreateForm({
                        ...createForm,
                        version: parseInt(e.target.value) || 1,
                      })
                    }
                    placeholder="1"
                  />
                </div>
                <div>
                  <Label htmlFor="content">Configuration Content</Label>
                  <Textarea
                    id="content"
                    value={createForm.content}
                    onChange={(e) =>
                      setCreateForm({ ...createForm, content: e.target.value })
                    }
                    placeholder="Enter configuration content (YAML, JSON, etc.)"
                    rows={8}
                  />
                </div>
              </div>
              <DialogFooter>
                <Button
                  variant="outline"
                  onClick={() => setCreateDialogOpen(false)}
                >
                  Cancel
                </Button>
                <Button
                  onClick={handleCreateConfig}
                  disabled={!createForm.content}
                >
                  Create Config
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      {/* Configs Table */}
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
              <Button onClick={() => setCreateDialogOpen(true)}>
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
