import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import TaskCreatePage from "./TaskCreatePage";

describe("TaskCreatePage", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  it("creates a complete manual task and resets for another entry", async () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-07-11T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张晨", phone: "13800138000", status: "active" }
    });
    let postedBody: Record<string, unknown> | null = null;
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/tasks/projects") {
        return Promise.resolve(new Response(JSON.stringify({ projects: ["客户验证", "商业沙盘"] }), { status: 200 }));
      }
      if (url === "/api/v1/tasks" && init?.method === "POST") {
        postedBody = JSON.parse(String(init.body));
        return Promise.resolve(new Response(JSON.stringify({
          id: 103,
          user_id: 7,
          title: "完成首轮客户访谈",
          description: "访谈 5 位目标客户并整理关键问题",
          project: "客户验证",
          status: "todo",
          priority: "high",
          due_at: "2026-07-20T02:00:00Z",
          tools: ["CRM", "任务中心"],
          learning: "客户访谈方法",
          created_at: "2026-07-10T08:00:00Z",
          updated_at: "2026-07-10T08:00:00Z"
        }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    render(
      <MemoryRouter>
        <TaskCreatePage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "新建任务" })).toBeInTheDocument();
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/tasks/projects",
      expect.objectContaining({ method: "GET" })
    ));
    fireEvent.change(screen.getByLabelText("任务标题"), { target: { value: "完成首轮客户访谈" } });
    fireEvent.change(screen.getByLabelText("任务描述"), { target: { value: "访谈 5 位目标客户并整理关键问题" } });
    fireEvent.change(screen.getByLabelText("所属项目"), { target: { value: "客户验证" } });
    fireEvent.change(screen.getByLabelText("截止时间"), { target: { value: "2026-07-20T10:00" } });
    fireEvent.change(screen.getByLabelText("优先级"), { target: { value: "high" } });
    fireEvent.change(screen.getByLabelText("建议工具"), { target: { value: "CRM，任务中心" } });
    fireEvent.change(screen.getByLabelText("补课内容"), { target: { value: "客户访谈方法" } });
    fireEvent.click(screen.getByRole("button", { name: "保存并继续添加" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/tasks",
      expect.objectContaining({ method: "POST" })
    ));
    expect(postedBody).toEqual({
      title: "完成首轮客户访谈",
      description: "访谈 5 位目标客户并整理关键问题",
      project: "客户验证",
      priority: "high",
      due_at: new Date("2026-07-20T10:00").toISOString(),
      tools: ["CRM", "任务中心"],
      learning: "客户访谈方法"
    });
    expect(await screen.findByText("已创建任务：完成首轮客户访谈")).toBeInTheDocument();
    expect(screen.getByLabelText("任务标题")).toHaveValue("");
  });
});
