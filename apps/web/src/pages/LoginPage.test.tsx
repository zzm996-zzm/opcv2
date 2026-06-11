import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import LoginPage from "./LoginPage";

describe("LoginPage", () => {
  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
  });

  it("requires agreement before login", () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    );

    fireEvent.change(screen.getByLabelText("昵称"), { target: { value: "张晨" } });
    fireEvent.change(screen.getByLabelText("手机号"), {
      target: { value: "13800138000" }
    });
    fireEvent.change(screen.getByLabelText("验证码"), {
      target: { value: "246810" }
    });

    expect(screen.getByRole("button", { name: "开始使用" })).toBeDisabled();
    fireEvent.click(screen.getByRole("checkbox", { name: "同意用户协议和隐私政策" }));
    expect(screen.getByRole("button", { name: "开始使用" })).toBeEnabled();
  });

  it("sends a code and logs in", async () => {
    const fetchMock = vi
      .spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ status: "sent" }), { status: 202 }))
      .mockResolvedValueOnce(
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

    fireEvent.change(screen.getByLabelText("昵称"), { target: { value: "张晨" } });
    fireEvent.change(screen.getByLabelText("手机号"), {
      target: { value: "13800138000" }
    });
    fireEvent.click(screen.getByRole("button", { name: "获取验证码" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
    fireEvent.change(screen.getByLabelText("验证码"), {
      target: { value: "246810" }
    });
    fireEvent.click(screen.getByRole("checkbox", { name: "同意用户协议和隐私政策" }));
    fireEvent.click(screen.getByRole("button", { name: "开始使用" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
    expect(await screen.findByText("登录成功，正在进入工作台")).toBeInTheDocument();
  });

  it("returns to the requested route after login", async () => {
    vi.useFakeTimers();
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
          <Route element={<h1>VIP获客</h1>} path="/leads" />
        </Routes>
      </MemoryRouter>
    );

    fireEvent.change(screen.getByLabelText("昵称"), { target: { value: "张晨" } });
    fireEvent.change(screen.getByLabelText("手机号"), {
      target: { value: "13800138000" }
    });
    fireEvent.change(screen.getByLabelText("验证码"), {
      target: { value: "246810" }
    });
    fireEvent.click(screen.getByRole("checkbox", { name: "同意用户协议和隐私政策" }));
    fireEvent.click(screen.getByRole("button", { name: "开始使用" }));

    await vi.runAllTimersAsync();
    expect(screen.getByRole("heading", { name: "VIP获客" })).toBeInTheDocument();
  });
});
