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
      .mockResolvedValueOnce(new Response(JSON.stringify({ total: 1, todo: 0, in_progress: 1, completed: 0, reminder: 0, overdue: 0 }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 99, status: "completed" }), { status: 200 }));

    await tasksApi.createTask({ title: "整理客户名单", project: "AI线索开发", priority: "high", tools: ["CRM"], learning: "线索评分" });
    await tasksApi.listTasks({ status: "in_progress", project: "商业沙盘", q: "接口", limit: 10 });
    await tasksApi.stats();
    await tasksApi.updateTask(99, { status: "completed" });

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "/api/v1/tasks",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ title: "整理客户名单", project: "AI线索开发", priority: "high", tools: ["CRM"], learning: "线索评分" })
      })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "/api/v1/tasks?status=in_progress&project=%E5%95%86%E4%B8%9A%E6%B2%99%E7%9B%98&q=%E6%8E%A5%E5%8F%A3&limit=10",
      expect.objectContaining({ method: "GET" })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(3, "/api/v1/tasks/stats", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(
      4,
      "/api/v1/tasks/99",
      expect.objectContaining({ method: "PATCH", body: JSON.stringify({ status: "completed" }) })
    );
  });
});
