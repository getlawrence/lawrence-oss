import { CheckCircle, AlertCircle } from "lucide-react";

import type { ValidateConfigResponse } from "@/api/configs";
import { Alert, AlertDescription } from "@/components/ui/alert";

interface ConfigValidationProps {
  validation: ValidateConfigResponse;
}

export function ConfigValidation({ validation }: ConfigValidationProps) {
  return (
    <Alert variant={validation.valid ? "default" : "destructive"}>
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
  );
}
