import { apiRequest } from "./apiRequest";

export type DashboardSummary = {
  metrics: Array<{ label: string; value: string; change: string }>;
  projects: Array<{ name: string; value: string; leads: string; stage: string }>;
  trend: Array<{ label: string; value: number }>;
  pipeline: Array<{ stage: string; count: string; percent: string }>;
  alerts: Array<{ title: string; detail: string }>;
  actions: Array<{ time: string; title: string }>;
};

export const dashboardApi = {
  getSummary() {
    return apiRequest<DashboardSummary>("/api/v1/dashboard/summary", {
      method: "GET"
    });
  }
};
