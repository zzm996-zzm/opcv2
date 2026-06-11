import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "./authSession";
import { analysisApi } from "./analysisApi";

describe("analysisApi", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("starts direction analysis with bearer token", async () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-11T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张晨", phone: "13800138000", status: "active" }
    });
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ session_id: 1, status: "needs_input", questions: [] }), {
        status: 200
      })
    );

    await analysisApi.startDirection({ intent: "想创业" });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/analysis/direction",
      expect.objectContaining({
        method: "POST",
        headers: expect.objectContaining({ Authorization: "Bearer access-token" })
      })
    );
  });
});
