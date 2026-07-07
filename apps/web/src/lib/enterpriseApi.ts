import { apiRequest } from "./apiRequest";

export type EnterpriseMetric = {
  key: string;
  label: string;
  value: string;
};

export type EnterprisePlan = {
  id: number;
  title: string;
  audience?: string;
  price_label?: string;
  focus: string[];
  result?: string;
};

export type EnterpriseDeliveryItem = {
  stage: string;
  count: number;
  detail?: string;
};

export type EnterpriseMilestone = {
  time_label: string;
  title: string;
  detail?: string;
};

export type EnterpriseCase = {
  id: number;
  company: string;
  result?: string;
};

export type EnterpriseOverview = {
  stats: EnterpriseMetric[];
  plans: EnterprisePlan[];
  delivery_board: EnterpriseDeliveryItem[];
  milestones: EnterpriseMilestone[];
  cases: EnterpriseCase[];
};

export type EnterpriseDiagnosisRequestInput = {
  need: string;
};

export type EnterpriseDiagnosisRequestUpdateInput = {
  status: "follow_up_created" | "in_delivery";
};

export type EnterpriseDiagnosisRequest = {
  id: number;
  user_id: number;
  need: string;
  status: string;
  created_at?: string;
  updated_at?: string;
};

export type EnterpriseDiagnosisRequestsResponse = {
  requests: EnterpriseDiagnosisRequest[];
};

export const enterpriseApi = {
  overview() {
    return apiRequest<EnterpriseOverview>("/api/v1/enterprise/overview");
  },

  listDiagnosisRequests(limit = 10) {
    return apiRequest<EnterpriseDiagnosisRequestsResponse>(`/api/v1/enterprise/diagnosis-requests?limit=${limit}`);
  },

  createDiagnosisRequest(input: EnterpriseDiagnosisRequestInput) {
    return apiRequest<EnterpriseDiagnosisRequest>("/api/v1/enterprise/diagnosis-requests", {
      method: "POST",
      body: JSON.stringify(input)
    });
  },

  updateDiagnosisRequest(id: number, input: EnterpriseDiagnosisRequestUpdateInput) {
    return apiRequest<EnterpriseDiagnosisRequest>(`/api/v1/enterprise/diagnosis-requests/${id}`, {
      method: "PATCH",
      body: JSON.stringify(input)
    });
  }
};
