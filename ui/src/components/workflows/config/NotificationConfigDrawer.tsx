import {
  Plus,
  Trash2,
  Mail,
  MessageSquare,
  Webhook,
  FileText,
} from "lucide-react";
import { useMemo, useState } from "react";

import type { NotificationNodeData } from "../types/flow-types";

import { ConfigDrawerLayout } from "./ConfigDrawerLayout";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { useFormState, useArrayField } from "@/hooks/useDrawerForm";

interface NotificationConfigDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  nodeData: NotificationNodeData | null;
  onSave: (data: NotificationNodeData) => void;
}

export function NotificationConfigDrawer({
  open,
  onOpenChange,
  nodeData,
  onSave,
}: NotificationConfigDrawerProps) {
  const initialState = useMemo(
    () => ({
      channel: (nodeData?.channel || "log") as
        | "email"
        | "slack"
        | "webhook"
        | "log",
      message: nodeData?.message || "",
      severity: (nodeData?.severity || "info") as
        | "info"
        | "warning"
        | "error"
        | "success",
      label: nodeData?.label || "",
      description: nodeData?.description || "",
    }),
    [nodeData],
  );

  const { state, updateField } = useFormState(initialState);
  const {
    items: recipients,
    addItem: addRecipient,
    removeItem: removeRecipient,
  } = useArrayField<string>(nodeData?.recipients || []);
  const [newRecipient, setNewRecipient] = useState<string>("");
  const [error, setError] = useState<string | null>(null);

  const handleAddRecipient = () => {
    if (newRecipient.trim()) {
      addRecipient(newRecipient.trim());
      setNewRecipient("");
    }
  };

  const handleRemoveRecipient = (index: number) => {
    removeRecipient(index);
  };

  const handleSave = () => {
    setError(null);

    if (!state.message.trim()) {
      setError("Message is required");
      return;
    }

    if (
      (state.channel === "email" || state.channel === "slack") &&
      recipients.length === 0
    ) {
      setError(
        "At least one recipient is required for email and slack channels",
      );
      return;
    }

    const updatedData: NotificationNodeData = {
      label: state.label || "Notification",
      description: state.description || undefined,
      channel: state.channel,
      message: state.message.trim(),
      severity: state.severity,
      recipients: recipients.length > 0 ? recipients : undefined,
    };

    onSave(updatedData);
  };

  return (
    <ConfigDrawerLayout
      open={open}
      onOpenChange={onOpenChange}
      title="Configure Notification"
      description="Configure notification settings for this workflow step"
      error={error}
      onSave={handleSave}
    >
      {/* Basic Info */}
      <div className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="label">Label</Label>
          <Input
            id="label"
            placeholder="Notification"
            value={state.label}
            onChange={(e) => updateField("label", e.target.value)}
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="description">Description (Optional)</Label>
          <Input
            id="description"
            placeholder="Brief description of this notification"
            value={state.description}
            onChange={(e) => updateField("description", e.target.value)}
          />
        </div>
      </div>

      {/* Channel Selection */}
      <div className="space-y-2">
        <Label htmlFor="channel">Channel</Label>
        <Select
          value={state.channel}
          onValueChange={(value: any) => updateField("channel", value)}
        >
          <SelectTrigger id="channel">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="log">
              <div className="flex items-center gap-2">
                <FileText className="h-4 w-4" />
                <span>Log</span>
              </div>
            </SelectItem>
            <SelectItem value="email">
              <div className="flex items-center gap-2">
                <Mail className="h-4 w-4" />
                <span>Email</span>
              </div>
            </SelectItem>
            <SelectItem value="slack">
              <div className="flex items-center gap-2">
                <MessageSquare className="h-4 w-4" />
                <span>Slack</span>
              </div>
            </SelectItem>
            <SelectItem value="webhook">
              <div className="flex items-center gap-2">
                <Webhook className="h-4 w-4" />
                <span>Webhook</span>
              </div>
            </SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Severity */}
      <div className="space-y-2">
        <Label htmlFor="severity">Severity</Label>
        <Select
          value={state.severity}
          onValueChange={(value: any) => updateField("severity", value)}
        >
          <SelectTrigger id="severity">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="info">Info</SelectItem>
            <SelectItem value="success">Success</SelectItem>
            <SelectItem value="warning">Warning</SelectItem>
            <SelectItem value="error">Error</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Message */}
      <div className="space-y-2">
        <Label htmlFor="message">Message *</Label>
        <Textarea
          id="message"
          placeholder="Enter notification message..."
          value={state.message}
          onChange={(e) => updateField("message", e.target.value)}
          rows={4}
          className="font-mono text-sm"
        />
        <p className="text-xs text-muted-foreground">
          You can use variables like {"${workflow.name}"} or {"${execution.id}"}
        </p>
      </div>

      {/* Recipients (for email and slack) */}
      {(state.channel === "email" || state.channel === "slack") && (
        <div className="space-y-2">
          <Label>Recipients *</Label>
          <div className="flex gap-2">
            <Input
              placeholder={
                state.channel === "email"
                  ? "email@example.com"
                  : "#channel or @user"
              }
              value={newRecipient}
              onChange={(e) => setNewRecipient(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  e.preventDefault();
                  handleAddRecipient();
                }
              }}
            />
            <Button type="button" onClick={handleAddRecipient} size="icon">
              <Plus className="h-4 w-4" />
            </Button>
          </div>
          {recipients.length > 0 && (
            <div className="space-y-1">
              {recipients.map((recipient, index) => (
                <div
                  key={index}
                  className="flex items-center justify-between bg-muted px-3 py-2 rounded-md"
                >
                  <span className="text-sm">{recipient}</span>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    onClick={() => handleRemoveRecipient(index)}
                    className="h-6 w-6"
                  >
                    <Trash2 className="h-3 w-3" />
                  </Button>
                </div>
              ))}
            </div>
          )}
          {recipients.length === 0 && (
            <p className="text-xs text-muted-foreground">
              Add at least one recipient
            </p>
          )}
        </div>
      )}

      {/* Webhook URL (for webhook channel) */}
      {state.channel === "webhook" && (
        <div className="space-y-2">
          <Label>Webhook URL</Label>
          <Input
            placeholder="https://example.com/webhook"
            value={newRecipient}
            onChange={(e) => setNewRecipient(e.target.value)}
          />
          <p className="text-xs text-muted-foreground">
            Webhook URL will be sent as a recipient
          </p>
        </div>
      )}
    </ConfigDrawerLayout>
  );
}
