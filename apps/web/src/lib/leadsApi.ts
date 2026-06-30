import { apiRequest } from "./apiRequest";

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

export const leadsApi = {
  createTask(input: { query: string; idempotencyKey: string }) {
    return apiRequest<LeadTask>("/api/v1/leads/tasks", {
      method: "POST",
      body: JSON.stringify({
        query: input.query,
        idempotency_key: input.idempotencyKey
      })
    });
  },

  listTasks(limit = 20) {
    return apiRequest<{ tasks: LeadTask[] }>(`/api/v1/leads/tasks?limit=${limit}`, {
      method: "GET"
    });
  }
};
