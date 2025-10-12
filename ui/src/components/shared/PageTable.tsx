import type { LucideIcon } from "lucide-react";
import type { ReactNode } from "react";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

export interface PageTableColumn {
  header: string | ReactNode;
  key: string;
}

export interface PageTableAction {
  label: string;
  icon: LucideIcon;
  onClick: () => void;
  disabled?: boolean;
  variant?: "default" | "outline" | "ghost" | "destructive";
}

export interface PageTableEmptyState {
  icon: LucideIcon;
  title: string;
  description: string;
  action?: {
    label: string;
    onClick: () => void;
  };
}

export interface PageTableProps<T> {
  // Page header
  pageTitle: string;
  pageDescription: string;
  pageActions?: PageTableAction[];

  // Card/Table
  cardTitle: string;
  cardDescription: string;
  cardHeaderExtra?: ReactNode;
  columns: PageTableColumn[];
  data: T[];
  renderRow: (item: T) => ReactNode;
  emptyState: PageTableEmptyState;

  // Row interaction
  onRowClick?: (item: T) => void;
  getRowKey: (item: T) => string;
}

export function PageTable<T>({
  pageTitle,
  pageDescription,
  pageActions = [],
  cardTitle,
  cardDescription,
  cardHeaderExtra,
  columns,
  data,
  renderRow,
  emptyState,
  onRowClick,
  getRowKey,
}: PageTableProps<T>) {
  const EmptyIcon = emptyState.icon;

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* Page Header */}
      <div className="flex justify-between items-center pb-4 border-b border-border">
        <div>
          <h1 className="text-3xl font-bold">{pageTitle}</h1>
          <p className="text-muted-foreground mt-1">{pageDescription}</p>
        </div>
        {pageActions.length > 0 && (
          <div className="flex space-x-2">
            {pageActions.map((action, index) => {
              const Icon = action.icon;
              return (
                <Button
                  key={index}
                  onClick={action.onClick}
                  disabled={action.disabled}
                  variant={action.variant}
                >
                  <Icon
                    className={`h-4 w-4 mr-2 ${
                      action.disabled &&
                      action.label.toLowerCase().includes("refresh")
                        ? "animate-spin"
                        : ""
                    }`}
                  />
                  {action.label}
                </Button>
              );
            })}
          </div>
        )}
      </div>

      {/* Data Table Card */}
      <Card>
        <CardHeader>
          <div className="flex justify-between items-start">
            <div>
              <CardTitle>{cardTitle}</CardTitle>
              <CardDescription>{cardDescription}</CardDescription>
            </div>
            {cardHeaderExtra && <div>{cardHeaderExtra}</div>}
          </div>
        </CardHeader>
        <CardContent>
          {data.length === 0 ? (
            <div className="text-center py-8">
              <EmptyIcon className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-gray-900 mb-2">
                {emptyState.title}
              </h3>
              <p className="text-gray-600 mb-4">{emptyState.description}</p>
              {emptyState.action && (
                <Button onClick={emptyState.action.onClick}>
                  {emptyState.action.label}
                </Button>
              )}
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  {columns.map((column) => (
                    <TableHead key={column.key}>{column.header}</TableHead>
                  ))}
                </TableRow>
              </TableHeader>
              <TableBody>
                {data.map((item) => (
                  <TableRow
                    key={getRowKey(item)}
                    onClick={() => onRowClick?.(item)}
                    className={
                      onRowClick ? "cursor-pointer hover:bg-muted/50" : ""
                    }
                  >
                    {renderRow(item)}
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
