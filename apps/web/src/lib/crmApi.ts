import { apiRequest } from "./apiRequest";

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

export const crmApi = {
  importLead(input: { leadResultId: number; name: string; phone?: string; email?: string; website?: string }) {
    return apiRequest<CrmCustomer>("/api/v1/crm/customers/import-lead", {
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
    return apiRequest<{ customers: CrmCustomer[] }>(`/api/v1/crm/customers/due?limit=${limit}`, {
      method: "GET"
    });
  },

  updateStage(customerId: number, input: { stage: CrmStage; note?: string }) {
    return apiRequest<CrmCustomer>(`/api/v1/crm/customers/${customerId}/stage`, {
      method: "POST",
      body: JSON.stringify(input)
    });
  },

  recordFollowUp(customerId: number, input: { note: string; nextFollowUpAt: string }) {
    return apiRequest<CrmFollowUp>(`/api/v1/crm/customers/${customerId}/follow-ups`, {
      method: "POST",
      body: JSON.stringify({
        note: input.note,
        next_follow_up_at: input.nextFollowUpAt
      })
    });
  },

  generateFollowUpCopy(customerId: number, input: { goal: string }) {
    return apiRequest<CrmFollowUpCopy>(`/api/v1/crm/customers/${customerId}/follow-up-copy`, {
      method: "POST",
      body: JSON.stringify(input)
    });
  }
};
