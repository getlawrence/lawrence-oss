import yaml from "js-yaml";
import { useState, useEffect, useCallback } from "react";

export interface YamlParseResult {
  valid: boolean;
  data: unknown | null;
  error: string | null;
}

interface UseYamlParserOptions {
  debounceMs?: number;
}

export function useYamlParser(
  yamlContent: string,
  options: UseYamlParserOptions = {},
) {
  const { debounceMs = 300 } = options;
  const [parseResult, setParseResult] = useState<YamlParseResult>({
    valid: true,
    data: null,
    error: null,
  });
  const [isParsing, setIsParsing] = useState(false);

  const parseYaml = useCallback((content: string) => {
    if (!content.trim()) {
      setParseResult({
        valid: true,
        data: null,
        error: null,
      });
      return;
    }

    try {
      const parsed = yaml.load(content);
      setParseResult({
        valid: true,
        data: parsed,
        error: null,
      });
    } catch (error) {
      setParseResult({
        valid: false,
        data: null,
        error: error instanceof Error ? error.message : "Failed to parse YAML",
      });
    } finally {
      setIsParsing(false);
    }
  }, []);

  useEffect(() => {
    setIsParsing(true);
    const timer = setTimeout(() => {
      parseYaml(yamlContent);
    }, debounceMs);

    return () => {
      clearTimeout(timer);
    };
  }, [yamlContent, debounceMs, parseYaml]);

  return { parseResult, isParsing };
}
