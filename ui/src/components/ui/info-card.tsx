import type { ReactNode } from "react";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export interface InfoItem {
  label: string;
  value: ReactNode;
}

interface InfoCardProps {
  /**
   * Card title
   */
  title: string;
  /**
   * Array of label-value pairs to display
   */
  items: InfoItem[];
  /**
   * Optional icon to display next to the title
   */
  icon?: ReactNode;
}

/**
 * Reusable card for displaying key-value information pairs
 *
 * @example
 * ```tsx
 * <InfoCard
 *   title="Agent Information"
 *   items={[
 *     { label: "ID", value: agent.id },
 *     { label: "Status", value: <Badge>{agent.status}</Badge> }
 *   ]}
 * />
 * ```
 */
export function InfoCard({ title, items, icon }: InfoCardProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg font-semibold flex items-center gap-2">
          {icon}
          {title}
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        {items.map((item, idx) => (
          <div key={idx} className="flex justify-between">
            <span className="text-sm text-gray-600">{item.label}:</span>
            <span className="text-sm">{item.value}</span>
          </div>
        ))}
      </CardContent>
    </Card>
  );
}
