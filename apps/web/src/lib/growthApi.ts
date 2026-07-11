import { apiRequest } from "./apiRequest";

export type GrowthAssumptions = {
  monthly_visits: number;
  lead_rate: number;
  deal_rate: number;
  average_order: number;
  acquisition_cost: number;
  delivery_cost: number;
};

export type GrowthResult = {
  monthly_revenue: number;
  leads: number;
  deals: number;
  payback_days: number;
  net_margin: number;
};

export type GrowthModel = {
  id: number;
  user_id: number;
  name: string;
  assumptions: GrowthAssumptions;
  result: GrowthResult;
  created_at: string;
  updated_at: string;
};

export type GrowthScenario = {
  name: string;
  revenue: number;
  cost: number;
  margin: number;
  highlight: string;
};

export type GrowthScenarios = {
  model_id: number;
  model_name: string;
  scenarios: GrowthScenario[];
  generated_at: string;
};

export type ForecastMonth = {
  month: string;
  revenue: number;
  phase: string;
  progress_percent: number;
};

export type GrowthForecast = {
  model_id: number;
  model_name: string;
  months: ForecastMonth[];
  generated_at: string;
};

export type CostItem = {
  name: string;
  amount: number;
  detail: string;
};

export type GrowthRecommendations = {
  model_id: number;
  model_name: string;
  headline: string;
  summary: string;
  cost_items: CostItem[];
  action_items: string[];
  generated_at: string;
};

export type GrowthQuestion = {
  key: keyof GrowthAssumptions;
  label: string;
  unit: string;
  min: number;
  max?: number;
};

export type GrowthDraft = {
  id: number;
  user_id: number;
  input: string;
  status: "needs_input" | "ready" | "calculated";
  assumptions: GrowthAssumptions;
  questions: GrowthQuestion[];
  answers: Partial<Record<keyof GrowthAssumptions, number>>;
  model_id?: number;
  created_at: string;
  updated_at: string;
};

export type GrowthSnapshot = {
  id: number;
  user_id: number;
  model_id: number;
  model_name: string;
  assumptions: GrowthAssumptions;
  result: GrowthResult;
  scenarios: GrowthScenarios;
  forecast: GrowthForecast;
  recommendations: GrowthRecommendations;
  created_at: string;
};

export const growthApi = {
  createDraft(input: string) {
    return apiRequest<GrowthDraft>("/api/v1/growth/drafts", {
      method: "POST",
      body: JSON.stringify({ input })
    });
  },

  getDraft(id: number) {
    return apiRequest<GrowthDraft>(`/api/v1/growth/drafts/${id}`, { method: "GET" });
  },

  answerDraft(id: number, answers: Partial<Record<keyof GrowthAssumptions, number>>) {
    return apiRequest<GrowthDraft>(`/api/v1/growth/drafts/${id}/answers`, {
      method: "POST",
      body: JSON.stringify({ answers })
    });
  },

  calculateDraft(id: number, name?: string) {
    return apiRequest<{ draft: GrowthDraft; model: GrowthModel; snapshot: GrowthSnapshot }>(`/api/v1/growth/drafts/${id}/calculate`, {
      method: "POST",
      body: JSON.stringify(name ? { name } : {})
    });
  },

  listSnapshots(id: number) {
    return apiRequest<{ snapshots: GrowthSnapshot[] }>(`/api/v1/growth/models/${id}/snapshots`, { method: "GET" });
  },

  createModel(input: {
    name: string;
    monthlyVisits: number;
    leadRate: number;
    dealRate: number;
    averageOrder: number;
    acquisitionCost: number;
    deliveryCost: number;
  }) {
    return apiRequest<GrowthModel>("/api/v1/growth/models", {
      method: "POST",
      body: JSON.stringify({
        name: input.name,
        monthly_visits: input.monthlyVisits,
        lead_rate: input.leadRate,
        deal_rate: input.dealRate,
        average_order: input.averageOrder,
        acquisition_cost: input.acquisitionCost,
        delivery_cost: input.deliveryCost
      })
    });
  },

  listModels() {
    return apiRequest<{ models: GrowthModel[] }>("/api/v1/growth/models", {
      method: "GET"
    });
  },

  getModel(id: number) {
    return apiRequest<GrowthModel>(`/api/v1/growth/models/${id}`, {
      method: "GET"
    });
  },

  modelScenarios(id: number) {
    return apiRequest<GrowthScenarios>(`/api/v1/growth/models/${id}/scenarios`, {
      method: "GET"
    });
  },

  modelForecast(id: number) {
    return apiRequest<GrowthForecast>(`/api/v1/growth/models/${id}/forecast`, {
      method: "GET"
    });
  },

  modelRecommendations(id: number) {
    return apiRequest<GrowthRecommendations>(`/api/v1/growth/models/${id}/recommendations`, {
      method: "GET"
    });
  }
};
