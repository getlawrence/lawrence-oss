import Editor from "@monaco-editor/react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

interface ConfigYamlEditorProps {
  value: string;
  onChange: (value: string) => void;
}

export function ConfigYamlEditor({ value, onChange }: ConfigYamlEditorProps) {
  return (
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
            value={value}
            onChange={(value) => onChange(value || "")}
            theme="vs-light"
            options={{
              minimap: { enabled: false },
              fontSize: 13,
              lineNumbers: "on",
              roundedSelection: false,
              scrollBeyondLastLine: false,
              readOnly: false,
              automaticLayout: true,
            }}
          />
        </div>
      </CardContent>
    </Card>
  );
}

