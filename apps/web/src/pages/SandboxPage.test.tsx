import { cleanup, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { authSession } from "../lib/authSession";
import App from "../App";

describe("SandboxPage", () => {
  afterEach(() => {
    cleanup();
    authSession.clear();
  });

  it("renders the business sandbox workbench", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-23T12:00:00Z",
      is_new_user: false,
      user: {
        id: 7,
        nickname: "张婧",
        phone: "",
        account: "zhangjing",
        status: "active"
      }
    });

    render(<MemoryRouter initialEntries={["/sandbox"]}><App /></MemoryRouter>);

    expect(screen.getByRole("heading", { name: "商业沙盘" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /开始推演/ })).toHaveAttribute("href", "/sandbox/setup");
    expect(screen.getByText("多角色推演")).toBeInTheDocument();
    expect(screen.queryByText("第一版正在实现")).not.toBeInTheDocument();
  });

  it("renders sandbox setup and role selection states", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-23T12:00:00Z",
      is_new_user: false,
      user: {
        id: 7,
        nickname: "张婧",
        phone: "",
        account: "zhangjing",
        status: "active"
      }
    });

    render(<MemoryRouter initialEntries={["/sandbox/roles"]}><App /></MemoryRouter>);

    expect(screen.getByRole("heading", { name: "你希望从谁的视角进行推演？" })).toBeInTheDocument();
    expect(screen.getByText("用户视角")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /确认角色/ })).toHaveAttribute("href", "/sandbox/start");
  });

  it("renders sandbox questions, report, history and quota states", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-23T12:00:00Z",
      is_new_user: false,
      user: {
        id: 7,
        nickname: "张婧",
        phone: "",
        account: "zhangjing",
        status: "active"
      }
    });

    render(<MemoryRouter initialEntries={["/sandbox/questions"]}><App /></MemoryRouter>);
    expect(screen.getByRole("heading", { name: "你的目标用户更具体是哪些上班族？" })).toBeInTheDocument();
    cleanup();

    render(<MemoryRouter initialEntries={["/sandbox/report"]}><App /></MemoryRouter>);
    expect(screen.getByRole("heading", { name: "AI 驱动中小企业知识管理平台" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "核心结论" })).toBeInTheDocument();
    cleanup();

    render(<MemoryRouter initialEntries={["/sandbox/history"]}><App /></MemoryRouter>);
    expect(screen.getByRole("heading", { name: "历史推演" })).toBeInTheDocument();
    expect(screen.getByText("AI智能客服SaaS平台")).toBeInTheDocument();
    cleanup();

    render(<MemoryRouter initialEntries={["/sandbox/quota"]}><App /></MemoryRouter>);
    expect(screen.getByRole("dialog", { name: "本月沙盘次数已用尽" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "升级套餐" })).toHaveAttribute("href", "/membership");
  });
});
