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
