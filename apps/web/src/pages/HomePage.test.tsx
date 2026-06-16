import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import HomePage from "./HomePage";

describe("HomePage", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  it("shows the signed-in user and logs out from the account menu", async () => {
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
    const fetchMock = vi
      .spyOn(globalThis, "fetch")
      .mockResolvedValue(new Response(null, { status: 204 }));

    render(
      <MemoryRouter>
        <HomePage />
      </MemoryRouter>
    );

    fireEvent.click(screen.getByRole("button", { name: "张晨的账号菜单" }));
    expect(screen.getByText("有效期至 2025-12-31")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "退出登录" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
    expect(await screen.findByRole("link", { name: "登录 / 注册" })).toBeInTheDocument();
  });

  it("clears the local session even when logout cannot reach the API", async () => {
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
    vi.spyOn(globalThis, "fetch").mockRejectedValue(new Error("offline"));

    render(
      <MemoryRouter>
        <HomePage />
      </MemoryRouter>
    );

    fireEvent.click(screen.getByRole("button", { name: "张晨的账号菜单" }));
    fireEvent.click(screen.getByRole("button", { name: "退出登录" }));

    expect(await screen.findByRole("link", { name: "登录 / 注册" })).toBeInTheDocument();
  });

  it("opens notification menu and Copilot utility states", () => {
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

    render(
      <MemoryRouter>
        <HomePage />
      </MemoryRouter>
    );

    fireEvent.click(screen.getByRole("button", { name: "通知" }));
    expect(screen.getByRole("dialog", { name: "通知下拉框" })).toBeInTheDocument();
    expect(screen.getByText("项目分析完成")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "打开 Copilot 设置" }));
    expect(screen.getByText("Copilot 设置")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Claude opus4.8" })).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "添加文件" }));
    expect(screen.getByText("市场分析报告.pdf")).toBeInTheDocument();
  });

  it("renders direct Copilot settings and file states", () => {
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

    const { unmount } = render(
      <MemoryRouter>
        <HomePage assistantState="settings" />
      </MemoryRouter>
    );

    expect(screen.getByText("Copilot 设置")).toBeInTheDocument();
    expect(screen.getByText("深度思考")).toBeInTheDocument();

    unmount();

    render(
      <MemoryRouter>
        <HomePage assistantState="files" />
      </MemoryRouter>
    );

    expect(screen.getByText("市场分析报告.pdf")).toBeInTheDocument();
  });

  it("renders direct home dropdown and collapsed assistant states", () => {
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

    const noticeRender = render(
      <MemoryRouter>
        <HomePage menuState="notice" />
      </MemoryRouter>
    );

    expect(screen.getByRole("dialog", { name: "通知下拉框" })).toBeInTheDocument();

    noticeRender.unmount();

    const accountRender = render(
      <MemoryRouter>
        <HomePage menuState="account" />
      </MemoryRouter>
    );

    expect(screen.getByRole("dialog", { name: "头像下拉框" })).toBeInTheDocument();

    accountRender.unmount();

    render(
      <MemoryRouter>
        <HomePage assistantState="collapsed" />
      </MemoryRouter>
    );

    expect(screen.getAllByRole("button", { name: /智活 Copilot/ }).length).toBeGreaterThan(0);
  });
});
