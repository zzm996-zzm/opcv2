import { apiRequest } from "./apiRequest";

export type SandboxReport = {
  score: number;
  summary: string;
  basis?: "model_simulation";
  disclaimer?: string;
  assumptions?: string[];
  evidence_sources?: Array<{ title: string; url: string; captured_at: string }>;
  metrics: Array<{ label: string; value: string }>;
  role_summaries: Array<{ role: string; view: string }>;
  risks: string[];
  next_actions: string[];
};

export type SandboxSession = {
  id: number;
  user_id: number;
  goal: string;
  target_users: string;
  product: string;
  roles: string[];
  status: "draft" | "queued" | "running" | "completed" | "failed" | "canceled";
  progress_percent: number;
  current_step: string;
  error_message?: string;
  run_attempt: number;
  report?: SandboxReport;
  created_at: string;
  updated_at: string;
};

export type SandboxRole = {
  key: string;
  label: string;
  description: string;
  badge: string;
};

export type SandboxDraftUpdate = Partial<{
  goal: string;
  target_users: string;
  product: string;
  roles: string[];
}>;

export type SandboxMessage = {
  id: number;
  session_id: number;
  user_id: number;
  role: string;
  question: string;
  answer: string;
  created_at: string;
};

export const sandboxApi = {
  listRoles() {
    return apiRequest<{ roles: SandboxRole[] }>("/api/v1/sandbox/roles", { method: "GET" });
  },

  createSession(input: { goal: string; targetUsers: string; product: string; roles: string[] }) {
    return apiRequest<SandboxSession>("/api/v1/sandbox/sessions", {
      method: "POST",
      body: JSON.stringify({
        goal: input.goal,
        target_users: input.targetUsers,
        product: input.product,
        roles: input.roles
      })
    });
  },

  runSession(id: number) {
    return apiRequest<SandboxSession>(`/api/v1/sandbox/sessions/${id}/run`, {
      method: "POST"
    });
  },

  retrySession(id: number) {
    return apiRequest<SandboxSession>(`/api/v1/sandbox/sessions/${id}/retry`, { method: "POST" });
  },

  cancelSession(id: number) {
    return apiRequest<SandboxSession>(`/api/v1/sandbox/sessions/${id}/cancel`, { method: "POST" });
  },

  getStatus(id: number) {
    return apiRequest<SandboxSession>(`/api/v1/sandbox/sessions/${id}/status`, { method: "GET" });
  },

  updateDraft(id: number, input: SandboxDraftUpdate) {
    return apiRequest<SandboxSession>(`/api/v1/sandbox/sessions/${id}/draft`, {
      method: "PATCH",
      body: JSON.stringify(input)
    });
  },

  listMessages(id: number) {
    return apiRequest<{ messages: SandboxMessage[] }>(`/api/v1/sandbox/sessions/${id}/messages`, { method: "GET" });
  },

  askRole(id: number, input: { role: string; question: string }) {
    return apiRequest<SandboxMessage>(`/api/v1/sandbox/sessions/${id}/messages`, {
      method: "POST",
      body: JSON.stringify(input)
    });
  },

  listSessions(limit = 20) {
    return apiRequest<{ sessions: SandboxSession[] }>(`/api/v1/sandbox/sessions?limit=${limit}`, {
      method: "GET"
    });
  },

  getSession(id: number) {
    return apiRequest<SandboxSession>(`/api/v1/sandbox/sessions/${id}`, {
      method: "GET"
    });
  }
};
