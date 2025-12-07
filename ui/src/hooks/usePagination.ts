import { useState, useCallback } from "react";

export interface PaginationParams {
  page: number;
  page_size: number;
}

export interface PaginationMeta {
  page: number;
  page_size: number;
  total_count: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
}

export interface PaginatedResponse<T> {
  data: T[];
  page: number;
  page_size: number;
  total_count: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
}

export function usePagination(initialPageSize: number = 50) {
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(initialPageSize);

  const goToPage = useCallback((newPage: number) => {
    setPage(newPage);
  }, []);

  const nextPage = useCallback(() => {
    setPage((prev) => prev + 1);
  }, []);

  const prevPage = useCallback(() => {
    setPage((prev) => Math.max(1, prev - 1));
  }, []);

  const changePageSize = useCallback((newPageSize: number) => {
    setPageSize(newPageSize);
    setPage(1); // Reset to first page when changing page size
  }, []);

  const reset = useCallback(() => {
    setPage(1);
    setPageSize(initialPageSize);
  }, [initialPageSize]);

  return {
    page,
    pageSize,
    goToPage,
    nextPage,
    prevPage,
    changePageSize,
    reset,
    params: {
      page,
      page_size: pageSize,
    } as PaginationParams,
  };
}
