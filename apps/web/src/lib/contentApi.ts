import { apiRequest } from "./apiRequest";

export type ContentArticle = {
  id: number;
  slug: string;
  title: string;
  summary?: string;
  body?: string;
  status: "draft" | "published";
  published_at?: string;
  created_at: string;
  updated_at: string;
};

export type ContentTool = {
  id: number;
  slug: string;
  name: string;
  description?: string;
  url?: string;
  status: "draft" | "published";
  created_at: string;
  updated_at: string;
};

export type CommunityConfig = {
  id: number;
  headline: string;
  description?: string;
  join_url?: string;
  created_at: string;
  updated_at: string;
};

export type BrandMetric = {
  id: number;
  key: string;
  label: string;
  value: string;
  created_at: string;
  updated_at: string;
};

export type BrandCase = {
  id: number;
  slug: string;
  title: string;
  summary?: string;
  url?: string;
  created_at: string;
  updated_at: string;
};

export const contentApi = {
  listArticles() {
    return apiRequest<{ articles: ContentArticle[] }>("/api/v1/content/articles");
  },

  getArticle(slug: string) {
    return apiRequest<ContentArticle>(`/api/v1/content/articles/${slug}`);
  },

  listTools() {
    return apiRequest<{ tools: ContentTool[] }>("/api/v1/content/tools");
  },

  getCommunityConfig() {
    return apiRequest<CommunityConfig>("/api/v1/content/community");
  },

  getBrand() {
    return apiRequest<{ metrics: BrandMetric[]; cases: BrandCase[] }>("/api/v1/content/brand");
  }
};
