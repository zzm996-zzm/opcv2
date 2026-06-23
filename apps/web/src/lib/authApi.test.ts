import { afterEach, describe, expect, it, vi } from "vitest";

import { authApi } from "./authApi";

describe("authApi", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("includes credentials for refresh cookie requests", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          access_token: "next-token",
          access_token_expires_at: new Date().toISOString(),
          user: { id: 42, nickname: "张晨", phone: "13800138000", status: "active" }
        }),
        { status: 200 }
      )
    );

    await authApi.refresh();

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/auth/refresh",
      expect.objectContaining({ method: "POST", credentials: "include" })
    );
  });

  it("posts account password login payload", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          access_token: "access-token",
          access_token_expires_at: "2026-06-18T12:00:00Z",
          is_new_user: false,
          user: { id: 42, nickname: "部署测试", account: "deploy_user", phone: "", status: "active" }
        }),
        { status: 200 }
      )
    );

    await authApi.login({
      account: "deploy_user",
      password: "secret123"
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/auth/login",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({
          account: "deploy_user",
          password: "secret123"
        })
      })
    );
  });

  it("posts account password registration payload", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          access_token: "access-token",
          access_token_expires_at: "2026-06-18T12:00:00Z",
          is_new_user: true,
          user: { id: 42, nickname: "部署测试", account: "deploy_user", phone: "", status: "active" }
        }),
        { status: 200 }
      )
    );

    await authApi.register({
      nickname: "部署测试",
      account: "deploy_user",
      password: "secret123",
      agreementAccepted: true
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/auth/register",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({
          nickname: "部署测试",
          account: "deploy_user",
          password: "secret123",
          agreement_accepted: true
        })
      })
    );
  });

  it("deduplicates concurrent session restore requests", async () => {
    let resolveRequest!: (response: Response) => void;
    const responsePromise = new Promise<Response>((resolve) => {
      resolveRequest = resolve;
    });
    const fetchMock = vi.spyOn(globalThis, "fetch").mockReturnValue(responsePromise);

    const first = authApi.restore();
    const second = authApi.restore();
    expect(fetchMock).toHaveBeenCalledTimes(1);

    resolveRequest(
      new Response(
        JSON.stringify({
          access_token: "access-token",
          access_token_expires_at: "2026-06-11T12:00:00Z",
          is_new_user: false,
          user: {
            id: 7,
            nickname: "张晨",
            phone: "13800138000",
            status: "active"
          }
        }),
        { status: 200 }
      )
    );

    await expect(first).resolves.toEqual(await second);
  });
});
