import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "./authSession";
import { tasksApi } from "./tasksApi";

describe("tasksApi", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  it("creates, lists, and updates tasks", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 99, title: "整理客户名单" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ tasks: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 99, status: "completed" }), { status: 200 }));

    await tasksApi.createTask({ title: "整理客户名单", project: "AI线索开发", priority: "high", tools: ["CRM"], learning: "线索评分" });
    await tasksApi.listTasks(10);
    await tasksApi.updateTask(99, { status: "completed" });

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "/api/v1/tasks",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ title: "整理客户名单", project: "AI线索开发", priority: "high", tools: ["CRM"], learning: "线索评分" })
      })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/tasks?limit=10", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(
      3,
      "/api/v1/tasks/99",
      expect.objectContaining({ method: "PATCH", body: JSON.stringify({ status: "completed" }) })
    );
  });
});
