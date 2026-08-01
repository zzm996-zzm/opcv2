import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "./lib/authSession";
import { membershipApi } from "./lib/membershipApi";
import App from "./App";

vi.mock("./lib/membershipApi", async (importActual) => {
  const actual = await importActual<typeof import("./lib/membershipApi")>();
  return {
    ...actual,
    membershipApi: {
      ...actual.membershipApi,
      featureAccess: vi.fn()
    }
  };
});

describe("App", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  function mockLockedLeadFeature() {
    vi.mocked(membershipApi.featureAccess).mockResolvedValue({
      features: [{
        key: "ai_lead_development",
        label: "AI线索开发",
        status: "locked",
        required_plan: "pro",
        message: "该模块当前版本仅开放入口展示，真实工作流暂未对外启用。",
        allow_read_only: false,
        allow_workflow: false
      }]
    });
  }

  it("keeps the application shell visible while authentication is restoring", () => {
    render(
      <MemoryRouter initialEntries={["/dashboard"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByRole("navigation", { name: "顶部全局功能区" })).toBeInTheDocument();
    expect(screen.getByRole("complementary", { name: "产品侧边导航" })).toBeInTheDocument();
    expect(screen.getByRole("status", { name: "恢复登录状态" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "打开智活 Copilot" })).not.toBeInTheDocument();
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

  it("shows one draggable Copilot orb on product routes and hides it on Copilot", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2099-06-11T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张晨", phone: "13800138000", status: "active" }
    });

    const productRoute = render(
      <MemoryRouter initialEntries={["/help"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getAllByRole("button", { name: "打开智活 Copilot" })).toHaveLength(1);
    productRoute.unmount();

    render(
      <MemoryRouter initialEntries={["/copilot"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.queryByRole("button", { name: "打开智活 Copilot" })).not.toBeInTheDocument();
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

  it("renders a product route for a signed-in user", async () => {
    mockLockedLeadFeature();
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2099-06-11T12:00:00Z",
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

    expect(await screen.findByRole("heading", { name: "AI线索开发" })).toBeInTheDocument();
  });

  it("redirects account settings alias to the existing settings page", async () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2099-06-11T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张晨", phone: "13800138000", status: "active" }
    });

    render(
      <MemoryRouter initialEntries={["/account/settings"]}>
        <App />
      </MemoryRouter>
    );

    expect(await screen.findByRole("heading", { name: "账号与资料设置" })).toBeInTheDocument();
  });

  it("renders an in-app not-found page for an unknown authenticated route", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2099-06-11T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张晨", phone: "13800138000", status: "active" }
    });

    render(
      <MemoryRouter initialEntries={["/does-not-exist"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "页面不存在" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "返回工作台" })).toHaveAttribute("href", "/");
  });

  it("renders the locked board-three navigation on visible but unavailable growth pages", async () => {
    mockLockedLeadFeature();
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2099-06-11T12:00:00Z",
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

    expect(await screen.findByRole("heading", { name: "AI线索开发" })).toBeInTheDocument();
    expect(screen.getByRole("navigation", { name: "顶部全局功能区" })).toBeInTheDocument();
    expect(screen.getByRole("complementary", { name: "产品侧边导航" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "智活 Copilot" })).toHaveAttribute("href", "/copilot");
    expect(screen.getByRole("link", { name: "GEO获客" })).toHaveAttribute("href", "/geo");
  });

  it("renders first-class V4 account routes for a signed-in user", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2099-06-11T12:00:00Z",
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
      access_token_expires_at: "2099-06-11T12:00:00Z",
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

    expect(screen.getByRole("heading", { name: "正在加载能力评估..." })).toBeInTheDocument();
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
    expect(screen.queryByRole("button", { name: "打开智活 Copilot" })).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "收起 Copilot" }));
    expect(screen.queryByRole("complementary", { name: "智活 Copilot 咨询助手" })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "打开智活 Copilot" })).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "打开智活 Copilot" }));
    expect(screen.getByRole("complementary", { name: "智活 Copilot 咨询助手" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "打开智活 Copilot" })).not.toBeInTheDocument();
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
    expect(screen.getByRole("heading", { name: "开始 AI 对比分析" })).toBeInTheDocument();
    expect(screen.queryByText("回答完成")).not.toBeInTheDocument();
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

    expect(screen.getByRole("heading", { name: "正在加载差距分析..." })).toBeInTheDocument();
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

    expect(screen.getByRole("heading", { name: "正在加载学习建议..." })).toBeInTheDocument();
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

    expect(screen.getByRole("heading", { name: "正在加载学习计划..." })).toBeInTheDocument();
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

    expect(screen.getByRole("heading", { name: "正在加载诊断报告..." })).toBeInTheDocument();
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

    expect(screen.getByRole("heading", { name: "全部课程" })).toBeInTheDocument();
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

    expect(screen.getByRole("heading", { name: "推荐课程" })).toBeInTheDocument();
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

    expect(screen.getByRole("heading", { name: "正在加载课程..." })).toBeInTheDocument();
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

    expect(screen.getByRole("heading", { name: "正在加载课程学习页..." })).toBeInTheDocument();
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
