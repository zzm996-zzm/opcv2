import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "./authSession";
import { leadsApi } from "./leadsApi";

describe("leadsApi", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  it("creates lead task with bearer token and idempotency key", async () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-25T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张晨", phone: "", status: "active" }
    });
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        id: 99,
        user_id: 7,
        query: "成都 教培",
        status: "queued",
        idempotency_key: "lead-task-99",
        credit_cost: 1,
        created_at: "2026-06-25T12:00:00Z",
        updated_at: "2026-06-25T12:00:00Z"
      }), { status: 200 })
    );

    await leadsApi.createTask({ query: "成都 教培", idempotencyKey: "lead-task-99" });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/leads/tasks",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ query: "成都 教培", idempotency_key: "lead-task-99" }),
        headers: expect.objectContaining({ Authorization: "Bearer access-token" })
      })
    );
  });

  it("lists lead tasks with limit", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ tasks: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        task: { id: 99, query: "成都 教培", status: "succeeded" },
        progress_percent: 100,
        message: "线索采集已完成。",
        results_count: 1
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ results: [] }), { status: 200 }));

    await leadsApi.listTasks(10);
    await leadsApi.getTask(99);
    await leadsApi.listResults(99, 10);

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/leads/tasks?limit=10",
      expect.objectContaining({ method: "GET" })
    );
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/leads/tasks/99",
      expect.objectContaining({ method: "GET" })
    );
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/leads/tasks/99/results?limit=10",
      expect.objectContaining({ method: "GET" })
    );
  });
});
