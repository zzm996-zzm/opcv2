import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import App from "../App";
import { authSession } from "../lib/authSession";
import type { SandboxReport, SandboxRole, SandboxRun } from "../lib/sandboxApi";

const roles: SandboxRole[] = [
  ["customer", "目标客户", false], ["investor", "投资人", false], ["competitor", "竞争对手", false], ["channel", "渠道方", false],
  ["supply", "供应链运营", false], ["expert", "行业专家", false], ["skeptic", "悲观者", true], ["partner", "合伙人", false]
].map(([role_code, display_name, is_required]) => ({ role_code: String(role_code), display_name: String(display_name), description: `${display_name}职责`, analysis_dimensions: ["dimension"], default_selected: role_code !== "partner", is_required: Boolean(is_required), default_model_route: "role_default", prompt_version: "sandbox_role_v1" }));

const report: SandboxReport = {
  summary: "建议先用三家门店验证付费意愿。", feasibility: { score: 82, level: "可小范围验证", basis: "客户支持，但悲观者指出交付成本风险。" }, purchase_probability: { value_pct: 68, basis: "基于角色输出", is_model_generated: true },
  opportunity: [{ point: "门店提效", reason: "客户存在重复运营工作" }], risk: [{ point: "交付成本", severity: "high", reason: "早期定制较多", mitigation: "限制试点范围" }],
  advice: [{ action: "访谈十家门店", why: "验证付费和需求优先级", priority: 1, effort: "1 周" }], role_takeaways: [{ role: "customer", stance: "support", key_points: ["有明确痛点"], dimension_scores: [] }],
  dimension_summary: [], disagreements: [{ topic: "是否立即扩张", views: [{ role: "customer", point: "支持试点" }, { role: "skeptic", point: "反对扩张" }], decision_needed: "先验证交付成本" }], missing_roles: [], scenarios: {}, assumptions: ["渠道成本待验证"], is_model_generated: true
};

function fixture(overrides: Partial<SandboxRun> = {}): SandboxRun {
  return { id: 42, name: "企业 AI 运营平台", product: { name: "企业 AI 运营平台" }, context: { target_customer: "连锁门店", channel: "行业伙伴" }, questions: [{ key: "selling_point", field: "product.selling_point", type: "text", question: "核心卖点是什么？", required: true }], assumptions: [], roles: ["customer", "investor", "skeptic"], orchestration_mode: "isolated_sessions", completeness: 0.6, rounds: 0, status: "clarifying", revision: 1, created_at: "2026-08-13T08:00:00Z", updated_at: "2026-08-13T08:00:00Z", next_questions: [{ key: "selling_point", field: "product.selling_point", type: "text", question: "核心卖点是什么？", required: true }], done: false, ...overrides };
}
function response(body: unknown, status = 200, headers?: HeadersInit) { return Promise.resolve(new Response(status === 204 ? null : JSON.stringify(body), { status, headers: headers ?? { "Content-Type": "application/json" } })); }
function signIn() { authSession.set({ access_token: "access-token", access_token_expires_at: "2026-12-31T23:59:59Z", is_new_user: false, user: { id: 7, nickname: "张婧", phone: "", status: "active" } }); }

describe("SandboxPage V1.2 flow", () => {
  afterEach(() => { cleanup(); authSession.clear(); vi.restoreAllMocks(); });

  it("opens the description step directly and validates empty input", () => {
    signIn(); render(<MemoryRouter initialEntries={["/sandbox/new"]}><App /></MemoryRouter>);
    expect(screen.getByRole("heading", { name: "商业沙盘" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /开始多角色推演/ })).toHaveAttribute("href", "/sandbox/new");
    expect(screen.getByRole("link", { name: /查看推演思路/ })).toHaveAttribute("href", "/sandbox/new");
    expect(screen.getByRole("link", { name: /生成推演大纲/ })).toHaveAttribute("href", "/sandbox/new");
    expect(screen.getByRole("link", { name: /历史推演记录/ })).toHaveAttribute("href", "/sandbox/history");
    fireEvent.click(screen.getByRole("button", { name: /开始推演/ }));
    expect(screen.getByRole("alert")).toHaveTextContent("请先描述要推演的项目或情况");
  });

  it("uses canonical links for run-specific Copilot actions", async () => {
    signIn();
    const completed = fixture({ status: "done", done: true, completeness: 1, report });
    vi.spyOn(globalThis, "fetch").mockImplementation((input) => String(input) === "/api/v1/sandbox-runs/42" ? response(completed) : Promise.reject(new Error(`unexpected ${String(input)}`)));
    render(<MemoryRouter initialEntries={["/sandbox-runs/42/report"]}><App /></MemoryRouter>);
    await screen.findByRole("heading", { name: "企业 AI 运营平台" });
    const copilot = screen.getByRole("complementary", { name: "智活 Copilot" });
    expect(within(copilot).getByRole("link", { name: /查看完整建议动作/ })).toHaveAttribute("href", "/sandbox-runs/42/report");
    expect(within(copilot).getByRole("link", { name: /返回推演记录/ })).toHaveAttribute("href", "/sandbox/history");
  });

  it("creates a V1.2 run and enters bounded clarification", async () => {
    signIn(); const created = fixture(); const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => String(input) === "/api/v1/sandbox-runs" && init?.method === "POST" ? response(created, 201) : String(input) === "/api/v1/sandbox-runs/42" ? response(created) : Promise.reject(new Error(`unexpected ${String(input)}`)));
    render(<MemoryRouter initialEntries={["/sandbox"]}><App /></MemoryRouter>);
    fireEvent.change(screen.getByLabelText("描述你要推演的项目或情况"), { target: { value: "企业 AI 运营平台" } }); fireEvent.click(screen.getByRole("button", { name: /开始推演/ }));
    expect(await screen.findByRole("heading", { name: "核心卖点是什么？" })).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/sandbox-runs", expect.objectContaining({ method: "POST", body: JSON.stringify({ name: "企业 AI 运营平台", product: { name: "企业 AI 运营平台" }, context: { extra: "企业 AI 运营平台" } }) }));
  });

  it("shows AI smart completion without exposing internal field paths", async () => {
    signIn();
    const initial = fixture({ context: { target_customer: "连锁门店", channel: "行业伙伴", extra: "我想做一款帮助连锁门店自动运营的 AI 平台" } });
    vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => String(input) === "/api/v1/sandbox-runs/42" && init?.method === "GET" ? response(initial) : Promise.reject(new Error(`unexpected ${String(input)}`)));
    render(<MemoryRouter initialEntries={["/sandbox/new?run=42&step=2"]}><App /></MemoryRouter>);

    expect(await screen.findByRole("heading", { name: "推演发起配置" })).toBeInTheDocument();
    expect(screen.getByText("AI 动态提问中")).toBeInTheDocument();
    expect(screen.getByText("AI 已识别到的关键信息")).toBeInTheDocument();
    expect(screen.getByText("连锁门店")).toBeInTheDocument();
    expect(screen.queryByText("product.selling_point")).not.toBeInTheDocument();
  });

  it("answers clarification using revision and advances to all eight roles", async () => {
    signIn(); const initial = fixture(); const ready = fixture({ status: "ready", revision: 2, completeness: 0.8, done: true, next_questions: [] });
    vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => { const url = String(input); if (url === "/api/v1/sandbox-runs/42" && init?.method === "GET") return response(initial); if (url === "/api/v1/sandbox-runs/42/answer") return response(ready); if (url === "/api/v1/sandbox-runs/roles") return response({ roles }); return Promise.reject(new Error(`unexpected ${url}`)); });
    render(<MemoryRouter initialEntries={["/sandbox/new?run=42&step=2"]}><App /></MemoryRouter>);
    fireEvent.change(await screen.findByLabelText("核心卖点是什么？"), { target: { value: "自动化门店运营" } }); fireEvent.click(screen.getByRole("button", { name: /完成补充/ }));
    expect(await screen.findByRole("heading", { name: "选择推演角色" })).toBeInTheDocument();
    expect(roles.every((role) => screen.getByRole("button", { name: new RegExp(role.display_name) }))).toBe(true);
    expect(screen.getAllByRole("button", { pressed: true })).toHaveLength(3);
    expect(screen.getByRole("button", { name: /悲观者/ })).toHaveAttribute("aria-pressed", "true");
  });

  it("enforces at least three roles and keeps skeptic selected", async () => {
    signIn(); const current = fixture({ status: "ready", done: true, completeness: 1, next_questions: [] }); const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => { const url = String(input); if (url === "/api/v1/sandbox-runs/42") return response(current); if (url === "/api/v1/sandbox-runs/roles") return response({ roles }); if (url === "/api/v1/sandbox-runs/42/roles" && init?.method === "POST") return response({ ...current, revision: 2 }); return Promise.reject(new Error(`unexpected ${url}`)); });
    render(<MemoryRouter initialEntries={["/sandbox/new?run=42&step=3"]}><App /></MemoryRouter>);
    const skeptic = await screen.findByRole("button", { name: /悲观者/ }); fireEvent.click(skeptic); expect(skeptic).toHaveAttribute("aria-pressed", "true");
    fireEvent.click(screen.getByRole("button", { name: /确认角色/ })); await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/sandbox-runs/42/roles", expect.objectContaining({ body: JSON.stringify({ revision: 1, roles: ["customer", "investor", "skeptic"] }) })));
  });

  it("starts from the compatibility route without membership quota calls", async () => {
    signIn(); const ready = fixture({ status: "ready", done: true, completeness: 1 }); const running = fixture({ status: "running", done: true, revision: 2, run_roles: [] }); const stream = "id: 1\nevent: run_queued\ndata: {\"id\":1,\"run_id\":42,\"event\":\"run_queued\"}\n\n";
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => { const url = String(input); if (url === "/api/v1/sandbox-runs/42" && init?.method === "GET") return response(ready); if (url === "/api/v1/sandbox-runs/roles") return response({ roles }); if (url === "/api/v1/sandbox-runs/42/start") return response(running, 202); if (url === "/api/v1/sandbox-runs/42/stream") return Promise.resolve(new Response(stream, { status: 200 })); return Promise.reject(new Error(`unexpected ${url}`)); });
    render(<MemoryRouter initialEntries={["/sandbox/start?session=42"]}><App /></MemoryRouter>); fireEvent.click(await screen.findByRole("button", { name: /开始推演/ }));
    expect(await screen.findByText("角色执行快照尚未生成")).toBeInTheDocument(); expect(fetchMock.mock.calls.some(([input]) => String(input).includes("membership/usage"))).toBe(false);
  });

  it("renders V1.2 report disagreements and downloads authenticated PDF", async () => {
    signIn(); const completed = fixture({ status: "done", done: true, completeness: 1, report }); const createURL = vi.spyOn(URL, "createObjectURL").mockReturnValue("blob:pdf"); vi.spyOn(URL, "revokeObjectURL").mockImplementation(() => undefined); const click = vi.spyOn(HTMLAnchorElement.prototype, "click").mockImplementation(() => undefined);
    vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => { const url = String(input); if (url === "/api/v1/sandbox-runs/42" && init?.method === "GET") return response(completed); if (url.endsWith("/report/export")) return response({ id: 700, run_id: 42, format: "pdf", download_url: "/api/v1/sandbox-runs/42/exports/700/download", expires_at: "", created_at: "" }, 201); if (url.endsWith("/exports/700/download")) return Promise.resolve(new Response("%PDF", { status: 200 })); return Promise.reject(new Error(`unexpected ${url}`)); });
    render(<MemoryRouter initialEntries={["/sandbox-runs/42/report"]}><App /></MemoryRouter>); expect(await screen.findByText("是否立即扩张")).toBeInTheDocument(); fireEvent.click(screen.getByRole("button", { name: /导出 PDF/ })); await waitFor(() => expect(click).toHaveBeenCalledOnce()); expect(createURL).toHaveBeenCalledOnce();
  });

  it("saves selected report advice and opens server-validated growth input", async () => {
    signIn(); const completed = fixture({ status: "done", done: true, completeness: 1, report }); const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => { const url = String(input); if (url === "/api/v1/sandbox-runs/42" && init?.method === "GET") return response(completed); if (url.endsWith("/report/tasks")) return response({ tasks: [{ id: 9, title: "访谈十家门店", source_type: "sandbox_session", source_id: 42 }] }, 201); if (url.endsWith("/report/growth-handoff")) return response({ url: "/growth-calculator?sandbox_run=42&price_cents=3900" }); return Promise.reject(new Error(`unexpected ${url}`)); });
    render(<MemoryRouter initialEntries={["/sandbox-runs/42/report"]}><App /></MemoryRouter>);
    fireEvent.click(await screen.findByRole("button", { name: "保存为任务" })); await screen.findByText("已按选择创建任务，可在任务中心继续跟进。");
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/sandbox-runs/42/report/tasks", expect.objectContaining({ body: JSON.stringify({ advice_indexes: [0] }) }));
    fireEvent.click(screen.getByRole("button", { name: "去增长测算" })); await waitFor(() => expect(screen.getByText("增长测算")).toBeInTheDocument());
  });

  it("creates a separate run when starting another simulation", async () => {
    signIn(); const completed = fixture({ status: "done", done: true, completeness: 1, report }); const created = fixture({ id: 77, name: "企业 AI 运营平台 - 再次推演", status: "clarifying" }); vi.spyOn(window, "confirm").mockReturnValue(true); const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => { const url = String(input); if (url === "/api/v1/sandbox-runs/42" && init?.method === "GET") return response(completed); if (url === "/api/v1/sandbox-runs" && init?.method === "POST") return response(created, 201); if (url === "/api/v1/sandbox-runs/77" && init?.method === "GET") return response(created); return Promise.reject(new Error(`unexpected ${url}`)); });
    render(<MemoryRouter initialEntries={["/sandbox-runs/42/report"]}><App /></MemoryRouter>);
    fireEvent.click(await screen.findByRole("button", { name: "再次推演" }));
    expect(await screen.findByRole("heading", { name: "核心卖点是什么？" })).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/sandbox-runs", expect.objectContaining({ method: "POST", body: JSON.stringify({ name: "企业 AI 运营平台 - 再次推演", product: completed.product, context: completed.context }) }));
  });

  it("renders historical reports whose optional collections are null", async () => {
    signIn(); const nullable = { ...report, missing_roles: null, assumptions: null, disagreements: null, role_takeaways: [{ ...report.role_takeaways[0], key_points: null }] } as unknown as SandboxReport; const completed = fixture({ status: "done", done: true, report: nullable }); vi.spyOn(globalThis, "fetch").mockImplementation((input) => String(input) === "/api/v1/sandbox-runs/42" ? response(completed) : Promise.reject(new Error(`unexpected ${String(input)}`)));
    render(<MemoryRouter initialEntries={["/sandbox-runs/42/report"]}><App /></MemoryRouter>);
    expect(await screen.findByRole("heading", { name: "企业 AI 运营平台" })).toBeInTheDocument(); expect(screen.getByText("本次推演基于以下假设")).toBeInTheDocument();
  });

  it("loads server-filtered history and uses canonical report links", async () => {
    signIn(); const completed = fixture({ status: "done", done: true, report }); const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input) => String(input).startsWith("/api/v1/sandbox-runs?") ? response({ runs: [completed], page: 1, limit: 20 }) : Promise.reject(new Error(`unexpected ${String(input)}`)));
    render(<MemoryRouter initialEntries={["/sandbox/history"]}><App /></MemoryRouter>); expect(await screen.findByText("企业 AI 运营平台")).toBeInTheDocument(); expect(screen.getByRole("link", { name: "查看报告" })).toHaveAttribute("href", "/sandbox-runs/42/report"); fireEvent.change(screen.getByLabelText("搜索产品名称"), { target: { value: "门店" } }); await waitFor(() => expect(fetchMock.mock.calls.some(([input]) => String(input).includes("product=%E9%97%A8%E5%BA%97"))).toBe(true));
  });

  it("sends the current sandbox run identifier to Copilot", async () => {
    signIn();
    const completed = fixture({ status: "done", done: true, completeness: 1, report });
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/sandbox-runs/42" && init?.method === "GET") return response(completed);
      if (url === "/api/v1/copilot/threads" && init?.method === "POST") return response({ id: 99, user_id: 7, title: "商业沙盘：下一步", mode: "chat", model: "deepseek", created_at: "", updated_at: "" });
      if (url === "/api/v1/copilot/threads/99/messages" && init?.method === "POST") return response({
        user_message: { id: 1, user_id: 7, thread_id: 99, role: "user", content: "下一步怎么验证", status: "completed", model: "deepseek", created_at: "" },
        assistant_message: { id: 2, user_id: 7, thread_id: 99, role: "assistant", content: "先访谈目标门店。", status: "completed", model: "deepseek", created_at: "" }
      });
      return Promise.reject(new Error(`unexpected ${url}`));
    });

    render(<MemoryRouter initialEntries={["/sandbox-runs/42/report"]}><App /></MemoryRouter>);
    await screen.findByRole("heading", { name: "企业 AI 运营平台" });
    const copilot = screen.getByRole("complementary", { name: "智活 Copilot" });
    fireEvent.change(await within(copilot).findByLabelText("向沙盘 Copilot 提问"), { target: { value: "下一步怎么验证" } });
    await waitFor(() => expect(within(copilot).getByRole("button", { name: "发送" })).not.toBeDisabled());
    fireEvent.click(within(copilot).getByRole("button", { name: "发送" }));
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/copilot/threads/99/messages", expect.objectContaining({ method: "POST" })));
    const sendCall = fetchMock.mock.calls.find(([input]) => String(input) === "/api/v1/copilot/threads/99/messages");
    expect(JSON.parse(String(sendCall?.[1]?.body))).toEqual(expect.objectContaining({
      active_filters: { module: "sandbox", view: "report", run_id: "42" }
    }));
  });

  it("offers both recovery exits for a missing run", async () => {
    signIn(); vi.spyOn(globalThis, "fetch").mockImplementation(() => response({ error: "sandbox_run_not_found" }, 404)); render(<MemoryRouter initialEntries={["/sandbox-runs/999"]}><App /></MemoryRouter>);
    expect(await screen.findByRole("heading", { name: "当前页面缺少沙盘会话" })).toBeInTheDocument(); expect(screen.getByRole("link", { name: "返回首页" })).toHaveAttribute("href", "/sandbox"); expect(screen.getByRole("link", { name: "创建新推演" })).toHaveAttribute("href", "/sandbox/new");
  });
});
