import { apiRequest } from "./apiRequest";

export type User = {
  id: number;
  nickname: string;
  phone: string;
  account?: string;
  wechat?: string;
  status: string;
  created_at?: string;
};

export type LoginResponse = {
  user: User;
  access_token: string;
  access_token_expires_at: string;
  is_new_user: boolean;
};

let restoreInFlight: Promise<LoginResponse> | null = null;

export const authApi = {
  sendCode(phone: string) {
    return apiRequest<{ status: string; cooldown_seconds: number }>("/api/v1/auth/sms/send", {
      method: "POST",
      body: JSON.stringify({ phone })
    });
  },

  login(input: {
    account: string;
    password: string;
  }) {
    return apiRequest<LoginResponse>("/api/v1/auth/login", {
      method: "POST",
      body: JSON.stringify({
        account: input.account,
        password: input.password
      })
    });
  },

  register(input: {
    nickname: string;
    account: string;
    password: string;
    agreementAccepted: boolean;
  }) {
    return apiRequest<LoginResponse>("/api/v1/auth/register", {
      method: "POST",
      body: JSON.stringify({
        nickname: input.nickname,
        account: input.account,
        password: input.password,
        agreement_accepted: input.agreementAccepted
      })
    });
  },

  refresh() {
    return apiRequest<LoginResponse>("/api/v1/auth/refresh", { method: "POST" });
  },

  restore() {
    if (!restoreInFlight) {
      restoreInFlight = apiRequest<LoginResponse>("/api/v1/auth/refresh", {
        method: "POST"
      }).finally(() => {
        restoreInFlight = null;
      });
    }
    return restoreInFlight;
  },

  logout() {
    return apiRequest<void>("/api/v1/auth/logout", { method: "POST" });
  }
};
