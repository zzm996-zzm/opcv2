import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import LoginPage from "./LoginPage";

describe("LoginPage", () => {
  afterEach(() => {
    authSession.clear();
    vi.useRealTimers();
    vi.restoreAllMocks();
  });

  it("requires agreement before account password login", () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    );

    fireEvent.change(screen.getByLabelText("账号"), {
      target: { value: "deploy_user" }
    });
    fireEvent.change(screen.getByLabelText("密码"), {
      target: { value: "secret123" }
    });

    expect(screen.getByRole("button", { name: "登录" })).toBeDisabled();
    fireEvent.click(screen.getByRole("checkbox", { name: "同意用户协议和隐私政策" }));
    expect(screen.getByRole("button", { name: "登录" })).toBeEnabled();
  });

  it("logs in with account and password", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          access_token: "access-token",
          access_token_expires_at: new Date(Date.now() + 60000).toISOString(),
          user: {
            id: 42,
            nickname: "部署测试",
            account: "deploy_user",
            phone: "",
            status: "active"
          },
          is_new_user: false
        }),
        { status: 200 }
      )
    );

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    );

    fireEvent.change(screen.getByLabelText("账号"), {
      target: { value: "deploy_user" }
    });
    fireEvent.change(screen.getByLabelText("密码"), {
      target: { value: "secret123" }
    });
    fireEvent.click(screen.getByRole("checkbox", { name: "同意用户协议和隐私政策" }));
    fireEvent.click(screen.getByRole("button", { name: "登录" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
    expect(await screen.findByText("登录成功，正在进入工作台")).toBeInTheDocument();
  });

  it("renders the registration fields from the UI design", () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    );

    fireEvent.click(screen.getByRole("tab", { name: "注册" }));

    expect(screen.getByLabelText("账号")).toBeInTheDocument();
    expect(screen.getByLabelText("密码")).toBeInTheDocument();
    expect(screen.getByLabelText("确认密码")).toBeInTheDocument();
    expect(screen.getByLabelText("邮箱")).toBeInTheDocument();
    expect(screen.getByLabelText("手机号")).toBeInTheDocument();
    expect(screen.getByLabelText("微信或企业微信")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "注册并创建账号" })).toBeDisabled();
  });

  it("registers with account and password without phone", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          access_token: "access-token",
          access_token_expires_at: new Date(Date.now() + 60000).toISOString(),
          user: {
            id: 42,
            nickname: "deploy_user",
            account: "deploy_user",
            phone: "",
            status: "active"
          },
          is_new_user: true
        }),
        { status: 200 }
      )
    );

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    );

    fireEvent.click(screen.getByRole("tab", { name: "注册" }));
    fireEvent.change(screen.getByLabelText("账号"), {
      target: { value: "deploy_user" }
    });
    fireEvent.change(screen.getByLabelText("密码"), {
      target: { value: "secret123" }
    });
    fireEvent.change(screen.getByLabelText("确认密码"), {
      target: { value: "secret123" }
    });
    fireEvent.change(screen.getByLabelText("手机号"), {
      target: { value: "13800138000" }
    });
    fireEvent.click(screen.getByRole("checkbox", { name: "同意用户协议和隐私政策" }));
    fireEvent.click(screen.getByRole("button", { name: "注册并创建账号" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
    expect(await screen.findByText("注册成功，正在进入工作台")).toBeInTheDocument();
  });

  it("shows a specific registration error when the account already exists", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ error: "account_exists" }), { status: 409 })
    );

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    );

    fireEvent.click(screen.getByRole("tab", { name: "注册" }));
    fireEvent.change(screen.getByLabelText("账号"), {
      target: { value: "deploy_user" }
    });
    fireEvent.change(screen.getByLabelText("密码"), {
      target: { value: "secret123" }
    });
    fireEvent.change(screen.getByLabelText("确认密码"), {
      target: { value: "secret123" }
    });
    fireEvent.change(screen.getByLabelText("手机号"), {
      target: { value: "13800138000" }
    });
    fireEvent.click(screen.getByRole("checkbox", { name: "同意用户协议和隐私政策" }));
    fireEvent.click(screen.getByRole("button", { name: "注册并创建账号" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("账号已存在，请直接登录");
  });

  it("returns to the requested route after login", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          access_token: "access-token",
          access_token_expires_at: new Date(Date.now() + 60000).toISOString(),
          user: {
            id: 42,
            nickname: "张晨",
            phone: "13800138000",
            status: "active"
          },
          is_new_user: false
        }),
        { status: 200 }
      )
    );

    render(
      <MemoryRouter
        initialEntries={[{ pathname: "/login", state: { returnTo: "/leads" } }]}
      >
        <Routes>
          <Route element={<LoginPage />} path="/login" />
          <Route element={<h1>AI线索开发</h1>} path="/leads" />
        </Routes>
      </MemoryRouter>
    );

    fireEvent.change(screen.getByLabelText("账号"), {
      target: { value: "deploy_user" }
    });
    fireEvent.change(screen.getByLabelText("密码"), {
      target: { value: "secret123" }
    });
    fireEvent.click(screen.getByRole("checkbox", { name: "同意用户协议和隐私政策" }));
    fireEvent.click(screen.getByRole("button", { name: "登录" }));

    expect(await screen.findByText("登录成功，正在进入工作台")).toBeInTheDocument();
    expect(await screen.findByRole("heading", { name: "AI线索开发" })).toBeInTheDocument();
  });
});
