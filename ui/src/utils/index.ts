// Simple utility functions
export const formatDate = (date: string | Date): string => {
  return new Date(date).toLocaleString();
};

export const formatRelativeTime = (date: string | Date): string => {
  const now = new Date();
  const target = new Date(date);
  const diffInSeconds = Math.floor((now.getTime() - target.getTime()) / 1000);

  if (diffInSeconds < 60) return "Just now";
  if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)}m ago`;
  if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)}h ago`;
  return `${Math.floor(diffInSeconds / 86400)}d ago`;
};

export const truncateText = (text: string, maxLength: number): string => {
  if (text.length <= maxLength) return text;
  return text.slice(0, maxLength) + "...";
};

/**
 * Format a number into a human-readable count with appropriate suffix
 * Examples:
 *   999 -> "999"
 *   1000 -> "1K"
 *   1500 -> "1.5K"
 *   69335 -> "69.3K"
 *   1000000 -> "1M"
 *   1500000 -> "1.5M"
 *   1000000000 -> "1B"
 */
export const formatCount = (count: number): string => {
  if (count < 1000) {
    return count.toString();
  }

  if (count < 1000000) {
    const formatted = (count / 1000).toFixed(1);
    // Remove trailing .0 if present
    return formatted.endsWith(".0")
      ? formatted.slice(0, -2) + "K"
      : formatted + "K";
  }

  if (count < 1000000000) {
    const formatted = (count / 1000000).toFixed(1);
    // Remove trailing .0 if present
    return formatted.endsWith(".0")
      ? formatted.slice(0, -2) + "M"
      : formatted + "M";
  }

  const formatted = (count / 1000000000).toFixed(1);
  // Remove trailing .0 if present
  return formatted.endsWith(".0")
    ? formatted.slice(0, -2) + "B"
    : formatted + "B";
};
