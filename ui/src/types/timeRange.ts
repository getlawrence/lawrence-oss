/**
 * Time range options for metrics and analytics
 */
export type TimeRange = "1m" | "5m" | "15m" | "1h" | "6h" | "24h";

/**
 * Default time range value
 */
export const DEFAULT_TIME_RANGE: TimeRange = "5m";

/**
 * All available time range options
 */
export const TIME_RANGE_OPTIONS = [
  { value: "1m" as const, label: "1 minute", shortLabel: "Last 1m" },
  { value: "5m" as const, label: "5 minutes", shortLabel: "Last 5m" },
  { value: "15m" as const, label: "15 minutes", shortLabel: "Last 15m" },
  { value: "1h" as const, label: "1 hour", shortLabel: "Last 1h" },
  { value: "6h" as const, label: "6 hours", shortLabel: "Last 6h" },
  { value: "24h" as const, label: "24 hours", shortLabel: "Last 24h" },
] as const;
