import { RefreshCw } from "lucide-react";
import { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router-dom";
import useSWR from "swr";

import {
  getConfigs,
  getConfig,
  createConfig,
  updateConfig,
  validateConfig,
  getConfigVersions,
  type Config,
  type ValidateConfigResponse,
} from "@/api/configs";
import { getGroups } from "@/api/groups";
import {
  ConfigsList,
  ConfigEditorHeader,
  ConfigTarget,
  ConfigValidation,
  ConfigYamlEditor,
  ConfigVersionHistory,
} from "@/components/configs";
import { Button } from "@/components/ui/button";

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

type PageMode = "list" | "create" | "edit";

interface ConfigsPageProps {
  mode?: PageMode;
  configId?: string;
}

export default function ConfigsPage({
  mode: propMode,
  configId: propConfigId,
}: ConfigsPageProps = {}) {
  const navigate = useNavigate();
  const params = useParams<{ configId?: string; mode?: string }>();

  // Determine mode from props or URL params
  const mode: PageMode =
    propMode ||
    (params.mode === "new"
      ? "create"
      : params.configId || params.mode === "edit"
        ? "edit"
        : "list");
  const configId = propConfigId || params.configId;

  const [refreshing, setRefreshing] = useState(false);
  const [editorContent, setEditorContent] = useState(DEFAULT_CONFIG);
  const [validation, setValidation] = useState<ValidateConfigResponse | null>(
    null,
  );
  const [isSaving, setIsSaving] = useState(false);
  const [isValidating, setIsValidating] = useState(false);
  const [showVersions, setShowVersions] = useState(false);
  const [selectedGroupId, setSelectedGroupId] = useState<string>("");

  const {
    data: configsData,
    error: configsError,
    mutate: mutateConfigs,
  } = useSWR("configs", () => getConfigs({ limit: 100 }), {
    refreshInterval: 30000,
  });

  const { data: groupsData } = useSWR("groups", getGroups);

  const { data: currentConfigData } = useSWR(
    mode === "edit" && configId ? `config-${configId}` : null,
    () => (configId ? getConfig(configId) : null),
  );

  const { data: versionsData, mutate: mutateVersions } = useSWR(
    selectedGroupId ? `config-versions-${selectedGroupId}` : null,
    () =>
      selectedGroupId ? getConfigVersions({ group_id: selectedGroupId }) : null,
  );

  // Load config into editor when in edit mode
  useEffect(() => {
    if (mode === "edit" && currentConfigData) {
      setEditorContent(currentConfigData.content);
      setSelectedGroupId(currentConfigData.group_id || "");
    } else if (mode === "create") {
      setEditorContent(DEFAULT_CONFIG);
      setSelectedGroupId("");
    }
  }, [mode, currentConfigData]);

  const handleRefresh = async () => {
    setRefreshing(true);
    await mutateConfigs();
    setRefreshing(false);
  };

  const handleValidate = async () => {
    setIsValidating(true);
    try {
      const result = await validateConfig({ content: editorContent });
      setValidation(result);
    } catch (error) {
      console.error("Validation failed:", error);
      setValidation({
        valid: false,
        errors: ["Failed to validate configuration"],
      });
    } finally {
      setIsValidating(false);
    }
  };

  const handleSave = async () => {
    setIsSaving(true);
    try {
      if (mode === "edit" && currentConfigData) {
        // Update existing config (creates new version)
        await updateConfig(currentConfigData.id, {
          content: editorContent,
          version: currentConfigData.version + 1,
        });
      } else {
        // Create new config
        await createConfig({
          group_id: selectedGroupId || undefined,
          config_hash: `hash_${Date.now()}`,
          content: editorContent,
          version: 1,
        });
      }
      await mutateConfigs();
      await mutateVersions();
      navigate("/configs");
    } catch (error) {
      console.error("Save failed:", error);
      alert("Failed to save configuration");
    } finally {
      setIsSaving(false);
    }
  };

  const handleLoadVersion = async (config: Config) => {
    setEditorContent(config.content);
    setShowVersions(false);
  };

  const handleEditConfig = (config: Config) => {
    navigate(`/configs/${config.id}/edit`);
  };

  const handleCreateNew = () => {
    navigate("/configs/new");
  };

  const handleBackToList = () => {
    navigate("/configs");
  };

  const configs = configsData?.configs || [];
  const groups = groupsData?.groups || [];
  const versions = versionsData?.versions || [];

  // List View
  if (mode === "list") {
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

    return (
      <ConfigsList
        configs={configs}
        refreshing={refreshing}
        onRefresh={handleRefresh}
        onCreateNew={handleCreateNew}
        onEditConfig={handleEditConfig}
      />
    );
  }

  // Editor View (Create or Edit)
  return (
    <div className="container mx-auto p-6 space-y-6">
      <ConfigEditorHeader
        mode={mode as "create" | "edit"}
        isValidating={isValidating}
        isSaving={isSaving}
        canSave={!!selectedGroupId}
        onBack={handleBackToList}
        onShowVersions={() => setShowVersions(true)}
        onValidate={handleValidate}
        onSave={handleSave}
      />

      <ConfigTarget
        mode={mode as "create" | "edit"}
        selectedGroupId={selectedGroupId}
        groups={groups}
        currentVersion={currentConfigData?.version}
        onGroupChange={setSelectedGroupId}
      />

      {validation && <ConfigValidation validation={validation} />}

      <ConfigYamlEditor value={editorContent} onChange={setEditorContent} />

      <ConfigVersionHistory
        open={showVersions}
        versions={versions}
        onOpenChange={setShowVersions}
        onLoadVersion={handleLoadVersion}
      />
    </div>
  );
}
