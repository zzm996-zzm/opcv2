import { apiRequest } from "./apiRequest";

export type MembershipPlan = {
  code: string;
  name: string;
  monthly_analysis_limit: number;
  lead_export_limit: number;
};

export type MembershipPlanQuota = {
  key: string;
  label: string;
  limit: number;
  unit: string;
};

export type MembershipPlanOption = {
  id: number;
  code: string;
  name: string;
  price_cents: number;
  billing_cycle: string;
  features: string[];
  quotas: MembershipPlanQuota[];
  recommended: boolean;
};

export type MembershipSnapshot = {
  plan: MembershipPlan;
  credit_balance: number;
};

export type MembershipUsageItem = {
  key: string;
  label: string;
  used: number;
  limit: number;
  unit: string;
  reset_at?: string;
};

export type MembershipOrder = {
  id: number;
  order_no: string;
  plan_code: string;
  amount_cents: number;
  status: "pending" | "paid" | string;
  created_at: string;
  paid_at?: string;
};

export type CheckoutInput = {
  plan_code: string;
  billing_cycle: string;
};

export type CheckoutResult = {
  order: MembershipOrder;
  payment: {
    mode: string;
    message?: string;
  };
};

export type RedeemResponse = {
  snapshot: MembershipSnapshot;
  already_redeemed: boolean;
};

export const membershipApi = {
  current() {
    return apiRequest<MembershipSnapshot>("/api/v1/membership/me");
  },

  listPlans() {
    return apiRequest<{ plans: MembershipPlanOption[] }>("/api/v1/membership/plans");
  },

  usage() {
    return apiRequest<{ usage: MembershipUsageItem[] }>("/api/v1/membership/usage");
  },

  listOrders(limit = 20) {
    return apiRequest<{ orders: MembershipOrder[] }>(`/api/v1/membership/orders?limit=${limit}`);
  },

  checkout(input: CheckoutInput) {
    return apiRequest<CheckoutResult>("/api/v1/membership/checkout", {
      method: "POST",
      body: JSON.stringify(input)
    });
  },

  redeem(code: string) {
    return apiRequest<RedeemResponse>("/api/v1/redemptions/redeem", {
      method: "POST",
      body: JSON.stringify({ code })
    });
  }
};
