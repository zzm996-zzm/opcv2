import { apiRequest } from "./apiRequest";

export type CopilotThread = {
  id: number;
  user_id: number;
  title: string;
  mode: "chat" | "compare";
  model?: string;
  archived_at?: string;
  created_at: string;
  updated_at: string;
};

export type CopilotMessage = {
  id: number;
  user_id: number;
  thread_id: number;
  role: "user" | "assistant" | "system";
  content: string;
  status: "completed" | "failed";
  model?: string;
  error_code?: string;
  input_tokens?: number;
  output_tokens?: number;
  metadata?: {
    kind?: "compare_question" | "compare_answer" | "compare_summary" | string;
  };
  created_at: string;
};

export type CopilotMemory = {
  id: number;
  user_id: number;
  key: string;
  value: string;
  confidence: number;
  source?: string;
  created_at: string;
  updated_at: string;
};

export type CopilotModelOption = {
  name: string;
  value: string;
  provider?: string;
  is_default: boolean;
};

export type ModelSmokeResult = {
  ok: boolean;
  model: string;
  reply: string;
  input_tokens?: number;
  output_tokens?: number;
};

export type CopilotAIRun = {
  id: number;
  user_id: number;
  feature: string;
  prompt_version: string;
  provider: string;
  model: string;
  status: "pending" | "completed" | "failed" | string;
  error_code?: string;
  error_message?: string;
  input_tokens?: number;
  output_tokens?: number;
  latency_ms?: number;
  created_at: string;
  updated_at: string;
};

export type SendMessageResult = {
  user_message: CopilotMessage;
  assistant_message: CopilotMessage;
};

export type CompareAnswer = {
  model: string;
  assistant_message: CopilotMessage;
  error_code?: string;
};

export type CompareMessagesResult = {
  user_message: CopilotMessage;
  answers: CompareAnswer[];
};

export type CompareSummaryResult = {
  summary_message: CopilotMessage;
};

export const copilotApi = {
  createThread(input: { title: string; mode?: "chat" | "compare"; model?: string }) {
    return apiRequest<CopilotThread>("/api/v1/copilot/threads", {
      method: "POST",
      body: JSON.stringify(input)
    });
  },

  listThreads(limit = 20) {
    return apiRequest<{ threads: CopilotThread[] }>(`/api/v1/copilot/threads?limit=${limit}`, {
      method: "GET"
    });
  },

  listModels() {
    return apiRequest<{ models: CopilotModelOption[] }>("/api/v1/copilot/models", {
      method: "GET"
    });
  },

  smokeModel(input: { model?: string; prompt?: string }) {
    return apiRequest<ModelSmokeResult>("/api/v1/copilot/models/smoke", {
      method: "POST",
      body: JSON.stringify(input)
    });
  },

  listAIRuns(limit = 20) {
    return apiRequest<{ runs: CopilotAIRun[] }>(`/api/v1/copilot/ai-runs?limit=${limit}`, {
      method: "GET"
    });
  },

  getThread(id: number) {
    return apiRequest<CopilotThread>(`/api/v1/copilot/threads/${id}`, {
      method: "GET"
    });
  },

  renameThread(id: number, title: string) {
    return apiRequest<CopilotThread>(`/api/v1/copilot/threads/${id}`, {
      method: "PATCH",
      body: JSON.stringify({ title })
    });
  },

  archiveThread(id: number) {
    return apiRequest<void>(`/api/v1/copilot/threads/${id}`, {
      method: "DELETE"
    });
  },

  listMessages(threadId: number, limit = 50) {
    return apiRequest<{ messages: CopilotMessage[] }>(`/api/v1/copilot/threads/${threadId}/messages?limit=${limit}`, {
      method: "GET"
    });
  },

  sendMessage(threadId: number, input: { content: string; model?: string }) {
    return apiRequest<SendMessageResult>(`/api/v1/copilot/threads/${threadId}/messages`, {
      method: "POST",
      body: JSON.stringify(input)
    });
  },

  compareMessages(threadId: number, input: { content: string; models?: string[] }) {
    return apiRequest<CompareMessagesResult>(`/api/v1/copilot/threads/${threadId}/compare`, {
      method: "POST",
      body: JSON.stringify(input)
    });
  },

  summarizeComparison(threadId: number, input: { content: string; model?: string; answers: CompareAnswer[] }) {
    return apiRequest<CompareSummaryResult>(`/api/v1/copilot/threads/${threadId}/compare/summary`, {
      method: "POST",
      body: JSON.stringify(input)
    });
  },

  listMemories(limit = 50) {
    return apiRequest<{ memories: CopilotMemory[] }>(`/api/v1/copilot/memories?limit=${limit}`, {
      method: "GET"
    });
  },

  saveMemory(input: { key: string; value: string; confidence?: number; source?: string }) {
    return apiRequest<CopilotMemory>("/api/v1/copilot/memories", {
      method: "POST",
      body: JSON.stringify(input)
    });
  },

  deleteMemory(id: number) {
    return apiRequest<void>(`/api/v1/copilot/memories/${id}`, {
      method: "DELETE"
    });
  }
};
