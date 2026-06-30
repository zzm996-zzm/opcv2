import { apiRequest } from "./apiRequest";

export type SandboxReport = {
  score: number;
  summary: string;
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
  status: "draft" | "completed";
  report?: SandboxReport;
  created_at: string;
  updated_at: string;
};

export const sandboxApi = {
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
