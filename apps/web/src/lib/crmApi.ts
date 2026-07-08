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

export type CrmActivity = {
  id: number;
  user_id: number;
  customer_id: number;
  type: string;
  note?: string;
  created_at: string;
};

export type CrmFollowUpCopy = {
  subject: string;
  body: string;
  channel: string;
};

export type CrmPipelineStats = {
  total: number;
  new: number;
  contacted: number;
  qualified: number;
  proposal: number;
  won: number;
  lost: number;
  due_today: number;
};

export type CrmCustomerFilters = {
  stage?: CrmStage;
  source?: "lead" | "enterprise";
  q?: string;
  limit?: number;
};

export type CrmFollowUpFilters = {
  customerId?: number;
  q?: string;
  due?: "today" | "week" | "overdue";
  limit?: number;
};

function customerQuery(filters: CrmCustomerFilters = {}) {
  const params = new URLSearchParams();
  if (filters.stage) params.set("stage", filters.stage);
  if (filters.source) params.set("source", filters.source);
  if (filters.q) params.set("q", filters.q);
  if (filters.limit) params.set("limit", String(filters.limit));
  const query = params.toString();
  return query ? `?${query}` : "";
}

function followUpQuery(filters: CrmFollowUpFilters = {}) {
  const params = new URLSearchParams();
  if (filters.customerId) params.set("customer_id", String(filters.customerId));
  if (filters.q) params.set("q", filters.q);
  if (filters.due) params.set("due", filters.due);
  if (filters.limit) params.set("limit", String(filters.limit));
  const query = params.toString();
  return query ? `?${query}` : "";
}

export const crmApi = {
  listCustomers(filters: CrmCustomerFilters = {}) {
    return apiRequest<{ customers: CrmCustomer[] }>(`/api/v1/crm/customers${customerQuery(filters)}`, {
      method: "GET"
    });
  },

  getCustomer(customerId: number) {
    return apiRequest<CrmCustomer>(`/api/v1/crm/customers/${customerId}`, {
      method: "GET"
    });
  },

  createCustomer(input: { name: string; phone?: string; email?: string; website?: string }) {
    return apiRequest<CrmCustomer>("/api/v1/crm/customers", {
      method: "POST",
      body: JSON.stringify(input)
    });
  },

  updateCustomer(customerId: number, input: { name?: string; phone?: string; email?: string; website?: string }) {
    return apiRequest<CrmCustomer>(`/api/v1/crm/customers/${customerId}`, {
      method: "PATCH",
      body: JSON.stringify(input)
    });
  },

  listActivities(customerId: number, limit = 20) {
    return apiRequest<{ activities: CrmActivity[] }>(`/api/v1/crm/customers/${customerId}/activities?limit=${limit}`, {
      method: "GET"
    });
  },

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

  listFollowUps(filters: CrmFollowUpFilters = {}) {
    return apiRequest<{ follow_ups: CrmFollowUp[] }>(`/api/v1/crm/follow-ups${followUpQuery(filters)}`, {
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

  rescheduleFollowUp(followUpId: number, input: { nextFollowUpAt: string }) {
    return apiRequest<CrmFollowUp>(`/api/v1/crm/follow-ups/${followUpId}`, {
      method: "PATCH",
      body: JSON.stringify({
        next_follow_up_at: input.nextFollowUpAt
      })
    });
  },

  generateFollowUpCopy(customerId: number, input: { goal: string }) {
    return apiRequest<CrmFollowUpCopy>(`/api/v1/crm/customers/${customerId}/follow-up-copy`, {
      method: "POST",
      body: JSON.stringify(input)
    });
  },

  pipelineStats() {
    return apiRequest<CrmPipelineStats>("/api/v1/crm/pipeline-stats", {
      method: "GET"
    });
  }
};
