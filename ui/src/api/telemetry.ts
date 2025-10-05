import { apiGet } from './base';

export interface TelemetryOverview {
  metricsCount: number;
  logsCount: number;
  tracesCount: number;
  servicesCount: number;
}

// Get telemetry overview (placeholder)
export const getTelemetryOverview = (): Promise<TelemetryOverview> => {
  return apiGet<TelemetryOverview>('/telemetry/overview');
};

// Get services (placeholder)
export const getServices = (): Promise<string[]> => {
  return apiGet<string[]>('/telemetry/services');
};
