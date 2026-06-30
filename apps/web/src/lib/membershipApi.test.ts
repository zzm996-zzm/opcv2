import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "./authSession";
import { membershipApi } from "./membershipApi";

describe("membershipApi", () => {
  afterEach(() => {
    authSession.clear();
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("loads the current membership with bearer access token", async () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-11T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张晨", phone: "13800138000", status: "active" }
    });
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          plan: { code: "free", name: "免费版", monthly_analysis_limit: 3, lead_export_limit: 0 },
          credit_balance: 25
        }),
        { status: 200 }
      )
    );

    const snapshot = await membershipApi.current();

    expect(snapshot.credit_balance).toBe(25);
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/membership/me",
      expect.objectContaining({
        headers: expect.objectContaining({ Authorization: "Bearer access-token" })
      })
    );
  });

  it("uses configured api base url", async () => {
    vi.stubEnv("VITE_API_BASE_URL", "https://api.example.com");
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          plan: { code: "free", name: "免费版", monthly_analysis_limit: 3, lead_export_limit: 0 },
          credit_balance: 25
        }),
        { status: 200 }
      )
    );

    await membershipApi.current();

    expect(fetchMock).toHaveBeenCalledWith(
      "https://api.example.com/api/v1/membership/me",
      expect.any(Object)
    );
  });
});
