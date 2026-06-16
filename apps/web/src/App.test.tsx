import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "./lib/authSession";
import App from "./App";

describe("App", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("renders the OPC product shell", () => {
    render(
      <MemoryRouter>
        <App />
      </MemoryRouter>
    );

    expect(
      screen.getByRole("heading", { name: "欢迎来到 智活AI" })
    ).toBeInTheDocument();
    expect(screen.getByRole("navigation", { name: "顶部全局功能区" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "工具箱" })).toHaveAttribute(
      "href",
      "/tools"
    );
  });

  it("redirects a signed-out user from a protected product route", async () => {
    authSession.finishRestore();

    render(
      <MemoryRouter initialEntries={["/leads"]}>
        <App />
      </MemoryRouter>
    );

    expect(await screen.findByRole("tab", { name: "登录" })).toHaveAttribute(
      "aria-selected",
      "true"
    );
  });

  it("renders a product route for a signed-in user", () => {
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
      <MemoryRouter initialEntries={["/leads"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "AI线索开发" })).toBeInTheDocument();
  });

  it("renders first-class V4 account routes for a signed-in user", () => {
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
      <MemoryRouter initialEntries={["/profile/settings"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "账号与资料设置" })).toBeInTheDocument();
  });

  it("renders the free-loop V4 product routes for a signed-in user", () => {
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
      <MemoryRouter initialEntries={["/projects"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "项目超市" })).toBeInTheDocument();
    expect(screen.queryByText("第一版正在实现")).not.toBeInTheDocument();
  });
});
