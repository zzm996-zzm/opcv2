import { apiRequest } from "./apiRequest";

export type SandboxReportEvidence = {
  title: string;
  url: string;
  captured_at: string;
};

export type SandboxReportMetric = {
  label: string;
  value: string;
};

export type SandboxRoleSummary = {
  role: string;
  view: string;
};

export type SandboxReportInsight = {
  title: string;
  detail: string;
  tags: string[];
};

export type SandboxActionPlanItem = {
  order: number;
  title: string;
  detail: string;
  duration: string;
};

export type SandboxGrowthPathItem = {
  stage: number;
  title: string;
  detail: string;
};

export type SandboxValidationMetric = {
  label: string;
  current: string;
  target: string;
  confidence_percent: number;
};

export type SandboxTimelineItem = {
  title: string;
  period: string;
};

export type SandboxReport = {
  score: number;
  summary: string;
  basis?: "model_simulation";
  disclaimer?: string;
  assumptions?: string[];
  evidence_sources?: SandboxReportEvidence[];
  metrics: SandboxReportMetric[];
  role_summaries: SandboxRoleSummary[];
  risks: string[];
  next_actions: string[];
  report_version?: string;
  consumer_probability?: number;
  risk_level?: string;
  recommendation_grade?: string;
  core_conclusions?: string[];
  opportunity_analysis?: SandboxReportInsight[];
  risk_analysis?: SandboxReportInsight[];
  action_plan?: SandboxActionPlanItem[];
  growth_path?: SandboxGrowthPathItem[];
  validation_metrics?: SandboxValidationMetric[];
  timeline?: SandboxTimelineItem[];
};

export type SandboxRecognizedField = {
  key: string;
  label: string;
  value: string;
};

export type SandboxIntakeQuestion = {
  key: string;
  title: string;
  hint: string;
  placeholder: string;
  required: boolean;
  max_length: number;
  position: number;
  answer?: string;
  skipped: boolean;
};

export type SandboxIntake = {
  status: "questions" | "ready";
  initial_idea: string;
  recognized_fields: SandboxRecognizedField[];
  questions: SandboxIntakeQuestion[];
  answered_count: number;
  total_questions: number;
};

export type SandboxRunSettings = {
  depth: "standard" | "deep";
  output_style: "structured_report" | "concise_report";
  generate_outline: boolean;
  variables: Record<string, string>;
};

export type SandboxSettingOption = {
  value: string;
  label: string;
  description: string;
  recommended?: boolean;
};

export type SandboxOptions = {
  roles: SandboxRole[];
  system_perspectives: SandboxRole[];
  depths: SandboxSettingOption[];
  output_styles: SandboxSettingOption[];
  defaults: SandboxRunSettings;
};

export type SandboxSession = {
  id: number;
  user_id: number;
  goal: string;
  target_users: string;
  product: string;
  roles: string[];
  status: "draft" | "queued" | "running" | "completed" | "failed" | "canceled";
  progress_percent: number;
  current_step: string;
  error_message?: string;
  run_attempt: number;
  intake: SandboxIntake;
  settings: SandboxRunSettings;
  report?: SandboxReport;
  created_at: string;
  updated_at: string;
  is_example?: boolean;
  example_key?: string;
};

export type SandboxRole = {
  key: string;
  label: string;
  description: string;
  badge: string;
};

export type SandboxDraftUpdate = Partial<{
  goal: string;
  target_users: string;
  product: string;
  roles: string[];
  settings: SandboxRunSettings;
}>;

export type SandboxMessage = {
  id: number;
  session_id: number;
  user_id: number;
  role: string;
  question: string;
  answer: string;
  created_at: string;
};

export const sandboxApi = {
  getOptions() {
    return apiRequest<SandboxOptions>("/api/v1/sandbox/options", { method: "GET" });
  },

  listRoles() {
    return apiRequest<{ roles: SandboxRole[] }>("/api/v1/sandbox/roles", { method: "GET" });
  },

  createIntake(initialIdea: string) {
    return apiRequest<SandboxSession>("/api/v1/sandbox/sessions/intake", {
      method: "POST",
      body: JSON.stringify({ initial_idea: initialIdea })
    });
  },

  answerIntakeQuestion(id: number, questionKey: string, input: { answer: string; skipped: boolean }) {
    return apiRequest<SandboxSession>(`/api/v1/sandbox/sessions/${id}/intake/questions/${encodeURIComponent(questionKey)}`, {
      method: "PUT",
      body: JSON.stringify(input)
    });
  },

  completeIntake(id: number) {
    return apiRequest<SandboxSession>(`/api/v1/sandbox/sessions/${id}/intake/complete`, {
      method: "POST"
    });
  },

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

  retrySession(id: number) {
    return apiRequest<SandboxSession>(`/api/v1/sandbox/sessions/${id}/retry`, { method: "POST" });
  },

  cancelSession(id: number) {
    return apiRequest<SandboxSession>(`/api/v1/sandbox/sessions/${id}/cancel`, { method: "POST" });
  },

  getStatus(id: number) {
    return apiRequest<SandboxSession>(`/api/v1/sandbox/sessions/${id}/status`, { method: "GET" });
  },

  updateDraft(id: number, input: SandboxDraftUpdate) {
    return apiRequest<SandboxSession>(`/api/v1/sandbox/sessions/${id}/draft`, {
      method: "PATCH",
      body: JSON.stringify(input)
    });
  },

  listMessages(id: number) {
    return apiRequest<{ messages: SandboxMessage[] }>(`/api/v1/sandbox/sessions/${id}/messages`, { method: "GET" });
  },

  askRole(id: number, input: { role: string; question: string }) {
    return apiRequest<SandboxMessage>(`/api/v1/sandbox/sessions/${id}/messages`, {
      method: "POST",
      body: JSON.stringify(input)
    });
  },

  listSessions(limit = 20) {
    return apiRequest<{ sessions: SandboxSession[] }>(`/api/v1/sandbox/sessions?limit=${limit}`, {
      method: "GET"
    });
  },

  listExamples() {
    return apiRequest<{ sessions: SandboxSession[] }>("/api/v1/sandbox/examples", { method: "GET" });
  },

  getSession(id: number) {
    return apiRequest<SandboxSession>(`/api/v1/sandbox/sessions/${id}`, {
      method: "GET"
    });
  }
};
