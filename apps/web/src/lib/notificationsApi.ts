import { apiRequest } from "./apiRequest";

export type NotificationItem = {
  id: number;
  user_id: number;
  type: string;
  title: string;
  summary: string;
  body: string;
  source_type: string;
  source_id?: number;
  action_label: string;
  action_url: string;
  read_at?: string;
  created_at: string;
};

export type NotificationFilters = {
  type?: string;
  status?: "all" | "unread" | "read" | string;
  limit?: number;
};

export type NotificationTypeCount = {
  type: string;
  count: number;
};

export type NotificationSummary = {
  unread: number;
  by_type: NotificationTypeCount[];
  latest: NotificationItem[];
};

function queryString(params: NotificationFilters) {
  const search = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== "") {
      search.set(key, String(value));
    }
  });
  const encoded = search.toString();
  return encoded ? `?${encoded}` : "";
}

export const notificationsApi = {
  list(filters: NotificationFilters = {}) {
    return apiRequest<{ notifications: NotificationItem[] }>(`/api/v1/notifications${queryString(filters)}`);
  },

  get(id: number) {
    return apiRequest<NotificationItem>(`/api/v1/notifications/${id}`);
  },

  markRead(id: number) {
    return apiRequest<NotificationItem>(`/api/v1/notifications/${id}/read`, {
      method: "PATCH"
    });
  },

  markAllRead() {
    return apiRequest<{ updated: number }>("/api/v1/notifications/read-all", {
      method: "POST"
    });
  },

  summary() {
    return apiRequest<NotificationSummary>("/api/v1/notifications/summary");
  }
};
