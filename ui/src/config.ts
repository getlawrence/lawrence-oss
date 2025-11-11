declare global {
  interface Window {
    __LAWRENCE_CONFIG__?: {
      backendUrl?: string;
    };
  }
}

const PLACEHOLDER = "__LAWRENCE_BACKEND_URL__";

const runtimeBackendUrl =
  typeof window !== "undefined"
    ? window.__LAWRENCE_CONFIG__?.backendUrl
    : undefined;

const sanitizedRuntimeUrl =
  runtimeBackendUrl && runtimeBackendUrl !== PLACEHOLDER
    ? runtimeBackendUrl
    : undefined;

const sameOriginUrl =
  typeof window !== "undefined" ? window.location.origin : undefined;

const buildTimeUrl =
  import.meta.env.VITE_BACKEND_URL &&
  import.meta.env.VITE_BACKEND_URL !== PLACEHOLDER
    ? import.meta.env.VITE_BACKEND_URL
    : undefined;

export const BACKEND_HOSTNAME =
  sanitizedRuntimeUrl ??
  sameOriginUrl ??
  buildTimeUrl ??
  "http://localhost:8080";

export const apiBaseUrl = `${BACKEND_HOSTNAME}/api/v1`;
