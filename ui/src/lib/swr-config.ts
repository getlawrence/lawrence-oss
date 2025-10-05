import type { SWRConfiguration } from 'swr';

import { apiGet } from '@/api/base';

// Global SWR configuration
export const swrConfig: SWRConfiguration = {
  // Use our existing API wrapper as the fetcher
  fetcher: (url: string) => apiGet(url),

  // Revalidation settings - more conservative defaults
  revalidateOnFocus: false, // Disable focus revalidation by default
  revalidateOnReconnect: true, // Keep reconnect revalidation
  revalidateIfStale: false, // Disable stale revalidation by default

  // Data freshness settings
  // Note: staleTime and cacheTime are set per-hook, not globally

  // Error handling
  errorRetryCount: 3,
  errorRetryInterval: 5000,
  shouldRetryOnError: (error: unknown) => {
    // Don't retry on 4xx errors (client errors)
    if (error && typeof error === 'object' && 'status' in error) {
      const { status } = error as { status: number };
      if (status >= 400 && status < 500) {
        return false;
      }
    }
    return true;
  },

  // Performance optimizations
  dedupingInterval: 2000, // Dedupe requests within 2 seconds
  focusThrottleInterval: 5000, // Throttle focus revalidation

  // Global error handler
  onError: (_error: unknown, _key: string) => {
    // TODO: Add proper error reporting (e.g., Sentry)
    // console.error('SWR Error:', { key, error });
  },

  // Global success handler
  onSuccess: (_data: unknown, _key: string) => {
    // Optional: Add global success logging
    if (process.env.NODE_ENV === 'development') {
      // console.log('SWR Success:', { key, data });
    }
  },
};

// Specialized configurations for different data types
export const swrConfigs = {
  // For real-time data that needs frequent updates
  realtime: {
    ...swrConfig,
    revalidateOnFocus: true,
    // staleTime: 30 * 1000, // 30 seconds - set per hook
  },

  // For user data that changes less frequently
  user: {
    ...swrConfig,
    refreshInterval: 0, // No automatic refresh
    revalidateOnFocus: true,
    // staleTime: 10 * 60 * 1000, // 10 minutes - set per hook
  },

  // For static/semi-static data
  static: {
    ...swrConfig,
    refreshInterval: 0,
    revalidateOnFocus: false,
    revalidateOnReconnect: false,
    // staleTime: 30 * 60 * 1000, // 30 minutes - set per hook
  },

  // For settings/configuration data
  settings: {
    ...swrConfig,
    refreshInterval: 0,
    revalidateOnFocus: true,
    revalidateOnReconnect: true,
    // staleTime: 15 * 60 * 1000, // 15 minutes - set per hook
  },
} as const;

// Type for configuration keys
export type SWRConfigType = keyof typeof swrConfigs;
