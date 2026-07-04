import { apiRequest } from "./apiRequest";

export type GeoMetric = {
  key: string;
  label: string;
  value: string;
};

export type GeoEngineCoverage = {
  name: string;
  coverage_percent: number;
  status?: string;
};

export type GeoLeadSignal = {
  title: string;
  detail: string;
};

export type GeoKeywordOpportunity = {
  id: number;
  query: string;
  intent?: string;
  coverage?: string;
  score: number;
  action?: string;
};

export type GeoContentTask = {
  id: number;
  type?: string;
  title: string;
  priority?: string;
  due_at?: string;
};

export type GeoOverview = {
  stats: GeoMetric[];
  engines: GeoEngineCoverage[];
  lead_signals: GeoLeadSignal[];
  keywords: GeoKeywordOpportunity[];
  content_tasks: GeoContentTask[];
};

export type GeoAnalysisRequestInput = {
  target: string;
};

export type GeoAnalysisRequest = {
  id: number;
  user_id?: number;
  target: string;
  status: string;
  error_message?: string;
  created_at: string;
  updated_at?: string;
};

export const geoApi = {
  overview() {
    return apiRequest<GeoOverview>("/api/v1/geo/overview");
  },

  createAnalysisRequest(input: GeoAnalysisRequestInput) {
    return apiRequest<GeoAnalysisRequest>("/api/v1/geo/analysis-requests", {
      method: "POST",
      body: JSON.stringify(input)
    });
  },

  listAnalysisRequests(limit = 20) {
    return apiRequest<{ requests: GeoAnalysisRequest[] }>(`/api/v1/geo/analysis-requests?limit=${limit}`);
  },

  getAnalysisRequest(id: number) {
    return apiRequest<GeoAnalysisRequest>(`/api/v1/geo/analysis-requests/${id}`);
  }
};
