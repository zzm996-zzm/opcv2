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

    await tasksApi.createTask({ title: "整理客户名单", project: "AI线索开发", priority: "high", tools: ["CRM"], learning: "线索评分", idempotencyKey: "task-99" });
    await tasksApi.listTasks({ status: "in_progress", project: "商业沙盘", q: "接口", limit: 10 });
    await tasksApi.stats();
    await tasksApi.updateTask(99, { status: "completed", version: 4 });

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "/api/v1/tasks",
      expect.objectContaining({
        method: "POST",
        headers: expect.objectContaining({ "Idempotency-Key": "task-99" }),
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
      expect.objectContaining({ method: "PATCH", body: JSON.stringify({ status: "completed", version: 4 }) })
    );
  });

  it("passes sort and group context and lists task activities", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ tasks: [], total: 0 }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ activities: [], total: 0 }), { status: 200 }));

    await tasksApi.listTasks({ sort: "due_at", group: "project", limit: 20 });
    await tasksApi.listTaskActivities(99, 20, 20);

    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/tasks?sort=due_at&group=project&limit=20", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/tasks/99/activities?limit=20&offset=20", expect.objectContaining({ method: "GET" }));
  });

  it("loads calendar ranges, batch edits, and task comments", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ tasks: [], unscheduled: [], total: 0 }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ updated: 1, failed: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ comments: [], total: 0 }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 12, content: "记录进展" }), { status: 201 }));

    await tasksApi.calendar({ from: "2026-08-01", to: "2026-09-01", q: "客户" });
    await tasksApi.batchUpdate({ ids: [8], assignee: "李明", tags: ["客户"] });
    await tasksApi.listTaskComments(8);
    await tasksApi.createTaskComment(8, "记录进展");

    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/tasks/calendar?from=2026-08-01&to=2026-09-01&q=%E5%AE%A2%E6%88%B7", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/tasks/batch-update", expect.objectContaining({
      method: "PATCH",
      body: JSON.stringify({ ids: [8], assignee: "李明", tags: ["客户"] })
    }));
    expect(fetchMock).toHaveBeenNthCalledWith(3, "/api/v1/tasks/8/comments?limit=20", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(4, "/api/v1/tasks/8/comments", expect.objectContaining({ method: "POST", body: JSON.stringify({ content: "记录进展", parent_comment_id: undefined }) }));
  });

  it("uploads, signs, lists, and deletes task attachments", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ attachments: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 3, name: "notes.txt" }), { status: 201 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ attachment: { id: 3 }, url: "/api/v1/tasks/8/attachments/3/download?expires=1&signature=x", expires_at: "2026-08-15T10:00:00Z" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(null, { status: 204 }));

    const file = new File(["notes"], "notes.txt", { type: "text/plain" });
    await tasksApi.listTaskAttachments(8);
    await tasksApi.uploadTaskAttachment(8, file);
    await tasksApi.getTaskAttachmentURL(8, 3);
    await tasksApi.deleteTaskAttachment(8, 3);

    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/tasks/8/attachments", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/tasks/8/attachments", expect.objectContaining({ method: "POST", body: expect.any(FormData) }));
    expect(fetchMock).toHaveBeenNthCalledWith(3, "/api/v1/tasks/8/attachments/3/url", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(4, "/api/v1/tasks/8/attachments/3", expect.objectContaining({ method: "DELETE" }));
  });

  it("restores a soft-deleted task", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(new Response(null, { status: 204 }));

    await tasksApi.restoreTask(99);

    expect(fetchMock).toHaveBeenCalledWith("/api/v1/tasks/99/restore", expect.objectContaining({ method: "POST" }));
  });

  it("generates a task plan from a goal", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({ tasks: [{ id: 101, title: "整理访谈名单" }] }), { status: 200 })
    );

    await tasksApi.generateTasks("验证教培客户需求", {
      sourceType: "learning_diagnosis",
      sourceId: 99,
      sourceTitle: "企业AI落地能力路径",
      sourceUrl: "/learning/plan"
    });

    expect(fetchMock).toHaveBeenCalledWith("/api/v1/tasks/generate", expect.objectContaining({
      method: "POST",
      body: JSON.stringify({
        goal: "验证教培客户需求",
        source_type: "learning_diagnosis",
        source_id: 99,
        source_title: "企业AI落地能力路径",
        source_url: "/learning/plan"
      })
    }));
  });

  it("adopts selected AI draft tasks with an idempotency key", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(new Response(JSON.stringify({ tasks: [{ id: 101, title: "整理访谈名单" }] }), { status: 201 }));

    await tasksApi.adoptTaskAIDraft(77, [{
      draftIndex: 0,
      title: "整理访谈名单",
      project: "客户验证",
      priority: "high",
      tags: ["访谈"],
      tools: ["CRM"]
    }], "task-ai-draft-77");

    expect(fetchMock).toHaveBeenCalledWith("/api/v1/tasks/ai-drafts/77/adopt", expect.objectContaining({
      method: "POST",
      headers: expect.objectContaining({ "Idempotency-Key": "task-ai-draft-77" }),
      body: JSON.stringify({
        tasks: [{
          draft_index: 0,
          title: "整理访谈名单",
          description: undefined,
          assignee: undefined,
          project: "客户验证",
          priority: "high",
          tags: ["访谈"],
          due_at: undefined,
          tools: ["CRM"],
          learning: undefined
        }]
      })
    }));
  });

  it("updates statuses and deletes tasks in batches", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ updated: 2 }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ deleted: 2 }), { status: 200 }));

    await tasksApi.batchUpdateStatus([41, 42], "completed");
    await tasksApi.batchDelete([41, 42]);

    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/tasks/batch", expect.objectContaining({
      method: "PATCH",
      body: JSON.stringify({ ids: [41, 42], status: "completed" })
    }));
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/tasks/batch", expect.objectContaining({
      method: "DELETE",
      body: JSON.stringify({ ids: [41, 42] })
    }));
  });

  it("creates a task linked to its source", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({ id: 101, title: "反击竞品更新" }), { status: 200 })
    );

    await tasksApi.createTask({
      title: "反击竞品更新",
      project: "竞品动态监测",
      priority: "high",
      tools: [],
      learning: "",
      sourceType: "competitor_scan",
      sourceId: 11,
      sourceTitle: "销售自动化提速",
      sourceUrl: "/competitor-data"
    });

    expect(fetchMock).toHaveBeenCalledWith("/api/v1/tasks", expect.objectContaining({
      method: "POST",
      body: JSON.stringify({
        title: "反击竞品更新",
        project: "竞品动态监测",
        priority: "high",
        tools: [],
        learning: "",
        source_type: "competitor_scan",
        source_id: 11,
        source_title: "销售自动化提速",
        source_url: "/competitor-data"
      })
    }));
  });

	it("lists, creates, updates, and deletes task subtasks", async () => {
	const fetchMock = vi.spyOn(globalThis, "fetch")
	  .mockResolvedValueOnce(new Response(JSON.stringify({ subtasks: [] }), { status: 200 }))
	  .mockResolvedValueOnce(new Response(JSON.stringify({ id: 7, title: "整理访谈提纲" }), { status: 200 }))
	  .mockResolvedValueOnce(new Response(JSON.stringify({ id: 7, completed: true }), { status: 200 }))
	  .mockResolvedValueOnce(new Response(null, { status: 204 }));

	await tasksApi.listSubtasks(99);
	await tasksApi.createSubtask(99, { title: "整理访谈提纲", assignee: "李明", parentSubtaskId: 6 });
	await tasksApi.updateSubtask(99, 7, { completed: true });
	await tasksApi.deleteSubtask(99, 7);

	expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/tasks/99/subtasks", expect.objectContaining({ method: "GET" }));
	expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/tasks/99/subtasks", expect.objectContaining({
	  method: "POST",
	  body: JSON.stringify({ title: "整理访谈提纲", assignee: "李明", parent_subtask_id: 6 })
	}));
	expect(fetchMock).toHaveBeenNthCalledWith(3, "/api/v1/tasks/99/subtasks/7", expect.objectContaining({
	  method: "PATCH",
	  body: JSON.stringify({ completed: true })
	}));
	expect(fetchMock).toHaveBeenNthCalledWith(4, "/api/v1/tasks/99/subtasks/7", expect.objectContaining({ method: "DELETE" }));
	});

	it("loads and saves list field preferences", async () => {
		const fetchMock = vi.spyOn(globalThis, "fetch")
			.mockResolvedValueOnce(new Response(JSON.stringify({ view: "list", columns: ["title", "status"], updated_at: "2026-08-15T10:00:00Z" }), { status: 200 }))
			.mockResolvedValueOnce(new Response(JSON.stringify({ view: "list", columns: ["title", "status", "assignee"], updated_at: "2026-08-15T10:01:00Z" }), { status: 200 }));

		await tasksApi.getViewPreference();
		await tasksApi.saveViewPreference(["title", "status", "assignee"]);

		expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/tasks/view-preferences?view=list", expect.objectContaining({ method: "GET" }));
		expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/tasks/view-preferences", expect.objectContaining({
			method: "PUT",
			body: JSON.stringify({ view: "list", columns: ["title", "status", "assignee"] })
		}));
	});

  it("gets, upserts, and deletes a task reminder", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ reminder: null }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 8, remind_at: "2026-07-18T10:00:00Z" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(null, { status: 204 }));

    await tasksApi.getReminder(99);
    await tasksApi.upsertReminder(99, { remindAt: "2026-07-18T10:00:00Z", recurrence: "weekly" });
    await tasksApi.deleteReminder(99);

    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/tasks/99/reminder", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/tasks/99/reminder", expect.objectContaining({
      method: "PUT",
      body: JSON.stringify({ remind_at: "2026-07-18T10:00:00Z", recurrence: "weekly" })
    }));
    expect(fetchMock).toHaveBeenNthCalledWith(3, "/api/v1/tasks/99/reminder", expect.objectContaining({ method: "DELETE" }));
  });
});
