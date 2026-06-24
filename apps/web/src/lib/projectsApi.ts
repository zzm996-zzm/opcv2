import { authSession } from "./authSession";

export type ProjectQuestion = {
  key: string;
  text: string;
  options: string[];
};

export type ProjectMatch = {
  rank: number;
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
};

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
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

export const projectsApi = {
  createMatch(input: { intent: string }) {
    return request<ProjectMatchResult>("/api/v1/projects/matches", {
      method: "POST",
      body: JSON.stringify(input)
    });
  },

  listMatches() {
    return request<{ matches: ProjectMatchSession[] }>("/api/v1/projects/matches", {
      method: "GET"
    });
  },

  favoriteMatch(id: number) {
    return request<ProjectFavorite>(`/api/v1/projects/matches/${id}/favorite`, {
      method: "POST"
    });
  }
};
