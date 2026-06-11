export type User = {
  id: number;
  nickname: string;
  phone: string;
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

async function request<T>(path: string, init: RequestInit): Promise<T> {
  const response = await fetch(path, {
    ...init,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...init.headers
    }
  });
  if (!response.ok) {
    const payload = (await response.json().catch(() => ({}))) as { error?: string };
    throw new Error(payload.error ?? "request_failed");
  }
  if (response.status === 204) {
    return undefined as T;
  }
  return (await response.json()) as T;
}

let restoreInFlight: Promise<LoginResponse> | null = null;

export const authApi = {
  sendCode(phone: string) {
    return request<{ status: string; cooldown_seconds: number }>("/api/v1/auth/sms/send", {
      method: "POST",
      body: JSON.stringify({ phone })
    });
  },

  login(input: {
    nickname: string;
    phone: string;
    code: string;
    agreementAccepted: boolean;
  }) {
    return request<LoginResponse>("/api/v1/auth/login", {
      method: "POST",
      body: JSON.stringify({
        nickname: input.nickname,
        phone: input.phone,
        code: input.code,
        agreement_accepted: input.agreementAccepted
      })
    });
  },

  refresh() {
    return request<LoginResponse>("/api/v1/auth/refresh", { method: "POST" });
  },

  restore() {
    if (!restoreInFlight) {
      restoreInFlight = request<LoginResponse>("/api/v1/auth/refresh", {
        method: "POST"
      }).finally(() => {
        restoreInFlight = null;
      });
    }
    return restoreInFlight;
  },

  logout() {
    return request<void>("/api/v1/auth/logout", { method: "POST" });
  }
};
