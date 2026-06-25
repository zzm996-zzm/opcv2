import { authSession } from "./authSession";

export type LeadTask = {
  id: number;
  user_id: number;
  query: string;
  status: "queued" | "running" | "succeeded" | "failed" | "cancelled" | "refunded";
  idempotency_key: string;
  credit_cost: number;
  error_code?: string;
  created_at: string;
  updated_at: string;
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

export const leadsApi = {
  createTask(input: { query: string; idempotencyKey: string }) {
    return request<LeadTask>("/api/v1/leads/tasks", {
      method: "POST",
      body: JSON.stringify({
        query: input.query,
        idempotency_key: input.idempotencyKey
      })
    });
  },

  listTasks(limit = 20) {
    return request<{ tasks: LeadTask[] }>(`/api/v1/leads/tasks?limit=${limit}`, {
      method: "GET"
    });
  }
};
