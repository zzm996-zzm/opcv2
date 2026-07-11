import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import TasksPage from "./TasksPage";

describe("TasksPage", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  function signIn() {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-11T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张晨", phone: "13800138000", status: "active" }
    });
  }

  function renderTasksPage() {
    signIn();
    render(
      <MemoryRouter>
        <TasksPage />
      </MemoryRouter>
    );
  }

  it("renders the task overview and an empty state without backend rows", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/tasks?limit=20") {
        return Promise.resolve(new Response(JSON.stringify({ tasks: [] }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/stats") {
        return Promise.resolve(new Response(JSON.stringify({ total: 0, todo: 0, in_progress: 0, completed: 0, reminder: 0, overdue: 0 }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderTasksPage();

    expect(screen.getByRole("heading", { name: "任务中心" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "生成任务表" })).toBeInTheDocument();
    expect(await screen.findByText("暂无任务数据")).toBeInTheDocument();
    expect(screen.queryByText("完成智能客服系统项目商业画布")).not.toBeInTheDocument();
  });

  it("links the primary new task action to the manual creation page", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/tasks?limit=20") {
        return Promise.resolve(new Response(JSON.stringify({ tasks: [] }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/stats") {
        return Promise.resolve(new Response(JSON.stringify({ total: 0, todo: 0, in_progress: 0, completed: 0, reminder: 0, overdue: 0 }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderTasksPage();

    expect(screen.getByRole("link", { name: "新建任务" })).toHaveAttribute("href", "/tasks/new");
    expect(await screen.findByText("暂无任务数据")).toBeInTheDocument();
  });

  it("loads tasks from API", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/tasks?limit=20") {
        return Promise.resolve(new Response(JSON.stringify({
          tasks: [
            {
              id: 41,
              user_id: 7,
              title: "联调商业沙盘接口",
              project: "商业沙盘",
              status: "in_progress",
              priority: "high",
              due_at: "2026-06-30T10:00:00Z",
              tools: ["沙盘推演", "任务中心"],
              learning: "后端接口联调",
              created_at: "2026-06-29T10:00:00Z",
              updated_at: "2026-06-30T09:00:00Z"
            }
          ]
        }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/stats") {
        return Promise.resolve(new Response(JSON.stringify({ total: 3, todo: 1, in_progress: 1, completed: 1, reminder: 0, overdue: 1 }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderTasksPage();

    expect(await screen.findByRole("heading", { name: "联调商业沙盘接口" })).toBeInTheDocument();
    expect(screen.getByText("商业沙盘 · 负责人 未指定 · 截止 06/30 18:00")).toBeInTheDocument();
    expect(screen.getByText("建议工具：沙盘推演 / 任务中心")).toBeInTheDocument();
    const inProgressStat = screen.getAllByText("进行中").find((node) => node.tagName.toLowerCase() === "small")?.closest("article");
    expect(inProgressStat).not.toBeNull();
    expect(within(inProgressStat as HTMLElement).getByText("1")).toBeInTheDocument();
  });

  it("filters tasks by backend status", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/tasks?limit=20") {
        return Promise.resolve(new Response(JSON.stringify({ tasks: [] }), { status: 200 }));
      }
      if (url === "/api/v1/tasks?status=in_progress&limit=20") {
        return Promise.resolve(new Response(JSON.stringify({
          tasks: [{
            id: 42,
            user_id: 7,
            title: "筛选后的进行中任务",
            project: "任务中心",
            status: "in_progress",
            priority: "medium",
            due_at: "2026-07-02T10:00:00Z",
            tools: ["任务中心"],
            learning: "任务筛选",
            created_at: "2026-07-02T09:00:00Z",
            updated_at: "2026-07-02T09:30:00Z"
          }]
        }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/stats") {
        return Promise.resolve(new Response(JSON.stringify({ total: 1, todo: 0, in_progress: 1, completed: 0, reminder: 0, overdue: 0 }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderTasksPage();

    fireEvent.click(await screen.findByRole("button", { name: "筛选进行中" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/tasks?status=in_progress&limit=20",
      expect.objectContaining({ method: "GET" })
    ));
    expect(await screen.findByRole("heading", { name: "筛选后的进行中任务" })).toBeInTheDocument();
  });

  it("shows backend list errors without rendering fallback tasks", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ error: "invalid_limit" }), { status: 400 })
    );

    renderTasksPage();

    expect(await screen.findByText("列表数量参数有误")).toBeInTheDocument();
    expect(screen.getByText("暂无任务数据")).toBeInTheDocument();
    expect(screen.queryByText("完成智能客服系统项目商业画布")).not.toBeInTheDocument();
  });

  it("updates task status from the task row action", async () => {
    let completed = false;
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/tasks?limit=20") {
        return Promise.resolve(new Response(JSON.stringify({
          tasks: [{
            id: 41,
            user_id: 7,
            title: "联调商业沙盘接口",
            project: "商业沙盘",
            status: "in_progress",
            priority: "high",
            due_at: "2026-06-30T10:00:00Z",
            tools: ["沙盘推演", "任务中心"],
            learning: "后端接口联调",
            created_at: "2026-06-29T10:00:00Z",
            updated_at: "2026-06-30T09:00:00Z"
          }]
        }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/stats") {
        return Promise.resolve(new Response(JSON.stringify({
          total: 1,
          todo: 0,
          in_progress: completed ? 0 : 1,
          completed: completed ? 1 : 0,
          reminder: 0,
          overdue: 0
        }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/41" && init?.method === "PATCH") {
        completed = true;
        return Promise.resolve(new Response(JSON.stringify({
          id: 41,
          user_id: 7,
          title: "联调商业沙盘接口",
          project: "商业沙盘",
          status: "completed",
          priority: "high",
          due_at: "2026-06-30T10:00:00Z",
          tools: ["沙盘推演", "任务中心"],
          learning: "后端接口联调",
          created_at: "2026-06-29T10:00:00Z",
          updated_at: "2026-06-30T10:00:00Z"
        }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderTasksPage();

    const taskHeading = await screen.findByRole("heading", { name: "联调商业沙盘接口" });
    const taskRow = taskHeading.closest("article") as HTMLElement;
    fireEvent.click(within(taskRow).getByRole("button", { name: "标记完成" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        "/api/v1/tasks/41",
        expect.objectContaining({ method: "PATCH", body: JSON.stringify({ status: "completed" }) })
      );
    });
    expect(within(taskRow).getByRole("button", { name: "重新打开" })).toBeEnabled();
    const completedStat = screen.getAllByText("已完成").find((node) => node.tagName.toLowerCase() === "small")?.closest("article");
    expect(completedStat).not.toBeNull();
    expect(within(completedStat as HTMLElement).getByText("1")).toBeInTheDocument();
  });

  it("starts todo tasks and removes transitioned tasks from active filters", async () => {
    let status = "todo";
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/tasks?limit=20") {
        return Promise.resolve(new Response(JSON.stringify({ tasks: [] }), { status: 200 }));
      }
      if (url === "/api/v1/tasks?status=todo&limit=20") {
        return Promise.resolve(new Response(JSON.stringify({
          tasks: [{
            id: 51,
            user_id: 7,
            title: "准备首轮客户访谈",
            project: "客户验证",
            status: "todo",
            priority: "high",
            tools: ["CRM"],
            learning: "客户访谈",
            created_at: "2026-07-09T09:00:00Z",
            updated_at: "2026-07-09T09:00:00Z"
          }]
        }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/stats") {
        return Promise.resolve(new Response(JSON.stringify({
          total: 1,
          todo: status === "todo" ? 1 : 0,
          in_progress: status === "in_progress" ? 1 : 0,
          completed: 0,
          reminder: 0,
          overdue: 0
        }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/51" && init?.method === "PATCH") {
        status = "in_progress";
        return Promise.resolve(new Response(JSON.stringify({
          id: 51,
          user_id: 7,
          title: "准备首轮客户访谈",
          project: "客户验证",
          status,
          priority: "high",
          tools: ["CRM"],
          learning: "客户访谈",
          created_at: "2026-07-09T09:00:00Z",
          updated_at: "2026-07-09T10:00:00Z"
        }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderTasksPage();
    fireEvent.click(await screen.findByRole("button", { name: "筛选待开始" }));
    const taskHeading = await screen.findByRole("heading", { name: "准备首轮客户访谈" });
    fireEvent.click(within(taskHeading.closest("article") as HTMLElement).getByRole("button", { name: "开始任务" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/tasks/51",
      expect.objectContaining({ method: "PATCH", body: JSON.stringify({ status: "in_progress" }) })
    ));
    expect(await screen.findByText("暂无任务数据")).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "准备首轮客户访谈" })).not.toBeInTheDocument();
  });

  it("shows task update errors from the row action", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/tasks?limit=20") {
        return Promise.resolve(new Response(JSON.stringify({
          tasks: [{
            id: 41,
            user_id: 7,
            title: "联调商业沙盘接口",
            project: "商业沙盘",
            status: "in_progress",
            priority: "high",
            due_at: "2026-06-30T10:00:00Z",
            tools: ["沙盘推演", "任务中心"],
            learning: "后端接口联调",
            created_at: "2026-06-29T10:00:00Z",
            updated_at: "2026-06-30T09:00:00Z"
          }]
        }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/stats") {
        return Promise.resolve(new Response(JSON.stringify({ total: 1, todo: 0, in_progress: 1, completed: 0, reminder: 0, overdue: 0 }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/41" && init?.method === "PATCH") {
        return Promise.resolve(new Response(JSON.stringify({ error: "request_failed" }), { status: 500 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderTasksPage();

    const taskHeading = await screen.findByRole("heading", { name: "联调商业沙盘接口" });
    const taskRow = taskHeading.closest("article") as HTMLElement;
    fireEvent.click(within(taskRow).getByRole("button", { name: "标记完成" }));

    expect(await screen.findByText("请求失败，请稍后重试")).toBeInTheDocument();
    expect(within(taskRow).getByText("进行中")).toBeInTheDocument();
  });

  it("creates a task from the goal input", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/tasks?limit=20") {
        return Promise.resolve(new Response(JSON.stringify({ tasks: [] }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/stats") {
        return Promise.resolve(new Response(JSON.stringify({ total: 0, todo: 0, in_progress: 0, completed: 0, reminder: 0, overdue: 0 }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/generate" && init?.method === "POST") {
        return Promise.resolve(new Response(JSON.stringify({ tasks: [
          {
            id: 88,
            user_id: 7,
            title: "梳理竞品反击动作",
            project: "竞品应对",
            status: "todo",
            priority: "high",
            tools: ["竞争对手数据"],
            learning: "竞品分析",
            created_at: "2026-07-07T10:00:00Z",
            updated_at: "2026-07-07T10:00:00Z"
          },
          {
            id: 89,
            user_id: 7,
            title: "安排客户验证",
            project: "竞品应对",
            status: "todo",
            priority: "medium",
            tools: ["CRM"],
            learning: "客户访谈",
            created_at: "2026-07-07T10:00:00Z",
            updated_at: "2026-07-07T10:00:00Z"
          }
        ] }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderTasksPage();

    fireEvent.change(screen.getByLabelText("描述任务目标"), { target: { value: "梳理竞品反击动作" } });
    fireEvent.click(screen.getByRole("button", { name: "生成任务表" }));

    expect(await screen.findByText("已生成 2 条任务")).toBeInTheDocument();
    expect(await screen.findByRole("heading", { name: "梳理竞品反击动作" })).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/tasks/generate",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ goal: "梳理竞品反击动作" })
      })
    );
  });

  it("keeps created tasks out of the list when they do not match the active status filter", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/tasks?limit=20") {
        return Promise.resolve(new Response(JSON.stringify({ tasks: [] }), { status: 200 }));
      }
      if (url === "/api/v1/tasks?status=completed&limit=20") {
        return Promise.resolve(new Response(JSON.stringify({ tasks: [] }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/stats") {
        return Promise.resolve(new Response(JSON.stringify({ total: 0, todo: 0, in_progress: 0, completed: 0, reminder: 0, overdue: 0 }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/generate" && init?.method === "POST") {
        return Promise.resolve(new Response(JSON.stringify({ tasks: [{
          id: 88,
          user_id: 7,
          title: "梳理竞品反击动作",
          project: "任务中心",
          status: "todo",
          priority: "medium",
          tools: ["任务中心"],
          learning: "梳理竞品反击动作",
          created_at: "2026-07-07T10:00:00Z",
          updated_at: "2026-07-07T10:00:00Z"
        }] }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderTasksPage();

    fireEvent.click(await screen.findByRole("button", { name: "筛选已完成" }));
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/tasks?status=completed&limit=20",
      expect.objectContaining({ method: "GET" })
    ));
    fireEvent.change(screen.getByLabelText("描述任务目标"), { target: { value: "梳理竞品反击动作" } });
    fireEvent.click(screen.getByRole("button", { name: "生成任务表" }));

    expect(await screen.findByText("已生成 1 条任务")).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "梳理竞品反击动作" })).not.toBeInTheDocument();
    expect(screen.getByText("暂无任务数据")).toBeInTheDocument();
  });

  it("switches to a full task board grouped by every status", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/tasks?limit=20") {
        return Promise.resolve(new Response(JSON.stringify({
          tasks: [
            {
              id: 91,
              user_id: 7,
              title: "整理客户访谈提纲",
              project: "客户验证",
              status: "todo",
              priority: "high",
              tools: ["CRM"],
              learning: "访谈方法",
              created_at: "2026-07-10T08:00:00Z",
              updated_at: "2026-07-10T08:00:00Z"
            },
            {
              id: 92,
              user_id: 7,
              title: "跟进试用反馈",
              project: "客户验证",
              status: "reminder",
              priority: "medium",
              tools: ["任务中心"],
              learning: "反馈分析",
              created_at: "2026-07-10T09:00:00Z",
              updated_at: "2026-07-10T09:00:00Z"
            }
          ]
        }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/stats") {
        return Promise.resolve(new Response(JSON.stringify({ total: 2, todo: 1, in_progress: 0, completed: 0, reminder: 1, overdue: 0 }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderTasksPage();

    const boardButton = screen.getByRole("button", { name: "看板" });
    fireEvent.click(boardButton);

    expect(boardButton).toHaveAttribute("aria-pressed", "true");
    const board = await screen.findByRole("region", { name: "任务看板" });
    expect(within(board).getByRole("heading", { name: "待开始" })).toBeInTheDocument();
    expect(within(board).getByRole("heading", { name: "提醒中" })).toBeInTheDocument();
    expect(within(board).getByRole("heading", { name: "整理客户访谈提纲" })).toBeInTheDocument();
    expect(within(board).getByRole("heading", { name: "跟进试用反馈" })).toBeInTheDocument();
    expect(screen.queryByLabelText("任务看板预览")).not.toBeInTheDocument();
  });

  it("groups calendar tasks by due date and restores the list view", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/tasks?limit=20") {
        return Promise.resolve(new Response(JSON.stringify({
          tasks: [
            {
              id: 93,
              user_id: 7,
              title: "提交商业计划书",
              project: "融资准备",
              status: "in_progress",
              priority: "high",
              due_at: "2026-07-15T10:00:00Z",
              tools: ["商业画布"],
              learning: "融资材料",
              created_at: "2026-07-10T08:00:00Z",
              updated_at: "2026-07-10T08:00:00Z"
            },
            {
              id: 94,
              user_id: 7,
              title: "补充竞品数据",
              project: "竞品分析",
              status: "todo",
              priority: "low",
              tools: ["竞品雷达"],
              learning: "竞品调研",
              created_at: "2026-07-10T09:00:00Z",
              updated_at: "2026-07-10T09:00:00Z"
            }
          ]
        }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/stats") {
        return Promise.resolve(new Response(JSON.stringify({ total: 2, todo: 1, in_progress: 1, completed: 0, reminder: 0, overdue: 0 }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderTasksPage();

    fireEvent.click(screen.getByRole("button", { name: "日历" }));

    const calendar = await screen.findByRole("region", { name: "任务日历" });
    expect(within(calendar).getByRole("heading", { name: "2026年7月15日" })).toBeInTheDocument();
    expect(within(calendar).getByRole("heading", { name: "待安排" })).toBeInTheDocument();
    expect(within(calendar).getByRole("heading", { name: "提交商业计划书" })).toBeInTheDocument();
    expect(within(calendar).getByRole("heading", { name: "补充竞品数据" })).toBeInTheDocument();
    expect(screen.queryByLabelText("任务看板预览")).not.toBeInTheDocument();

    const listButton = screen.getByRole("button", { name: "列表" });
    fireEvent.click(listButton);

    expect(listButton).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByLabelText("任务列表")).toBeInTheDocument();
    expect(screen.getByLabelText("任务看板预览")).toBeInTheDocument();
    expect(screen.queryByRole("region", { name: "任务日历" })).not.toBeInTheDocument();
  });

  it("loads the latest task detail and saves edited fields", async () => {
    const task = {
      id: 95,
      user_id: 7,
      title: "准备客户访谈",
      assignee: "张晨",
      project: "客户验证",
      status: "todo",
      priority: "medium",
      tags: ["客户", "访谈"],
      due_at: "2026-07-18T10:00:00Z",
      tools: ["CRM"],
      learning: "访谈方法",
      source_type: "competitor_scan",
      source_id: 11,
      source_title: "竞品扫描：商业沙盘竞品",
      source_url: "/competitor-data",
      created_at: "2026-07-10T08:00:00Z",
      updated_at: "2026-07-10T08:00:00Z"
    };
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/tasks?limit=20") {
        return Promise.resolve(new Response(JSON.stringify({ tasks: [task] }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/stats") {
        return Promise.resolve(new Response(JSON.stringify({ total: 1, todo: 1, in_progress: 0, completed: 0, reminder: 0, overdue: 0 }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/95" && init?.method === "GET") {
        return Promise.resolve(new Response(JSON.stringify({ ...task, title: "准备首轮客户访谈" }), { status: 200 }));
      }
	  if (url === "/api/v1/tasks/95/subtasks" && init?.method === "GET") {
		return Promise.resolve(new Response(JSON.stringify({ subtasks: [] }), { status: 200 }));
	  }
	  if (url === "/api/v1/tasks/95/reminder" && init?.method === "GET") {
		return Promise.resolve(new Response(JSON.stringify({ reminder: null }), { status: 200 }));
	  }
      if (url === "/api/v1/tasks/95" && init?.method === "PATCH") {
        return Promise.resolve(new Response(JSON.stringify({
          ...task,
          title: "完成客户访谈提纲",
          status: "in_progress",
          priority: "high",
          due_at: undefined,
          tools: ["CRM", "任务中心"],
          learning: "访谈复盘",
          updated_at: "2026-07-10T10:00:00Z"
        }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderTasksPage();

    const taskHeading = await screen.findByRole("heading", { name: "准备客户访谈" });
    expect(screen.getByText(/负责人 张晨/)).toBeInTheDocument();
    expect(within(taskHeading.closest("article") as HTMLElement).getByText("客户")).toBeInTheDocument();
    expect(within(taskHeading.closest("article") as HTMLElement).getByRole("link", { name: "来源：竞品扫描：商业沙盘竞品" })).toHaveAttribute("href", "/competitor-data");
    fireEvent.click(within(taskHeading.closest("article") as HTMLElement).getByRole("button", { name: "查看任务详情" }));

    const dialog = await screen.findByRole("dialog", { name: "任务详情" });
    expect(await within(dialog).findByDisplayValue("准备首轮客户访谈")).toBeInTheDocument();
    expect(within(dialog).getByRole("link", { name: "查看来源：竞品扫描：商业沙盘竞品" })).toHaveAttribute("href", "/competitor-data");
    fireEvent.change(within(dialog).getByLabelText("任务标题"), { target: { value: "完成客户访谈提纲" } });
    fireEvent.change(within(dialog).getByLabelText("负责人"), { target: { value: "李明" } });
    fireEvent.change(within(dialog).getByLabelText("标签"), { target: { value: "客户，执行" } });
    fireEvent.change(within(dialog).getByLabelText("任务状态"), { target: { value: "in_progress" } });
    fireEvent.change(within(dialog).getByLabelText("优先级"), { target: { value: "high" } });
    fireEvent.change(within(dialog).getByLabelText("截止时间"), { target: { value: "" } });
    fireEvent.change(within(dialog).getByLabelText("建议工具"), { target: { value: "CRM，任务中心" } });
    fireEvent.change(within(dialog).getByLabelText("补课内容"), { target: { value: "访谈复盘" } });
    fireEvent.click(within(dialog).getByRole("button", { name: "保存修改" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/tasks/95",
      expect.objectContaining({
        method: "PATCH",
        body: JSON.stringify({
          title: "完成客户访谈提纲",
          assignee: "李明",
          project: "客户验证",
          status: "in_progress",
          priority: "high",
          tags: ["客户", "执行"],
          clear_due_at: true,
          tools: ["CRM", "任务中心"],
          learning: "访谈复盘"
        })
      })
    ));
    expect(await screen.findByText("已更新任务：完成客户访谈提纲")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "完成客户访谈提纲" })).toBeInTheDocument();
    expect(screen.queryByRole("dialog", { name: "任务详情" })).not.toBeInTheDocument();
  });

  it("loads, creates, completes, and deletes subtasks in task detail", async () => {
	const task = {
	  id: 97,
	  user_id: 7,
	  title: "准备客户访谈",
	  project: "客户验证",
	  status: "todo",
	  priority: "medium",
	  tools: ["CRM"],
	  learning: "访谈方法",
	  created_at: "2026-07-10T08:00:00Z",
	  updated_at: "2026-07-10T08:00:00Z"
	};
	const initialSubtask = {
	  id: 7,
	  task_id: 97,
	  user_id: 7,
	  title: "确认访谈名单",
	  assignee: "张晨",
	  completed: false,
	  created_at: "2026-07-10T09:00:00Z",
	  updated_at: "2026-07-10T09:00:00Z"
	};
	const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
	  const url = String(input);
	  if (url === "/api/v1/tasks?limit=20") {
		return Promise.resolve(new Response(JSON.stringify({ tasks: [task], total: 1, limit: 20, offset: 0 }), { status: 200 }));
	  }
	  if (url === "/api/v1/tasks/stats") {
		return Promise.resolve(new Response(JSON.stringify({ total: 1, todo: 1, in_progress: 0, completed: 0, reminder: 0, overdue: 0 }), { status: 200 }));
	  }
	  if (url === "/api/v1/tasks/97" && init?.method === "GET") {
		return Promise.resolve(new Response(JSON.stringify(task), { status: 200 }));
	  }
	  if (url === "/api/v1/tasks/97/subtasks" && init?.method === "GET") {
		return Promise.resolve(new Response(JSON.stringify({ subtasks: [initialSubtask] }), { status: 200 }));
	  }
	  if (url === "/api/v1/tasks/97/reminder" && init?.method === "GET") {
		return Promise.resolve(new Response(JSON.stringify({ reminder: null }), { status: 200 }));
	  }
	  if (url === "/api/v1/tasks/97/subtasks" && init?.method === "POST") {
		return Promise.resolve(new Response(JSON.stringify({ ...initialSubtask, id: 8, title: "整理访谈提纲", assignee: "" }), { status: 200 }));
	  }
	  if (url === "/api/v1/tasks/97/subtasks/7" && init?.method === "PATCH") {
		return Promise.resolve(new Response(JSON.stringify({ ...initialSubtask, completed: true }), { status: 200 }));
	  }
	  if (url === "/api/v1/tasks/97/subtasks/8" && init?.method === "DELETE") {
		return Promise.resolve(new Response(null, { status: 204 }));
	  }
	  return Promise.reject(new Error(`unexpected request: ${url}`));
	});

	renderTasksPage();

	const taskHeading = await screen.findByRole("heading", { name: "准备客户访谈" });
	fireEvent.click(within(taskHeading.closest("article") as HTMLElement).getByRole("button", { name: "查看任务详情" }));
	const dialog = await screen.findByRole("dialog", { name: "任务详情" });
	expect(await within(dialog).findByText("确认访谈名单")).toBeInTheDocument();

	fireEvent.change(within(dialog).getByLabelText("新建子任务"), { target: { value: "整理访谈提纲" } });
	fireEvent.click(within(dialog).getByRole("button", { name: "添加子任务" }));
	expect(await within(dialog).findByText("整理访谈提纲")).toBeInTheDocument();

	const completedCheckbox = within(dialog).getByRole("checkbox", { name: "完成子任务 确认访谈名单" });
	fireEvent.click(completedCheckbox);
	await waitFor(() => expect(completedCheckbox).toBeChecked());

	fireEvent.click(within(dialog).getByRole("button", { name: "删除子任务 整理访谈提纲" }));
	await waitFor(() => expect(within(dialog).queryByText("整理访谈提纲")).not.toBeInTheDocument());
	expect(fetchMock).toHaveBeenCalledWith("/api/v1/tasks/97/subtasks/7", expect.objectContaining({ method: "PATCH", body: JSON.stringify({ completed: true }) }));
	expect(fetchMock).toHaveBeenCalledWith("/api/v1/tasks/97/subtasks/8", expect.objectContaining({ method: "DELETE" }));
  });

  it("sets and cancels a one-time task reminder", async () => {
    const task = {
      id: 98,
      user_id: 7,
      title: "提交客户访谈报告",
      project: "客户验证",
      status: "todo",
      priority: "high",
      tools: ["CRM"],
      learning: "访谈复盘",
      created_at: "2026-07-10T08:00:00Z",
      updated_at: "2026-07-10T08:00:00Z"
    };
    const remindAt = new Date("2026-07-18T18:00").toISOString();
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/tasks?limit=20") {
        return Promise.resolve(new Response(JSON.stringify({ tasks: [task], total: 1, limit: 20, offset: 0 }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/stats") {
        return Promise.resolve(new Response(JSON.stringify({ total: 1, todo: 1, in_progress: 0, completed: 0, reminder: 0, overdue: 0 }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/98" && init?.method === "GET") {
        return Promise.resolve(new Response(JSON.stringify(task), { status: 200 }));
      }
      if (url === "/api/v1/tasks/98/subtasks" && init?.method === "GET") {
        return Promise.resolve(new Response(JSON.stringify({ subtasks: [] }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/98/reminder" && init?.method === "GET") {
        return Promise.resolve(new Response(JSON.stringify({ reminder: null }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/98/reminder" && init?.method === "PUT") {
        const payload = JSON.parse(String(init.body)) as { recurrence?: string };
        if (payload.recurrence === "daily") {
          return Promise.resolve(new Response(JSON.stringify({ error: "membership_required" }), { status: 402 }));
        }
        return Promise.resolve(new Response(JSON.stringify({
          id: 8,
          task_id: 98,
          user_id: 7,
          remind_at: remindAt,
          recurrence: "once",
          created_at: "2026-07-10T10:00:00Z",
          updated_at: "2026-07-10T10:00:00Z"
        }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/98/reminder" && init?.method === "DELETE") {
        return Promise.resolve(new Response(null, { status: 204 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderTasksPage();

    const taskHeading = await screen.findByRole("heading", { name: "提交客户访谈报告" });
    fireEvent.click(within(taskHeading.closest("article") as HTMLElement).getByRole("button", { name: "查看任务详情" }));
    const dialog = await screen.findByRole("dialog", { name: "任务详情" });
    expect(await within(dialog).findByText("暂无提醒")).toBeInTheDocument();

    fireEvent.change(within(dialog).getByLabelText("提醒时间"), { target: { value: "2026-07-18T18:00" } });
    fireEvent.click(within(dialog).getByRole("button", { name: "保存提醒" }));

    expect(await within(dialog).findByText("已设置站内提醒")).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/tasks/98/reminder", expect.objectContaining({
      method: "PUT",
      body: JSON.stringify({ remind_at: remindAt, recurrence: "once" })
    }));

    fireEvent.click(within(dialog).getByRole("button", { name: "取消提醒" }));
    await waitFor(() => expect(within(dialog).getByText("暂无提醒")).toBeInTheDocument());
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/tasks/98/reminder", expect.objectContaining({ method: "DELETE" }));

    fireEvent.change(within(dialog).getByLabelText("提醒时间"), { target: { value: "2026-07-18T18:00" } });
    fireEvent.change(within(dialog).getByLabelText("提醒频率"), { target: { value: "daily" } });
    fireEvent.click(within(dialog).getByRole("button", { name: "保存提醒" }));

    expect(await within(dialog).findByText("循环提醒仅限会员使用，请升级后重试")).toBeInTheDocument();
    expect(within(dialog).getByRole("link", { name: "升级会员" })).toHaveAttribute("href", "/membership");
  });

  it("deletes a task after a second confirmation", async () => {
    let deleted = false;
    const task = {
      id: 96,
      user_id: 7,
      title: "清理过期跟进任务",
      project: "客户验证",
      status: "todo",
      priority: "low",
      tools: ["CRM"],
      learning: "客户跟进",
      created_at: "2026-07-10T08:00:00Z",
      updated_at: "2026-07-10T08:00:00Z"
    };
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/tasks?limit=20") {
        return Promise.resolve(new Response(JSON.stringify({ tasks: [task] }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/stats") {
        return Promise.resolve(new Response(JSON.stringify({
          total: deleted ? 0 : 1,
          todo: deleted ? 0 : 1,
          in_progress: 0,
          completed: 0,
          reminder: 0,
          overdue: 0
        }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/96" && init?.method === "GET") {
        return Promise.resolve(new Response(JSON.stringify(task), { status: 200 }));
      }
	  if (url === "/api/v1/tasks/96/subtasks" && init?.method === "GET") {
		return Promise.resolve(new Response(JSON.stringify({ subtasks: [] }), { status: 200 }));
	  }
	  if (url === "/api/v1/tasks/96/reminder" && init?.method === "GET") {
		return Promise.resolve(new Response(JSON.stringify({ reminder: null }), { status: 200 }));
	  }
      if (url === "/api/v1/tasks/96" && init?.method === "DELETE") {
        deleted = true;
        return Promise.resolve(new Response(null, { status: 204 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderTasksPage();

    const taskHeading = await screen.findByRole("heading", { name: "清理过期跟进任务" });
    fireEvent.click(within(taskHeading.closest("article") as HTMLElement).getByRole("button", { name: "查看任务详情" }));
    const dialog = await screen.findByRole("dialog", { name: "任务详情" });
    expect(await within(dialog).findByDisplayValue("清理过期跟进任务")).toBeInTheDocument();

    fireEvent.click(within(dialog).getByRole("button", { name: "删除任务" }));
    fireEvent.click(within(dialog).getByRole("button", { name: "确认删除" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/tasks/96",
      expect.objectContaining({ method: "DELETE" })
    ));
    expect(await screen.findByText("已删除任务：清理过期跟进任务")).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "清理过期跟进任务" })).not.toBeInTheDocument();
    expect(screen.queryByRole("dialog", { name: "任务详情" })).not.toBeInTheDocument();
  });

  it("searches tasks and combines project and priority filters", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/tasks?limit=20") {
        return Promise.resolve(new Response(JSON.stringify({ tasks: [], total: 0, limit: 20, offset: 0 }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/projects") {
        return Promise.resolve(new Response(JSON.stringify({ projects: ["商业沙盘", "客户验证"] }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/tags") {
        return Promise.resolve(new Response(JSON.stringify({ tags: ["用户研究", "访谈"] }), { status: 200 }));
      }
      if (url === "/api/v1/tasks?q=%E5%AE%A2%E6%88%B7&limit=20") {
        return Promise.resolve(new Response(JSON.stringify({ tasks: [], total: 0, limit: 20, offset: 0 }), { status: 200 }));
      }
      if (url === "/api/v1/tasks?project=%E5%95%86%E4%B8%9A%E6%B2%99%E7%9B%98&q=%E5%AE%A2%E6%88%B7&limit=20") {
        return Promise.resolve(new Response(JSON.stringify({ tasks: [], total: 0, limit: 20, offset: 0 }), { status: 200 }));
      }
      if (url === "/api/v1/tasks?project=%E5%95%86%E4%B8%9A%E6%B2%99%E7%9B%98&priority=high&q=%E5%AE%A2%E6%88%B7&limit=20") {
        return Promise.resolve(new Response(JSON.stringify({ tasks: [], total: 0, limit: 20, offset: 0 }), { status: 200 }));
      }
      if (url === "/api/v1/tasks?project=%E5%95%86%E4%B8%9A%E6%B2%99%E7%9B%98&priority=high&tag=%E7%94%A8%E6%88%B7%E7%A0%94%E7%A9%B6&q=%E5%AE%A2%E6%88%B7&limit=20") {
        return Promise.resolve(new Response(JSON.stringify({
          tasks: [{
            id: 101,
            user_id: 7,
            title: "高优先级客户沙盘任务",
            project: "商业沙盘",
            status: "todo",
            priority: "high",
            tags: ["用户研究"],
            tools: ["商业沙盘"],
            learning: "客户分析",
            created_at: "2026-07-10T08:00:00Z",
            updated_at: "2026-07-10T08:00:00Z"
          }],
          total: 1,
          limit: 20,
          offset: 0
        }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/stats") {
        return Promise.resolve(new Response(JSON.stringify({ total: 1, todo: 1, in_progress: 0, completed: 0, reminder: 0, overdue: 0 }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderTasksPage();

    fireEvent.change(screen.getByLabelText("搜索任务"), { target: { value: "客户" } });
    fireEvent.click(screen.getByRole("button", { name: "搜索" }));
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/tasks?q=%E5%AE%A2%E6%88%B7&limit=20",
      expect.objectContaining({ method: "GET" })
    ));

    const projectFilter = screen.getByRole("combobox", { name: "按项目筛选" });
    fireEvent.focus(projectFilter);
    expect(await screen.findByRole("option", { name: "商业沙盘" })).toBeInTheDocument();
    fireEvent.change(projectFilter, { target: { value: "商业沙盘" } });
    fireEvent.change(screen.getByRole("combobox", { name: "按优先级筛选" }), { target: { value: "high" } });

    const tagFilter = screen.getByRole("combobox", { name: "按标签筛选" });
    fireEvent.focus(tagFilter);
    expect(await screen.findByRole("option", { name: "用户研究" })).toBeInTheDocument();
    fireEvent.change(tagFilter, { target: { value: "用户研究" } });

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/tasks?project=%E5%95%86%E4%B8%9A%E6%B2%99%E7%9B%98&priority=high&tag=%E7%94%A8%E6%88%B7%E7%A0%94%E7%A9%B6&q=%E5%AE%A2%E6%88%B7&limit=20",
      expect.objectContaining({ method: "GET" })
    ));
    expect(await screen.findByRole("heading", { name: "高优先级客户沙盘任务" })).toBeInTheDocument();
  });

  it("paginates tasks using the backend total", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/tasks?limit=20") {
        return Promise.resolve(new Response(JSON.stringify({ tasks: [], total: 21, limit: 20, offset: 0 }), { status: 200 }));
      }
      if (url === "/api/v1/tasks?limit=20&offset=20") {
        return Promise.resolve(new Response(JSON.stringify({
          tasks: [{
            id: 102,
            user_id: 7,
            title: "第二页任务",
            project: "任务中心",
            status: "todo",
            priority: "medium",
            tools: ["任务中心"],
            learning: "分页验证",
            created_at: "2026-07-01T08:00:00Z",
            updated_at: "2026-07-01T08:00:00Z"
          }],
          total: 21,
          limit: 20,
          offset: 20
        }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/stats") {
        return Promise.resolve(new Response(JSON.stringify({ total: 21, todo: 21, in_progress: 0, completed: 0, reminder: 0, overdue: 0 }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderTasksPage();

    expect(await screen.findByText("第 1 / 2 页 · 共 21 条")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "下一页" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/tasks?limit=20&offset=20",
      expect.objectContaining({ method: "GET" })
    ));
    expect(await screen.findByRole("heading", { name: "第二页任务" })).toBeInTheDocument();
    expect(screen.getByText("第 2 / 2 页 · 共 21 条")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "下一页" })).toBeDisabled();
  });

  it("selects the current page and updates task statuses in a batch", async () => {
    let statsRequests = 0;
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/tasks?limit=20") {
        return Promise.resolve(new Response(JSON.stringify({
          tasks: [
            {
              id: 201,
              user_id: 7,
              title: "整理销售话术",
              project: "销售增长",
              status: "todo",
              priority: "high",
              tools: ["CRM"],
              learning: "销售方法",
              created_at: "2026-07-10T08:00:00Z",
              updated_at: "2026-07-10T08:00:00Z"
            },
            {
              id: 202,
              user_id: 7,
              title: "复盘客户反馈",
              project: "客户验证",
              status: "in_progress",
              priority: "medium",
              tools: ["客户管理"],
              learning: "客户访谈",
              created_at: "2026-07-10T09:00:00Z",
              updated_at: "2026-07-10T09:00:00Z"
            }
          ],
          total: 2,
          limit: 20,
          offset: 0
        }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/stats") {
        statsRequests += 1;
        return Promise.resolve(new Response(JSON.stringify({
          total: 2,
          todo: statsRequests === 1 ? 1 : 0,
          in_progress: statsRequests === 1 ? 1 : 0,
          completed: statsRequests === 1 ? 0 : 2,
          reminder: 0,
          overdue: 0
        }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/batch" && init?.method === "PATCH") {
        return Promise.resolve(new Response(JSON.stringify({ updated: 2 }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderTasksPage();

    await screen.findByRole("heading", { name: "整理销售话术" });
    fireEvent.click(screen.getByRole("checkbox", { name: "全选当前页任务" }));
    expect(screen.getByText("已选择 2 项")).toBeInTheDocument();

    fireEvent.change(screen.getByRole("combobox", { name: "批量设置状态" }), { target: { value: "completed" } });
    fireEvent.click(screen.getByRole("button", { name: "应用状态" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/tasks/batch",
      expect.objectContaining({
        method: "PATCH",
        body: JSON.stringify({ ids: [201, 202], status: "completed" })
      })
    ));
    expect(await screen.findByText("已更新 2 条任务状态")).toBeInTheDocument();
    expect(screen.queryByText("已选择 2 项")).not.toBeInTheDocument();
    expect(screen.getAllByText("已完成").filter((node) => node.classList.contains("task-state"))).toHaveLength(2);
    await waitFor(() => expect(statsRequests).toBe(2));
  });

  it("requires confirmation before batch deletion and clears selection after page changes", async () => {
    let statsRequests = 0;
    let deleted = false;
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/tasks?limit=20") {
        if (deleted) {
          return Promise.resolve(new Response(JSON.stringify({
            tasks: [{
              id: 303,
              user_id: 7,
              title: "补入当前页的任务",
              project: "任务中心",
              status: "todo",
              priority: "low",
              tools: ["任务中心"],
              learning: "分页补位",
              created_at: "2026-07-10T10:00:00Z",
              updated_at: "2026-07-10T10:00:00Z"
            }],
            total: 21,
            limit: 20,
            offset: 0
          }), { status: 200 }));
        }
        return Promise.resolve(new Response(JSON.stringify({
          tasks: [{
            id: 301,
            user_id: 7,
            title: "待删除任务",
            project: "任务中心",
            status: "todo",
            priority: "medium",
            tools: ["任务中心"],
            learning: "批量操作",
            created_at: "2026-07-10T08:00:00Z",
            updated_at: "2026-07-10T08:00:00Z"
          }],
          total: 22,
          limit: 20,
          offset: 0
        }), { status: 200 }));
      }
      if (url === "/api/v1/tasks?limit=20&offset=20") {
        return Promise.resolve(new Response(JSON.stringify({
          tasks: [{
            id: 302,
            user_id: 7,
            title: "第二页保留任务",
            project: "任务中心",
            status: "todo",
            priority: "low",
            tools: ["任务中心"],
            learning: "分页",
            created_at: "2026-07-10T09:00:00Z",
            updated_at: "2026-07-10T09:00:00Z"
          }],
          total: 21,
          limit: 20,
          offset: 20
        }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/stats") {
        statsRequests += 1;
        return Promise.resolve(new Response(JSON.stringify({ total: 22, todo: 22, in_progress: 0, completed: 0, reminder: 0, overdue: 0 }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/batch" && init?.method === "DELETE") {
        deleted = true;
        return Promise.resolve(new Response(JSON.stringify({ deleted: 1 }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderTasksPage();

    fireEvent.click(await screen.findByRole("checkbox", { name: "选择任务 待删除任务" }));
    fireEvent.click(screen.getByRole("button", { name: "下一页" }));
    expect(await screen.findByRole("heading", { name: "第二页保留任务" })).toBeInTheDocument();
    expect(screen.queryByText(/已选择/)).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "上一页" }));
    fireEvent.click(await screen.findByRole("checkbox", { name: "选择任务 待删除任务" }));
    fireEvent.click(screen.getByRole("button", { name: "批量删除" }));
    expect(fetchMock).not.toHaveBeenCalledWith("/api/v1/tasks/batch", expect.objectContaining({ method: "DELETE" }));
    expect(screen.getByRole("button", { name: "确认删除 1 项" })).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "确认删除 1 项" }));
    expect(await screen.findByText("已删除 1 条任务")).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "待删除任务" })).not.toBeInTheDocument();
    expect(await screen.findByRole("heading", { name: "补入当前页的任务" })).toBeInTheDocument();
    expect(screen.getByText("第 1 / 2 页 · 共 21 条")).toBeInTheDocument();
    await waitFor(() => expect(statsRequests).toBe(2));
  });
});
