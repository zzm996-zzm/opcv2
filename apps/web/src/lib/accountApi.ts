import { apiRequest } from "./apiRequest";

export type AccountProfile = {
  id: number;
  nickname: string;
  phone: string;
  email: string;
  wechat: string;
  company: string;
  industry: string;
  role: string;
  onboarding_completed: boolean;
  created_at: string;
  updated_at: string;
};

export type AccountBinding = {
  type: string;
  masked_value: string;
  bound: boolean;
};

export type AccountProfilePayload = {
  profile: AccountProfile;
  bindings: AccountBinding[];
};

export type AccountProfileUpdate = Partial<{
  nickname: string;
  email: string;
  wechat: string;
  company: string;
  industry: string;
  role: string;
}>;

export type OnboardingSection = {
  key: string;
  title: string;
  fields: Record<string, string>;
};

export type OnboardingState = {
  completed: boolean;
  sections: OnboardingSection[];
};

export type AccountPreferences = {
  notifications_enabled: boolean;
  default_model: string;
  language: string;
  timezone: string;
};

export type AccountPreferencesUpdate = Partial<AccountPreferences>;

export type AccountQuota = {
  key: string;
  label: string;
  used: number;
  limit: number;
  unit: string;
  reset_at?: string;
};

export type AccountContentItem = {
  id: string;
  type: string;
  title: string;
  summary: string;
  url: string;
  created_at: string;
};

export type AccountDeletionStatus = {
  status: string;
};

export const accountApi = {
  getProfile() {
    return apiRequest<AccountProfilePayload>("/api/v1/account/profile");
  },

  updateProfile(input: AccountProfileUpdate) {
    return apiRequest<AccountProfilePayload>("/api/v1/account/profile", {
      method: "PATCH",
      body: JSON.stringify(input)
    });
  },

  getOnboarding() {
    return apiRequest<OnboardingState>("/api/v1/account/onboarding");
  },

  saveOnboarding(input: OnboardingState) {
    return apiRequest<OnboardingState>("/api/v1/account/onboarding", {
      method: "PUT",
      body: JSON.stringify(input)
    });
  },

  completeOnboarding() {
    return apiRequest<OnboardingState>("/api/v1/account/onboarding/complete", {
      method: "POST"
    });
  },

  getPreferences() {
    return apiRequest<AccountPreferences>("/api/v1/account/preferences");
  },

  updatePreferences(input: AccountPreferencesUpdate) {
    return apiRequest<AccountPreferences>("/api/v1/account/preferences", {
      method: "PATCH",
      body: JSON.stringify(input)
    });
  },

  getQuotas() {
    return apiRequest<{ quotas: AccountQuota[] }>("/api/v1/account/quotas");
  },

  listContent(limit = 20) {
    return apiRequest<{ items: AccountContentItem[] }>(`/api/v1/account/content?limit=${limit}`);
  },

  deleteAccount() {
    return apiRequest<AccountDeletionStatus>("/api/v1/account", {
      method: "DELETE"
    });
  }
};
