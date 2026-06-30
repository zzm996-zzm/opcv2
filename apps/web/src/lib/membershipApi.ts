import { apiRequest } from "./apiRequest";

export type MembershipPlan = {
  code: string;
  name: string;
  monthly_analysis_limit: number;
  lead_export_limit: number;
};

export type MembershipSnapshot = {
  plan: MembershipPlan;
  credit_balance: number;
};

export type RedeemResponse = {
  snapshot: MembershipSnapshot;
  already_redeemed: boolean;
};

export const membershipApi = {
  current() {
    return apiRequest<MembershipSnapshot>("/api/v1/membership/me");
  },

  redeem(code: string) {
    return apiRequest<RedeemResponse>("/api/v1/redemptions/redeem", {
      method: "POST",
      body: JSON.stringify({ code })
    });
  }
};
