import { apiRequest } from "./apiRequest";

export type TaskStatus = "todo" | "in_progress" | "completed" | "reminder";
export type TaskPriority = "low" | "medium" | "high";

export type Task = {
  id: number;
  user_id: number;
  title: string;
  project: string;
  status: TaskStatus;
  priority: TaskPriority;
  due_at?: string;
  tools: string[];
  learning: string;
  created_at: string;
  updated_at: string;
};

export type CreateTaskInput = {
  title: string;
  project: string;
  priority: TaskPriority;
  dueAt?: string;
  tools: string[];
  learning: string;
};

export type UpdateTaskInput = Partial<{
  title: string;
  project: string;
  status: TaskStatus;
  priority: TaskPriority;
  dueAt: string;
  tools: string[];
  learning: string;
}>;

function toCreatePayload(input: CreateTaskInput) {
  return {
    title: input.title,
    project: input.project,
    priority: input.priority,
    due_at: input.dueAt,
    tools: input.tools,
    learning: input.learning
  };
}

function toUpdatePayload(input: UpdateTaskInput) {
  return {
    title: input.title,
    project: input.project,
    status: input.status,
    priority: input.priority,
    due_at: input.dueAt,
    tools: input.tools,
    learning: input.learning
  };
}

export const tasksApi = {
  createTask(input: CreateTaskInput) {
    return apiRequest<Task>("/api/v1/tasks", {
      method: "POST",
      body: JSON.stringify(toCreatePayload(input))
    });
  },

  listTasks(limit = 20) {
    return apiRequest<{ tasks: Task[] }>(`/api/v1/tasks?limit=${limit}`, {
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
  }
};
