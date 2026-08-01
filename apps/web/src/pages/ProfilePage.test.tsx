import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { accountApi } from "../lib/accountApi";
import { authSession } from "../lib/authSession";
import ProfilePage from "./ProfilePage";

vi.mock("../lib/accountApi", () => ({
  accountApi: {
    getProfile: vi.fn(),
    updateProfile: vi.fn(),
    getOnboarding: vi.fn(),
    saveOnboarding: vi.fn(),
    completeOnboarding: vi.fn(),
    getPreferences: vi.fn(),
    updatePreferences: vi.fn(),
    getQuotas: vi.fn(),
    listContent: vi.fn(),
    deleteAccount: vi.fn()
  }
}));

describe("ProfilePage", () => {
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

  function mockAccount() {
    vi.mocked(accountApi.getProfile).mockResolvedValue({
      profile: {
        id: 7,
        nickname: "张晨",
        phone: "13800138000",
        email: "founder@example.com",
        wechat: "zhihuo_ai",
        company: "深度科技有限公司",
        industry: "人工智能",
        role: "创始人",
        onboarding_completed: false,
        created_at: "2026-07-01T10:00:00Z",
        updated_at: "2026-07-02T10:00:00Z"
      },
      bindings: [
        { type: "phone", masked_value: "138****8000", bound: true },
        { type: "wechat", masked_value: "zhihuo_ai", bound: true }
      ]
    });
    vi.mocked(accountApi.getOnboarding).mockResolvedValue({
      completed: false,
      sections: [
        { key: "identity", title: "基本身份", fields: { 姓名: "张晨", 身份角色: "创始人" } },
        { key: "company", title: "我的业务 / 公司", fields: { 公司名称: "深度科技有限公司", 所在行业: "人工智能" } }
      ]
    });
    vi.mocked(accountApi.getQuotas).mockResolvedValue({
      quotas: [{ key: "analysis", label: "AI分析额度", used: 8, limit: 30, unit: "次/月" }]
    });
    vi.mocked(accountApi.listContent).mockResolvedValue({
      items: [{
        id: "project-1",
        type: "project",
        title: "本地AI获客顾问项目匹配",
        summary: "基于资源与预算生成的项目匹配记录",
        url: "/projects/detail",
        created_at: "2026-07-02T10:15:00Z"
      }]
    });
    vi.mocked(accountApi.getPreferences).mockResolvedValue({
      notifications_enabled: true,
      default_model: "deepseek",
      language: "zh-CN",
      timezone: "Asia/Shanghai"
    });
  }

  it("renders the profile overview and quota cards from API", async () => {
    signIn();
    mockAccount();

    render(
      <MemoryRouter>
        <ProfilePage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "个人中心" })).toBeInTheDocument();
    expect(await screen.findByRole("heading", { name: "张晨" })).toBeInTheDocument();
    expect(screen.getByText("深度科技有限公司")).toBeInTheDocument();
    expect(screen.getByText("AI分析额度")).toBeInTheDocument();
    expect(screen.getByText("8")).toBeInTheDocument();
    expect(accountApi.getProfile).toHaveBeenCalled();
    expect(accountApi.getQuotas).toHaveBeenCalled();
  });

  it("renders account profile settings from profile bindings and onboarding", async () => {
    signIn();
    mockAccount();

    render(
      <MemoryRouter>
        <ProfilePage mode="settings" />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "账号与资料设置" })).toBeInTheDocument();
    expect(await screen.findByText("用户画像资料编辑")).toBeInTheDocument();
    expect(screen.getByText("公司名称：深度科技有限公司")).toBeInTheDocument();
    expect(screen.getByText("138****8000")).toBeInTheDocument();
    expect(screen.getByText("zhihuo_ai")).toBeInTheDocument();
  });

  it("renders unbound account settings state", async () => {
    mockAccount();
    vi.mocked(accountApi.getProfile).mockResolvedValueOnce({
      profile: {
        id: 7,
        nickname: "张晨",
        phone: "",
        email: "",
        wechat: "",
        company: "",
        industry: "",
        role: "",
        onboarding_completed: false,
        created_at: "2026-07-01T10:00:00Z",
        updated_at: "2026-07-02T10:00:00Z"
      },
      bindings: [
        { type: "phone", masked_value: "未填写", bound: false },
        { type: "wechat", masked_value: "未绑定", bound: false }
      ]
    });

    render(
      <MemoryRouter>
        <ProfilePage binding="unbound" mode="settings" />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "账号与资料设置" })).toBeInTheDocument();
    expect(await screen.findByText("可选联系方式")).toBeInTheDocument();
    expect(screen.getAllByText("未绑定").length).toBeGreaterThan(0);
    expect(screen.getAllByRole("button", { name: "去填写" }).length).toBeGreaterThan(0);
  });

  it("renders my content records from API", async () => {
    mockAccount();
    render(
      <MemoryRouter>
        <ProfilePage mode="content" />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "我的内容" })).toBeInTheDocument();
    expect(await screen.findByText("本地AI获客顾问项目匹配")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /增长测算/ })).toBeInTheDocument();
    expect(accountApi.listContent).toHaveBeenCalledWith(20);
  });

  it("shows reference profile states when APIs return empty collections", async () => {
    vi.mocked(accountApi.getProfile).mockResolvedValue({
      profile: {
        id: 7,
        nickname: "",
        phone: "",
        email: "",
        wechat: "",
        company: "",
        industry: "",
        role: "",
        onboarding_completed: false,
        created_at: "2026-07-01T10:00:00Z",
        updated_at: "2026-07-02T10:00:00Z"
      },
      bindings: []
    });
    vi.mocked(accountApi.getOnboarding).mockResolvedValue({ completed: false, sections: [] });
    vi.mocked(accountApi.getQuotas).mockResolvedValue({ quotas: [] });
    vi.mocked(accountApi.listContent).mockResolvedValue({ items: [] });
    vi.mocked(accountApi.getPreferences).mockResolvedValue({
      notifications_enabled: true,
      default_model: "deepseek",
      language: "zh-CN",
      timezone: "Asia/Shanghai"
    });

    const { rerender } = render(
      <MemoryRouter>
        <ProfilePage />
      </MemoryRouter>
    );

    expect(await screen.findByText("AI 智算额度")).toBeInTheDocument();
    expect(screen.getByText(/查看了项目拆解结果/)).toBeInTheDocument();

    rerender(
      <MemoryRouter>
        <ProfilePage mode="settings" />
      </MemoryRouter>
    );

    expect(await screen.findByText("公司名称：智活AI科技有限公司")).toBeInTheDocument();
    expect(screen.getByText("138 **** 5678")).toBeInTheDocument();

    rerender(
      <MemoryRouter>
        <ProfilePage mode="content" />
      </MemoryRouter>
    );

    expect(await screen.findByText("智能客服系统项目匹配")).toBeInTheDocument();
  });

  it("renders and saves preference settings", async () => {
    mockAccount();
    vi.mocked(accountApi.updatePreferences).mockResolvedValue({
      notifications_enabled: true,
      default_model: "deepseek",
      language: "zh-CN",
      timezone: "Asia/Shanghai"
    });

    render(
      <MemoryRouter>
        <ProfilePage mode="preferences" />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "偏好设置" })).toBeInTheDocument();
    expect(await screen.findByText("通知设置")).toBeInTheDocument();
    expect(screen.getByText("默认模型")).toBeInTheDocument();
    expect(screen.getByText("当前默认模型：deepseek")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "保存设置" }));

    await waitFor(() => expect(accountApi.updatePreferences).toHaveBeenCalledWith({
      notifications_enabled: true,
      default_model: "deepseek",
      language: "zh-CN",
      timezone: "Asia/Shanghai"
    }));
  });

  it("exposes interactive notification switches and model selection", async () => {
    mockAccount();
    vi.mocked(accountApi.updatePreferences).mockResolvedValue({
      notifications_enabled: false,
      default_model: "chatgpt-5.5",
      language: "zh-CN",
      timezone: "Asia/Shanghai"
    });

    render(
      <MemoryRouter>
        <ProfilePage mode="preferences" />
      </MemoryRouter>
    );

    const switches = await screen.findAllByRole("switch");
    expect(switches[0]).toHaveAttribute("aria-checked", "true");
    fireEvent.click(switches[0]);
    expect(switches[0]).toHaveAttribute("aria-checked", "false");

    const chatgpt = screen.getByRole("radio", { name: "ChatGPT 5.5" });
    fireEvent.click(chatgpt);
    expect(chatgpt).toHaveAttribute("aria-checked", "true");

    fireEvent.click(screen.getByRole("button", { name: "保存设置" }));
    await waitFor(() => expect(accountApi.updatePreferences).toHaveBeenCalledWith({
      notifications_enabled: false,
      default_model: "chatgpt-5.5",
      language: "zh-CN",
      timezone: "Asia/Shanghai"
    }));
  });

  it.each([
    ["password", "修改密码", "确认修改"],
    ["logout", "退出登录", "确认退出"],
    ["delete", "注销账号", "确认注销"],
    ["complete", "完善资料", "保存并完成"]
  ] as const)("renders the %s account overlay", (_kind, title, action) => {
    mockAccount();
    render(
      <MemoryRouter>
        <ProfilePage mode="settings" overlay={_kind} />
      </MemoryRouter>
    );

    expect(screen.getByRole("dialog", { name: title })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: action })).toBeInTheDocument();
  });

  it("deletes the account from the delete overlay", async () => {
    mockAccount();
    vi.mocked(accountApi.deleteAccount).mockResolvedValue({ status: "pending_deletion" });

    render(
      <MemoryRouter>
        <ProfilePage mode="settings" overlay="delete" />
      </MemoryRouter>
    );

    fireEvent.click(screen.getByLabelText("我已阅读并同意注销须知"));
    fireEvent.click(screen.getByRole("button", { name: "确认注销" }));

    await waitFor(() => expect(accountApi.deleteAccount).toHaveBeenCalled());
    expect(await screen.findByText("账号注销已提交")).toBeInTheDocument();
  });
});
