import { useState, useCallback } from "react";
import useSWR from "swr";

import {
  executeLawrenceQL,
  validateQuery,
  getQuerySuggestions,
  getQueryTemplates,
  getQueryFunctions,
  type LawrenceQLRequest,
  type LawrenceQLResponse,
  type QueryTemplate,
  type FunctionInfo,
} from "../api/lawrence-ql";

export interface QueryHistoryItem {
  query: string;
  timestamp: Date;
  executionTime?: number;
  rowCount?: number;
}

/**
 * Hook for managing Lawrence QL queries
 */
export function useLawrenceQL() {
  const [query, setQuery] = useState("");
  const [isExecuting, setIsExecuting] = useState(false);
  const [results, setResults] = useState<LawrenceQLResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [history, setHistory] = useState<QueryHistoryItem[]>([]);

  // Execute query
  const execute = useCallback(
    async (customQuery?: string, options?: Partial<LawrenceQLRequest>) => {
      const queryToExecute = customQuery || query;
      if (!queryToExecute.trim()) {
        setError("Query cannot be empty");
        return;
      }

      setIsExecuting(true);
      setError(null);

      try {
        const request: LawrenceQLRequest = {
          query: queryToExecute,
          ...options,
        };

        const response = await executeLawrenceQL(request);
        setResults(response);

        // Add to history
        setHistory((prev) => [
          {
            query: queryToExecute,
            timestamp: new Date(),
            executionTime: response.meta.execution_time,
            rowCount: response.meta.row_count,
          },
          ...prev.slice(0, 49), // Keep last 50 queries
        ]);
      } catch (err) {
        setError(
          err instanceof Error ? err.message : "Failed to execute query",
        );
        setResults(null);
      } finally {
        setIsExecuting(false);
      }
    },
    [query],
  );

  // Validate query
  const validate = useCallback(
    async (queryToValidate?: string) => {
      const q = queryToValidate || query;
      if (!q.trim()) return { valid: false, error: "Query cannot be empty" };

      try {
        return await validateQuery(q);
      } catch (err) {
        return {
          valid: false,
          error: err instanceof Error ? err.message : "Validation failed",
        };
      }
    },
    [query],
  );

  // Get suggestions
  const getSuggestions = useCallback(
    async (cursorPos: number, customQuery?: string) => {
      const q = customQuery || query;
      try {
        const response = await getQuerySuggestions(q, cursorPos);
        return response.suggestions;
      } catch (err) {
        console.error("Failed to get suggestions:", err);
        return [];
      }
    },
    [query],
  );

  // Clear results
  const clearResults = useCallback(() => {
    setResults(null);
    setError(null);
  }, []);

  // Clear history
  const clearHistory = useCallback(() => {
    setHistory([]);
  }, []);

  return {
    query,
    setQuery,
    isExecuting,
    results,
    error,
    history,
    execute,
    validate,
    getSuggestions,
    clearResults,
    clearHistory,
  };
}

/**
 * Hook for loading query templates
 */
export function useQueryTemplates() {
  const { data, error, isLoading } = useSWR<{ templates: QueryTemplate[] }>(
    "/telemetry/query/templates",
    getQueryTemplates,
  );

  return {
    templates: data?.templates || [],
    error,
    isLoading,
  };
}

/**
 * Hook for loading available functions
 */
export function useQueryFunctions() {
  const { data, error, isLoading } = useSWR<{ functions: FunctionInfo[] }>(
    "/telemetry/query/functions",
    getQueryFunctions,
  );

  return {
    functions: data?.functions || [],
    error,
    isLoading,
  };
}

/**
 * Hook for managing saved queries
 */
export function useSavedQueries() {
  const [savedQueries, setSavedQueries] = useState<QueryTemplate[]>(() => {
    const saved = localStorage.getItem("lawrence-ql-saved-queries");
    return saved ? JSON.parse(saved) : [];
  });

  const saveQuery = useCallback(
    (query: string, name: string, description?: string) => {
      const newQuery: QueryTemplate = {
        id: Date.now().toString(),
        name,
        description: description || "",
        query,
        category: "saved",
      };

      const updated = [...savedQueries, newQuery];
      setSavedQueries(updated);
      localStorage.setItem(
        "lawrence-ql-saved-queries",
        JSON.stringify(updated),
      );
    },
    [savedQueries],
  );

  const deleteQuery = useCallback(
    (id: string) => {
      const updated = savedQueries.filter((q) => q.id !== id);
      setSavedQueries(updated);
      localStorage.setItem(
        "lawrence-ql-saved-queries",
        JSON.stringify(updated),
      );
    },
    [savedQueries],
  );

  return {
    savedQueries,
    saveQuery,
    deleteQuery,
  };
}
