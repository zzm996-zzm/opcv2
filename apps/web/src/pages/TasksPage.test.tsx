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

  it("renders the task overview and AI generation entry", () => {
    renderTasksPage();

    expect(screen.getByRole("heading", { name: "任务中心" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "生成任务表" })).toBeInTheDocument();
    expect(screen.getByText("完成智能客服系统项目商业画布")).toBeInTheDocument();
  });

  it("loads tasks from API", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
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
      }), { status: 200 })
    );

    renderTasksPage();

    expect(await screen.findByRole("heading", { name: "联调商业沙盘接口" })).toBeInTheDocument();
    expect(screen.getByText("商业沙盘 · 截止 06/30 18:00")).toBeInTheDocument();
    expect(screen.getByText("建议工具：沙盘推演 / 任务中心")).toBeInTheDocument();
    const inProgressStat = screen.getAllByText("进行中").find((node) => node.tagName.toLowerCase() === "small")?.closest("article");
    expect(inProgressStat).not.toBeNull();
    expect(within(inProgressStat as HTMLElement).getByText("1")).toBeInTheDocument();
  });

  it("shows backend list errors while keeping fallback tasks visible", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ error: "invalid_limit" }), { status: 400 })
    );

    renderTasksPage();

    expect(await screen.findByText("列表数量参数有误")).toBeInTheDocument();
    expect(screen.getByText("完成智能客服系统项目商业画布")).toBeInTheDocument();
  });

  it("updates task status from the task row action", async () => {
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
      if (url === "/api/v1/tasks/41" && init?.method === "PATCH") {
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
    expect(within(taskRow).getByRole("button", { name: "已完成" })).toBeDisabled();
  });
});
