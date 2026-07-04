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

export const enterpriseApi = {
  overview() {
    return apiRequest<EnterpriseOverview>("/api/v1/enterprise/overview");
  }
};
