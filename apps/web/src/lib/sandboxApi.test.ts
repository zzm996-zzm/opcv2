import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "./authSession";
import { sandboxApi } from "./sandboxApi";

describe("sandboxApi", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  it("creates and runs a sandbox session with bearer token", async () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-30T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张晨", phone: "", status: "active" }
    });
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 99, status: "draft" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 99, status: "completed", report: { score: 83 } }), { status: 200 }));

    await sandboxApi.createSession({ goal: "验证项目", targetUsers: "教培机构", product: "AI客服", roles: ["用户"] });
    await sandboxApi.runSession(99);

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "/api/v1/sandbox/sessions",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ goal: "验证项目", target_users: "教培机构", product: "AI客服", roles: ["用户"] }),
        headers: expect.objectContaining({ Authorization: "Bearer access-token" })
      })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "/api/v1/sandbox/sessions/99/run",
      expect.objectContaining({ method: "POST" })
    );
  });
});
