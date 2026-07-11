import { apiRequest } from "./apiRequest";

export type TaskStatus = "todo" | "in_progress" | "completed" | "reminder";
export type TaskPriority = "low" | "medium" | "high";
export type ReminderRecurrence = "once" | "daily" | "weekly";
export type TaskSourceType =
  | "analysis_session"
  | "project_match"
  | "sandbox_session"
  | "competitor_scan"
  | "competitor_monitoring"
  | "enterprise_diagnosis"
  | "lead_task"
  | "crm_customer"
  | "learning_diagnosis";

export type Task = {
  id: number;
  user_id: number;
  title: string;
  description?: string;
  assignee?: string;
  project: string;
  status: TaskStatus;
  priority: TaskPriority;
  tags?: string[];
  due_at?: string;
  tools: string[];
  learning: string;
  source_type?: TaskSourceType;
  source_id?: number;
  source_title?: string;
  source_url?: string;
  created_at: string;
  updated_at: string;
};

export type TaskFilters = {
  status?: TaskStatus;
  project?: string;
  priority?: TaskPriority;
  tag?: string;
  q?: string;
  limit?: number;
  offset?: number;
};

export type TaskPage = {
  tasks: Task[];
  total: number;
  limit: number;
  offset: number;
};

export type TaskStats = {
  total: number;
  todo: number;
  in_progress: number;
  completed: number;
  reminder: number;
  overdue: number;
};

export type GenerateTasksResult = {
  tasks: Task[];
};

export type GenerateTaskSource = {
  sourceType: TaskSourceType;
  sourceId?: number;
  sourceTitle: string;
  sourceUrl: string;
};

export type TaskSubtask = {
  id: number;
  task_id: number;
  user_id: number;
  title: string;
  assignee?: string;
  due_at?: string;
  completed: boolean;
  created_at: string;
  updated_at: string;
};

export type TaskReminder = {
  id: number;
  task_id: number;
  user_id: number;
  remind_at: string;
  recurrence: ReminderRecurrence;
  sent_at?: string;
  created_at: string;
  updated_at: string;
};

export type CreateSubtaskInput = {
  title: string;
  assignee?: string;
  dueAt?: string;
};

export type UpdateSubtaskInput = Partial<{
  title: string;
  assignee: string;
  dueAt: string;
  clearDueAt: boolean;
  completed: boolean;
}>;

export type UpsertReminderInput = {
  remindAt: string;
  recurrence: ReminderRecurrence;
};

export type CreateTaskInput = {
  title: string;
  description?: string;
  assignee?: string;
  project: string;
  priority: TaskPriority;
  tags?: string[];
  dueAt?: string;
  tools: string[];
  learning: string;
  sourceType?: TaskSourceType;
  sourceId?: number;
  sourceTitle?: string;
  sourceUrl?: string;
};

export type UpdateTaskInput = Partial<{
  title: string;
  description: string;
  assignee: string;
  project: string;
  status: TaskStatus;
  priority: TaskPriority;
  tags: string[];
  dueAt: string;
  clearDueAt: boolean;
  tools: string[];
  learning: string;
}>;

function toCreatePayload(input: CreateTaskInput) {
  return {
    title: input.title,
    description: input.description,
    assignee: input.assignee,
    project: input.project,
    priority: input.priority,
    tags: input.tags,
    due_at: input.dueAt,
    tools: input.tools,
    learning: input.learning,
    source_type: input.sourceType,
    source_id: input.sourceId,
    source_title: input.sourceTitle,
    source_url: input.sourceUrl
  };
}

function toUpdatePayload(input: UpdateTaskInput) {
  return {
    title: input.title,
    description: input.description,
    assignee: input.assignee,
    project: input.project,
    status: input.status,
    priority: input.priority,
    tags: input.tags,
    due_at: input.dueAt,
    clear_due_at: input.clearDueAt,
    tools: input.tools,
    learning: input.learning
  };
}

function toSubtaskPayload(input: CreateSubtaskInput | UpdateSubtaskInput) {
  return {
    title: input.title,
    assignee: input.assignee,
    due_at: input.dueAt,
    clear_due_at: "clearDueAt" in input ? input.clearDueAt : undefined,
    completed: "completed" in input ? input.completed : undefined
  };
}

function queryString(params: Record<string, string | number | undefined>) {
  const search = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== "") {
      search.set(key, String(value));
    }
  });
  const encoded = search.toString();
  return encoded ? `?${encoded}` : "";
}

export const tasksApi = {
  createTask(input: CreateTaskInput) {
    return apiRequest<Task>("/api/v1/tasks", {
      method: "POST",
      body: JSON.stringify(toCreatePayload(input))
    });
  },

  generateTasks(goal: string, source?: GenerateTaskSource) {
    return apiRequest<GenerateTasksResult>("/api/v1/tasks/generate", {
      method: "POST",
      body: JSON.stringify({
        goal,
        ...(source ? {
          source_type: source.sourceType,
          source_id: source.sourceId,
          source_title: source.sourceTitle,
          source_url: source.sourceUrl
        } : {})
      })
    });
  },

  listTasks(filters: TaskFilters | number = {}) {
    const normalized = typeof filters === "number" ? { limit: filters } : filters;
    return apiRequest<TaskPage>(`/api/v1/tasks${queryString(normalized)}`, {
      method: "GET"
    });
  },

  listProjects() {
    return apiRequest<{ projects: string[] }>("/api/v1/tasks/projects", {
      method: "GET"
    });
  },

  listTags() {
    return apiRequest<{ tags: string[] }>("/api/v1/tasks/tags", {
      method: "GET"
    });
  },

  stats() {
    return apiRequest<TaskStats>("/api/v1/tasks/stats", {
      method: "GET"
    });
  },

  getTask(id: number) {
    return apiRequest<Task>(`/api/v1/tasks/${id}`, {
      method: "GET"
    });
  },

  updateTask(id: number, input: UpdateTaskInput) {
    return apiRequest<Task>(`/api/v1/tasks/${id}`, {
      method: "PATCH",
      body: JSON.stringify(toUpdatePayload(input))
    });
  },

  batchUpdateStatus(ids: number[], status: TaskStatus) {
    return apiRequest<{ updated: number }>("/api/v1/tasks/batch", {
      method: "PATCH",
      body: JSON.stringify({ ids, status })
    });
  },

  batchDelete(ids: number[]) {
    return apiRequest<{ deleted: number }>("/api/v1/tasks/batch", {
      method: "DELETE",
      body: JSON.stringify({ ids })
    });
  },

  deleteTask(id: number) {
    return apiRequest<void>(`/api/v1/tasks/${id}`, {
      method: "DELETE"
    });
  },

  listSubtasks(taskID: number) {
    return apiRequest<{ subtasks: TaskSubtask[] }>(`/api/v1/tasks/${taskID}/subtasks`, {
      method: "GET"
    });
  },

  createSubtask(taskID: number, input: CreateSubtaskInput) {
    return apiRequest<TaskSubtask>(`/api/v1/tasks/${taskID}/subtasks`, {
      method: "POST",
      body: JSON.stringify(toSubtaskPayload(input))
    });
  },

  updateSubtask(taskID: number, subtaskID: number, input: UpdateSubtaskInput) {
    return apiRequest<TaskSubtask>(`/api/v1/tasks/${taskID}/subtasks/${subtaskID}`, {
      method: "PATCH",
      body: JSON.stringify(toSubtaskPayload(input))
    });
  },

  deleteSubtask(taskID: number, subtaskID: number) {
    return apiRequest<void>(`/api/v1/tasks/${taskID}/subtasks/${subtaskID}`, {
      method: "DELETE"
    });
  },

  getReminder(taskID: number) {
    return apiRequest<{ reminder: TaskReminder | null }>(`/api/v1/tasks/${taskID}/reminder`, {
      method: "GET"
    });
  },

  upsertReminder(taskID: number, input: UpsertReminderInput) {
    return apiRequest<TaskReminder>(`/api/v1/tasks/${taskID}/reminder`, {
      method: "PUT",
      body: JSON.stringify({ remind_at: input.remindAt, recurrence: input.recurrence })
    });
  },

  deleteReminder(taskID: number) {
    return apiRequest<void>(`/api/v1/tasks/${taskID}/reminder`, {
      method: "DELETE"
    });
  }
};
