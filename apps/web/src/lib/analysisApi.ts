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
  }
};
