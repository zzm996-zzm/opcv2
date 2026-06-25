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

async function request<T>(path: string): Promise<T> {
  const response = await fetch(path, {
    credentials: "include",
    headers: {
      "Content-Type": "application/json"
    }
  });
  if (!response.ok) {
    const payload = (await response.json().catch(() => ({}))) as { error?: string };
    throw new Error(payload.error ?? "request_failed");
  }
  return (await response.json()) as T;
}

export const contentApi = {
  listArticles() {
    return request<{ articles: ContentArticle[] }>("/api/v1/content/articles");
  },

  getArticle(slug: string) {
    return request<ContentArticle>(`/api/v1/content/articles/${slug}`);
  },

  listTools() {
    return request<{ tools: ContentTool[] }>("/api/v1/content/tools");
  },

  getCommunityConfig() {
    return request<CommunityConfig>("/api/v1/content/community");
  },

  getBrand() {
    return request<{ metrics: BrandMetric[]; cases: BrandCase[] }>("/api/v1/content/brand");
  }
};
