import { useState, useEffect } from 'react';
import useSWR from 'swr';
import Editor from '@monaco-editor/react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import {
  Save,
  FileText,
  CheckCircle,
  AlertCircle,
  RefreshCw,
  History,
} from 'lucide-react';
import {
  getConfigs,
  createConfig,
  updateConfig,
  validateConfig,
  getConfigVersions,
  type Config,
  type ValidateConfigResponse,
} from '@/api/configs';
import { getGroups } from '@/api/groups';

const DEFAULT_CONFIG = `receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    timeout: 10s
    send_batch_size: 1024

exporters:
  otlp:
    endpoint: localhost:4317
    tls:
      insecure: true

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp]
    metrics:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp]
    logs:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp]
`;

export default function ConfigEditorPage() {
  const [selectedConfig, setSelectedConfig] = useState<Config | null>(null);
  const [editorContent, setEditorContent] = useState(DEFAULT_CONFIG);
  const [validation, setValidation] = useState<ValidateConfigResponse | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  const [isValidating, setIsValidating] = useState(false);
  const [showVersions, setShowVersions] = useState(false);
  const [selectedGroupId, setSelectedGroupId] = useState<string>('');

  const { data: configsData, mutate: mutateConfigs } = useSWR('configs', getConfigs);
  const { data: groupsData } = useSWR('groups', getGroups);
  const { data: versionsData, mutate: mutateVersions } = useSWR(
    selectedGroupId ? `config-versions-${selectedGroupId}` : null,
    () => selectedGroupId ? getConfigVersions({ group_id: selectedGroupId }) : null
  );

  // Load selected config into editor
  useEffect(() => {
    if (selectedConfig) {
      setEditorContent(selectedConfig.content);
    }
  }, [selectedConfig]);

  const handleValidate = async () => {
    setIsValidating(true);
    try {
      const result = await validateConfig({ content: editorContent });
      setValidation(result);
    } catch (error) {
      console.error('Validation failed:', error);
      setValidation({
        valid: false,
        errors: ['Failed to validate configuration'],
      });
    } finally {
      setIsValidating(false);
    }
  };

  const handleSave = async () => {
    setIsSaving(true);
    try {
      if (selectedConfig) {
        // Update existing config (creates new version)
        await updateConfig(selectedConfig.id, {
          content: editorContent,
          version: selectedConfig.version + 1,
        });
      } else {
        // Create new config
        await createConfig({
          group_id: selectedGroupId || undefined,
          config_hash: '', // Will be calculated by backend
          content: editorContent,
          version: 1,
        });
      }
      await mutateConfigs();
      await mutateVersions();
      alert('Configuration saved successfully!');
    } catch (error) {
      console.error('Save failed:', error);
      alert('Failed to save configuration');
    } finally {
      setIsSaving(false);
    }
  };

  const handleLoadVersion = async (config: Config) => {
    setSelectedConfig(config);
    setEditorContent(config.content);
    setShowVersions(false);
  };

  const configs = configsData?.configs || [];
  const groups = groupsData?.groups || [];
  const versions = versionsData?.versions || [];

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">Config Editor</h1>
          <p className="text-gray-600">Create and manage OpenTelemetry collector configurations</p>
        </div>
        <div className="flex gap-2">
          <Dialog open={showVersions} onOpenChange={setShowVersions}>
            <DialogTrigger asChild>
              <Button variant="outline">
                <History className="h-4 w-4 mr-2" />
                Version History
              </Button>
            </DialogTrigger>
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
                    onClick={() => handleLoadVersion(version)}
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
          <Button onClick={handleValidate} disabled={isValidating}>
            {isValidating ? (
              <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <CheckCircle className="h-4 w-4 mr-2" />
            )}
            Validate
          </Button>
          <Button onClick={handleSave} disabled={isSaving || !selectedGroupId}>
            {isSaving ? (
              <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <Save className="h-4 w-4 mr-2" />
            )}
            Save
          </Button>
        </div>
      </div>

      {/* Config Selection */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Select Configuration Target</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="text-sm font-medium mb-2 block">Group</label>
              <Select value={selectedGroupId} onValueChange={setSelectedGroupId}>
                <SelectTrigger>
                  <SelectValue placeholder="Select a group" />
                </SelectTrigger>
                <SelectContent>
                  {groups.map((group) => (
                    <SelectItem key={group.id} value={group.id}>
                      {group.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div>
              <label className="text-sm font-medium mb-2 block">Existing Config</label>
              <Select
                value={selectedConfig?.id || ''}
                onValueChange={(id) => {
                  const config = configs.find(c => c.id === id);
                  setSelectedConfig(config || null);
                }}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Start with new config" />
                </SelectTrigger>
                <SelectContent>
                  {configs.map((config) => (
                    <SelectItem key={config.id} value={config.id}>
                      Version {config.version} - {new Date(config.created_at).toLocaleDateString()}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          {selectedConfig && (
            <Alert>
              <FileText className="h-4 w-4" />
              <AlertDescription>
                Editing config version {selectedConfig.version} - Saving will create version {selectedConfig.version + 1}
              </AlertDescription>
            </Alert>
          )}
        </CardContent>
      </Card>

      {/* Validation Results */}
      {validation && (
        <Alert variant={validation.valid ? 'default' : 'destructive'}>
          {validation.valid ? (
            <CheckCircle className="h-4 w-4" />
          ) : (
            <AlertCircle className="h-4 w-4" />
          )}
          <AlertDescription>
            {validation.valid ? (
              <div>
                <div className="font-semibold">Configuration is valid!</div>
                {validation.warnings && validation.warnings.length > 0 && (
                  <div className="mt-2">
                    <div className="text-sm font-medium">Warnings:</div>
                    <ul className="list-disc list-inside text-sm">
                      {validation.warnings.map((warning, idx) => (
                        <li key={idx}>{warning}</li>
                      ))}
                    </ul>
                  </div>
                )}
              </div>
            ) : (
              <div>
                <div className="font-semibold">Configuration has errors:</div>
                <ul className="list-disc list-inside text-sm mt-2">
                  {validation.errors?.map((error, idx) => (
                    <li key={idx}>{error}</li>
                  ))}
                </ul>
              </div>
            )}
          </AlertDescription>
        </Alert>
      )}

      {/* Editor */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">YAML Configuration</CardTitle>
          <CardDescription>
            Edit your OpenTelemetry collector configuration in YAML format
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="border rounded-lg overflow-hidden">
            <Editor
              height="60vh"
              defaultLanguage="yaml"
              value={editorContent}
              onChange={(value) => setEditorContent(value || '')}
              theme="vs-light"
              options={{
                minimap: { enabled: false },
                fontSize: 13,
                lineNumbers: 'on',
                roundedSelection: false,
                scrollBeyondLastLine: false,
                readOnly: false,
                automaticLayout: true,
              }}
            />
          </div>
        </CardContent>
      </Card>

      {/* Quick Reference */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Quick Reference</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-3 gap-4 text-sm">
            <div>
              <h4 className="font-semibold mb-2">Receivers</h4>
              <ul className="space-y-1 text-gray-600">
                <li>• otlp</li>
                <li>• jaeger</li>
                <li>• prometheus</li>
                <li>• zipkin</li>
              </ul>
            </div>
            <div>
              <h4 className="font-semibold mb-2">Processors</h4>
              <ul className="space-y-1 text-gray-600">
                <li>• batch</li>
                <li>• memory_limiter</li>
                <li>• attributes</li>
                <li>• resource</li>
              </ul>
            </div>
            <div>
              <h4 className="font-semibold mb-2">Exporters</h4>
              <ul className="space-y-1 text-gray-600">
                <li>• otlp</li>
                <li>• logging</li>
                <li>• prometheus</li>
                <li>• jaeger</li>
              </ul>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
