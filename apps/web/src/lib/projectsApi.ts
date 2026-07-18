import { apiRequest, apiStreamRequest } from "./apiRequest";

export type ProjectQuestion = {
  key: string;
  text: string;
  options: string[];
};

export type ProjectMatch = {
  rank: number;
  opportunity_slug?: string;
  title: string;
  score: number;
  tags: string[];
  budget: string;
  reasons: string[];
  risk: string;
};

export type ProjectMatchResult = {
  session_id: number;
  status: "needs_input" | "completed";
  questions?: ProjectQuestion[];
  projects?: ProjectMatch[];
};

export type ProjectMatchSession = {
  id: number;
  user_id: number;
  intent: string;
  answers?: { key: string; value: string }[];
  status: "needs_input" | "completed";
  questions?: ProjectQuestion[];
  result?: ProjectMatchResult;
  created_at: string;
  updated_at: string;
};

export type ProjectFavorite = {
  id: number;
  user_id: number;
  session_id: number;
  created_at?: string;
  session?: ProjectMatchSession;
};

export type ProjectOpportunity = {
  id: number;
  slug: string;
  title: string;
  summary: string;
  industry: string;
  tags: string[];
  budget_band: string;
  difficulty: string;
  resource_requirements: string[];
  sections?: ProjectOpportunitySection[];
  published_at?: string;
  updated_at?: string;
};
export type ProjectContentItem = {
  title?: string;
  value?: string;
  detail?: string;
  meta?: string;
  tone?: string;
  progress?: number;
  tags?: string[];
};
export type ProjectContentBlock = {
  type: string;
  title?: string;
  subtitle?: string;
  columns?: number;
  items?: ProjectContentItem[];
  series?: ProjectContentItem[];
};
export type ProjectOpportunitySection = {
  key?: string;
  title: string;
  body: string;
  items: string[];
  blocks?: ProjectContentBlock[];
};
export type ProjectCase = {
  id:number;
  slug:string;
  opportunity_id?:number;
  opportunity_slug?:string;
  opportunity_title?:string;
  industry?:string;
  title:string;
  summary:string;
  case_type:string;
  outcome:string;
  key_actions?:string[];
  lessons?:string[];
  pitfalls?:string[];
  source_title:string;
  source_url:string;
  captured_at:string;
};
export type ProjectComparison = { id:number; user_id:number; items:ProjectOpportunity[]; created_at:string };

export const projectsApi = {
  listOpportunities(filters: { query?: string; industry?: string } = {}) {
    const query = new URLSearchParams();
    if (filters.query) query.set("q", filters.query);
    if (filters.industry) query.set("industry", filters.industry);
    const suffix = query.toString();
    return apiRequest<{ opportunities: ProjectOpportunity[] }>(`/api/v1/projects/opportunities${suffix ? `?${suffix}` : ""}`, { method: "GET" });
  },

  getOpportunity(slug: string) {
    return apiRequest<ProjectOpportunity>(`/api/v1/projects/opportunities/${encodeURIComponent(slug)}`, { method: "GET" });
  },
  listCases(filters: { caseType?: string; industry?: string; opportunitySlug?: string } = {}) {
    const query = new URLSearchParams();
    if (filters.caseType) query.set("type", filters.caseType);
    if (filters.industry) query.set("industry", filters.industry);
    if (filters.opportunitySlug) query.set("opportunity_slug", filters.opportunitySlug);
    const suffix = query.toString() ? `?${query.toString()}` : "";
    return apiRequest<{ cases: ProjectCase[] }>(`/api/v1/projects/cases${suffix}`, { method: "GET" });
  },
  createMatch(input: { intent: string }) {
    return apiRequest<ProjectMatchResult>("/api/v1/projects/matches", {
      method: "POST",
      body: JSON.stringify(input)
    });
  },

  listMatches() {
    return apiRequest<{ matches: ProjectMatchSession[] }>("/api/v1/projects/matches", {
      method: "GET"
    });
  },

  getMatch(id: number) {
    return apiRequest<ProjectMatchSession>(`/api/v1/projects/matches/${id}`, {
      method: "GET"
    });
  },
  answerMatch(id: number, answers: { key:string; value:string }[]) {
    return apiRequest<ProjectMatchResult>(`/api/v1/projects/matches/${id}/answers`, { method:"POST", body:JSON.stringify({ answers }) });
  },
  createComparison(opportunitySlugs: string[]) { return apiRequest<ProjectComparison>("/api/v1/projects/comparisons", { method:"POST", body:JSON.stringify({ opportunity_slugs:opportunitySlugs }) }); },
  getComparison(id: number) { return apiRequest<ProjectComparison>(`/api/v1/projects/comparisons/${id}`, { method:"GET" }); },
  createExport(sourceType: "match" | "comparison", sourceId: number) { return apiRequest<{ id:number; status:string; download_url:string }>("/api/v1/projects/exports", { method:"POST", body:JSON.stringify({ source_type:sourceType, source_id:sourceId }) }); },
  async downloadExport(id: number) {
    const response = await apiStreamRequest(`/api/v1/projects/exports/${id}/download`, { method: "GET" });
    return response.blob();
  },

  favoriteMatch(id: number) {
    return apiRequest<ProjectFavorite>(`/api/v1/projects/matches/${id}/favorite`, {
      method: "POST"
    });
  },

  listFavorites() {
    return apiRequest<{ favorites: ProjectFavorite[] }>("/api/v1/projects/favorites", {
      method: "GET"
    });
  },

  unfavoriteMatch(id: number) {
    return apiRequest<void>(`/api/v1/projects/matches/${id}/favorite`, {
      method: "DELETE"
    });
  }
};
