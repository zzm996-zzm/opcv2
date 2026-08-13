import { apiRequest, apiStreamRequest } from "./apiRequest";

export type SandboxRunStatus = "draft" | "clarifying" | "ready" | "running" | "partial" | "done" | "failed" | "ai_no_result";
export type SandboxRoleStatus = "pending" | "queued" | "running" | "done" | "failed" | "cancelled";
export type SandboxStance = "support" | "neutral" | "oppose";

export type SandboxProduct = {
  name: string;
  price_cents?: number;
  selling_point?: string;
  cost?: string;
  stage?: string;
};

export type SandboxRunContext = {
  target_customer?: string;
  channel?: string;
  market?: string;
  extra?: string;
};

export type SandboxQuestion = {
  key: string;
  field: string;
  type: string;
  question: string;
  options?: string[];
  required: boolean;
  answer?: string;
  skipped?: boolean;
};

export type SandboxRole = {
  role_code: string;
  display_name: string;
  description: string;
  analysis_dimensions: string[];
  default_selected: boolean;
  is_required: boolean;
  default_model_route: string;
  prompt_version: string;
};

export type SandboxDimensionScore = {
  code: string;
  score: number;
  basis: string;
  confidence: number;
  evidence_refs: string[];
};

export type SandboxRoleRisk = { point: string; severity: string; basis: string };
export type SandboxRecommendation = { action: string; why: string };

export type SandboxRoleOutput = {
  role_code: string;
  stance: SandboxStance;
  verdict: string;
  content: string;
  dimension_scores: SandboxDimensionScore[];
  key_findings: string[];
  risks: SandboxRoleRisk[];
  recommendations: SandboxRecommendation[];
  questions_to_validate: string[];
  assumptions: string[];
  kill_criteria?: string[];
  is_model_generated: boolean;
};

export type SandboxRunRole = {
  run_id: number;
  role_code: string;
  seq: number;
  role_session_id: string;
  model_route: string;
  prompt_version: string;
  analysis_dimensions: string[];
  input_context_hash: string;
  status: SandboxRoleStatus;
  stance?: SandboxStance;
  output?: SandboxRoleOutput;
  input_tokens: number;
  output_tokens: number;
  latency_ms?: number;
  retry_count: number;
  error_code?: string;
  started_at?: string;
  finished_at?: string;
};

export type SandboxReport = {
  summary: string;
  feasibility: { score: number; level: string; basis: string };
  purchase_probability: { value_pct: number; basis: string; is_model_generated: boolean };
  opportunity: { point: string; reason: string }[];
  risk: { point: string; severity: string; reason: string; mitigation: string }[];
  advice: { action: string; why: string; priority: number; effort: string }[];
  role_takeaways: { role: string; stance: SandboxStance; key_points: string[]; dimension_scores: SandboxDimensionScore[] }[];
  dimension_summary: { dimension: string; score: number; consensus: string; supporting_roles: string[]; opposing_roles: string[] }[];
  disagreements: { topic: string; views: { role: string; point: string }[]; decision_needed: string }[];
  missing_roles: string[];
  scenarios: Record<string, { desc: string; condition: string }>;
  assumptions: string[];
  is_model_generated: boolean;
};

export type SandboxRun = {
  id: number;
  name?: string;
  product: SandboxProduct;
  context: SandboxRunContext;
  questions?: SandboxQuestion[];
  assumptions?: string[];
  roles: string[];
  orchestration_mode: "isolated_sessions";
  model_routing_snapshot?: Record<string, unknown>;
  evidence_pack?: Record<string, unknown>[];
  input_context_hash?: string;
  completeness: number;
  rounds: number;
  status: SandboxRunStatus;
  started_at?: string;
  finished_at?: string;
  revision: number;
  run_roles?: SandboxRunRole[];
  report?: SandboxReport;
  created_at: string;
  updated_at: string;
  next_questions?: SandboxQuestion[];
  done: boolean;
};

export type SandboxHome = {
  roles: SandboxRole[];
  recent_runs: SandboxRun[];
  capabilities: string[];
};

export type SandboxRunPage = {
  runs: SandboxRun[];
  page: number;
  limit: number;
};

export type SandboxProgressEvent = {
  id: number;
  run_id: number;
  role?: string;
  event: string;
  payload?: Record<string, unknown>;
  created_at?: string;
};

export type SandboxExport = {
  id: number;
  run_id: number;
  format: "pdf" | "json" | "link";
  download_url: string;
  expires_at: string;
  created_at: string;
};

export type SandboxFollowUp = {
  id: number;
  run_id: number;
  role_code: string;
  question: string;
  answer: string;
  created_at: string;
};

export type SandboxTaskHandoff = {
  tasks: { id: number; title: string; source_type: "sandbox_session"; source_id: string }[];
};

export type SandboxGrowthHandoff = {
  url: string;
  pricing?: number;
  channel?: string;
};

export type SandboxAnalyticsInput = {
  event: string;
  event_id: string;
  run_id?: number;
  properties?: Record<string, unknown>;
};

const sandboxVisitorKey = "opcv2:sandbox-analytics-visitor";
function analyticsID(prefix: string) { const value = typeof crypto !== "undefined" && typeof crypto.randomUUID === "function" ? crypto.randomUUID() : `${Date.now()}-${Math.random().toString(36).slice(2)}`; return `${prefix}-${value}`; }
function sandboxVisitor() { try { const stored = window.localStorage.getItem(sandboxVisitorKey); if (stored) return stored; const value = analyticsID("visitor"); window.localStorage.setItem(sandboxVisitorKey, value); return value; } catch { return analyticsID("visitor"); } }

export type CreateSandboxRunInput = {
  name?: string;
  product: SandboxProduct;
  context: SandboxRunContext;
};

export type AnswerSandboxRunInput = {
  revision: number;
  answers?: { key: string; value: string; skipped?: boolean }[];
  skip?: boolean;
};

export type SandboxRunFilters = {
  page?: number;
  limit?: number;
  status?: SandboxRunStatus;
  product?: string;
};

function runPath(id: number) {
  return `/api/v1/sandbox-runs/${id}`;
}

export const sandboxApi = {
  getHome() {
    return apiRequest<SandboxHome>("/api/v1/sandbox/home", { method: "GET" });
  },

  listRoles() {
    return apiRequest<{ roles: SandboxRole[] }>("/api/v1/sandbox-runs/roles", { method: "GET" });
  },

  createRun(input: CreateSandboxRunInput) {
    return apiRequest<SandboxRun>("/api/v1/sandbox-runs", { method: "POST", body: JSON.stringify(input) });
  },

  answerRun(id: number, input: AnswerSandboxRunInput) {
    return apiRequest<SandboxRun>(`${runPath(id)}/answer`, { method: "POST", body: JSON.stringify(input) });
  },

  setRoles(id: number, input: { revision: number; roles: string[] }) {
    return apiRequest<SandboxRun>(`${runPath(id)}/roles`, { method: "POST", body: JSON.stringify(input) });
  },

  startRun(id: number) {
    return apiRequest<SandboxRun>(`${runPath(id)}/start`, { method: "POST" });
  },

  stopRun(id: number) {
    return apiRequest<SandboxRun>(`${runPath(id)}/stop`, { method: "POST" });
  },

  retryRole(id: number, roleCode: string) {
    return apiRequest<SandboxRun>(`${runPath(id)}/roles/${encodeURIComponent(roleCode)}/retry`, { method: "POST" });
  },

  getRun(id: number) {
    return apiRequest<SandboxRun>(runPath(id), { method: "GET" });
  },

  listRuns(filters: SandboxRunFilters = {}) {
    const query = new URLSearchParams();
    if (filters.page) query.set("page", String(filters.page));
    if (filters.limit) query.set("limit", String(filters.limit));
    if (filters.status) query.set("status", filters.status);
    if (filters.product?.trim()) query.set("product", filters.product.trim());
    const suffix = query.size ? `?${query.toString()}` : "";
    return apiRequest<SandboxRunPage>(`/api/v1/sandbox-runs${suffix}`, { method: "GET" });
  },

  renameRun(id: number, input: { name: string; revision: number }) {
    return apiRequest<SandboxRun>(runPath(id), { method: "PATCH", body: JSON.stringify(input) });
  },

  deleteRun(id: number) {
    return apiRequest<void>(runPath(id), { method: "DELETE" });
  },

  getReport(id: number) {
    return apiRequest<SandboxReport>(`${runPath(id)}/report`, { method: "GET" });
  },

  generateReport(id: number) {
    return apiRequest<SandboxReport>(`${runPath(id)}/report`, { method: "POST" });
  },

  async streamRun(id: number, afterEventId: number, onEvent: (event: SandboxProgressEvent) => void, signal?: AbortSignal) {
    const response = await apiStreamRequest(`${runPath(id)}/stream`, {
      method: "GET",
      signal,
      headers: afterEventId > 0 ? { "Last-Event-ID": String(afterEventId) } : undefined
    });
    if (!response.body) throw new Error("streaming_not_supported");
    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = "";
    const processBlock = (block: string) => {
      let idValue = 0;
      let eventName = "message";
      const data: string[] = [];
      for (const line of block.split("\n")) {
        if (line.startsWith("id:")) idValue = Number(line.slice(3).trim());
        else if (line.startsWith("event:")) eventName = line.slice(6).trim();
        else if (line.startsWith("data:")) data.push(line.slice(5).trim());
      }
      if (!Number.isFinite(idValue) || idValue <= 0 || data.length === 0) return;
      const event = JSON.parse(data.join("\n")) as SandboxProgressEvent;
      onEvent({ ...event, id: idValue, event: event.event || eventName });
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

  createExport(id: number, format: SandboxExport["format"] = "pdf") {
    return apiRequest<SandboxExport>(`${runPath(id)}/report/export`, { method: "POST", body: JSON.stringify({ format }) });
  },

  async downloadExport(downloadUrl: string) {
    if (!/^\/api\/v1\/sandbox-runs\/\d+\/exports\/\d+\/download$/.test(downloadUrl)) throw new Error("invalid_export_url");
    const response = await apiStreamRequest(downloadUrl, { method: "GET" });
    return response.blob();
  },

  listFollowUps(id: number) {
    return apiRequest<{ follow_ups: SandboxFollowUp[] }>(`${runPath(id)}/follow-ups`, { method: "GET" });
  },

  askRole(id: number, input: { role_code: string; question: string }) {
    return apiRequest<SandboxFollowUp>(`${runPath(id)}/follow-ups`, { method: "POST", body: JSON.stringify(input) });
  },

  createTasks(id: number, input: { advice_indexes: number[] }) {
    return apiRequest<SandboxTaskHandoff>(`${runPath(id)}/report/tasks`, { method: "POST", body: JSON.stringify(input) });
  },

  createGrowthHandoff(id: number) {
    return apiRequest<SandboxGrowthHandoff>(`${runPath(id)}/report/growth-handoff`, { method: "POST" });
  },

  track(input: SandboxAnalyticsInput) {
    return apiRequest<void>("/api/v1/sandbox/analytics", { method: "POST", body: JSON.stringify({ ...input, event_id: input.event_id || analyticsID("event"), visitor_key: sandboxVisitor(), route: typeof window === "undefined" ? "/sandbox" : `${window.location.pathname}${window.location.search}` }) });
  }
};
