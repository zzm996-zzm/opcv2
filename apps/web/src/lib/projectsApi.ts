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

export type ProjectMatchEvidence = {
  source_type: "knowledge_base" | "web";
  source_id?: string;
  url?: string;
  title: string;
  publisher?: string;
  excerpt: string;
  quality: number;
  untrusted_content: boolean;
};

export type ProjectMatchGeneration = {
  match_id: number;
  status: "queued" | "running" | "completed" | "partial" | "failed" | "canceled";
  attempt: number;
  progress_percent: number;
  current_step: string;
  result?: ProjectMatchResult & { evidence?: ProjectMatchEvidence[]; evidence_status?: string };
  error_code?: string;
};

export type ProjectClarificationQuestion = {
  id: string;
  field: string;
  type: string;
  question: string;
  options?: string[];
  required: boolean;
  reason?: string;
};

export type ProjectMatchWorkflow = {
  match_id: number;
  need: string;
  status: "clarifying" | "ready" | ProjectMatchGeneration["status"];
  analysis_summary?: string;
  parsed_profile?: Record<string, unknown>;
  completeness: number;
  missing_fields?: string[];
  questions?: ProjectClarificationQuestion[];
  assumptions?: string[];
  revision: number;
  file_ids?: number[];
  generation?: ProjectMatchGeneration;
  created_at?: string;
  updated_at?: string;
};

export type ProjectMatchProgressEvent = {
  id: number;
  event: string;
  data: Record<string, unknown>;
};

export type ProjectMatchFile = {
  id: number;
  match_id?: number;
  name: string;
  mime_type: string;
  detected_mime: string;
  size_bytes: number;
  sha256: string;
  parse_status: "uploading" | "scanning" | "parsing" | "ready" | "failed" | "deleted";
  extracted_text?: string;
  extracted_json?: Record<string, unknown>;
  error_code?: string;
  expires_at: string;
  created_at: string;
  updated_at: string;
};

export type ProjectMatchSession = {
  id: number;
  user_id: number;
  intent: string;
  answers?: { key: string; value: string }[];
  status: "needs_input" | "completed";
  questions?: ProjectQuestion[];
  result?: ProjectMatchResult;
  files?: ProjectMatchFile[];
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

export type ProjectPublicConfig = {
  feature_paywall_enabled: boolean;
};

export type ProjectCatalogFavorite = {
  project_id: number;
  slug: string;
  title: string;
  created_at?: string;
};

export type ProjectCompareItem = {
  project_id: number;
  slug: string;
  title: string;
  added_at?: string;
};

export type ProjectExport = {
  id: number;
  status: "queued" | "running" | "ready" | "failed";
  format: "pdf" | "json" | "link";
  download_url?: string;
  expires_at: string;
  error_code?: string;
};

export type UserProject = {
  id: number;
  user_id: number;
  name: string;
  description?: string;
  status: "draft" | "active" | "archived" | string;
  source_type?: string;
  created_at: string;
  updated_at: string;
};

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
  listUserProjects(limit = 50) {
    return apiRequest<{ projects: UserProject[] }>(`/api/v1/projects/user-projects?limit=${limit}`, { method: "GET" });
  },

  getPublicConfig() {
    return apiRequest<ProjectPublicConfig>("/api/v1/config", { method: "GET" });
  },

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

  listProjectFavorites() {
    return apiRequest<{ favorites: ProjectCatalogFavorite[] }>("/api/v1/projects/project-favorites", { method: "GET" });
  },

  favoriteProject(ref: string) {
    return apiRequest<ProjectCatalogFavorite>(`/api/v1/projects/${encodeURIComponent(ref)}/favorite`, { method: "POST" });
  },

  unfavoriteProject(ref: string) {
    return apiRequest<void>(`/api/v1/projects/${encodeURIComponent(ref)}/favorite`, { method: "DELETE" });
  },

  listProjectCompareItems() {
    return apiRequest<{ items: ProjectCompareItem[] }>("/api/v1/projects/compare", { method: "GET" });
  },

  addProjectCompareItem(ref: string) {
    return apiRequest<ProjectCompareItem>("/api/v1/projects/compare-items", {
      method: "POST",
      body: JSON.stringify({ project_id: ref })
    });
  },

  removeProjectCompareItem(ref: string) {
    return apiRequest<void>(`/api/v1/projects/compare-items/${encodeURIComponent(ref)}`, { method: "DELETE" });
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
  uploadProjectMatchFile(file: File) {
    const body = new FormData();
    body.append("file", file);
    return apiRequest<ProjectMatchFile>("/api/v1/project-match-files", { method: "POST", body });
  },

  getProjectMatchFile(id: number) {
    return apiRequest<ProjectMatchFile>(`/api/v1/project-match-files/${id}`, { method: "GET" });
  },

  listProjectMatchFiles(matchId?: number) {
    const suffix = matchId ? `?match_id=${matchId}` : "";
    return apiRequest<{ files: ProjectMatchFile[] }>(`/api/v1/project-match-files${suffix}`, { method: "GET" });
  },

  retryProjectMatchFile(id: number) {
    return apiRequest<ProjectMatchFile>(`/api/v1/project-match-files/${id}/retry`, { method: "POST" });
  },

  deleteProjectMatchFile(id: number) {
    return apiRequest<void>(`/api/v1/project-match-files/${id}`, { method: "DELETE" });
  },

  createMatch(input: { intent: string; file_ids?: number[] }) {
    return apiRequest<ProjectMatchResult>("/api/v1/projects/matches", {
      method: "POST",
      body: JSON.stringify(input)
    });
  },

  createProjectMatch(input: { need: string; file_ids?: number[] }, idempotencyKey: string) {
    return apiRequest<ProjectMatchWorkflow>("/api/v1/project-matches", {
      method: "POST",
      headers: { "Idempotency-Key": idempotencyKey },
      body: JSON.stringify(input)
    });
  },

  listProjectMatches(limit = 20) {
    return apiRequest<{ matches: ProjectMatchWorkflow[] }>(`/api/v1/project-matches?limit=${encodeURIComponent(String(limit))}`, { method: "GET" });
  },

  getProjectMatch(id: number) {
    return apiRequest<ProjectMatchWorkflow>(`/api/v1/project-matches/${id}`, { method: "GET" });
  },

  answerProjectMatch(id: number, input: { revision: number; answers: { question_id: string; field: string; value: string }[] }, idempotencyKey: string) {
    return apiRequest<ProjectMatchWorkflow>(`/api/v1/project-matches/${id}/answer`, {
      method: "POST",
      headers: { "Idempotency-Key": idempotencyKey },
      body: JSON.stringify(input)
    });
  },

  generateProjectMatch(id: number) {
    return apiRequest<ProjectMatchGeneration>(`/api/v1/project-matches/${id}/generate`, { method: "POST" });
  },

  cancelProjectMatch(id: number) {
    return apiRequest<ProjectMatchGeneration>(`/api/v1/project-matches/${id}/cancel`, { method: "POST" });
  },

  async streamProjectMatch(id: number, afterEventId: number, onEvent: (event: ProjectMatchProgressEvent) => void, signal?: AbortSignal) {
    const response = await apiStreamRequest(`/api/v1/project-matches/${id}/stream`, {
      method: "GET",
      signal,
      headers: afterEventId > 0 ? { "Last-Event-ID": String(afterEventId) } : undefined
    });
    if (!response.body) throw new Error("streaming_not_supported");
    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = "";
    const processBlock = (block: string) => {
      let event = "message";
      let eventId = 0;
      const data: string[] = [];
      for (const line of block.split("\n")) {
        if (line.startsWith("id:")) eventId = Number(line.slice(3).trim());
        if (line.startsWith("event:")) event = line.slice(6).trim();
        if (line.startsWith("data:")) data.push(line.slice(5).trim());
      }
      if (!Number.isFinite(eventId) || eventId <= 0) return;
      onEvent({ id: eventId, event, data: data.length ? JSON.parse(data.join("\n")) as Record<string, unknown> : {} });
    };
    while (true) {
      const { done, value } = await reader.read();
      buffer += decoder.decode(value, { stream: !done }).replace(/\r\n/g, "\n");
      let boundary = buffer.indexOf("\n\n");
      while (boundary >= 0) {
        processBlock(buffer.slice(0, boundary));
        buffer = buffer.slice(boundary + 2);
        boundary = buffer.indexOf("\n\n");
      }
      if (done) break;
    }
    if (buffer.trim()) processBlock(buffer.trim());
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
  createExport(sourceType: "match" | "comparison", sourceId: number) { return apiRequest<ProjectExport>("/api/v1/projects/exports", { method:"POST", body:JSON.stringify({ source_type:sourceType, source_id:sourceId, format:"pdf" }) }); },
  getExport(id: number) { return apiRequest<ProjectExport>(`/api/v1/projects/exports/${id}`, { method:"GET" }); },
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
