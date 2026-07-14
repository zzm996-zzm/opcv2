import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import App from "../App";
import { authSession } from "../lib/authSession";
import { membershipApi } from "../lib/membershipApi";

vi.mock("../lib/membershipApi", async (importActual) => {
  const actual = await importActual<typeof import("../lib/membershipApi")>();
  return {
    ...actual,
    membershipApi: {
      ...actual.membershipApi,
      featureAccess: vi.fn()
    }
  };
});

describe("LeadDevelopmentPage", () => {
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

  function renderLeadRoute() {
    signIn();
    vi.mocked(membershipApi.featureAccess).mockResolvedValue({
      features: [{ key: "ai_lead_development", label: "AI线索开发", status: "available", allow_read_only: true, allow_workflow: true }]
    });
    render(
      <MemoryRouter initialEntries={["/leads"]}>
        <App />
      </MemoryRouter>
    );
  }

  function membershipUsageResponse(used = 8, limit = 30) {
    return new Response(JSON.stringify({
      usage: [{ key: "lead_tasks", label: "AI线索任务", used, limit, unit: "次/月" }]
    }), { status: 200 });
  }

  it("renders an empty lead workbench instead of static sample leads", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/leads/tasks?limit=20") {
        return Promise.resolve(new Response(JSON.stringify({ tasks: [] }), { status: 200 }));
      }
      if (url === "/api/v1/membership/usage") {
        return Promise.resolve(membershipUsageResponse());
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });
    renderLeadRoute();

    expect(await screen.findByRole("button", { name: "新建线索任务" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "AI线索开发" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "高意向线索" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "开发路径" })).toBeInTheDocument();
    expect(await screen.findByText("暂无高意向线索")).toBeInTheDocument();
    expect(screen.getByText("暂无CRM统计")).toBeInTheDocument();
    expect(screen.getByText("暂无待跟进客户")).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "星桥教育集团" })).not.toBeInTheDocument();
    expect(screen.queryByText("第一版正在实现")).not.toBeInTheDocument();
    expect(await screen.findByText("22/30")).toBeInTheDocument();
  });

  it("loads lead tasks from API", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/leads/tasks?limit=20") {
        return Promise.resolve(new Response(JSON.stringify({
          tasks: [
            {
              id: 99,
              user_id: 7,
              query: "华东 连锁教培 智能客服",
              status: "queued",
              idempotency_key: "lead-task-99",
              credit_cost: 1,
              created_at: "2026-06-25T12:00:00Z",
              updated_at: "2026-06-25T12:00:00Z"
            }
          ]
        }), { status: 200 }));
      }
      if (url === "/api/v1/membership/usage") {
        return Promise.resolve(membershipUsageResponse());
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });
    renderLeadRoute();

    expect(await screen.findByRole("heading", { name: "华东 连锁教培 智能客服" })).toBeInTheDocument();
    expect(screen.getAllByText("排队中").length).toBeGreaterThan(0);
  });

  it("loads completed lead task results from API", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/leads/tasks?limit=20") {
        return Promise.resolve(new Response(JSON.stringify({
          tasks: [
            {
              id: 99,
              user_id: 7,
              query: "成都 教培 私域转化",
              status: "succeeded",
              idempotency_key: "lead-task-99",
              credit_cost: 1,
              created_at: "2026-06-25T12:00:00Z",
              updated_at: "2026-06-25T12:05:00Z"
            }
          ]
        }), { status: 200 }));
      }
      if (url === "/api/v1/leads/tasks/99") {
        return Promise.resolve(new Response(JSON.stringify({
          task: {
            id: 99,
            user_id: 7,
            query: "成都 教培 私域转化",
            status: "succeeded",
            idempotency_key: "lead-task-99",
            credit_cost: 1,
            created_at: "2026-06-25T12:00:00Z",
            updated_at: "2026-06-25T12:05:00Z"
          },
          progress_percent: 100,
          message: "线索采集已完成，可以查看结果并导入 CRM。",
          results_count: 1
        }), { status: 200 }));
      }
      if (url === "/api/v1/leads/tasks/99/results?limit=20") {
        return Promise.resolve(new Response(JSON.stringify({
          results: [
            {
              id: 7,
              task_id: 99,
              name: "成都启明星教育",
              phone: "028-12345678",
              website: "https://example.com",
              evidence: [
                { type: "website", title: "官网出现暑期招生咨询入口", url: "https://example.com" },
                { type: "search", title: "公开页面提到私域转化", url: "https://example.com/case" }
              ],
              created_at: "2026-06-25T12:05:00Z"
            }
          ]
        }), { status: 200 }));
      }
      if (url === "/api/v1/membership/usage") {
        return Promise.resolve(membershipUsageResponse());
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderLeadRoute();

    expect(await screen.findByRole("heading", { name: "成都启明星教育" })).toBeInTheDocument();
    expect(screen.getByText("官网出现暑期招生咨询入口")).toBeInTheDocument();
    expect(screen.getByText("线索采集已完成，可以查看结果并导入 CRM。 已发现 1 条候选线索。")).toBeInTheDocument();
  });

  it("imports one lead result into CRM", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/leads/tasks?limit=20") {
        return Promise.resolve(new Response(JSON.stringify({
          tasks: [{
            id: 99,
            user_id: 7,
            query: "成都 教培 私域转化",
            status: "succeeded",
            idempotency_key: "lead-task-99",
            credit_cost: 1,
            created_at: "2026-06-25T12:00:00Z",
            updated_at: "2026-06-25T12:05:00Z"
          }]
        }), { status: 200 }));
      }
      if (url === "/api/v1/leads/tasks/99") {
        return Promise.resolve(new Response(JSON.stringify({
          task: { id: 99, user_id: 7, query: "成都 教培 私域转化", status: "succeeded" },
          progress_percent: 100,
          message: "线索采集已完成，可以查看结果并导入 CRM。",
          results_count: 1
        }), { status: 200 }));
      }
      if (url === "/api/v1/leads/tasks/99/results?limit=20") {
        return Promise.resolve(new Response(JSON.stringify({
          results: [{
            id: 7,
            task_id: 99,
            name: "成都启明星教育",
            phone: "028-12345678",
            email: "hello@example.com",
            website: "https://example.com",
            created_at: "2026-06-25T12:05:00Z"
          }]
        }), { status: 200 }));
      }
      if (url === "/api/v1/crm/customers/import-lead" && init?.method === "POST") {
        return Promise.resolve(new Response(JSON.stringify({ id: 100, name: "成都启明星教育" }), { status: 200 }));
      }
      if (url === "/api/v1/membership/usage") {
        return Promise.resolve(membershipUsageResponse());
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderLeadRoute();

    expect(await screen.findByRole("heading", { name: "成都启明星教育" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "加入CRM" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/crm/customers/import-lead", expect.objectContaining({
      method: "POST",
      body: JSON.stringify({
        lead_result_id: 7,
        name: "成都启明星教育",
        phone: "028-12345678",
        email: "hello@example.com",
        website: "https://example.com"
      })
    })));
    expect(await screen.findByText("成都启明星教育 已加入 CRM")).toBeInTheDocument();
  });

  it("batch imports selected lead results into CRM", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/leads/tasks?limit=20") {
        return Promise.resolve(new Response(JSON.stringify({
          tasks: [{
            id: 99,
            user_id: 7,
            query: "成都 教培 私域转化",
            status: "succeeded",
            idempotency_key: "lead-task-99",
            credit_cost: 1,
            created_at: "2026-06-25T12:00:00Z",
            updated_at: "2026-06-25T12:05:00Z"
          }]
        }), { status: 200 }));
      }
      if (url === "/api/v1/leads/tasks/99") {
        return Promise.resolve(new Response(JSON.stringify({
          task: { id: 99, user_id: 7, query: "成都 教培 私域转化", status: "succeeded" },
          progress_percent: 100,
          message: "线索采集已完成，可以查看结果并导入 CRM。",
          results_count: 2
        }), { status: 200 }));
      }
      if (url === "/api/v1/leads/tasks/99/results?limit=20") {
        return Promise.resolve(new Response(JSON.stringify({
          results: [
            { id: 7, task_id: 99, name: "成都启明星教育", phone: "028-12345678", created_at: "2026-06-25T12:05:00Z" },
            { id: 8, task_id: 99, name: "星桥教育集团", email: "hello@star.test", created_at: "2026-06-25T12:06:00Z" }
          ]
        }), { status: 200 }));
      }
      if (url === "/api/v1/crm/customers/import-lead" && init?.method === "POST") {
        return Promise.resolve(new Response(JSON.stringify({ id: 100 }), { status: 200 }));
      }
      if (url === "/api/v1/membership/usage") {
        return Promise.resolve(membershipUsageResponse());
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderLeadRoute();

    expect(await screen.findByRole("heading", { name: "成都启明星教育" })).toBeInTheDocument();
    fireEvent.click(screen.getByLabelText("选择线索 成都启明星教育"));
    fireEvent.click(screen.getByLabelText("选择线索 星桥教育集团"));
    fireEvent.click(screen.getByRole("button", { name: "批量加入CRM (2)" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/crm/customers/import-lead", expect.objectContaining({
      method: "POST",
      body: JSON.stringify({ lead_result_id: 8, name: "星桥教育集团", email: "hello@star.test" })
    })));
    expect(await screen.findByText("已批量加入 2 条线索到 CRM")).toBeInTheDocument();
  });

  it("creates lead task from target profile", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/leads/tasks?limit=20") {
        return Promise.resolve(new Response(JSON.stringify({ tasks: [] }), { status: 200 }));
      }
      if (url === "/api/v1/membership/usage") {
        return Promise.resolve(membershipUsageResponse(init?.method === "POST" ? 9 : 8));
      }
      if (url === "/api/v1/leads/tasks" && init?.method === "POST") {
        return Promise.resolve(new Response(JSON.stringify({
          id: 100,
          user_id: 7,
          query: "成都 教培 私域转化",
          status: "queued",
          idempotency_key: "lead-task-100",
          credit_cost: 1,
          created_at: "2026-06-25T12:10:00Z",
          updated_at: "2026-06-25T12:10:00Z"
        }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });
    renderLeadRoute();

    fireEvent.change(await screen.findByLabelText("描述目标客户画像"), {
      target: { value: "成都 教培 私域转化" }
    });
    fireEvent.click(screen.getByRole("button", { name: "生成线索池" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/leads/tasks",
      expect.objectContaining({ method: "POST" })
    ));
    expect(await screen.findByRole("heading", { name: "成都 教培 私域转化" })).toBeInTheDocument();
  });

  it("blocks lead task creation when monthly quota is depleted", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/leads/tasks?limit=20") {
        return Promise.resolve(new Response(JSON.stringify({ tasks: [] }), { status: 200 }));
      }
      if (url === "/api/v1/membership/usage") {
        return Promise.resolve(membershipUsageResponse(30, 30));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });
    renderLeadRoute();

    expect(await screen.findByText("0/30")).toBeInTheDocument();
    fireEvent.change(await screen.findByLabelText("描述目标客户画像"), {
      target: { value: "成都 教培 私域转化" }
    });

    const submitButton = screen.getByRole("button", { name: "生成线索池" });
    expect(submitButton).toBeDisabled();
    expect(screen.getByRole("link", { name: "升级套餐" })).toHaveAttribute("href", "/membership");
    expect(fetchMock).not.toHaveBeenCalledWith("/api/v1/leads/tasks", expect.any(Object));
  });

  it("shows locked state without loading lead workflows", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ tasks: [] }), { status: 200 })
    );
    signIn();
    vi.mocked(membershipApi.featureAccess).mockResolvedValue({
      features: [{
        key: "ai_lead_development",
        label: "AI线索开发",
        status: "locked",
        required_plan: "pro",
        upgrade_url: "/membership",
        contact_url: "/enterprise",
        allow_read_only: false,
        allow_workflow: false
      }]
    });

    render(
      <MemoryRouter initialEntries={["/leads"]}>
        <App />
      </MemoryRouter>
    );

    expect(await screen.findByText("当前不会读取业务数据，也不会创建任务、客户或分析请求。")).toBeInTheDocument();
    expect(screen.queryByLabelText("描述目标客户画像")).not.toBeInTheDocument();
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
