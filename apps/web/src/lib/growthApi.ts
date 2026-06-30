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

export const growthApi = {
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
  }
};
