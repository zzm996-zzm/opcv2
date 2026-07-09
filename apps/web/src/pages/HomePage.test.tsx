import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import { homeApi } from "../lib/homeApi";
import HomePage from "./HomePage";

vi.mock("../lib/homeApi", () => ({
  homeApi: {
    summary: vi.fn()
  }
}));

describe("HomePage", () => {
  afterEach(() => {
    authSession.clear();
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

  function mockHomeSummary() {
    vi.mocked(homeApi.summary).mockResolvedValue({
      metrics: [
        { label: "进行中任务", value: "3", icon: "folder" },
        { label: "待办任务", value: "4", icon: "inbox" },
        { label: "今日跟进", value: "8", icon: "trend" }
      ],
      hero_cards: [
        { title: "项目雷达", summary: "发现高潜力机会", url: "/projects" },
        { title: "落地任务", summary: "推进今日待办", url: "/tasks" }
      ],
      recommendations: [
        { title: "本地AI获客顾问", summary: "适合轻资产启动", url: "/projects/detail" },
        { title: "跟进今日客户", summary: "星河教育 等 2 位客户待跟进", url: "/crm" }
      ],
      action_items: [
        { type: "project", priority: "medium", title: "本地AI获客顾问", summary: "适合轻资产启动", url: "/projects/detail", cta: "查看项目" },
        { type: "crm", priority: "high", title: "跟进今日客户", summary: "星河教育 等 2 位客户待跟进", url: "/crm", cta: "去跟进" }
      ],
      recent_tasks: [
        { id: 41, title: "联调首页聚合接口", project: "工作台", status: "in_progress", due_at: "2026-07-02T10:00:00Z" }
      ],
      notification_summary: {
        unread: 1,
        latest: [
          { id: 7, type: "task", title: "任务提醒", summary: "联调首页聚合接口即将截止", action_url: "/tasks", created_at: "2026-07-02T09:30:00Z" }
        ]
      },
      account_summary: {
        plan_name: "会员版",
        credit_balance: 88,
        quota_warnings: [{ key: "analysis", label: "AI分析额度", used: 28, limit: 30, message: "AI分析额度即将用完" }]
      }
    });
  }

  it("loads home summary from API", async () => {
    signIn();
    mockHomeSummary();

    render(
      <MemoryRouter>
        <HomePage />
      </MemoryRouter>
    );

    expect(await screen.findByRole("heading", { name: "项目雷达" })).toBeInTheDocument();
    expect(screen.getByText("本地AI获客顾问")).toBeInTheDocument();
    expect(screen.getByText("跟进今日客户")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /跟进今日客户/ })).toHaveAttribute("href", "/crm");
    expect(screen.getByText("CRM客户 · 去跟进")).toBeInTheDocument();
    expect(screen.getByText("优先处理")).toBeInTheDocument();
    expect(screen.getByText("联调首页聚合接口")).toBeInTheDocument();
    expect(screen.getByText("会员版")).toBeInTheDocument();
    expect(screen.getByText("88 积分")).toBeInTheDocument();
    expect(screen.getByText("AI分析额度即将用完")).toBeInTheDocument();
    expect(screen.getByText("今日跟进")).toBeInTheDocument();
    expect(screen.getByText("8")).toBeInTheDocument();
    expect(homeApi.summary).toHaveBeenCalled();
  });

  it("renders designed empty states instead of static dashboard records", async () => {
    signIn();
    vi.mocked(homeApi.summary).mockResolvedValue({
      metrics: [],
      hero_cards: [],
      recommendations: [],
      action_items: [],
      recent_tasks: [],
      notification_summary: {
        unread: 0,
        latest: []
      },
      account_summary: {
        plan_name: "基础版",
        credit_balance: 0,
        quota_warnings: []
      }
    });

    render(
      <MemoryRouter>
        <HomePage menuState="notice" />
      </MemoryRouter>
    );

    expect(await screen.findByText("暂无待处理行动")).toBeInTheDocument();
    expect(screen.getByText("暂无待办任务")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /先去项目超市/ })).toHaveAttribute("href", "/projects");
    expect(screen.getByRole("link", { name: /咨询 Copilot/ })).toHaveAttribute("href", "/copilot");
    expect(screen.getByText("暂无通知")).toBeInTheDocument();
    expect(screen.queryByText("完成【AI 智能硬件】项目商业画布")).not.toBeInTheDocument();
    expect(screen.queryByText("智能匹配")).not.toBeInTheDocument();
    expect(screen.queryByText("竞品价格监测数据已更新完成")).not.toBeInTheDocument();
  });

  it("shows the signed-in user and logs out from the account menu", async () => {
    signIn();
    mockHomeSummary();
    const fetchMock = vi
      .spyOn(globalThis, "fetch")
      .mockResolvedValue(new Response(null, { status: 204 }));

    render(
      <MemoryRouter>
        <HomePage />
      </MemoryRouter>
    );

    fireEvent.click(screen.getByRole("button", { name: "张晨的账号菜单" }));
    const accountMenu = await screen.findByRole("dialog", { name: "头像下拉框" });
    expect(within(accountMenu).getByText("88 积分")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "退出登录" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
    expect(await screen.findByRole("link", { name: "登录 / 注册" })).toBeInTheDocument();
  });

  it("clears the local session even when logout cannot reach the API", async () => {
    signIn();
    mockHomeSummary();
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

  it("opens notification menu and Copilot utility states", async () => {
    signIn();
    mockHomeSummary();

    render(
      <MemoryRouter>
        <HomePage />
      </MemoryRouter>
    );

    fireEvent.click(screen.getByRole("button", { name: "通知" }));
    expect(screen.getByRole("dialog", { name: "通知下拉框" })).toBeInTheDocument();
    expect(await screen.findByText("任务提醒")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "打开智活 Copilot" }));
    fireEvent.click(screen.getByRole("button", { name: "打开 Copilot 设置" }));
    expect(screen.getByText("Copilot 设置")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Claude opus4.8" })).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "添加文件" }));
    expect(screen.getByText("市场分析报告.pdf")).toBeInTheDocument();
  });

  it("renders direct Copilot settings and file states", () => {
    signIn();
    mockHomeSummary();

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
    signIn();
    mockHomeSummary();

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
