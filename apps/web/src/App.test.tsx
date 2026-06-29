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

  it("renders the CDK top navigation on reference pages", () => {
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

    expect(screen.getByRole("navigation", { name: "CDK 顶部导航" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "VIP获客" })).toHaveClass("active");
    expect(screen.getByRole("link", { name: "会员计划" })).toHaveAttribute("href", "/membership");
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

  it("renders the learning assessment route for a signed-in user", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-16T12:00:00Z",
      is_new_user: false,
      user: {
        id: 7,
        nickname: "张婧",
        phone: "13800138000",
        status: "active"
      }
    });

    render(
      <MemoryRouter initialEntries={["/learning/assessment"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByText("当前分析阶段：能力评估")).toBeInTheDocument();
  });

  it("renders the insights route for a signed-in user", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-18T12:00:00Z",
      is_new_user: false,
      user: {
        id: 7,
        nickname: "张婧",
        phone: "",
        account: "zhangjing",
        status: "active"
      }
    });

    render(
      <MemoryRouter initialEntries={["/insights"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "咨询通" })).toBeInTheDocument();
    expect(screen.queryByText("第一版正在实现")).not.toBeInTheDocument();
  });

  it("renders the tool recommendation plan route for a signed-in user", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-21T12:00:00Z",
      is_new_user: false,
      user: {
        id: 7,
        nickname: "张晨",
        phone: "",
        account: "zhangchen",
        status: "active"
      }
    });

    render(
      <MemoryRouter initialEntries={["/tools/recommendation-plan"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: /整套工具方案/ })).toBeInTheDocument();
    expect(screen.queryByText("第一版正在实现")).not.toBeInTheDocument();
  });

  it("renders the Copilot comparison route for a signed-in user", () => {
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

    render(
      <MemoryRouter initialEntries={["/copilot/compare"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "AI 对比分析" })).toBeInTheDocument();
    expect(screen.getAllByText("回答完成")).toHaveLength(3);
    expect(screen.queryByText("第一版正在实现")).not.toBeInTheDocument();
  });

  it("renders the business sandbox route for a signed-in user", () => {
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

    render(
      <MemoryRouter initialEntries={["/sandbox"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "商业沙盘" })).toBeInTheDocument();
    expect(screen.queryByText("第一版正在实现")).not.toBeInTheDocument();
  });

  it("renders the competitor data route for a signed-in user", () => {
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

    render(
      <MemoryRouter initialEntries={["/competitor-data"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "竞品全盘数据破解" })).toBeInTheDocument();
    expect(screen.queryByText("第一版正在实现")).not.toBeInTheDocument();
  });

  it("renders the learning gap analysis route for a signed-in user", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-16T12:00:00Z",
      is_new_user: false,
      user: {
        id: 7,
        nickname: "张婧",
        phone: "13800138000",
        status: "active"
      }
    });

    render(
      <MemoryRouter initialEntries={["/learning/gap-analysis"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByText("目标要求 vs 当前水平")).toBeInTheDocument();
  });

  it("renders the learning recommendation route for a signed-in user", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-16T12:00:00Z",
      is_new_user: false,
      user: {
        id: 7,
        nickname: "张婧",
        phone: "13800138000",
        status: "active"
      }
    });

    render(
      <MemoryRouter initialEntries={["/learning/recommendation"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByText("优先补强方向")).toBeInTheDocument();
  });

  it("renders the learning plan route for a signed-in user", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-16T12:00:00Z",
      is_new_user: false,
      user: {
        id: 7,
        nickname: "张婧",
        phone: "13800138000",
        status: "active"
      }
    });

    render(
      <MemoryRouter initialEntries={["/learning/plan"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByText("预计总学习时长")).toBeInTheDocument();
  });

  it("renders the learning report route for a signed-in user", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-16T12:00:00Z",
      is_new_user: false,
      user: {
        id: 7,
        nickname: "张婧",
        phone: "13800138000",
        status: "active"
      }
    });

    render(
      <MemoryRouter initialEntries={["/learning/report"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByText("诊断概览")).toBeInTheDocument();
  });

  it("renders the learning courses route for a signed-in user", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-16T12:00:00Z",
      is_new_user: false,
      user: {
        id: 7,
        nickname: "张婧",
        phone: "13800138000",
        status: "active"
      }
    });

    render(
      <MemoryRouter initialEntries={["/learning/courses"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByText("本周热门课程")).toBeInTheDocument();
  });

  it("renders the learning recommended courses route for a signed-in user", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-16T12:00:00Z",
      is_new_user: false,
      user: {
        id: 7,
        nickname: "张婧",
        phone: "13800138000",
        status: "active"
      }
    });

    render(
      <MemoryRouter initialEntries={["/learning/recommended-courses"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByText("推荐逻辑说明")).toBeInTheDocument();
  });

  it("renders the learning course intro route for a signed-in user", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-16T12:00:00Z",
      is_new_user: false,
      user: {
        id: 7,
        nickname: "张婧",
        phone: "13800138000",
        status: "active"
      }
    });

    render(
      <MemoryRouter initialEntries={["/learning/courses/intro"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByText("章节目录")).toBeInTheDocument();
  });

  it("renders the learning course detail route for a signed-in user", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-16T12:00:00Z",
      is_new_user: false,
      user: {
        id: 7,
        nickname: "张婧",
        phone: "13800138000",
        status: "active"
      }
    });

    render(
      <MemoryRouter initialEntries={["/learning/courses/detail"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByText("课程大纲")).toBeInTheDocument();
  });

  it("renders the learning history route for a signed-in user", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-16T12:00:00Z",
      is_new_user: false,
      user: {
        id: 7,
        nickname: "张婧",
        phone: "13800138000",
        status: "active"
      }
    });

    render(
      <MemoryRouter initialEntries={["/learning/history"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByText("最近学习记录")).toBeInTheDocument();
  });

  it("renders the community route for a signed-in user", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-17T12:00:00Z",
      is_new_user: false,
      user: {
        id: 7,
        nickname: "张婧",
        phone: "13800138000",
        status: "active"
      }
    });

    render(
      <MemoryRouter initialEntries={["/community"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByText("创业成长互助社区")).toBeInTheDocument();
  });

  it("renders the member community join route for a signed-in user", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-17T12:00:00Z",
      is_new_user: false,
      user: {
        id: 7,
        nickname: "张婧",
        phone: "13800138000",
        status: "active"
      }
    });

    render(
      <MemoryRouter initialEntries={["/community/members"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByRole("dialog", { name: "加入会员社群" })).toBeInTheDocument();
  });

  it("renders the enterprise community join route for a signed-in user", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-17T12:00:00Z",
      is_new_user: false,
      user: {
        id: 7,
        nickname: "张婧",
        phone: "13800138000",
        status: "active"
      }
    });

    render(
      <MemoryRouter initialEntries={["/community/enterprise"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByRole("dialog", { name: "加入企业社群" })).toBeInTheDocument();
  });
});
