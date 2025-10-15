import { Plus, RefreshCw, Server, Trash2, Users } from "lucide-react";
import { useState } from "react";
import useSWR from "swr";

import { getGroups, createGroup, deleteGroup } from "@/api/groups";
import type { Group, CreateGroupRequest } from "@/api/groups";
import { GroupDetailsDrawer } from "@/components/GroupDetailsDrawer";
import { PageTable } from "@/components/shared/PageTable";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { TableCell } from "@/components/ui/table";

export default function GroupsPage() {
  const [refreshing, setRefreshing] = useState(false);
  const [createDrawerOpen, setCreateDrawerOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedGroup, setSelectedGroup] = useState<Group | null>(null);
  const [selectedGroupId, setSelectedGroupId] = useState<string | null>(null);
  const [groupDrawerOpen, setGroupDrawerOpen] = useState(false);
  const [createForm, setCreateForm] = useState<CreateGroupRequest>({
    name: "",
    labels: {},
  });

  const {
    data: groupsData,
    error: groupsError,
    mutate: mutateGroups,
  } = useSWR("groups", getGroups, { refreshInterval: 30000 });

  const handleRefresh = async () => {
    setRefreshing(true);
    await mutateGroups();
    setRefreshing(false);
  };

  const handleCreateGroup = async () => {
    try {
      await createGroup(createForm);
      setCreateDrawerOpen(false);
      setCreateForm({ name: "", labels: {} });
      await mutateGroups();
    } catch (error) {
      console.error("Failed to create group:", error);
    }
  };

  const handleDeleteGroup = async () => {
    if (!selectedGroup) return;
    try {
      await deleteGroup(selectedGroup.id);
      setDeleteDialogOpen(false);
      setSelectedGroup(null);
      await mutateGroups();
    } catch (error) {
      console.error("Failed to delete group:", error);
    }
  };

  const handleGroupClick = (groupId: string) => {
    setSelectedGroupId(groupId);
    setGroupDrawerOpen(true);
  };

  const handleDeleteClick = (group: Group, e: React.MouseEvent) => {
    e.stopPropagation();
    setSelectedGroup(group);
    setDeleteDialogOpen(true);
  };

  if (groupsError) {
    return (
      <div className="container mx-auto p-6">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-red-600 mb-4">
            Error Loading Groups
          </h1>
          <p className="text-muted-foreground">{groupsError.message}</p>
          <Button onClick={handleRefresh} className="mt-4">
            <RefreshCw className="h-4 w-4 mr-2" />
            Retry
          </Button>
        </div>
      </div>
    );
  }

  const groups = groupsData?.groups || [];

  return (
    <>
      <PageTable
        pageTitle="Groups"
        pageDescription="Organize agents into groups for easier management"
        pageActions={[
          {
            label: "Refresh",
            icon: RefreshCw,
            onClick: handleRefresh,
            disabled: refreshing,
            variant: "ghost" as const,
          },
          {
            label: "Create Group",
            icon: Plus,
            onClick: () => setCreateDrawerOpen(true),
            variant: "default" as const,
          },
        ]}
        cardTitle={`Groups (${groups.length})`}
        cardDescription="All agent groups and their details"
        columns={[
          { header: "Name", key: "name" },
          { header: "Agents", key: "agents" },
          { header: "Config", key: "config" },
          { header: "Created", key: "created" },
          { header: "Updated", key: "updated" },
          { header: "Labels", key: "labels" },
          { header: "Actions", key: "actions" },
        ]}
        data={groups}
        getRowKey={(group: Group) => group.id}
        onRowClick={(group: Group) => handleGroupClick(group.id)}
        renderRow={(group: Group) => (
          <>
            <TableCell className="font-medium">{group.name}</TableCell>
            <TableCell>
              <div className="flex items-center gap-2">
                <Server className="h-4 w-4 text-muted-foreground" />
                <span>{group.agent_count}</span>
              </div>
            </TableCell>
            <TableCell>
              {group.config_name ? (
                <span className="text-sm font-mono text-muted-foreground">
                  {group.config_name}
                </span>
              ) : (
                <span className="text-xs text-muted-foreground">
                  No config
                </span>
              )}
            </TableCell>
            <TableCell>
              {new Date(group.created_at).toLocaleDateString()}
            </TableCell>
            <TableCell>
              {new Date(group.updated_at).toLocaleDateString()}
            </TableCell>
            <TableCell>
              <div className="flex flex-wrap gap-1">
                {Object.entries(group.labels).map(([key, value]) => (
                  <span
                    key={key}
                    className="text-xs bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded"
                  >
                    {key}={value}
                  </span>
                ))}
                {Object.keys(group.labels).length === 0 && (
                  <span className="text-xs text-muted-foreground">
                    No labels
                  </span>
                )}
              </div>
            </TableCell>
            <TableCell>
              <Button
                variant="ghost"
                size="sm"
                onClick={(e) => handleDeleteClick(group, e)}
                className="text-red-600 hover:text-red-700 hover:bg-red-50 dark:text-red-400 dark:hover:text-red-300 dark:hover:bg-red-950/30"
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </TableCell>
          </>
        )}
        emptyState={{
          icon: Users,
          title: "No Groups Found",
          description: "Create your first group to organize your agents.",
          action: {
            label: "Create Group",
            onClick: () => setCreateDrawerOpen(true),
          },
        }}
      />

      {/* Create Group Drawer */}
      <Sheet open={createDrawerOpen} onOpenChange={setCreateDrawerOpen}>
        <SheetContent>
          <SheetHeader>
            <SheetTitle>Create New Group</SheetTitle>
            <SheetDescription>
              Create a new group to organize your agents.
            </SheetDescription>
          </SheetHeader>
          <div className="space-y-4 mt-6">
            <div>
              <Label htmlFor="name">Group Name</Label>
              <Input
                id="name"
                value={createForm.name}
                onChange={(e) =>
                  setCreateForm({ ...createForm, name: e.target.value })
                }
                placeholder="Enter group name"
              />
            </div>
          </div>
          <SheetFooter className="mt-6">
            <Button
              variant="outline"
              onClick={() => setCreateDrawerOpen(false)}
            >
              Cancel
            </Button>
            <Button onClick={handleCreateGroup} disabled={!createForm.name}>
              Create Group
            </Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Group</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete the group "{selectedGroup?.name}"?
              This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setDeleteDialogOpen(false)}
            >
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleDeleteGroup}>
              Delete Group
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <GroupDetailsDrawer
        groupId={selectedGroupId}
        open={groupDrawerOpen}
        onOpenChange={setGroupDrawerOpen}
      />
    </>
  );
}
