import { apiBaseUrl } from "../config";

// Common types for API responses
export interface ApiResponse<T = unknown> {
  success?: boolean;
  error?: string;
  data?: T;
}

// Base API configuration
export const apiConfig = {
  baseUrl: apiBaseUrl,
  defaultHeaders: {
    "Content-Type": "application/json",
  },
};

// Simple request function for OSS version (no authentication)
export const simpleRequest = async <T = unknown>(
  endpoint: string,
  options: RequestInit = {},
): Promise<T> => {
  const url = `${apiConfig.baseUrl}${endpoint}`;

  const defaultOptions: RequestInit = {
    headers: {
      ...apiConfig.defaultHeaders,
      ...options.headers,
    },
  };

  const response = await fetch(url, {
    ...defaultOptions,
    ...options,
  });

  if (!response.ok) {
    // Try to get detailed error message from response body
    let errorMessage = `API request failed: ${response.status} ${response.statusText}`;
    try {
      const errorData = await response.json();
      if (errorData.error) {
        errorMessage = errorData.error;
        if (errorData.details) {
          errorMessage += `: ${errorData.details}`;
        }
      }
    } catch (e) {
      // If we can't parse the error response, use the default message
    }
    const error = new Error(errorMessage);
    (error as any).status = response.status;
    throw error;
  }

  return response.json();
};

// HTTP method helpers
export const apiGet = <T = unknown>(
  endpoint: string,
  params?: Record<string, string>,
): Promise<T> => {
  const url = params ? `${endpoint}?${new URLSearchParams(params)}` : endpoint;
  return simpleRequest<T>(url, { method: "GET" });
};

export const apiPost = <T = unknown>(
  endpoint: string,
  data?: unknown,
): Promise<T> => {
  return simpleRequest<T>(endpoint, {
    method: "POST",
    body: data ? JSON.stringify(data) : undefined,
  });
};

export const apiPut = <T = unknown>(
  endpoint: string,
  data?: unknown,
): Promise<T> => {
  return simpleRequest<T>(endpoint, {
    method: "PUT",
    body: data ? JSON.stringify(data) : undefined,
  });
};

export const apiDelete = <T = unknown>(endpoint: string): Promise<T> => {
  return simpleRequest<T>(endpoint, { method: "DELETE" });
};

export const apiPatch = <T = unknown>(
  endpoint: string,
  data?: unknown,
): Promise<T> => {
  return simpleRequest<T>(endpoint, {
    method: "PATCH",
    body: data ? JSON.stringify(data) : undefined,
  });
};
