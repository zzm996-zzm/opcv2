import { apiRequest } from "./apiRequest";

export type HomeCard = {
  title: string;
  summary?: string;
  url?: string;
};

export type HomeRecentTask = {
  id: number;
  title: string;
  project: string;
  status: string;
  due_at?: string;
};

export type HomeNotificationItem = {
  id: number;
  type: string;
  title: string;
  summary?: string;
  action_url?: string;
  created_at: string;
};

export type HomeNotificationSummary = {
  unread: number;
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
  hero_cards: HomeCard[];
  recommendations: HomeCard[];
  recent_tasks: HomeRecentTask[];
  notification_summary: HomeNotificationSummary;
  account_summary: HomeAccountSummary;
};

export const homeApi = {
  summary() {
    return apiRequest<HomeSummary>("/api/v1/home/summary");
  }
};
