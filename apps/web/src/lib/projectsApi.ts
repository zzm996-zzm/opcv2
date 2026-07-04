import { apiRequest } from "./apiRequest";

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

export const projectsApi = {
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

  favoriteMatch(id: number) {
    return apiRequest<ProjectFavorite>(`/api/v1/projects/matches/${id}/favorite`, {
      method: "POST"
    });
  }
};
