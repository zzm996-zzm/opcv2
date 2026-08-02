import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import App from "../App";
import { authSession } from "../lib/authSession";
import type { SandboxOptions, SandboxSession } from "../lib/sandboxApi";

const options: SandboxOptions = {
  roles: [
    { key: "user", label: "用户视角", description: "评估产品体验与价值", badge: "推荐优先" },
    { key: "investor", label: "投资人视角", description: "评估市场潜力与回报", badge: "热门选择" },
    { key: "channel", label: "代理商 / 渠道方视角", description: "评估落地可行性", badge: "渠道必选" },
    { key: "competitor", label: "竞争对手视角", description: "评估竞争格局", badge: "深度分析" },
    { key: "operator", label: "运营视角", description: "评估执行与增长", badge: "运营必选" }
  ],
  system_perspectives: [],
  depths: [
    { value: "standard", label: "标准", description: "覆盖核心机会与风险", recommended: true },
    { value: "deep", label: "深度", description: "增加验证指标和时间线" }
  ],
  output_styles: [
    { value: "structured_report", label: "结构化报告", description: "按结论、机会和风险组织", recommended: true },
    { value: "concise_report", label: "精简摘要", description: "突出关键结论" }
  ],
  defaults: { depth: "standard", output_style: "structured_report", generate_outline: true, variables: {} }
};

function fixture(overrides: Partial<SandboxSession> = {}): SandboxSession {
  return {
    id: 42,
    user_id: 7,
    goal: "验证企业 AI 运营平台是否值得投入",
    target_users: "连锁门店老板",
    product: "企业 AI 运营平台",
    roles: [],
    status: "draft",
    progress_percent: 0,
    current_step: "questions",
    run_attempt: 0,
    intake: {
      status: "questions",
      initial_idea: "为连锁门店提供 AI 运营与客户转化平台",
      recognized_fields: [
        { key: "goal", label: "推演目标", value: "验证项目是否值得投入" },
        { key: "target_users", label: "目标用户", value: "连锁门店老板" },
        { key: "product", label: "产品方案", value: "企业 AI 运营平台" }
      ],
      questions: [
        { key: "pain", title: "目标客户最急需解决的问题是什么？", hint: "描述最影响成交的问题", placeholder: "请输入客户痛点", required: true, max_length: 1000, position: 1, skipped: false }
      ],
      answered_count: 0,
      total_questions: 1
    },
    settings: { ...options.defaults },
    created_at: "2026-07-20T08:00:00Z",
    updated_at: "2026-07-20T08:00:00Z",
    ...overrides
  };
}

function completedFixture(overrides: Partial<SandboxSession> = {}): SandboxSession {
  return fixture({
    status: "completed",
    current_step: "completed",
    progress_percent: 100,
    run_attempt: 1,
    roles: ["用户视角", "投资人视角"],
    intake: { ...fixture().intake, status: "ready", answered_count: 1, questions: [{ ...fixture().intake.questions[0], answer: "咨询回复和复购运营效率低" }] },
    report: {
      score: 88,
      summary: "项目具备小范围试点价值，应先验证付费意愿和交付成本。",
      report_version: "sandbox_report_v3",
      consumer_probability: 72,
      risk_level: "medium",
      recommendation_grade: "A-",
      metrics: [{ label: "市场吸引力", value: "8.4" }],
      role_summaries: [
        { role: "用户视角", view: "门店老板关注降本与响应效率。" },
        { role: "投资人视角", view: "需要验证留存和单位经济模型。" }
      ],
      risks: ["渠道教育成本偏高"],
      next_actions: ["先做 3 家门店试点"],
      core_conclusions: ["先用真实试点验证付费意愿"],
      opportunity_analysis: [{ title: "门店运营提效", detail: "自动整理客户信息和跟进建议。", tags: ["效率"] }],
      risk_analysis: [{ title: "交付成本", detail: "早期定制需求可能推高成本。", tags: ["交付"] }],
      action_plan: [{ order: 1, title: "客户访谈", detail: "访谈十家目标门店", duration: "1 周" }],
      growth_path: [{ stage: 1, title: "单场景验证", detail: "完成三家试点" }],
      validation_metrics: [{ label: "响应时间", current: "15 分钟", target: "1 分钟内", confidence_percent: 70 }],
      timeline: [{ title: "客户访谈与需求确认", period: "第 1 周" }]
    },
    ...overrides
  });
}

function response(body: unknown, status = 200) {
  return Promise.resolve(new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } }));
}

function signIn() {
  authSession.set({
    access_token: "access-token",
    access_token_expires_at: "2026-12-31T23:59:59Z",
    is_new_user: false,
    user: { id: 7, nickname: "张婧", phone: "", account: "zhangjing", status: "active" }
  });
}

describe("SandboxPage production flow", () => {
  afterEach(() => {
    cleanup();
    authSession.clear();
    vi.restoreAllMocks();
  });

  it("validates an empty homepage prompt without presenting the primary action as disabled", () => {
    signIn();
    render(
      <MemoryRouter initialEntries={["/sandbox"]}>
        <App />
      </MemoryRouter>
    );

    const startButton = screen.getByRole("button", { name: /开始推演/ });
    expect(startButton).toBeEnabled();
    fireEvent.click(startButton);
    expect(screen.getByRole("alert")).toHaveTextContent("请先描述要推演的项目或情况。");
  });

  it("reopens the homepage Copilot after it was closed on another route", async () => {
    signIn();
    render(
      <MemoryRouter initialEntries={["/sandbox"]}>
        <App />
      </MemoryRouter>
    );

    fireEvent.click(screen.getByRole("button", { name: "收起 Copilot" }));
    fireEvent.click(screen.getByRole("link", { name: "项目超市" }));
    fireEvent.click(screen.getByRole("link", { name: "商业沙盘" }));

    expect(await screen.findByRole("complementary", { name: "智活 Copilot" })).toBeInTheDocument();
  });

  it("creates an API-backed intake from the homepage and enters questions", async () => {
    signIn();
    const created = fixture();
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/sandbox/sessions/intake" && init?.method === "POST") return response(created);
      if (url === "/api/v1/sandbox/sessions/42" && init?.method === "GET") return response(created);
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    render(<MemoryRouter initialEntries={["/sandbox"]}><App /></MemoryRouter>);
    expect(screen.getByRole("heading", { name: "商业沙盘" })).toBeInTheDocument();
    expect(screen.getByText("多角色推演")).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText("描述你要推演的项目或情况"), { target: { value: "验证门店 AI 运营平台" } });
    fireEvent.click(screen.getByRole("button", { name: /开始推演/ }));

    expect(await screen.findByRole("heading", { name: "目标客户最急需解决的问题是什么？" })).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/sandbox/sessions/intake", expect.objectContaining({
      method: "POST",
      body: JSON.stringify({ initial_idea: "验证门店 AI 运营平台" })
    }));
  });

  it("persists the answer, completes intake, and renders the answer summary", async () => {
    signIn();
    let current = fixture();
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/sandbox/sessions/42" && init?.method === "GET") return response(current);
      if (url === "/api/v1/sandbox/sessions/42/intake/questions/pain" && init?.method === "PUT") {
        current = fixture({ intake: { ...current.intake, answered_count: 1, questions: [{ ...current.intake.questions[0], answer: "咨询回复慢，复购跟进依赖人工" }] }, updated_at: "2026-07-20T08:01:00Z" });
        return response(current);
      }
      if (url === "/api/v1/sandbox/sessions/42/intake/complete" && init?.method === "POST") {
        current = fixture({ intake: { ...current.intake, status: "ready", answered_count: 1 }, current_step: "ready", updated_at: "2026-07-20T08:02:00Z" });
        return response(current);
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    render(<MemoryRouter initialEntries={["/sandbox/questions?session=42"]}><App /></MemoryRouter>);
    fireEvent.change(await screen.findByLabelText("目标客户最急需解决的问题是什么？"), { target: { value: "咨询回复慢，复购跟进依赖人工" } });
    fireEvent.click(screen.getByRole("button", { name: /完成并查看摘要/ }));

    expect(await screen.findByText("你已输入的信息")).toBeInTheDocument();
    expect(screen.getByText("咨询回复慢，复购跟进依赖人工")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /选择推演角色/ })).toHaveAttribute("href", "/sandbox/roles?session=42");
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/sandbox/sessions/42/intake/complete", expect.objectContaining({ method: "POST" }));
  });

  it("loads selectable roles from options and saves the selection", async () => {
    signIn();
    let current = fixture({ intake: { ...fixture().intake, status: "ready" }, current_step: "ready" });
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/sandbox/options") return response(options);
      if (url === "/api/v1/sandbox/sessions/42") return response(current);
      if (url === "/api/v1/sandbox/sessions/42/draft" && init?.method === "PATCH") {
        current = { ...current, roles: ["用户视角"], updated_at: "2026-07-20T08:03:00Z" };
        return response(current);
      }
      if (url === "/api/v1/membership/usage") return response({ usage: [] });
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    render(<MemoryRouter initialEntries={["/sandbox/roles?session=42"]}><App /></MemoryRouter>);
    fireEvent.click(await screen.findByRole("button", { name: /用户视角/ }));
    fireEvent.click(screen.getByRole("button", { name: /确认角色，进入下一步/ }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/sandbox/sessions/42/draft", expect.objectContaining({
      method: "PATCH",
      body: JSON.stringify({ roles: ["用户视角"] })
    })));
    expect(await screen.findByRole("heading", { name: /推演设置信息/ })).toBeInTheDocument();
  });

  it("saves run settings and keeps the queued response when entering the run page", async () => {
    signIn();
    let current = fixture({ roles: ["用户视角", "投资人视角"], intake: { ...fixture().intake, status: "ready" }, current_step: "ready" });
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/sandbox/options") return response(options);
      if (url === "/api/v1/sandbox/sessions/42" && init?.method === "GET") return response(current);
      if (url === "/api/v1/membership/usage") return response({ usage: [{ key: "sandbox_runs", label: "商业沙盘", used: 0, limit: 1, unit: "次/月" }] });
      if (url === "/api/v1/sandbox/sessions/42/draft" && init?.method === "PATCH") {
        current = { ...current, settings: { ...current.settings, depth: "deep" }, updated_at: "2026-07-20T08:03:00Z" };
        return response(current);
      }
      if (url === "/api/v1/sandbox/sessions/42/run" && init?.method === "POST") {
        current = { ...current, status: "queued", current_step: "queued", run_attempt: 1, updated_at: "2026-07-20T08:04:00Z" };
        return response(current, 202);
      }
      if (url === "/api/v1/sandbox/sessions/42/messages") return response({ messages: [] });
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    render(<MemoryRouter initialEntries={["/sandbox/start?session=42"]}><App /></MemoryRouter>);
    fireEvent.change(await screen.findByLabelText("推演深度"), { target: { value: "deep" } });
    fireEvent.click(screen.getByRole("button", { name: /开始推演/ }));

    expect(await screen.findByText("等待执行")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /取消本轮推演/ })).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/sandbox/sessions/42/draft", expect.objectContaining({
      method: "PATCH",
      body: expect.stringContaining('"depth":"deep"')
    }));
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/sandbox/sessions/42/run", expect.objectContaining({ method: "POST" }));
  });

  it("opens the real quota dialog when usage is depleted", async () => {
    signIn();
    const current = fixture({ roles: ["用户视角"], intake: { ...fixture().intake, status: "ready" }, current_step: "ready" });
    vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/sandbox/options") return response(options);
      if (url === "/api/v1/sandbox/sessions/42") return response(current);
      if (url === "/api/v1/membership/usage") return response({ usage: [{ key: "sandbox_runs", label: "商业沙盘", used: 1, limit: 1, unit: "次/月" }] });
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    render(<MemoryRouter initialEntries={["/sandbox/start?session=42"]}><App /></MemoryRouter>);
    fireEvent.click(await screen.findByRole("button", { name: /开始推演/ }));

    expect(await screen.findByRole("dialog", { name: "本月沙盘次数已用尽" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /升级套餐/ })).toHaveAttribute("href", "/membership/upgrade");
    fireEvent.click(screen.getByRole("button", { name: "关闭额度弹窗" }));
    expect(screen.queryByRole("dialog", { name: "本月沙盘次数已用尽" })).not.toBeInTheDocument();
  });

  it("renders the API report and exposes a working report download command", async () => {
    signIn();
    const current = completedFixture();
    vi.spyOn(globalThis, "fetch").mockImplementation((input) => String(input) === "/api/v1/sandbox/sessions/42" ? response(current) : Promise.reject(new Error(`unexpected request: ${String(input)}`)));
    const createURL = vi.spyOn(URL, "createObjectURL").mockReturnValue("blob:report");
    vi.spyOn(URL, "revokeObjectURL").mockImplementation(() => undefined);
    const clickDownload = vi.spyOn(HTMLAnchorElement.prototype, "click").mockImplementation(() => undefined);

    render(<MemoryRouter initialEntries={["/sandbox/sessions/42/report"]}><App /></MemoryRouter>);

    expect(await screen.findByRole("heading", { name: "企业 AI 运营平台" })).toBeInTheDocument();
    expect(screen.getByText("门店运营提效")).toBeInTheDocument();
    expect(screen.getByText("客户访谈与需求确认")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: /导出报告/ }));
    expect(createURL).toHaveBeenCalledOnce();
    expect(clickDownload).toHaveBeenCalledOnce();
  });

  it("loads API history and applies keyword and status filters", async () => {
    signIn();
    const sessions = [completedFixture(), fixture({ id: 43, product: "餐饮门店助手", goal: "验证餐饮获客工具", created_at: "2026-07-19T08:00:00Z" })];
    vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      if (String(input) === "/api/v1/sandbox/sessions?limit=100") return response({ sessions });
      if (String(input) === "/api/v1/sandbox/examples") return response({ sessions: [] });
      return Promise.reject(new Error(`unexpected request: ${String(input)}`));
    });

    render(<MemoryRouter initialEntries={["/sandbox/history"]}><App /></MemoryRouter>);
    expect(await screen.findByText("企业 AI 运营平台")).toBeInTheDocument();
    expect(screen.getByText("餐饮门店助手")).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText("搜索项目名称或关键词"), { target: { value: "餐饮" } });
    expect(screen.queryByText("企业 AI 运营平台")).not.toBeInTheDocument();
    expect(screen.getByText("餐饮门店助手")).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("全部状态"), { target: { value: "completed" } });
    expect(screen.getByText("没有符合条件的推演记录")).toBeInTheDocument();
  });

  it("shows clearly marked API examples for an empty account and opens an example report", async () => {
    signIn();
    const example = completedFixture({ id: 900001, product: "AI智能客服SaaS平台", is_example: true, example_key: "ai-customer-service" });
    vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      if (String(input) === "/api/v1/sandbox/sessions?limit=100") return response({ sessions: [] });
      if (String(input) === "/api/v1/sandbox/examples") return response({ sessions: [example] });
      return Promise.reject(new Error(`unexpected request: ${String(input)}`));
    });

    render(<MemoryRouter initialEntries={["/sandbox/history"]}><App /></MemoryRouter>);
    expect(await screen.findByText("AI智能客服SaaS平台")).toBeInTheDocument();
    expect(screen.getByText(/仅用于体验页面和报告结构/)).toBeInTheDocument();
    const reportLink = screen.getByRole("link", { name: "查看报告" });
    expect(reportLink).toHaveAttribute("href", "/sandbox/report?example=ai-customer-service");
    fireEvent.click(reportLink);
    expect(await screen.findByRole("heading", { name: "AI智能客服SaaS平台" })).toBeInTheDocument();
    expect(screen.getByText("示例数据")).toBeInTheDocument();
  });

  it("posts a role follow-up and displays the persisted answer", async () => {
    signIn();
    const current = completedFixture();
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/sandbox/sessions/42") return response(current);
      if (url === "/api/v1/sandbox/sessions/42/messages" && init?.method === "GET") return response({ messages: [] });
      if (url === "/api/v1/sandbox/sessions/42/messages" && init?.method === "POST") return response({ id: 1, session_id: 42, user_id: 7, role: "投资人视角", question: "最关注什么？", answer: "最关注客户留存和单位经济模型。", created_at: "2026-07-20T09:00:00Z" });
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    render(<MemoryRouter initialEntries={["/sandbox/run?session=42"]}><App /></MemoryRouter>);
    fireEvent.click(await screen.findByRole("button", { name: "投资人视角" }));
    fireEvent.change(screen.getByLabelText("追加追问"), { target: { value: "最关注什么？" } });
    fireEvent.click(screen.getByRole("button", { name: /发送追问/ }));

    expect(await screen.findByText(/最关注客户留存和单位经济模型/)).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/sandbox/sessions/42/messages", expect.objectContaining({ method: "POST", body: JSON.stringify({ role: "投资人视角", question: "最关注什么？" }) }));
  });

  it("does not render a half-configured state without a session", () => {
    signIn();
    render(<MemoryRouter initialEntries={["/sandbox/roles"]}><App /></MemoryRouter>);

    expect(screen.getByRole("heading", { name: "当前页面缺少沙盘会话" })).toBeInTheDocument();
    expect(screen.getByText("返回商业沙盘首页").closest("a")).toHaveAttribute("href", "/sandbox");
  });
});
