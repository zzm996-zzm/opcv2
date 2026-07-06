import { afterEach, describe, expect, it, vi } from "vitest";

import { accountApi } from "./accountApi";

describe("accountApi", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("reads and updates profile", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ profile: { id: 42, nickname: "张晨" }, bindings: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ profile: { id: 42, nickname: "张晨" }, bindings: [] }), { status: 200 }));

    await accountApi.getProfile();
    await accountApi.updateProfile({ nickname: "张晨", email: "founder@example.com" });

    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/account/profile", expect.any(Object));
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "/api/v1/account/profile",
      expect.objectContaining({
        method: "PATCH",
        body: JSON.stringify({ nickname: "张晨", email: "founder@example.com" })
      })
    );
  });

  it("reads profile context", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({
        user_id: 42,
        completed: true,
        groups: [{ key: "identity", title: "基本身份", fields: { nickname: "张晨" } }]
      }), { status: 200 }));

    const context = await accountApi.getProfileContext();

    expect(context.groups[0].key).toBe("identity");
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/account/profile-context", expect.any(Object));
  });

  it("saves and completes onboarding", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ completed: false, sections: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ completed: false, sections: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ completed: true, sections: [] }), { status: 200 }));

    await accountApi.getOnboarding();
    await accountApi.saveOnboarding({ completed: false, sections: [{ key: "identity", title: "基本身份", fields: { role: "创始人" } }] });
    await accountApi.completeOnboarding();

    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/account/onboarding", expect.any(Object));
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "/api/v1/account/onboarding",
      expect.objectContaining({ method: "PUT" })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      3,
      "/api/v1/account/onboarding/complete",
      expect.objectContaining({ method: "POST" })
    );
  });

  it("reads preferences quotas content and deletes account", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ language: "zh-CN" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ language: "zh-CN" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ quotas: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ items: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ status: "pending_deletion" }), { status: 202 }));

    await accountApi.getPreferences();
    await accountApi.updatePreferences({ default_model: "deepseek" });
    await accountApi.getQuotas();
    await accountApi.listContent(10);
    await accountApi.deleteAccount();

    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/account/preferences", expect.objectContaining({ method: "PATCH" }));
    expect(fetchMock).toHaveBeenNthCalledWith(4, "/api/v1/account/content?limit=10", expect.any(Object));
    expect(fetchMock).toHaveBeenNthCalledWith(5, "/api/v1/account", expect.objectContaining({ method: "DELETE" }));
  });
});
