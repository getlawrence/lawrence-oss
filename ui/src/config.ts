// OSS Configuration - simplified for local development
export const BACKEND_HOSTNAME =
  import.meta.env.VITE_BACKEND_URL ||
  (import.meta.env.MODE === 'production' ? 'http://localhost:8080' : 'http://localhost:8080');
export const apiBaseUrl = `${BACKEND_HOSTNAME}/api/v1`;
export const socketIOUrl = BACKEND_HOSTNAME;
export const API_BASE_URL = apiBaseUrl;

// Application URLs
export const APP_BASE_URL =
  import.meta.env.VITE_APP_BASE_URL ||
  (import.meta.env.MODE === 'production' ? 'http://localhost:5173' : 'http://localhost:5173');

// OSS version doesn't use Clerk authentication
export const CLERK_PUBLISHABLE_KEY = null;
