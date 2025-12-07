import { Plus, RefreshCw, FileText, Hash, Edit } from "lucide-react";

import type { Config, PaginatedResponse } from "@/api/configs";
import { PageTable } from "@/components/shared/PageTable";
import { Pagination } from "@/components/shared/Pagination";
import { TruncatedId } from "@/components/shared/TruncatedId";
import { Button } from "@/components/ui/button";
import { TableCell } from "@/components/ui/table";

interface ConfigsListProps {
  data: PaginatedResponse<Config>;
  refreshing: boolean;
  onRefresh: () => void;
  onCreateNew: () => void;
  onEditConfig: (config: Config) => void;
  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
}

export function ConfigsList({
  data,
  refreshing,
  onRefresh,
  onCreateNew,
  onEditConfig,
  onPageChange,
  onPageSizeChange,
}: ConfigsListProps) {
  return (
    <PageTable
      pageTitle="Configurations"
      pageDescription="Manage agent and group configurations"
      pageActions={[
        {
          label: "Refresh",
          icon: RefreshCw,
          onClick: onRefresh,
          disabled: refreshing,
          variant: "ghost" as const,
        },
        {
          label: "Create Config",
          icon: Plus,
          onClick: onCreateNew,
          variant: "default" as const,
        },
      ]}
      cardTitle={`Configurations (${data.total_count})`}
      cardDescription="All agent and group configurations"
      cardFooter={
        <Pagination
          page={data.page}
          pageSize={data.page_size}
          totalCount={data.total_count}
          totalPages={data.total_pages}
          hasNext={data.has_next}
          hasPrev={data.has_prev}
          onPageChange={onPageChange}
          onPageSizeChange={onPageSizeChange}
        />
      }
      columns={[
        { header: "Name", key: "name" },
        { header: "ID", key: "id" },
        { header: "Agent/Group", key: "target" },
        { header: "Version", key: "version" },
        { header: "Hash", key: "hash" },
        { header: "Created", key: "created" },
        { header: "Actions", key: "actions" },
      ]}
      data={data.data}
      getRowKey={(config) => config.id}
      renderRow={(config) => (
        <>
          <TableCell className="max-w-[200px]">
            <span className="font-medium truncate block">
              {config.name && config.name.trim() !== ""
                ? config.name
                : "Unnamed Config"}
            </span>
          </TableCell>
          <TableCell>
            <TruncatedId id={config.id} maxLength={8} />
          </TableCell>
          <TableCell>
            {config.agent_id ? (
              <div className="flex items-center gap-1">
                <span className="text-blue-600 text-sm">Agent:</span>
                <TruncatedId
                  id={config.agent_id}
                  maxLength={8}
                  className="text-blue-600"
                />
              </div>
            ) : config.group_id ? (
              <div className="flex items-center gap-1">
                <span className="text-green-600 text-sm">Group:</span>
                <TruncatedId
                  id={config.group_id}
                  maxLength={8}
                  className="text-green-600"
                />
              </div>
            ) : (
              <span className="text-gray-500">Global</span>
            )}
          </TableCell>
          <TableCell>
            <span className="bg-blue-100 text-blue-800 px-2 py-1 rounded text-xs">
              v{config.version}
            </span>
          </TableCell>
          <TableCell>
            <div className="flex items-center gap-1">
              <Hash className="h-3 w-3 text-muted-foreground" />
              <TruncatedId
                id={config.config_hash}
                maxLength={8}
                showCopyButton={false}
              />
            </div>
          </TableCell>
          <TableCell>
            {new Date(config.created_at).toLocaleDateString()}
          </TableCell>
          <TableCell>
            <Button
              variant="ghost"
              size="sm"
              onClick={(e) => {
                e.stopPropagation();
                onEditConfig(config);
              }}
            >
              <Edit className="h-4 w-4" />
            </Button>
          </TableCell>
        </>
      )}
      emptyState={{
        icon: FileText,
        title: "No Configurations Found",
        description: "Create your first configuration to manage your agents.",
        action: {
          label: "Create Configuration",
          onClick: onCreateNew,
        },
      }}
    />
  );
}
