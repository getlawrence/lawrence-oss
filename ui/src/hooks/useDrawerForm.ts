import { useCallback, useState } from "react";

/**
 * Custom hook for managing drawer form state following React best practices.
 *
 * This hook eliminates the anti-pattern of copying props into state via useEffect.
 * Instead, it uses proper state initialization with a stable key to reset state.
 *
 * Key Benefits:
 * - No useEffect prop syncing
 * - Follows React's recommended patterns
 * - Proper state initialization
 * - Type-safe form handling
 *
 * Pattern:
 * 1. Use `defaultValues` as initial state (not props)
 * 2. Parent manages when to "reset" by changing the `key` prop
 * 3. Form manages its own state after initialization
 *
 * @example
 * ```tsx
 * // In parent component:
 * <FormDrawer
 *   key={nodeData?.id || 'new'} // Change key to reset form
 *   nodeData={nodeData}
 * />
 *
 * // In drawer component:
 * function FormDrawer({ nodeData }) {
 *   const [label, setLabel] = useState(nodeData?.label || "");
 *   const [value, setValue] = useState(nodeData?.value || 0);
 *   // State initialized once from props, then managed independently
 * }
 * ```
 */

/**
 * Hook to manage form validation errors
 */
export function useFormValidation<TErrors extends Record<string, string>>() {
  const [errors, setErrors] = useState<Partial<TErrors>>({});

  const setError = (field: keyof TErrors, message: string) => {
    setErrors((prev) => ({ ...prev, [field]: message }));
  };

  const clearError = (field: keyof TErrors) => {
    setErrors((prev) => {
      const next = { ...prev };
      delete next[field];
      return next;
    });
  };

  const clearAllErrors = () => {
    setErrors({});
  };

  const hasErrors = Object.keys(errors).length > 0;
  const getError = (field: keyof TErrors) => errors[field];

  return {
    errors,
    setError,
    clearError,
    clearAllErrors,
    hasErrors,
    getError,
  };
}

/**
 * Helper to generate stable keys for drawer forms based on data
 */
export function getDrawerKey<T extends { id?: string } | null>(
  data: T,
  fallback: string = "new",
): string {
  if (!data) return fallback;
  // Use id if available, otherwise use fallback
  return (data as any).id || fallback;
}

/**
 * Hook to manage array fields (like recipients, conditions, etc.)
 */
export function useArrayField<T>(initialValue: T[] = []) {
  const [items, setItems] = useState<T[]>(initialValue);

  const addItem = (item: T) => {
    setItems((prev) => [...prev, item]);
  };

  const removeItem = (index: number) => {
    setItems((prev) => prev.filter((_, i) => i !== index));
  };

  const updateItem = (index: number, item: T) => {
    setItems((prev) =>
      prev.map((existing, i) => (i === index ? item : existing)),
    );
  };

  const clearItems = () => {
    setItems([]);
  };

  return {
    items,
    setItems,
    addItem,
    removeItem,
    updateItem,
    clearItems,
    count: items.length,
    isEmpty: items.length === 0,
  };
}

/**
 * Type-safe form state manager using single state object
 * Use this instead of multiple useState calls
 */
export function useFormState<T extends Record<string, any>>(initialState: T) {
  const [state, setState] = useState<T>(initialState);

  const updateField = useCallback(
    <K extends keyof T>(field: K, value: T[K]) => {
      setState((prev) => ({ ...prev, [field]: value }));
    },
    [],
  );

  const updateFields = useCallback((updates: Partial<T>) => {
    setState((prev) => ({ ...prev, ...updates }));
  }, []);

  const resetState = useCallback(
    (newState?: T) => {
      setState(newState || initialState);
    },
    [initialState],
  );

  return {
    state,
    updateField,
    updateFields,
    resetState,
  };
}

/**
 * Helper to derive initial form state from data with defaults
 */
export function getInitialFormState<
  TData extends Record<string, any>,
  TDefaults extends Record<string, any>,
>(data: TData | null, defaults: TDefaults): TData & TDefaults {
  if (!data) return defaults as TData & TDefaults;
  return { ...defaults, ...data };
}
