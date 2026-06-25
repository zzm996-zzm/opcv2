import { authSession } from "./authSession";

export type CrmStage = "new" | "contacted" | "qualified" | "proposal" | "won" | "lost";

export type CrmCustomer = {
  id: number;
  user_id: number;
  import_key: string;
  name: string;
  phone?: string;
  email?: string;
  website?: string;
  stage: CrmStage;
  source: string;
  next_follow_up_at?: string;
  created_at: string;
  updated_at: string;
};

export type CrmFollowUp = {
  id: number;
  user_id: number;
  customer_id: number;
  note: string;
  next_follow_up_at: string;
  created_at: string;
};

export type CrmFollowUpCopy = {
  subject: string;
  body: string;
  channel: string;
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

export const crmApi = {
  importLead(input: { leadResultId: number; name: string; phone?: string; email?: string; website?: string }) {
    return request<CrmCustomer>("/api/v1/crm/customers/import-lead", {
      method: "POST",
      body: JSON.stringify({
        lead_result_id: input.leadResultId,
        name: input.name,
        phone: input.phone,
        email: input.email,
        website: input.website
      })
    });
  },

  listDueCustomers(limit = 20) {
    return request<{ customers: CrmCustomer[] }>(`/api/v1/crm/customers/due?limit=${limit}`, {
      method: "GET"
    });
  },

  updateStage(customerId: number, input: { stage: CrmStage; note?: string }) {
    return request<CrmCustomer>(`/api/v1/crm/customers/${customerId}/stage`, {
      method: "POST",
      body: JSON.stringify(input)
    });
  },

  recordFollowUp(customerId: number, input: { note: string; nextFollowUpAt: string }) {
    return request<CrmFollowUp>(`/api/v1/crm/customers/${customerId}/follow-ups`, {
      method: "POST",
      body: JSON.stringify({
        note: input.note,
        next_follow_up_at: input.nextFollowUpAt
      })
    });
  },

  generateFollowUpCopy(customerId: number, input: { goal: string }) {
    return request<CrmFollowUpCopy>(`/api/v1/crm/customers/${customerId}/follow-up-copy`, {
      method: "POST",
      body: JSON.stringify(input)
    });
  }
};
