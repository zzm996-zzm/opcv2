import { apiRequest } from "./apiRequest";

export type ContentArticle = {
  id: number;
  slug: string;
  title: string;
  summary?: string;
  body?: string;
  status: "draft" | "published";
  source_name?: string;
  source_url?: string;
  author?: string;
  category?: string;
  tags: string[];
  citations: ContentArticleCitation[];
  source_published_at?: string;
  published_at?: string;
  created_at: string;
  updated_at: string;
};

export type ContentArticleCitation = {
  id: string;
  label: string;
  source_name: string;
  source_url: string;
  excerpt?: string;
  published_at?: string;
};

export type ArticleFilters = {
  category?: string;
  q?: string;
  limit?: number;
};

export type ContentTool = {
  id: number;
  slug: string;
  name: string;
  description?: string;
  url?: string;
  status: "draft" | "published";
  category?: string;
  tags: string[];
  provider_name?: string;
  price_label?: string;
  platforms: string[];
  features: string[];
  use_cases: string[];
  limitations: string[];
  source_url?: string;
  source_updated_at?: string;
  sort_weight: number;
  created_at: string;
  updated_at: string;
};

export type ToolRecommendationInput = {
  goal: string;
  scenario: string;
  category?: string;
  limit?: number;
};

export type ToolRecommendations = {
  basis: "catalog_match" | string;
  criteria: ToolRecommendationInput;
  tools: ContentTool[];
};

export type ToolFilters = {
  category?: string;
  q?: string;
  sort?: string;
  limit?: number;
};

export type FavoriteToolResult = {
  slug: string;
  favorited: boolean;
};

export type BookmarkArticleResult = {
  slug: string;
  bookmarked: boolean;
};

export type CommunityConfig = {
  id: number;
  headline: string;
  description?: string;
  join_url?: string;
  qr_variants: CommunityQRVariant[];
  created_at: string;
  updated_at: string;
};

export type CommunityQRVariant = {
  key: string;
  label: string;
  description?: string;
  image_url?: string;
  join_url?: string;
  status: "draft" | "published";
};

export type InsightAnswer = {
  answer: string;
  citations: Array<{
    id: string;
    article_slug: string;
    label: string;
    source_name: string;
    source_url: string;
    excerpt?: string;
  }>;
  assumptions: string[];
  basis: "catalog_citations" | string;
  disclaimer: string;
};

export type CommunityJoinInput = {
  community: string;
  contact: string;
  note?: string;
};

export type CommunityJoinRequest = {
  id: number;
  user_id?: number;
  community: string;
  contact?: string;
  note?: string;
  status: string;
  created_at: string;
};

export type HelpTopic = {
  key: string;
  name: string;
};

export type HelpArticleFilters = {
  topic?: string;
  q?: string;
  limit?: number;
};

export type HelpArticle = {
  id?: number;
  slug: string;
  topic: string;
  title: string;
  summary?: string;
  body?: string;
  created_at?: string;
  updated_at?: string;
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

export const contentApi = {
  listArticles(filters: ArticleFilters = {}) {
    return apiRequest<{ articles: ContentArticle[] }>(`/api/v1/content/articles${queryString(filters)}`);
  },

  getArticle(slug: string) {
    return apiRequest<ContentArticle>(`/api/v1/content/articles/${encodeURIComponent(slug)}`);
  },

  bookmarkArticle(slug: string) {
    return apiRequest<BookmarkArticleResult>(`/api/v1/content/articles/${slug}/bookmark`, {
      method: "POST"
    });
  },

  unbookmarkArticle(slug: string) {
    return apiRequest<BookmarkArticleResult>(`/api/v1/content/articles/${slug}/bookmark`, {
      method: "DELETE"
    });
  },

  listTools(filters: ToolFilters = {}) {
    return apiRequest<{ tools: ContentTool[] }>(`/api/v1/content/tools${queryString(filters)}`);
  },

  getTool(slug: string) {
    return apiRequest<ContentTool>(`/api/v1/content/tools/${encodeURIComponent(slug)}`);
  },

  recommendTools(input: ToolRecommendationInput) {
    return apiRequest<ToolRecommendations>("/api/v1/content/tools/recommendations", {
      method: "POST",
      body: JSON.stringify(input)
    });
  },

  favoriteTool(slug: string) {
    return apiRequest<FavoriteToolResult>(`/api/v1/content/tools/${slug}/favorite`, {
      method: "POST"
    });
  },

  unfavoriteTool(slug: string) {
    return apiRequest<FavoriteToolResult>(`/api/v1/content/tools/${slug}/favorite`, {
      method: "DELETE"
    });
  },

  getCommunityConfig() {
    return apiRequest<CommunityConfig>("/api/v1/content/community");
  },

  answerInsightQuestion(question: string, articleSlugs: string[]) {
    return apiRequest<InsightAnswer>("/api/v1/content/insights/qa", {
      method: "POST",
      body: JSON.stringify({ question, article_slugs: articleSlugs })
    });
  },

  joinCommunity(input: CommunityJoinInput) {
    return apiRequest<CommunityJoinRequest>("/api/v1/community/join-requests", {
      method: "POST",
      body: JSON.stringify(input)
    });
  },

  listHelpTopics() {
    return apiRequest<{ topics: HelpTopic[] }>("/api/v1/help/topics");
  },

  listHelpArticles(filters: HelpArticleFilters = {}) {
    return apiRequest<{ articles: HelpArticle[] }>(`/api/v1/help/articles${queryString(filters)}`);
  },

  getHelpArticle(slug: string) {
    return apiRequest<HelpArticle>(`/api/v1/help/articles/${slug}`);
  },

  getBrand() {
    return apiRequest<{ metrics: BrandMetric[]; cases: BrandCase[] }>("/api/v1/content/brand");
  }
};
