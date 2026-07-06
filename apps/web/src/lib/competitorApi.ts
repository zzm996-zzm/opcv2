import { apiRequest } from "./apiRequest";

export type CompetitorCard = {
  name: string;
  category: string;
  score: number;
  signal: string;
  risk: string;
  tags: string[];
};

export type CompetitorConclusion = {
  title: string;
  detail: string;
};

export type CompetitorEvidenceSource = {
  source_type: string;
  title: string;
  url: string;
  summary: string;
  captured_at: string;
};

export type CompetitorScan = {
  id: number;
  user_id: number;
  targets: string[];
  focus: string;
  status: "queued" | "running" | "succeeded" | "failed" | "completed";
  progress_percent: number;
  current_step: string;
  error_message?: string;
  competitors: CompetitorCard[];
  conclusions: CompetitorConclusion[];
  evidence_sources?: CompetitorEvidenceSource[];
  created_at: string;
  updated_at: string;
};

export type CompetitorWatchItem = {
  name: string;
  category: string;
  status: string;
  threat: string;
  last_seen_at: string;
  channels: string[];
  signal: string;
};

export type CompetitorEvent = {
  occurred_at: string;
  company: string;
  title: string;
  detail: string;
  level: string;
};

export const competitorApi = {
  createScan(input: { targets: string[]; focus: string }) {
    return apiRequest<CompetitorScan>("/api/v1/competitor/scans", {
      method: "POST",
      body: JSON.stringify(input)
    });
  },

  listScans(limit = 20) {
    return apiRequest<{ scans: CompetitorScan[] }>(`/api/v1/competitor/scans?limit=${limit}`, {
      method: "GET"
    });
  },

  getScan(id: number) {
    return apiRequest<CompetitorScan>(`/api/v1/competitor/scans/${id}`, {
      method: "GET"
    });
  },

  retryScan(id: number) {
    return apiRequest<CompetitorScan>(`/api/v1/competitor/scans/${id}/retry`, {
      method: "POST"
    });
  },

  getMonitoring(limit = 20) {
    return apiRequest<{ watchlist: CompetitorWatchItem[]; events: CompetitorEvent[] }>(`/api/v1/competitor/monitoring?limit=${limit}`, {
      method: "GET"
    });
  }
};
