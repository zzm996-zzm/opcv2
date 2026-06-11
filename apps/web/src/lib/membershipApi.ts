import { authSession } from "./authSession";

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

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const token = authSession.get().accessToken;
  const response = await fetch(path, {
    ...init,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...init.headers
    }
  });
  if (!response.ok) {
    const payload = (await response.json().catch(() => ({}))) as { error?: string };
    throw new Error(payload.error ?? "request_failed");
  }
  return (await response.json()) as T;
}

export const membershipApi = {
  current() {
    return request<MembershipSnapshot>("/api/v1/membership/me");
  },

  redeem(code: string) {
    return request<RedeemResponse>("/api/v1/redemptions/redeem", {
      method: "POST",
      body: JSON.stringify({ code })
    });
  }
};
