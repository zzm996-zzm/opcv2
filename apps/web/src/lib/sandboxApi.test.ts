import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "./authSession";
import { sandboxApi } from "./sandboxApi";

describe("sandboxApi V1.2", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  it("creates, clarifies, selects roles, and starts a V1.2 run", async () => {
    authSession.set({ access_token: "access-token", access_token_expires_at: "2026-08-30T12:00:00Z", is_new_user: false, user: { id: 7, nickname: "张晨", phone: "", status: "active" } });
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation(() => Promise.resolve(new Response(JSON.stringify({ id: 99, status: "ready", revision: 2 }), { status: 200 })));

    await sandboxApi.createRun({ name: "低卡代餐", product: { name: "低卡代餐", price_cents: 3900 }, context: { target_customer: "白领" } });
    await sandboxApi.answerRun(99, { revision: 1, answers: [{ key: "channel", value: "小红书" }] });
    await sandboxApi.setRoles(99, { revision: 2, roles: ["customer", "investor", "skeptic"] });
    await sandboxApi.startRun(99);

    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/sandbox-runs", expect.objectContaining({ method: "POST", headers: expect.objectContaining({ Authorization: "Bearer access-token" }) }));
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/sandbox-runs/99/answer", expect.objectContaining({ body: JSON.stringify({ revision: 1, answers: [{ key: "channel", value: "小红书" }] }) }));
    expect(fetchMock).toHaveBeenNthCalledWith(3, "/api/v1/sandbox-runs/99/roles", expect.objectContaining({ body: JSON.stringify({ revision: 2, roles: ["customer", "investor", "skeptic"] }) }));
    expect(fetchMock).toHaveBeenNthCalledWith(4, "/api/v1/sandbox-runs/99/start", expect.objectContaining({ method: "POST" }));
  });

  it("loads home, roles, a run, and filtered paginated history", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation(() => Promise.resolve(new Response(JSON.stringify({ runs: [] }), { status: 200 })));

    await sandboxApi.getHome();
    await sandboxApi.listRoles();
    await sandboxApi.getRun(99);
    await sandboxApi.listRuns({ page: 2, limit: 20, status: "done", product: "AI 客服" });

    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/sandbox/home", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/sandbox-runs/roles", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(3, "/api/v1/sandbox-runs/99", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(4, "/api/v1/sandbox-runs?page=2&limit=20&status=done&product=AI+%E5%AE%A2%E6%9C%8D", expect.objectContaining({ method: "GET" }));
  });

  it("stops, retries one role, renames, and deletes a run", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation(() => Promise.resolve(new Response(JSON.stringify({ id: 99 }), { status: 200 })));

    await sandboxApi.stopRun(99);
    await sandboxApi.retryRole(99, "skeptic");
    await sandboxApi.renameRun(99, { name: "新名称", revision: 4 });
    await sandboxApi.deleteRun(99);

    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/sandbox-runs/99/stop", expect.objectContaining({ method: "POST" }));
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/sandbox-runs/99/roles/skeptic/retry", expect.objectContaining({ method: "POST" }));
    expect(fetchMock).toHaveBeenNthCalledWith(3, "/api/v1/sandbox-runs/99", expect.objectContaining({ method: "PATCH", body: JSON.stringify({ name: "新名称", revision: 4 }) }));
    expect(fetchMock).toHaveBeenNthCalledWith(4, "/api/v1/sandbox-runs/99", expect.objectContaining({ method: "DELETE" }));
  });

  it("generates a report and uses authenticated SSE replay", async () => {
    authSession.set({ access_token: "access-token", access_token_expires_at: "2026-08-30T12:00:00Z", is_new_user: false, user: { id: 7, nickname: "张晨", phone: "", status: "active" } });
    const stream = ["id: 8", "event: role_done", "data: {\"id\":8,\"run_id\":99,\"role\":\"customer\",\"event\":\"role_done\"}", "", "id: 9", "event: run_done", "data: {\"id\":9,\"run_id\":99,\"event\":\"run_done\"}", ""].join("\n");
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ summary: "可进入验证" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(stream, { status: 200, headers: { "Content-Type": "text/event-stream" } }));
    const events: { id: number; event: string }[] = [];

    await sandboxApi.generateReport(99);
    await sandboxApi.streamRun(99, 7, (event) => events.push({ id: event.id, event: event.event }));

    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/sandbox-runs/99/stream", expect.objectContaining({
      method: "GET",
      headers: expect.objectContaining({ Authorization: "Bearer access-token", "Last-Event-ID": "7" })
    }));
    expect(events).toEqual([{ id: 8, event: "role_done" }, { id: 9, event: "run_done" }]);
  });

  it("creates and downloads a real authenticated PDF export", async () => {
    authSession.set({ access_token: "access-token", access_token_expires_at: "2026-08-30T12:00:00Z", is_new_user: false, user: { id: 7, nickname: "张晨", phone: "", status: "active" } });
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 700, run_id: 99, format: "pdf", download_url: "/api/v1/sandbox-runs/99/exports/700/download" }), { status: 201 }))
      .mockResolvedValueOnce(new Response("%PDF-1.7\nrendered", { status: 200, headers: { "Content-Type": "application/pdf" } }));

    const created = await sandboxApi.createExport(99, "pdf");
    const blob = await sandboxApi.downloadExport(created.download_url);

    expect(await blob.text()).toContain("%PDF-1.7");
    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/sandbox-runs/99/report/export", expect.objectContaining({ body: JSON.stringify({ format: "pdf" }) }));
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/sandbox-runs/99/exports/700/download", expect.objectContaining({ headers: expect.objectContaining({ Authorization: "Bearer access-token" }) }));
  });

  it("uses server workflows for follow-up, tasks, growth, and analytics", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation(() => Promise.resolve(new Response(JSON.stringify({}), { status: 200 })));

    await sandboxApi.listFollowUps(99);
    await sandboxApi.askRole(99, { role_code: "investor", question: "最先验证什么？" });
    await sandboxApi.createTasks(99, { advice_indexes: [0, 2] });
    await sandboxApi.createGrowthHandoff(99);
    await sandboxApi.track({ event: "sandbox_report_view", event_id: "evt-1", run_id: 99, properties: { status: "done" } });

    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/sandbox-runs/99/follow-ups", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/sandbox-runs/99/follow-ups", expect.objectContaining({ method: "POST" }));
    expect(fetchMock).toHaveBeenNthCalledWith(3, "/api/v1/sandbox-runs/99/report/tasks", expect.objectContaining({ body: JSON.stringify({ advice_indexes: [0, 2] }) }));
    expect(fetchMock).toHaveBeenNthCalledWith(4, "/api/v1/sandbox-runs/99/report/growth-handoff", expect.objectContaining({ method: "POST" }));
    const analyticsRequest = fetchMock.mock.calls[4]?.[1];
    expect(fetchMock.mock.calls[4]?.[0]).toBe("/api/v1/sandbox/analytics");
    expect(JSON.parse(String(analyticsRequest?.body))).toEqual(expect.objectContaining({ event: "sandbox_report_view", event_id: "evt-1", run_id: 99, route: "/", visitor_key: expect.stringMatching(/^visitor-/), properties: { status: "done" } }));
  });
});
