import { authSession } from "./authSession";

export type Question = {
  key: string;
  text: string;
  options: string[];
};

export type DirectionCard = {
  name: string;
  score: number;
  reasons: string[];
  market_evidence: string;
  difficulty: {
    level: string;
    notes: string[];
  };
  benchmarks: string[];
  actions: string[];
  upsell: string;
};

export type DirectionResult = {
  session_id: number;
  status: "needs_input" | "completed";
  questions?: Question[];
  cards?: DirectionCard[];
};

export type AnalysisSession = {
  id: number;
  user_id: number;
  mode: "direction" | string;
  intent: string;
  status: "needs_input" | "completed";
  questions?: Question[];
  result?: DirectionResult;
  created_at: string;
  updated_at: string;
};

export type AnalysisActionItem = {
  id: number;
  user_id: number;
  session_id: number;
  day_index: number;
  title: string;
  detail: string;
  completed: boolean;
  created_at: string;
  updated_at: string;
};

async function request<T>(path: string, init: RequestInit): Promise<T> {
  const token = authSession.get().accessToken;
  const response = await fetch(path, {
    ...init,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...init.headers
    }
  });
  if (!response.ok) {
    const payload = (await response.json().catch(() => ({}))) as { error?: string };
    throw new Error(payload.error ?? "request_failed");
  }
  return (await response.json()) as T;
}

export const analysisApi = {
  startDirection(input: { intent: string }) {
    return request<DirectionResult>("/api/v1/analysis/direction", {
      method: "POST",
      body: JSON.stringify(input)
    });
  },

  listSessions(limit = 20) {
    return request<{ sessions: AnalysisSession[] }>(`/api/v1/analysis/sessions?limit=${limit}`, {
      method: "GET"
    });
  },

  getSession(id: number) {
    return request<AnalysisSession>(`/api/v1/analysis/sessions/${id}`, {
      method: "GET"
    });
  },

  listActionItems(sessionId: number) {
    return request<{ items: AnalysisActionItem[] }>(`/api/v1/analysis/sessions/${sessionId}/action-items`, {
      method: "GET"
    });
  },

  updateActionItem(sessionId: number, itemId: number, completed: boolean) {
    return request<AnalysisActionItem>(`/api/v1/analysis/sessions/${sessionId}/action-items/${itemId}`, {
      method: "PATCH",
      body: JSON.stringify({ completed })
    });
  }
};
