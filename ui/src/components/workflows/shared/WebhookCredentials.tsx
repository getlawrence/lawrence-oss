import { Check, Copy, Eye, EyeOff, ExternalLink } from "lucide-react";
import { useState } from "react";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { apiBaseUrl } from "@/config";

interface WebhookCredentialsProps {
  webhookUrl: string;
  webhookSecret: string;
  workflowId?: string;
  className?: string;
}

export function WebhookCredentials({
  webhookUrl,
  webhookSecret,
  className,
}: WebhookCredentialsProps) {
  const [urlCopied, setUrlCopied] = useState(false);
  const [secretCopied, setSecretCopied] = useState(false);
  const [secretVisible, setSecretVisible] = useState(false);

  // Construct full webhook URL
  const fullWebhookUrl = webhookUrl.startsWith("http")
    ? webhookUrl
    : `${apiBaseUrl.replace("/api/v1", "")}${webhookUrl}`;

  const handleCopyUrl = async () => {
    try {
      await navigator.clipboard.writeText(fullWebhookUrl);
      setUrlCopied(true);
      setTimeout(() => setUrlCopied(false), 2000);
    } catch (err) {
      console.error("Failed to copy URL:", err);
    }
  };

  const handleCopySecret = async () => {
    try {
      await navigator.clipboard.writeText(webhookSecret);
      setSecretCopied(true);
      setTimeout(() => setSecretCopied(false), 2000);
    } catch (err) {
      console.error("Failed to copy secret:", err);
    }
  };

  const maskedSecret = secretVisible
    ? webhookSecret
    : "•".repeat(Math.min(webhookSecret.length, 32));

  return (
    <div className={className}>
      <Alert className="mb-4">
        <AlertDescription className="text-sm">
          Use these credentials to configure your external service to send
          webhook requests to this workflow.
        </AlertDescription>
      </Alert>

      <div className="space-y-4">
        {/* Webhook URL */}
        <div className="space-y-2">
          <Label className="text-sm font-medium">Webhook URL</Label>
          <div className="flex gap-2">
            <Input
              value={fullWebhookUrl}
              readOnly
              className="h-9 text-sm font-mono bg-muted"
            />
            <Button
              variant="outline"
              size="icon"
              className="h-9 w-9 shrink-0"
              onClick={handleCopyUrl}
              title="Copy webhook URL"
            >
              {urlCopied ? (
                <Check className="h-4 w-4 text-green-600" />
              ) : (
                <Copy className="h-4 w-4" />
              )}
            </Button>
            <Button
              variant="outline"
              size="icon"
              className="h-9 w-9 shrink-0"
              onClick={() => window.open(fullWebhookUrl, "_blank")}
              title="Open in new tab"
            >
              <ExternalLink className="h-4 w-4" />
            </Button>
          </div>
          <p className="text-xs text-muted-foreground">
            Send POST requests to this URL to trigger the workflow
          </p>
        </div>

        {/* Webhook Secret */}
        <div className="space-y-2">
          <Label className="text-sm font-medium">Webhook Secret</Label>
          <div className="flex gap-2">
            <div className="relative flex-1">
              <Input
                value={maskedSecret}
                readOnly
                className="h-9 text-sm font-mono bg-muted pr-10"
              />
              <Button
                variant="ghost"
                size="icon"
                className="absolute right-0 top-0 h-9 w-9"
                onClick={() => setSecretVisible(!secretVisible)}
                title={secretVisible ? "Hide secret" : "Show secret"}
              >
                {secretVisible ? (
                  <EyeOff className="h-4 w-4" />
                ) : (
                  <Eye className="h-4 w-4" />
                )}
              </Button>
            </div>
            <Button
              variant="outline"
              size="icon"
              className="h-9 w-9 shrink-0"
              onClick={handleCopySecret}
              title="Copy webhook secret"
            >
              {secretCopied ? (
                <Check className="h-4 w-4 text-green-600" />
              ) : (
                <Copy className="h-4 w-4" />
              )}
            </Button>
          </div>
          <p className="text-xs text-muted-foreground">
            Include this secret in the{" "}
            <code className="bg-muted px-1 py-0.5 rounded text-[10px]">
              X-Webhook-Secret
            </code>{" "}
            header
          </p>
        </div>

        {/* Example cURL */}
        <div className="space-y-2">
          <Label className="text-sm font-medium">Example Request</Label>
          <div className="bg-muted p-3 rounded-md">
            <pre className="text-xs font-mono overflow-x-auto">
              <code>{`curl -X POST ${fullWebhookUrl} \\
  -H "Content-Type: application/json" \\
  -H "X-Webhook-Secret: ${secretVisible ? webhookSecret : maskedSecret}" \\
  -d '{"key": "value"}'`}</code>
            </pre>
          </div>
        </div>
      </div>
    </div>
  );
}
