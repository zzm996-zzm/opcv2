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
  industry?: string;
  category?: string;
  track?: string;
  cover_url?: string;
  tags: string[];
  budget_band: string;
  difficulty: string;
  resource_requirements: string[];
  sections?: ProjectOpportunitySection[];
  detail?: { sections?: ProjectOpportunitySection[] } & Record<string, unknown>;
  heat?: number;
  is_featured?: boolean;
  is_favorited?: boolean;
  is_unlocked?: boolean;
  locked_blocks?: string[];
  source_url?: string;
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

export type ProjectHome = {
  hero: { title: string; subtitle: string; desc: string; image_url?: string };
  quick_tags: { code: string; name: string; filters: Record<string, unknown> }[];
  entries: { code: string; title: string; desc: string; image_url?: string; route: string; is_recommended: boolean }[];
  featured: ProjectOpportunity[];
};

export type ProjectPage = {
  items: ProjectOpportunity[];
  page: number;
  page_size: number;
  total: number;
};

export type EvidenceCaseItem = {
  id: number;
  title: string;
  cover_url?: string;
  result_summary: string;
  industry?: string;
  scale?: string;
  type: "success" | "fail";
  primary_source_url: string;
  source_count: number;
  verified_at?: string;
  published_at?: string;
  project_id?: number;
};

export type EvidenceSource = {
  id: number;
  title?: string;
  publisher?: string;
  url: string;
  published_at?: string;
  fetched_at: string;
  quality?: number;
  kind: string;
  is_primary: boolean;
  claim_fields: string[];
};

export type EvidenceCaseDetail = EvidenceCaseItem & {
  content_md: string;
  facts: { field: string; value: string; source_refs: number[] }[];
  analyses: { point: string; detail?: string; is_model_generated: boolean; source_refs: number[] }[];
  sources: EvidenceSource[];
};

export type EvidenceCasePage = {
  items: EvidenceCaseItem[];
  page: number;
  page_size: number;
  total: number;
};

export const projectsApi = {
  getHome() {
    return apiRequest<ProjectHome>("/api/v1/projects/home", { method: "GET" });
  },

  listProjects(filters: {
    keyword?: string;
    category?: string;
    track?: string;
    budget?: string;
    difficulty?: string;
    resource?: string;
    sort?: "latest" | "heat";
    isFeatured?: boolean;
    page?: number;
    pageSize?: number;
  } = {}) {
    const query = new URLSearchParams();
    if (filters.keyword) query.set("keyword", filters.keyword);
    if (filters.category) query.set("category", filters.category);
    if (filters.track) query.set("track", filters.track);
    if (filters.budget) query.set("budget", filters.budget);
    if (filters.difficulty) query.set("difficulty", filters.difficulty);
    if (filters.resource) query.set("resource", filters.resource);
    if (filters.sort) query.set("sort", filters.sort);
    if (filters.isFeatured !== undefined) query.set("is_featured", String(filters.isFeatured));
    if (filters.page) query.set("page", String(filters.page));
    if (filters.pageSize) query.set("page_size", String(filters.pageSize));
    const suffix = query.toString();
    return apiRequest<ProjectPage>(`/api/v1/projects${suffix ? `?${suffix}` : ""}`, { method: "GET" });
  },

  async getProject(ref: string) {
    const project = await apiRequest<ProjectOpportunity>(`/api/v1/projects/${encodeURIComponent(ref)}`, { method: "GET" });
    if (!project.sections && project.detail?.sections) project.sections = project.detail.sections;
    if (!project.industry) project.industry = project.track ?? project.category ?? "";
    return project;
  },

  listEvidenceCases(filters: { caseType?: string; industry?: string; scale?: string; page?: number; pageSize?: number } = {}) {
    const query = new URLSearchParams();
    if (filters.caseType) query.set("type", filters.caseType);
    if (filters.industry) query.set("industry", filters.industry);
    if (filters.scale) query.set("scale", filters.scale);
    if (filters.page) query.set("page", String(filters.page));
    if (filters.pageSize) query.set("page_size", String(filters.pageSize));
    const suffix = query.toString();
    return apiRequest<EvidenceCasePage>(`/api/v1/project-cases${suffix ? `?${suffix}` : ""}`, { method: "GET" });
  },

  getEvidenceCase(ref: string) {
    return apiRequest<EvidenceCaseDetail>(`/api/v1/project-cases/${encodeURIComponent(ref)}`, { method: "GET" });
  },

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
