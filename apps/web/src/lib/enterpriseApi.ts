import { apiRequest } from "./apiRequest";
import type { CrmCustomer } from "./crmApi";

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

export type EnterprisePublicStat = {
  key: string;
  label: string;
  value: string;
  note?: string;
};

export type EnterprisePublicProofPoint = {
  title: string;
  detail?: string;
};

export type EnterprisePublicServiceStep = {
  title: string;
  detail?: string;
};

export type EnterprisePublicOverview = {
  headline: string;
  subheadline?: string;
  description?: string;
  proof_points: EnterprisePublicProofPoint[];
  stats: EnterprisePublicStat[];
  service_steps: EnterprisePublicServiceStep[];
  source_name?: string;
  source_url?: string;
  source_updated_at?: string;
  updated_at?: string;
};

export type EnterprisePublicCaseMetric = {
  label: string;
  value: string;
};

export type EnterprisePublicCase = {
  id: number;
  slug: string;
  company: string;
  title: string;
  summary?: string;
  result?: string;
  industry?: string;
  services: string[];
  metrics: EnterprisePublicCaseMetric[];
  body?: string;
  source_name?: string;
  source_url?: string;
  source_updated_at?: string;
  updated_at?: string;
};

export type EnterprisePublicCasesResponse = {
  cases: EnterprisePublicCase[];
};

export type EnterpriseContactConfig = {
  consultant_name?: string;
  title?: string;
  description?: string;
  phone?: string;
  email?: string;
  wechat?: string;
  qr_image_url?: string;
  contact_url?: string;
  source_name?: string;
  updated_at?: string;
};

export type EnterpriseInquiryInput = {
  company?: string;
  name: string;
  phone?: string;
  email?: string;
  wechat?: string;
  need: string;
  budget?: string;
  timeline?: string;
  source_page?: string;
};

export type EnterpriseInquiry = EnterpriseInquiryInput & {
  id: number;
  status: string;
  crm_customer_id?: number;
  created_at?: string;
  updated_at?: string;
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
  status: "follow_up_created" | "in_delivery" | "completed";
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
  publicOverview() {
    return apiRequest<EnterprisePublicOverview>("/api/v1/enterprise/public-overview");
  },

  publicCases(limit = 20) {
    return apiRequest<EnterprisePublicCasesResponse>(`/api/v1/enterprise/cases?limit=${limit}`);
  },

  publicCase(slug: string) {
    return apiRequest<EnterprisePublicCase>(`/api/v1/enterprise/cases/${encodeURIComponent(slug)}`);
  },

  contactConfig() {
    return apiRequest<EnterpriseContactConfig>("/api/v1/enterprise/contact-config");
  },

  createInquiry(input: EnterpriseInquiryInput) {
    return apiRequest<EnterpriseInquiry>("/api/v1/enterprise/inquiries", {
      method: "POST",
      body: JSON.stringify(input)
    });
  },

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
  },

  importDiagnosisRequestCustomer(id: number) {
    return apiRequest<CrmCustomer>(`/api/v1/enterprise/diagnosis-requests/${id}/crm-customer`, {
      method: "POST"
    });
  }
};
