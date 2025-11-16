import {
  AlertCircle,
  ArrowLeft,
  Save,
  Webhook,
  Workflow as WorkflowIcon,
  X,
} from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import useSWR from "swr";

import {
  getWorkflow,
  createWorkflow,
  updateWorkflow,
  type Workflow as ApiWorkflow,
} from "@/api/workflows";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { WorkflowFlowEditor } from "@/components/workflows/core/WorkflowFlowEditor";
import { WebhookCredentials } from "@/components/workflows/shared/WebhookCredentials";
import { type Workflow } from "@/components/workflows/types/flow-types";
import {
  workflowToFlow,
  flowToWorkflow,
  validateFlow,
} from "@/components/workflows/utils/flow-utils";
import { useFormState } from "@/hooks/useDrawerForm";

export default function WorkflowEditor() {
  const navigate = useNavigate();
  const { id } = useParams();
  const isNew = !id;

  const { data: workflow } = useSWR(
    isNew ? null : `/api/v1/workflows/${id}`,
    () =>
      id ? getWorkflow(id) : Promise.reject(new Error("Invalid workflow id")),
  );

  const initialState = useMemo(
    () => {
      let flow: Workflow | null = null;
      if (workflow) {
        try {
          flow = workflowToFlow(workflow) as Workflow;
        } catch (error) {
          console.error("Failed to convert workflow to flow:", error);
        }
      }
      return {
        name: workflow?.name || "",
        description: workflow?.description || "",
        flow,
      };
    },
    [workflow]
  );

  const { state, updateField, resetState } = useFormState(initialState);
  const [createdWorkflow, setCreatedWorkflow] = useState<ApiWorkflow | null>(null);
  const [webhookCredentialsOpen, setWebhookCredentialsOpen] = useState(false);
  const [webhookSheetOpen, setWebhookSheetOpen] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string>("");

  // Reset form state when workflow data loads (for editing existing workflow)
  // This is appropriate useEffect usage - resetting state when SWR data changes
  useEffect(() => {
    if (workflow && !isNew) {
      let flow: Workflow | null = null;
      try {
        flow = workflowToFlow(workflow) as Workflow;
      } catch (error) {
        console.error("Failed to convert workflow to flow:", error);
      }
      resetState({
        name: workflow.name,
        description: workflow.description || "",
        flow,
      });
    }
  }, [workflow, isNew, resetState]);

  const handleFlowSave = useCallback(
    (updatedFlow: Workflow) => {
      updateField("flow", updatedFlow);
    },
    [updateField]
  );

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMessage(""); // Clear any previous errors

    if (!state.name.trim()) {
      setErrorMessage("Please enter a workflow name");
      return;
    }

    if (!state.flow) {
      setErrorMessage("Please create a workflow before saving");
      return;
    }

    // Validate flow
    const validation = validateFlow(state.flow as any);
    if (!validation.valid) {
      setErrorMessage("Flow validation errors:\n" + validation.errors.join("\n"));
      return;
    }

    // Convert flow to workflow format
    const workflowData: Partial<ApiWorkflow> = {
      name: state.name,
      description: state.description,
      status: "active",
      ...flowToWorkflow(state.flow as any, { name: state.name, description: state.description }),
    };

    try {
      let savedWorkflow: ApiWorkflow;
      if (isNew) {
        savedWorkflow = await createWorkflow(workflowData);
        // If it's a webhook workflow, show credentials dialog
        if (
          savedWorkflow.type === "webhook" &&
          savedWorkflow.webhook_url &&
          savedWorkflow.webhook_secret
        ) {
          setCreatedWorkflow(savedWorkflow);
          setWebhookCredentialsOpen(true);
        } else {
          navigate("/workflows");
        }
      } else {
        savedWorkflow = await updateWorkflow(id!, workflowData);
        // Refresh workflow data to get updated webhook info
        if (savedWorkflow.type === "webhook") {
          const updated = await getWorkflow(id!);
          setCreatedWorkflow(updated);
        }
        navigate("/workflows");
      }
    } catch (error) {
      console.error("Failed to save workflow:", error);
      setErrorMessage("Failed to save workflow. Please try again.");
    }
  };

  return (
    <div className="h-screen flex flex-col bg-background">
      {/* Error Alert */}
      {errorMessage && (
        <div className="flex-shrink-0 px-4 pt-2">
          <Alert variant="destructive" className="relative">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription className="whitespace-pre-line pr-8">
              {errorMessage}
            </AlertDescription>
            <button
              onClick={() => setErrorMessage("")}
              className="absolute right-4 top-4 rounded-sm opacity-70 hover:opacity-100"
            >
              <X className="h-4 w-4" />
            </button>
          </Alert>
        </div>
      )}

      {/* Header */}
      <div className="flex-shrink-0 border-b bg-background">
        <div className="flex items-center justify-between px-2 py-2">
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => navigate("/workflows")}
              className="px-2"
            >
              <ArrowLeft className="h-4 w-4 mr-1" />
              Back
            </Button>
            <div className="flex items-center gap-2">
              <div className="p-1.5 rounded-lg bg-primary/10">
                <WorkflowIcon className="h-4 w-4 text-primary" />
              </div>
              <Input
                placeholder="Workflow name *"
                value={state.name}
                onChange={(e) => updateField("name", e.target.value)}
                required
                className="h-8 border-0 shadow-none focus-visible:ring-0 px-2 font-semibold text-lg w-64"
              />
            </div>
          </div>

          <form onSubmit={handleSubmit} className="flex items-center gap-2">
            {/* Webhook Credentials Button - only show for existing webhook workflows */}
            {workflow &&
              workflow.type === "webhook" &&
              workflow.webhook_url &&
              workflow.webhook_secret && (
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  className="h-8"
                  onClick={() => setWebhookSheetOpen(true)}
                  title="View webhook credentials"
                >
                  <Webhook className="h-4 w-4 mr-2" />
                  Webhook
                </Button>
              )}
            <Button type="submit" size="sm" className="h-8">
              <Save className="h-4 w-4 mr-2" />
              {isNew ? "Create" : "Update"}
            </Button>
          </form>
        </div>
      </div>

      {/* Flow Editor - Full Height */}
      <div className="flex-1 overflow-hidden">
        <WorkflowFlowEditor
          initialFlow={state.flow || undefined}
          onSave={handleFlowSave}
        />
      </div>

      {/* Webhook Credentials Dialog (shown after creating a webhook workflow) */}
      <Dialog
        open={webhookCredentialsOpen}
        onOpenChange={setWebhookCredentialsOpen}
      >
        <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Webhook Workflow Created</DialogTitle>
            <DialogDescription>
              Your webhook workflow has been created. Use these credentials to
              configure your external service.
            </DialogDescription>
          </DialogHeader>
          {createdWorkflow &&
            createdWorkflow.webhook_url &&
            createdWorkflow.webhook_secret && (
              <WebhookCredentials
                webhookUrl={createdWorkflow.webhook_url}
                webhookSecret={createdWorkflow.webhook_secret}
                workflowId={createdWorkflow.id}
                className="mt-4"
              />
            )}
          <div className="flex justify-end gap-2 mt-6">
            <Button
              variant="outline"
              onClick={() => {
                setWebhookCredentialsOpen(false);
                navigate("/workflows");
              }}
            >
              Close
            </Button>
            <Button
              onClick={() => {
                setWebhookCredentialsOpen(false);
                navigate(`/workflows/${createdWorkflow?.id}/edit`);
              }}
            >
              View Workflow
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      {/* Webhook Credentials Sheet (for viewing credentials when editing) */}
      {workflow &&
        workflow.type === "webhook" &&
        workflow.webhook_url &&
        workflow.webhook_secret && (
          <Sheet open={webhookSheetOpen} onOpenChange={setWebhookSheetOpen}>
            <SheetContent className="w-full sm:max-w-2xl overflow-y-auto p-6">
              <SheetHeader>
                <SheetTitle>Webhook Credentials</SheetTitle>
                <SheetDescription>
                  Use these credentials to configure your external service to
                  send webhook requests to this workflow.
                </SheetDescription>
              </SheetHeader>
              <div className="mt-6">
                <WebhookCredentials
                  webhookUrl={workflow.webhook_url}
                  webhookSecret={workflow.webhook_secret}
                  workflowId={workflow.id}
                />
              </div>
            </SheetContent>
          </Sheet>
        )}
    </div>
  );
}
