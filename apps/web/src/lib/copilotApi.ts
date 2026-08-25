import { ApiRequestError, apiRequest, apiStreamRequest } from "./apiRequest";

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
    task_id?: number;
    current_view?: string;
    active_filters?: Record<string, string>;
    tool_preview?: CopilotToolPreview;
    tool_result?: {
      tool: "create_task" | "create_project" | "project_match" | string;
      status: string;
      entity_id: number;
      title: string;
      url: string;
      message: string;
    };
  };
  created_at: string;
};

export type CopilotToolPreview = {
  id: string;
  source_message_id: number;
  call: {
    tool: "create_task" | "create_project" | "project_match" | string;
    arguments: {
      title?: string;
      description?: string;
      priority?: string;
      tags?: string[];
      intent?: string;
    };
  };
  status: "pending" | "executing" | "confirmed" | "cancelled" | "expired" | string;
  expires_at: string;
  error?: string;
};

export type ToolConfirmationResult = {
  message: CopilotMessage;
  tool_result?: NonNullable<CopilotMessage["metadata"]>["tool_result"];
};

export type CopilotMemory = {
  id: number;
  user_id: number;
  key: string;
  value: string;
  confidence: number;
  source?: string;
  status?: "pending" | "active" | "inactive";
  created_at: string;
  updated_at: string;
};

export type CopilotFile = {
  id: number;
  user_id: number;
  name: string;
  mime_type: string;
  size_bytes: number;
  content?: string;
  status: "ready" | "failed" | string;
  source: "pasted" | "upload" | string;
  sha256?: string;
  extracted_chars: number;
  error_code?: string;
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

export type StreamMessageHandlers = {
  onUserMessage?: (message: CopilotMessage) => void;
  onDelta?: (delta: string) => void;
  onAssistantMessage?: (message: CopilotMessage) => void;
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

  sendMessage(threadId: number, input: { content: string; model?: string; reference_ids?: number[]; request_id?: string; task_id?: number; current_view?: string; active_filters?: Record<string, string> }, signal?: AbortSignal) {
    return apiRequest<SendMessageResult>(`/api/v1/copilot/threads/${threadId}/messages`, {
      method: "POST",
      signal,
      body: JSON.stringify(input)
    });
  },

  confirmTool(threadId: number, messageId: number, decision: "confirm" | "cancel") {
    return apiRequest<ToolConfirmationResult>(`/api/v1/copilot/threads/${threadId}/messages/${messageId}/tool-confirmation`, {
      method: "POST",
      body: JSON.stringify({ decision })
    });
  },

  async streamMessage(
    threadId: number,
    input: { content: string; model?: string; reference_ids?: number[]; request_id?: string; task_id?: number; current_view?: string; active_filters?: Record<string, string> },
    handlers: StreamMessageHandlers = {},
    signal?: AbortSignal
  ): Promise<SendMessageResult> {
    const response = await apiStreamRequest(`/api/v1/copilot/threads/${threadId}/messages/stream`, {
      method: "POST",
      signal,
      body: JSON.stringify(input)
    });
    if (!response.body) throw new ApiRequestError("streaming_not_supported", response.status);
    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = "";
    let userMessage: CopilotMessage | undefined;
    let assistantMessage: CopilotMessage | undefined;

    const processBlock = (block: string) => {
      let event = "message";
      const data: string[] = [];
      for (const line of block.split("\n")) {
        if (line.startsWith("event:")) event = line.slice(6).trim();
        if (line.startsWith("data:")) data.push(line.slice(5).trim());
      }
      if (data.length === 0) return;
      const payload = JSON.parse(data.join("\n")) as {
        error?: string;
        delta?: string;
        user_message?: CopilotMessage;
        assistant_message?: CopilotMessage;
      };
      if (event === "error") throw new ApiRequestError(payload.error ?? "stream_failed", response.status);
      if (event === "user_message" && payload.user_message) {
        userMessage = payload.user_message;
        handlers.onUserMessage?.(payload.user_message);
      }
      if (event === "delta" && payload.delta) handlers.onDelta?.(payload.delta);
      if (event === "assistant_message" && payload.assistant_message) {
        assistantMessage = payload.assistant_message;
        handlers.onAssistantMessage?.(payload.assistant_message);
      }
    };

    while (true) {
      const { done, value } = await reader.read();
      buffer += decoder.decode(value, { stream: !done });
      buffer = buffer.replace(/\r\n/g, "\n");
      let boundary = buffer.indexOf("\n\n");
      while (boundary >= 0) {
        processBlock(buffer.slice(0, boundary));
        buffer = buffer.slice(boundary + 2);
        boundary = buffer.indexOf("\n\n");
      }
      if (done) break;
    }
    if (buffer.trim()) processBlock(buffer.trim());
    if (!userMessage || !assistantMessage) throw new ApiRequestError("stream_failed", response.status);
    return { user_message: userMessage, assistant_message: assistantMessage };
  },

  compareMessages(threadId: number, input: { content: string; models?: string[]; request_id?: string }, signal?: AbortSignal) {
    return apiRequest<CompareMessagesResult>(`/api/v1/copilot/threads/${threadId}/compare`, {
      method: "POST",
      signal,
      body: JSON.stringify(input)
    });
  },

  summarizeComparison(threadId: number, input: { content: string; model?: string; answers: CompareAnswer[]; request_id?: string }) {
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

  updateMemory(id: number, input: { key?: string; value?: string; status?: CopilotMemory["status"] }) {
    return apiRequest<CopilotMemory>(`/api/v1/copilot/memories/${id}`, {
      method: "PATCH",
      body: JSON.stringify(input)
    });
  },

  deleteMemory(id: number) {
    return apiRequest<void>(`/api/v1/copilot/memories/${id}`, {
      method: "DELETE"
    });
  },

  listFiles(limit = 50) {
    return apiRequest<{ files: CopilotFile[] }>(`/api/v1/copilot/files?limit=${limit}`, {
      method: "GET"
    });
  },

  saveFile(input: { name: string; mime_type?: string; content: string }) {
    return apiRequest<CopilotFile>("/api/v1/copilot/files", {
      method: "POST",
      body: JSON.stringify(input)
    });
  },

  uploadFile(file: File) {
    const body = new FormData();
    body.append("file", file);
    return apiRequest<CopilotFile>("/api/v1/copilot/files/upload", {
      method: "POST",
      body
    });
  },

  getFile(id: number) {
    return apiRequest<CopilotFile>(`/api/v1/copilot/files/${id}`, { method: "GET" });
  },

  deleteFile(id: number) {
    return apiRequest<void>(`/api/v1/copilot/files/${id}`, { method: "DELETE" });
  }
};
