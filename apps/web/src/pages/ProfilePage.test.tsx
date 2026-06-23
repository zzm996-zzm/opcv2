import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { authSession } from "../lib/authSession";
import ProfilePage from "./ProfilePage";

describe("ProfilePage", () => {
  afterEach(() => {
    authSession.clear();
  });

  it("renders the profile overview and quota cards", () => {
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
        <ProfilePage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "个人中心" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "张晨" })).toBeInTheDocument();
    expect(screen.getByText("AI智算额度")).toBeInTheDocument();
  });

  it("renders account profile settings", () => {
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
        <ProfilePage mode="settings" />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "账号与资料设置" })).toBeInTheDocument();
    expect(screen.getByText("资料完成度")).toBeInTheDocument();
    expect(screen.getByText("用户资料")).toBeInTheDocument();
    expect(screen.getByText("智活科技有限公司")).toBeInTheDocument();
    expect(screen.getByText("138****5678")).toBeInTheDocument();
  });

  it("renders unbound account settings state", () => {
    render(
      <MemoryRouter>
        <ProfilePage binding="unbound" mode="settings" />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "账号与资料设置" })).toBeInTheDocument();
    expect(screen.getByText("可选联系方式")).toBeInTheDocument();
    expect(screen.getByText("未绑定")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "去填写" })).toBeInTheDocument();
  });

  it("renders my content records", () => {
    render(
      <MemoryRouter>
        <ProfilePage mode="content" />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "我的内容" })).toBeInTheDocument();
    expect(screen.getByText("智能客服系统项目匹配")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /增长测算/ })).toBeInTheDocument();
  });

  it("renders preference settings", () => {
    render(
      <MemoryRouter>
        <ProfilePage mode="preferences" />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "偏好设置" })).toBeInTheDocument();
    expect(screen.getByText("通知设置")).toBeInTheDocument();
    expect(screen.getByText("默认模型")).toBeInTheDocument();
  });

  it.each([
    ["password", "修改密码", "确认修改"],
    ["logout", "退出登录", "确认退出"],
    ["delete", "注销账号", "确认注销"],
    ["complete", "完善资料", "保存并完成"]
  ] as const)("renders the %s account overlay", (_kind, title, action) => {
    render(
      <MemoryRouter>
        <ProfilePage mode="settings" overlay={_kind} />
      </MemoryRouter>
    );

    expect(screen.getByRole("dialog", { name: title })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: action })).toBeInTheDocument();
  });
});
