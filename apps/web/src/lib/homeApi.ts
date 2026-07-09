import { apiRequest } from "./apiRequest";

export type HomeCard = {
  title: string;
  summary?: string;
  url?: string;
};

export type HomeMetric = {
  label: string;
  value: string;
  icon?: string;
};

export type HomeActionItem = {
  type: string;
  priority: string;
  title: string;
  summary?: string;
  url?: string;
  cta?: string;
};

export type HomeRecentTask = {
  id: number;
  title: string;
  project: string;
  status: string;
  priority?: string;
  due_at?: string;
  is_overdue: boolean;
};

export type HomeNotificationItem = {
  id: number;
  type: string;
  title: string;
  summary?: string;
  action_label?: string;
  action_url?: string;
  read_at?: string;
  created_at: string;
};

export type HomeNotificationTypeCount = {
  type: string;
  count: number;
};

export type HomeNotificationSummary = {
  unread: number;
  by_type?: HomeNotificationTypeCount[];
  latest: HomeNotificationItem[];
};

export type HomeQuotaWarning = {
  key: string;
  label: string;
  used: number;
  limit: number;
  message: string;
};

export type HomeAccountSummary = {
  plan_name: string;
  credit_balance: number;
  quota_warnings: HomeQuotaWarning[];
};

export type HomeSummary = {
  metrics: HomeMetric[];
  hero_cards: HomeCard[];
  recommendations: HomeCard[];
  action_items?: HomeActionItem[];
  recent_tasks: HomeRecentTask[];
  notification_summary: HomeNotificationSummary;
  account_summary: HomeAccountSummary;
};

export const homeApi = {
  summary() {
    return apiRequest<HomeSummary>("/api/v1/home/summary");
  }
};
