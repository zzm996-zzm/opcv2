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

  it("lists analysis sessions with limit", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ sessions: [] }), { status: 200 })
    );

    await analysisApi.listSessions(12);

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/analysis/sessions?limit=12",
      expect.objectContaining({ method: "GET" })
    );
  });

  it("gets one analysis session", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ id: 99, user_id: 7, intent: "想创业", status: "completed" }), { status: 200 })
    );

    await analysisApi.getSession(99);

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/analysis/sessions/99",
      expect.objectContaining({ method: "GET" })
    );
  });

  it("lists and updates action items", async () => {
    const fetchMock = vi
      .spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ items: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 7, completed: true }), { status: 200 }));

    await analysisApi.listActionItems(99);
    await analysisApi.updateActionItem(99, 7, true);

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "/api/v1/analysis/sessions/99/action-items",
      expect.objectContaining({ method: "GET" })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "/api/v1/analysis/sessions/99/action-items/7",
      expect.objectContaining({
        method: "PATCH",
        body: JSON.stringify({ completed: true })
      })
    );
  });
});
