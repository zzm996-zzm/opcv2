import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import { notificationsApi } from "../lib/notificationsApi";
import MessagesPage from "./MessagesPage";

vi.mock("../lib/notificationsApi", () => ({
  notificationsApi: {
    list: vi.fn(),
    get: vi.fn(),
    markRead: vi.fn(),
    markAllRead: vi.fn(),
    delete: vi.fn(),
    summary: vi.fn()
  }
}));

describe("MessagesPage", () => {
  afterEach(() => {
    authSession.clear();
    vi.clearAllMocks();
    vi.restoreAllMocks();
  });

  function signIn() {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-11T12:00:00Z",
      is_new_user: false,
      user: {
        id: 7,
        nickname: "张晨",
        phone: "13800138000",
        status: "active"
      }
    });
  }

  function notification() {
    return {
      id: 7,
      user_id: 7,
      type: "task",
      title: "任务提醒：AI 智能硬件项目拆解完成",
      summary: "拆解报告已生成",
      body: "你发起的项目拆解已完成，点击查看拆解报告。",
      source_type: "analysis",
      source_id: 42,
      action_label: "查看拆解报告",
      action_url: "/analysis",
      created_at: "2026-07-02T09:30:00Z"
    };
  }

  it("renders notifications from API and marks all read", async () => {
    signIn();
    vi.mocked(notificationsApi.list).mockResolvedValue({ notifications: [notification()], total: 1, limit: 20, offset: 0 });
    vi.mocked(notificationsApi.summary).mockResolvedValue({
      unread: 1,
      by_type: [{ type: "task", count: 1 }],
      latest: [notification()]
    });
    vi.mocked(notificationsApi.markAllRead).mockResolvedValue({ updated: 1 });

    render(
      <MemoryRouter initialEntries={["/messages"]}>
        <MessagesPage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "消息中心" })).toBeInTheDocument();
    expect(await screen.findByRole("tab", { name: "全部 1" })).toHaveAttribute("aria-selected", "true");
    expect(screen.getByText("任务提醒：AI 智能硬件项目拆解完成")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /任务提醒：AI 智能硬件项目拆解完成/ })).toHaveAttribute("href", "/messages/7");
    expect(notificationsApi.list).toHaveBeenCalledWith({ limit: 20 });
    expect(notificationsApi.summary).toHaveBeenCalled();

    fireEvent.click(screen.getByRole("button", { name: "全部已读" }));

    await waitFor(() => expect(notificationsApi.markAllRead).toHaveBeenCalled());
    expect(await screen.findByText("已标记 1 条消息为已读")).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "任务 0" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "全部已读" })).toBeDisabled();
  });

  it("filters notifications by category", async () => {
    signIn();
    const systemNotification = {
      ...notification(),
      id: 8,
      type: "system",
      title: "系统维护通知"
    };
    vi.mocked(notificationsApi.list)
      .mockResolvedValueOnce({ notifications: [notification(), systemNotification], total: 2, limit: 20, offset: 0 })
      .mockResolvedValueOnce({ notifications: [systemNotification], total: 1, limit: 20, offset: 0 })
      .mockResolvedValueOnce({ notifications: [], total: 0, limit: 20, offset: 0 });
    vi.mocked(notificationsApi.summary).mockResolvedValue({
      unread: 2,
      by_type: [{ type: "task", count: 1 }, { type: "system", count: 1 }],
      latest: [notification(), systemNotification]
    });

    render(
      <MemoryRouter initialEntries={["/messages"]}>
        <MessagesPage />
      </MemoryRouter>
    );

    expect(await screen.findByText("系统维护通知")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("tab", { name: "系统 1" }));

    await waitFor(() => expect(notificationsApi.list).toHaveBeenLastCalledWith({ type: "system", limit: 20 }));
    expect(screen.getByRole("tab", { name: "系统 1" })).toHaveAttribute("aria-selected", "true");
    expect(screen.queryByText("任务提醒：AI 智能硬件项目拆解完成")).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("tab", { name: "CRM 0" }));
    expect(await screen.findByText("暂无CRM消息")).toBeInTheDocument();
    expect(notificationsApi.list).toHaveBeenLastCalledWith({ type: "crm", limit: 20 });
  });

  it("paginates notifications and resets page when category changes", async () => {
    signIn();
    const secondPageNotification = {
      ...notification(),
      id: 27,
      title: "第二页任务提醒"
    };
    vi.mocked(notificationsApi.list)
      .mockResolvedValueOnce({ notifications: [notification()], total: 25, limit: 20, offset: 0 })
      .mockResolvedValueOnce({ notifications: [secondPageNotification], total: 25, limit: 20, offset: 20 })
      .mockResolvedValueOnce({ notifications: [notification()], total: 1, limit: 20, offset: 0 });
    vi.mocked(notificationsApi.summary).mockResolvedValue({
      unread: 1,
      by_type: [{ type: "task", count: 1 }],
      latest: [notification()]
    });

    render(
      <MemoryRouter initialEntries={["/messages"]}>
        <MessagesPage />
      </MemoryRouter>
    );

    expect(await screen.findByText("任务提醒：AI 智能硬件项目拆解完成")).toBeInTheDocument();
    expect(screen.getByText("25条 · 20条/页")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "下一页" }));

    expect(await screen.findByText("第二页任务提醒")).toBeInTheDocument();
    expect(notificationsApi.list).toHaveBeenLastCalledWith({ limit: 20, offset: 20 });
    expect(screen.getByRole("button", { name: "2" })).toHaveAttribute("aria-current", "page");

    fireEvent.click(screen.getByRole("tab", { name: "任务 1" }));
    await waitFor(() => expect(notificationsApi.list).toHaveBeenLastCalledWith({ type: "task", limit: 20 }));
    expect(screen.getByRole("button", { name: "1" })).toHaveAttribute("aria-current", "page");
  });

  it("renders a message detail page and marks it read", async () => {
    signIn();
    vi.mocked(notificationsApi.get).mockResolvedValue(notification());
    vi.mocked(notificationsApi.markRead).mockResolvedValue({ ...notification(), read_at: "2026-07-02T10:00:00Z" });

    render(
      <MemoryRouter initialEntries={["/messages/7"]}>
        <Routes>
          <Route element={<MessagesPage />} path="/messages/:messageId" />
        </Routes>
      </MemoryRouter>
    );

    expect(await screen.findByRole("heading", { name: "任务提醒：AI 智能硬件项目拆解完成" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "查看拆解报告" })).toHaveAttribute("href", "/analysis");
    expect(screen.getByText("你发起的项目拆解已完成，点击查看拆解报告。")).toBeInTheDocument();
    await waitFor(() => expect(notificationsApi.markRead).toHaveBeenCalledWith(7));
    expect(await screen.findByText("已标记为已读")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /已读/ })).toBeDisabled();
  });

  it("allows retrying when automatic read marking fails", async () => {
    signIn();
    vi.mocked(notificationsApi.get).mockResolvedValue(notification());
    vi.mocked(notificationsApi.markRead)
      .mockRejectedValueOnce(new Error("offline"))
      .mockResolvedValueOnce({ ...notification(), read_at: "2026-07-02T10:00:00Z" });

    render(
      <MemoryRouter initialEntries={["/messages/7"]}>
        <Routes>
          <Route element={<MessagesPage />} path="/messages/:messageId" />
        </Routes>
      </MemoryRouter>
    );

    expect(await screen.findByText("暂时无法标记已读")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: /标记已读/ }));

    await waitFor(() => expect(notificationsApi.markRead).toHaveBeenCalledTimes(2));
    expect(await screen.findByText("已标记为已读")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /已读/ })).toBeDisabled();
  });

  it("confirms deletion and returns to the message list", async () => {
    signIn();
    vi.mocked(notificationsApi.get).mockResolvedValue({ ...notification(), read_at: "2026-07-02T10:00:00Z" });
    vi.mocked(notificationsApi.delete).mockResolvedValue(undefined);

    render(
      <MemoryRouter initialEntries={["/messages/7"]}>
        <Routes>
          <Route element={<MessagesPage />} path="/messages/:messageId" />
          <Route element={<h1>消息中心列表</h1>} path="/messages" />
        </Routes>
      </MemoryRouter>
    );

    expect(await screen.findByRole("heading", { name: "任务提醒：AI 智能硬件项目拆解完成" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: /删除消息/ }));
    expect(notificationsApi.delete).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: /确认删除/ }));

    await waitFor(() => expect(notificationsApi.delete).toHaveBeenCalledWith(7));
    expect(await screen.findByRole("heading", { name: "消息中心列表" })).toBeInTheDocument();
  });
});
