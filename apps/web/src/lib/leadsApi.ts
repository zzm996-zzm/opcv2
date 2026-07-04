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

export type LeadEvidence = {
  type: string;
  title: string;
  url: string;
};

export type LeadResult = {
  id: number;
  task_id: number;
  name: string;
  phone?: string;
  email?: string;
  website?: string;
  evidence?: LeadEvidence[];
  created_at: string;
};

export type LeadTaskDetail = {
  task: LeadTask;
  progress_percent: number;
  message: string;
  results_count: number;
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
  },

  getTask(id: number) {
    return apiRequest<LeadTaskDetail>(`/api/v1/leads/tasks/${id}`, {
      method: "GET"
    });
  },

  listResults(taskId: number, limit = 20) {
    return apiRequest<{ results: LeadResult[] }>(`/api/v1/leads/tasks/${taskId}/results?limit=${limit}`, {
      method: "GET"
    });
  }
};
