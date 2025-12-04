import { Plus, Play, RefreshCw, Trash2, Workflow as WorkflowIcon, History } from "lucide-react";
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import useSWR from "swr";

import {
  getWorkflows,
  deleteWorkflow,
  executeWorkflow,
  type Workflow,
} from "@/api/workflows";
import { PageTable } from "@/components/shared/PageTable";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { TableCell } from "@/components/ui/table";
import { WorkflowExecutionsDrawer } from "@/components/workflows/execution/WorkflowExecutionsDrawer";

export default function WorkflowsPage() {
  const navigate = useNavigate();
  const [refreshing, setRefreshing] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedWorkflow, setSelectedWorkflow] = useState<Workflow | null>(
    null,
  );
  const [executionsDrawerOpen, setExecutionsDrawerOpen] = useState(false);
  const [selectedWorkflowId, setSelectedWorkflowId] = useState<string | null>(
    null,
  );

  const { data, error, mutate } = useSWR("workflows", getWorkflows, {
    refreshInterval: 30000,
  });

  const handleRefresh = async () => {
    setRefreshing(true);
    await mutate();
    setRefreshing(false);
  };

  const handleExecute = async (workflow: Workflow) => {
    try {
      await executeWorkflow(workflow.id);
      await mutate();
    } catch (err) {
      console.error("Failed to execute workflow:", err);
    }
  };

  const handleDelete = async () => {
    if (!selectedWorkflow) return;
    try {
      await deleteWorkflow(selectedWorkflow.id);
      setDeleteDialogOpen(false);
      setSelectedWorkflow(null);
      await mutate();
    } catch (err) {
      console.error("Failed to delete workflow:", err);
    }
  };

  const handleDeleteClick = (workflow: Workflow, e: React.MouseEvent) => {
    e.stopPropagation();
    setSelectedWorkflow(workflow);
    setDeleteDialogOpen(true);
  };

  const workflows = data?.workflows || [];

  if (error) {
    return (
      <div className="container mx-auto p-6">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-red-600 mb-4">
            Error Loading Workflows
          </h1>
          <p className="text-muted-foreground">{error.message}</p>
          <Button onClick={handleRefresh} className="mt-4">
            <RefreshCw className="h-4 w-4 mr-2" />
            Retry
          </Button>
        </div>
      </div>
    );
  }

  return (
    <>
      <PageTable
        pageTitle="Workflows"
        pageDescription="Automate collector configuration updates"
        pageActions={[
          {
            label: "Refresh",
            icon: RefreshCw,
            onClick: handleRefresh,
            disabled: refreshing,
            variant: "ghost" as const,
          },
          {
            label: "Create Workflow",
            icon: Plus,
            onClick: () => navigate("/workflows/new"),
            variant: "default" as const,
          },
        ]}
        cardTitle={`Workflows (${workflows.length})`}
        cardDescription="All configured workflows and their status"
        columns={[
          { header: "Name", key: "name" },
          { header: "Type", key: "type" },
          { header: "Status", key: "status" },
          { header: "Schedule", key: "schedule" },
          { header: "Runs", key: "runs" },
          { header: "Last Run", key: "last_run" },
          { header: "Next Run", key: "next_run" },
          { header: "Actions", key: "actions" },
        ]}
        data={workflows}
        getRowKey={(workflow: Workflow) => workflow.id}
        onRowClick={(workflow: Workflow) =>
          navigate(`/workflows/${workflow.id}/edit`)
        }
        renderRow={(workflow: Workflow) => (
          <>
            <TableCell className="font-medium">
              <div className="flex items-center gap-2">
                <div>
                  <div>{workflow.name}</div>
                  {workflow.description && (
                    <div className="text-xs text-muted-foreground mt-0.5">
                      {workflow.description}
                    </div>
                  )}
                </div>
              </div>
            </TableCell>
            <TableCell>
              <Badge variant="outline">{workflow.type}</Badge>
            </TableCell>
            <TableCell>
              <Badge
                variant={workflow.status === "active" ? "default" : "secondary"}
              >
                {workflow.status}
              </Badge>
            </TableCell>
            <TableCell>
              {workflow.schedule ? (
                <code className="text-xs bg-muted px-2 py-1 rounded">
                  {workflow.schedule.cron_expression}
                </code>
              ) : (
                <span className="text-xs text-muted-foreground">—</span>
              )}
            </TableCell>
            <TableCell>
              <div className="flex flex-col gap-1">
                <span className="text-sm">{workflow.run_count}</span>
                {workflow.error_count > 0 && (
                  <span className="text-xs text-red-600">
                    {workflow.error_count} errors
                  </span>
                )}
              </div>
            </TableCell>
            <TableCell>
              {workflow.last_run ? (
                <span className="text-sm">
                  {new Date(workflow.last_run).toLocaleString()}
                </span>
              ) : (
                <span className="text-xs text-muted-foreground">Never</span>
              )}
            </TableCell>
            <TableCell>
              {workflow.next_run && workflow.type === "schedule" ? (
                <span className="text-sm">
                  {new Date(workflow.next_run).toLocaleString()}
                </span>
              ) : (
                <span className="text-xs text-muted-foreground">—</span>
              )}
            </TableCell>
            <TableCell>
              <div className="flex gap-2">
                {workflow.run_count > 0 && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={(e) => {
                      e.stopPropagation();
                      setSelectedWorkflowId(workflow.id);
                      setExecutionsDrawerOpen(true);
                    }}
                  >
                    <History className="h-4 w-4" />
                  </Button>
                )}
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={(e) => {
                    e.stopPropagation();
                    handleExecute(workflow);
                  }}
                >
                  <Play className="h-4 w-4" />
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={(e) => handleDeleteClick(workflow, e)}
                  className="text-red-600 hover:text-red-700 hover:bg-red-50 dark:text-red-400 dark:hover:text-red-300 dark:hover:bg-red-950/30"
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            </TableCell>
          </>
        )}
        emptyState={{
          icon: WorkflowIcon,
          title: "No Workflows Found",
          description:
            "Create your first workflow to automate collector configuration updates.",
          action: {
            label: "Create Workflow",
            onClick: () => navigate("/workflows/new"),
          },
        }}
      />

      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Workflow</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete "{selectedWorkflow?.name}"? This
              action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setDeleteDialogOpen(false)}
            >
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleDelete}>
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <WorkflowExecutionsDrawer
        workflowId={selectedWorkflowId}
        open={executionsDrawerOpen}
        onOpenChange={(open: boolean) => {
          setExecutionsDrawerOpen(open);
          if (!open) {
            setSelectedWorkflowId(null);
          }
        }}
      />
    </>
  );
}
