import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import App from "../App";
import { authSession } from "../lib/authSession";

describe("GrowthCalculatorPage", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  function signIn() {
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
  }

  function renderGrowthRoute(route = "/growth-calculator") {
    signIn();
    render(
      <MemoryRouter initialEntries={[route]}>
        <App />
      </MemoryRouter>
    );
  }

  it("renders the growth calculator workbench with empty backend state", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ models: [] }), { status: 200 })
    );

    renderGrowthRoute();

    expect(screen.getByRole("heading", { name: "增长测算" })).toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: "业务与增长问题" })).toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: "业务与增长问题" })).toHaveAttribute("maxlength", "2000");
    expect(screen.getByRole("button", { name: "开始测算" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "增长漏斗" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "情景对比" })).toBeInTheDocument();
    expect(await screen.findByText("暂无测算模型")).toBeInTheDocument();
    expect(screen.getByText("暂无情景对比")).toBeInTheDocument();
    expect(screen.queryByText("保守方案")).not.toBeInTheDocument();
    expect(screen.queryByText("第一版正在实现")).not.toBeInTheDocument();

    const quarterlyButton = screen.getByRole("button", { name: "季度" });
    expect(quarterlyButton).toHaveAttribute("aria-pressed", "false");
    fireEvent.click(quarterlyButton);
    expect(quarterlyButton).toHaveAttribute("aria-pressed", "true");
  });

  it.each([
    "/growth-calculator/questions",
    "/growth-calculator/report"
  ])("serves the live growth workbench at %s", async (route) => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ models: [] }), { status: 200 })
    );

    renderGrowthRoute(route);

    expect(screen.getByRole("heading", { name: "增长测算" })).toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: "业务与增长问题" })).toBeInTheDocument();
    expect(await screen.findByText("暂无测算模型")).toBeInTheDocument();
  });

  it("loads the report model selected in the URL", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/growth/models/99") {
        return Promise.resolve(new Response(JSON.stringify({
          id: 99, user_id: 7, name: "指定报告模型",
          assumptions: { monthly_visits: 12000, lead_rate: 0.08, deal_rate: 0.15, average_order: 6000, acquisition_cost: 80, delivery_cost: 120000 },
          result: { monthly_revenue: 864000, leads: 960, deals: 144, payback_days: 7, net_margin: 0.77 },
          created_at: "2026-08-15T08:00:00Z", updated_at: "2026-08-15T08:00:00Z"
        }), { status: 200 }));
      }
      if (url.endsWith("/scenarios")) return Promise.resolve(new Response(JSON.stringify({ scenarios: [] }), { status: 200 }));
      if (url.endsWith("/forecast")) return Promise.resolve(new Response(JSON.stringify({ months: [] }), { status: 200 }));
      if (url.endsWith("/recommendations")) return Promise.resolve(new Response(JSON.stringify({ action_items: [], cost_items: [] }), { status: 200 }));
      if (url.endsWith("/snapshots")) return Promise.resolve(new Response(JSON.stringify({ snapshots: [] }), { status: 200 }));
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderGrowthRoute("/growth-calculator/report?model_id=99");

    expect(await screen.findByText("指定报告模型")).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/growth/models/99", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).not.toHaveBeenCalledWith("/api/v1/growth/models", expect.anything());
  });

  it("exports the selected report through the authenticated API", async () => {
    const createObjectURL = vi.fn(() => "blob:growth-report");
    const revokeObjectURL = vi.fn();
    Object.defineProperty(URL, "createObjectURL", { configurable: true, value: createObjectURL });
    Object.defineProperty(URL, "revokeObjectURL", { configurable: true, value: revokeObjectURL });
    const click = vi.spyOn(HTMLAnchorElement.prototype, "click").mockImplementation(() => undefined);
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/growth/models/99") {
        return Promise.resolve(new Response(JSON.stringify({
          id: 99, user_id: 7, name: "待导出模型",
          assumptions: { monthly_visits: 12000, lead_rate: 0.08, deal_rate: 0.15, average_order: 6000, acquisition_cost: 80, delivery_cost: 120000 },
          result: { monthly_revenue: 864000, leads: 960, deals: 144, payback_days: 7, net_margin: 0.77 },
          created_at: "2026-08-15T08:00:00Z", updated_at: "2026-08-15T08:00:00Z"
        }), { status: 200 }));
      }
      if (url.endsWith("/scenarios")) return Promise.resolve(new Response(JSON.stringify({ scenarios: [] }), { status: 200 }));
      if (url.endsWith("/forecast")) return Promise.resolve(new Response(JSON.stringify({ months: [] }), { status: 200 }));
      if (url.endsWith("/recommendations")) return Promise.resolve(new Response(JSON.stringify({ action_items: [], cost_items: [] }), { status: 200 }));
      if (url.endsWith("/snapshots")) return Promise.resolve(new Response(JSON.stringify({ snapshots: [] }), { status: 200 }));
      if (url === "/api/v1/growth/models/99/export" && init?.method === "POST") {
        return Promise.resolve(new Response(JSON.stringify({ model: { id: 99 } }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderGrowthRoute("/growth-calculator/report?model_id=99");
    fireEvent.click(await screen.findByRole("button", { name: "导出报告" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/growth/models/99/export",
      expect.objectContaining({ method: "POST", body: JSON.stringify({ format: "json" }) })
    ));
    expect(createObjectURL).toHaveBeenCalledOnce();
    expect(click).toHaveBeenCalledOnce();
    expect(revokeObjectURL).toHaveBeenCalledWith("blob:growth-report");
    expect(await screen.findByText("报告已导出")).toBeInTheDocument();
  });

  it("loads the latest growth model from API", async () => {
    const generatedTasks = [
      {
        id: 801, user_id: 7, title: "优化高意向线索跟进", description: "将跟进延迟压缩到 24 小时内", assignee: "",
        project: "增长测算", status: "todo", priority: "high", tags: ["增长"], tools: [], learning: "", progress: 0, version: 1,
        created_at: "2026-06-30T08:30:00Z", updated_at: "2026-06-30T08:30:00Z"
      },
      {
        id: 802, user_id: 7, title: "提高成交转化率", description: "先优化成交再增加预算", assignee: "",
        project: "增长测算", status: "todo", priority: "medium", tags: ["转化"], tools: [], learning: "", progress: 0, version: 1,
        created_at: "2026-06-30T08:30:00Z", updated_at: "2026-06-30T08:30:00Z"
      }
    ];
    vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/growth/models") {
        return Promise.resolve(new Response(JSON.stringify({
          models: [
            {
              id: 9,
              user_id: 7,
              name: "商业沙盘标准模型",
              assumptions: {
                monthly_visits: 18000,
                lead_rate: 0.08,
                deal_rate: 0.16,
                average_order: 1200,
                acquisition_cost: 58,
                delivery_cost: 420
              },
              result: {
                monthly_revenue: 276480,
                leads: 1440,
                deals: 230,
                payback_days: 13,
                net_margin: 0.42
              },
              created_at: "2026-06-30T08:00:00Z",
              updated_at: "2026-06-30T08:30:00Z"
            }
          ]
        }), { status: 200 }));
      }
      if (url === "/api/v1/growth/models/9/scenarios") {
        return Promise.resolve(new Response(JSON.stringify({
          model_id: 9,
          model_name: "商业沙盘标准模型",
          scenarios: [
            { name: "保守方案", revenue: 180000, cost: 64000, margin: 0.31, highlight: "先控预算" },
            { name: "标准方案", revenue: 276480, cost: 84000, margin: 0.42, highlight: "当前推荐" },
            { name: "进攻方案", revenue: 412000, cost: 128000, margin: 0.38, highlight: "加速放量" }
          ],
          generated_at: "2026-06-30T08:30:00Z"
        }), { status: 200 }));
      }
      if (url === "/api/v1/growth/models/9/forecast") {
        return Promise.resolve(new Response(JSON.stringify({
          model_id: 9,
          model_name: "商业沙盘标准模型",
          months: [
            { month: "第1月", revenue: 160000, phase: "验证渠道", progress_percent: 30 },
            { month: "第2月", revenue: 220000, phase: "优化转化", progress_percent: 48 },
            { month: "第3月", revenue: 276480, phase: "稳定投放", progress_percent: 64 }
          ],
          generated_at: "2026-06-30T08:30:00Z"
        }), { status: 200 }));
      }
      if (url === "/api/v1/growth/models/9/recommendations") {
        return Promise.resolve(new Response(JSON.stringify({
          model_id: 9,
          model_name: "商业沙盘标准模型",
          headline: "先优化高意向成交",
          summary: "当前模型每月预计产生 1440 条线索、230 个成交。",
          cost_items: [
            { name: "内容生产", amount: 18000, detail: "案例页和行业内容" },
            { name: "投放预算", amount: 83520, detail: "搜索词测试" }
          ],
          action_items: ["把 CRM 跟进延迟压缩到 24 小时内", "先提高成交率再增加预算"],
          generated_at: "2026-06-30T08:30:00Z"
        }), { status: 200 }));
      }
      if (url === "/api/v1/growth/models/9/snapshots") {
        return Promise.resolve(new Response(JSON.stringify({ snapshots: [{
          id: 501,
          user_id: 7,
          model_id: 9,
          model_name: "商业沙盘标准模型 · 初版",
          assumptions: { monthly_visits: 10000, lead_rate: 0.05, deal_rate: 0.1, average_order: 2000, acquisition_cost: 50, delivery_cost: 40000 },
          result: { monthly_revenue: 100000, leads: 500, deals: 50, payback_days: 20, net_margin: 0.35 },
          scenarios: { model_id: 9, model_name: "商业沙盘标准模型 · 初版", scenarios: [], generated_at: "2026-06-20T08:00:00Z" },
          forecast: { model_id: 9, model_name: "商业沙盘标准模型 · 初版", months: [], generated_at: "2026-06-20T08:00:00Z" },
          recommendations: { model_id: 9, model_name: "商业沙盘标准模型 · 初版", headline: "先验证渠道", summary: "初版测算", cost_items: [], action_items: ["验证首个获客渠道"], generated_at: "2026-06-20T08:00:00Z" },
          created_at: "2026-06-20T08:00:00Z"
        }] }), { status: 200 }));
      }
      if (url === "/api/v1/growth/models/9/risks") {
        return Promise.resolve(new Response(JSON.stringify({
          model_id: 9,
          model_name: "商业沙盘标准模型",
          overall_level: "medium",
          risks: [{ key: "deal_rate", name: "成交转化率", level: "medium", current_value: "16.0%", threshold: "低于 15% 需关注", reason: "转化效率需要持续观察", suggestion: "继续优化跟进", }],
          generated_at: "2026-06-30T08:30:00Z"
        }), { status: 200 }));
      }
      if (url === "/api/v1/growth/models/9/action-plan") {
        return Promise.resolve(new Response(JSON.stringify({
          model_id: 9,
          model_name: "商业沙盘标准模型",
          phases: [{ key: "0-30", name: "0–30 天", goal: "验证渠道", items: [{ id: "validate-channel", title: "验证首个高意向获客渠道", detail: "验证", owner_role: "增长负责人", target_metric: "有效线索成本", expected_result: "确认入口" }] }],
          generated_at: "2026-06-30T08:30:00Z"
        }), { status: 200 }));
      }
      if (url === "/api/v1/growth/models/9/inputs") {
        return Promise.resolve(new Response(JSON.stringify({
          model_id: 9,
          model_name: "商业沙盘标准模型",
          completeness_percent: 100,
          fields: [{ key: "deal_rate", label: "成交转化率", value: 0.16, unit: "%", source: "测算参数（规则提取）", confidence: "待校准", confirmed_by_user: false }],
          generated_at: "2026-06-30T08:30:00Z"
        }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/generate" && init?.method === "POST") {
        return Promise.resolve(new Response(JSON.stringify({
          draft: {
            id: 701, user_id: 7, goal: "执行增长优化", tasks: generatedTasks, status: "draft",
            created_at: "2026-06-30T08:30:00Z", updated_at: "2026-06-30T08:30:00Z"
          },
          tasks: generatedTasks
        }), { status: 200 }));
      }
      if (url === "/api/v1/tasks/ai-drafts/701/adopt" && init?.method === "POST") {
        return Promise.resolve(new Response(JSON.stringify({ tasks: generatedTasks }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderGrowthRoute();

    expect(await screen.findByText("商业沙盘标准模型")).toBeInTheDocument();
    expect(screen.getAllByText("¥276,480").length).toBeGreaterThan(0);
    expect(screen.getByDisplayValue("18,000")).toBeInTheDocument();
    expect(screen.getByDisplayValue("8.0%")).toBeInTheDocument();
    expect(await screen.findByText("¥412,000")).toBeInTheDocument();
    expect(screen.getByText("案例页和行业内容")).toBeInTheDocument();
    expect(screen.getByText("先优化高意向成交")).toBeInTheDocument();
    expect(screen.getByText("把 CRM 跟进延迟压缩到 24 小时内")).toBeInTheDocument();
    expect(screen.getByText("整体风险：中风险")).toBeInTheDocument();
    expect(screen.getByText("验证首个高意向获客渠道")).toBeInTheDocument();
    expect(screen.getByText("信息完整度 100%")).toBeInTheDocument();
    expect(screen.getByText("测算参数（规则提取）")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "生成任务草稿" }));
    await waitFor(() => expect(fetch).toHaveBeenCalledWith(
      "/api/v1/tasks/generate",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({
          goal: "执行商业沙盘标准模型的增长优化：把 CRM 跟进延迟压缩到 24 小时内；先提高成交率再增加预算",
          source_type: "growth_model",
          source_id: 9,
          source_title: "商业沙盘标准模型",
          source_url: "/growth-calculator"
        })
      })
    ));
    expect(await screen.findByRole("dialog", { name: "增长任务草稿预览" })).toBeInTheDocument();
    expect(screen.getByText("已生成 2 条任务草稿，请确认后创建")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "确认创建任务" }));
    await waitFor(() => expect(fetch).toHaveBeenCalledWith(
      "/api/v1/tasks/ai-drafts/701/adopt",
      expect.objectContaining({
        method: "POST",
        headers: expect.objectContaining({ "Idempotency-Key": "growth-model-9-task-draft-701" })
      })
    ));
    expect(await screen.findByText("已创建 2 个增长任务")).toBeInTheDocument();

    fireEvent.change(await screen.findByRole("combobox", { name: "历史测算" }), { target: { value: "501" } });
    expect(await screen.findByText("商业沙盘标准模型 · 初版")).toBeInTheDocument();
    expect(screen.getAllByText("¥100,000").length).toBeGreaterThan(0);
    expect(screen.getByText("验证首个获客渠道")).toBeInTheDocument();
  });

  it("shows backend load errors without rendering fallback model values", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ error: "invalid_request" }), { status: 400 })
    );

    renderGrowthRoute();

    expect(await screen.findByText("请求参数有误，请检查后重试")).toBeInTheDocument();
    expect(screen.getByText("暂无测算模型")).toBeInTheDocument();
    expect(screen.queryByText("智能客服系统 · 标准方案")).not.toBeInTheDocument();
    expect(screen.queryByText("¥18.6万")).not.toBeInTheDocument();
  });

  it("clarifies a growth draft and calculates a persisted model", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/growth/models" && init?.method === "GET") {
        return Promise.resolve(new Response(JSON.stringify({ models: [] }), { status: 200 }));
      }
      if (url === "/api/v1/growth/drafts" && init?.method === "POST") {
        return Promise.resolve(new Response(JSON.stringify({
          id: 71,
          status: "needs_input",
          input: "企业培训服务，每月访问量 12000，线索转化率 8%，成交率 15%，客单价 6000 元",
          assumptions: { monthly_visits: 12000, lead_rate: 0.08, deal_rate: 0.15, average_order: 6000, acquisition_cost: 0, delivery_cost: 0 },
          questions: [
            { key: "acquisition_cost", label: "单条线索获客成本", unit: "元", min: 0 },
            { key: "delivery_cost", label: "每月交付成本", unit: "元", min: 0 }
          ],
          answers: {}
        }), { status: 200 }));
      }
      if (url === "/api/v1/growth/drafts/71/answers" && init?.method === "POST") {
        return Promise.resolve(new Response(JSON.stringify({ id: 71, status: "ready", questions: [], answers: {} }), { status: 200 }));
      }
      if (url === "/api/v1/growth/drafts/71/calculate" && init?.method === "POST") {
        return Promise.resolve(new Response(JSON.stringify({
          draft: { id: 71, status: "calculated", questions: [], answers: {} },
          model: {
            id: 12, user_id: 7, name: "企业培训服务增长测算",
            assumptions: { monthly_visits: 12000, lead_rate: 0.08, deal_rate: 0.15, average_order: 6000, acquisition_cost: 80, delivery_cost: 120000 },
            result: { monthly_revenue: 864000, leads: 960, deals: 144, payback_days: 7, net_margin: 0.77 },
            created_at: "2026-07-11T08:00:00Z", updated_at: "2026-07-11T08:00:00Z"
          }
        }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderGrowthRoute();
    fireEvent.change(screen.getByRole("textbox", { name: "业务与增长问题" }), {
      target: { value: "企业培训服务，每月访问量 12000，线索转化率 8%，成交率 15%，客单价 6000 元" }
    });
    fireEvent.click(screen.getByRole("button", { name: "开始测算" }));

    fireEvent.change(await screen.findByRole("spinbutton", { name: "单条线索获客成本" }), { target: { value: "80" } });
    fireEvent.change(screen.getByRole("spinbutton", { name: "每月交付成本" }), { target: { value: "120000" } });
    fireEvent.click(screen.getByRole("button", { name: "提交补充信息" }));

    fireEvent.click(await screen.findByRole("button", { name: "生成测算结果" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        "/api/v1/growth/drafts/71/answers",
        expect.objectContaining({
          method: "POST",
          body: JSON.stringify({
            answers: { acquisition_cost: 80, delivery_cost: 120000 }
          })
        })
      );
    });
    expect(await screen.findByText("企业培训服务增长测算")).toBeInTheDocument();
    expect(screen.getAllByText("¥864,000").length).toBeGreaterThan(0);
    expect(screen.getByText("7天")).toBeInTheDocument();
    expect(screen.getAllByText("模型测算").length).toBeGreaterThan(0);
  });
});
