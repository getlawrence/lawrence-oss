import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { TIME_RANGE_OPTIONS, type TimeRange } from "@/types/timeRange";

interface TimeRangeSelectProps {
  /**
   * Current selected time range value
   */
  value: TimeRange;
  /**
   * Callback when time range changes
   */
  onValueChange: (value: TimeRange) => void;
  /**
   * Width class for the trigger button (default: "w-32")
   */
  className?: string;
  /**
   * Whether to use short labels (e.g., "Last 1m" instead of "1 minute")
   */
  useShortLabels?: boolean;
  /**
   * Limit available options (e.g., ["1m", "5m", "15m", "1h"] to exclude longer ranges)
   */
  maxRange?: TimeRange;
}

/**
 * Reusable time range selector component
 * 
 * @example
 * ```tsx
 * const [timeRange, setTimeRange] = useState<TimeRange>("5m");
 * 
 * <TimeRangeSelect
 *   value={timeRange}
 *   onValueChange={setTimeRange}
 * />
 * ```
 */
export function TimeRangeSelect({
  value,
  onValueChange,
  className = "w-32",
  useShortLabels = false,
  maxRange,
}: TimeRangeSelectProps) {
  // Filter options based on maxRange if provided
  const options = maxRange
    ? TIME_RANGE_OPTIONS.filter(
        (option) =>
          TIME_RANGE_OPTIONS.findIndex((o) => o.value === option.value) <=
          TIME_RANGE_OPTIONS.findIndex((o) => o.value === maxRange)
      )
    : TIME_RANGE_OPTIONS;

  return (
    <Select value={value} onValueChange={onValueChange}>
      <SelectTrigger className={className}>
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        {options.map((option) => (
          <SelectItem key={option.value} value={option.value}>
            {useShortLabels ? option.shortLabel : option.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}

