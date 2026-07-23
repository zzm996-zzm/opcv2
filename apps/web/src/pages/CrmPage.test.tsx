import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
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

describe("CrmPage", () => {
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
    vi.mocked(membershipApi.featureAccess).mockResolvedValue({
      features: [{ key: "crm", label: "CRM客户管理", status: "available", allow_read_only: true, allow_workflow: true }]
    });
  }

  function renderCrmRoute() {
    signIn();
    render(
      <MemoryRouter initialEntries={["/crm"]}>
        <App />
      </MemoryRouter>
    );
  }

  it("renders an empty CRM workbench instead of static sample customers", async () => {
    vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ customers: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        total: 0,
        new: 0,
        contacted: 0,
        qualified: 0,
        proposal: 0,
        won: 0,
        lost: 0,
        due_today: 0
      }), { status: 200 }));
    renderCrmRoute();

    expect(await screen.findByRole("button", { name: "新建客户" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "CRM客户管理" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "客户列表" })).toBeInTheDocument();
    expect(await screen.findByText("暂无CRM客户")).toBeInTheDocument();
    expect(screen.getByText("暂无客户详情")).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "星桥教育集团" })).not.toBeInTheDocument();
    expect(screen.queryByText("第一版正在实现")).not.toBeInTheDocument();
  });

  it("creates a manual CRM customer from the workbench", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((url, init) => {
      if (String(url) === "/api/v1/crm/customers" && init?.method === "POST") {
        return Promise.resolve(new Response(JSON.stringify({
          id: 100,
          user_id: 7,
          import_key: "manual:1783514400000000000",
          name: "成都启明星教育",
          phone: "028-12345678",
          email: "hello@example.com",
          website: "https://example.com",
          stage: "new",
          source: "manual",
          created_at: "2026-07-08T10:00:00Z",
          updated_at: "2026-07-08T10:00:00Z"
        }), { status: 200 }));
      }
      if (String(url).includes("/activities")) {
        return Promise.resolve(new Response(JSON.stringify({ activities: [] }), { status: 200 }));
      }
      if (String(url).includes("/pipeline-stats")) {
        return Promise.resolve(new Response(JSON.stringify({
          total: 1,
          new: 1,
          contacted: 0,
          qualified: 0,
          proposal: 0,
          won: 0,
          lost: 0,
          due_today: 0
        }), { status: 200 }));
      }
      return Promise.resolve(new Response(JSON.stringify({ customers: [] }), { status: 200 }));
    });
    renderCrmRoute();

    expect(await screen.findByText("暂无CRM客户")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "新建客户" }));
    fireEvent.change(screen.getByLabelText("客户名称"), {
      target: { value: "成都启明星教育" }
    });
    fireEvent.change(screen.getByLabelText("电话"), {
      target: { value: "028-12345678" }
    });
    fireEvent.change(screen.getByLabelText("邮箱"), {
      target: { value: "hello@example.com" }
    });
    fireEvent.change(screen.getByLabelText("网站"), {
      target: { value: "https://example.com" }
    });
    fireEvent.click(screen.getByRole("button", { name: "创建客户" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/crm/customers", expect.objectContaining({
      method: "POST",
      body: JSON.stringify({
        name: "成都启明星教育",
        phone: "028-12345678",
        email: "hello@example.com",
        website: "https://example.com"
      })
    })));
    expect(await screen.findByText("客户已创建")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "成都启明星教育" })).toBeInTheDocument();
    expect(screen.getAllByText("手工录入").length).toBeGreaterThan(0);
  });

  it("imports a lead result from the CRM workbench", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((url, init) => {
      if (String(url) === "/api/v1/crm/customers/import-lead" && init?.method === "POST") {
        return Promise.resolve(new Response(JSON.stringify({
          id: 101,
          user_id: 7,
          import_key: "lead_result:77",
          name: "成都启明星教育",
          phone: "028-12345678",
          stage: "new",
          source: "lead",
          created_at: "2026-07-08T10:00:00Z",
          updated_at: "2026-07-08T10:00:00Z"
        }), { status: 200 }));
      }
      if (String(url).includes("/activities")) {
        return Promise.resolve(new Response(JSON.stringify({ activities: [] }), { status: 200 }));
      }
      if (String(url).includes("/pipeline-stats")) {
        return Promise.resolve(new Response(JSON.stringify({
          total: 1,
          new: 1,
          contacted: 0,
          qualified: 0,
          proposal: 0,
          won: 0,
          lost: 0,
          due_today: 0
        }), { status: 200 }));
      }
      return Promise.resolve(new Response(JSON.stringify({ customers: [] }), { status: 200 }));
    });
    renderCrmRoute();

    expect(await screen.findByText("暂无CRM客户")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "导入客户" }));
    fireEvent.change(screen.getByLabelText("线索结果ID"), { target: { value: "77" } });
    fireEvent.change(screen.getByLabelText("导入客户名称"), { target: { value: "成都启明星教育" } });
    fireEvent.change(screen.getByLabelText("导入电话"), { target: { value: "028-12345678" } });
    fireEvent.click(screen.getByRole("button", { name: "导入线索" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/crm/customers/import-lead", expect.objectContaining({
      method: "POST",
      body: JSON.stringify({
        lead_result_id: 77,
        name: "成都启明星教育",
        phone: "028-12345678",
        email: "",
        website: ""
      })
    })));
    expect(await screen.findByText("线索客户已导入")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "成都启明星教育" })).toBeInTheDocument();
  });

  it("loads CRM customers and pipeline stats from API", async () => {
    vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({
          customers: [
            {
              id: 100,
              user_id: 7,
              import_key: "lead:99",
              name: "成都启明星教育",
              phone: "028-12345678",
              email: "hello@example.com",
              stage: "contacted",
              source: "lead",
              next_follow_up_at: "2026-06-25T14:00:00Z",
              created_at: "2026-06-24T12:00:00Z",
              updated_at: "2026-06-25T12:00:00Z"
            }
          ]
        }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        total: 3,
        new: 1,
        contacted: 1,
        qualified: 1,
        proposal: 0,
        won: 1,
        lost: 0,
        due_today: 2
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        activities: [{
          id: 1,
          user_id: 7,
          customer_id: 100,
          type: "customer_updated",
          note: "客户资料已更新",
          created_at: "2026-06-25T12:00:00Z"
        }]
      }), { status: 200 }));
    renderCrmRoute();

    expect(await screen.findByRole("heading", { name: "成都启明星教育" })).toBeInTheDocument();
    expect(screen.getAllByText(/需求确认/).length).toBeGreaterThan(0);
    const stats = screen.getByLabelText("CRM关键指标");
    expect(within(stats).getByText("客户总数")).toBeInTheDocument();
    expect(within(stats).getByText("3")).toBeInTheDocument();
    expect(await screen.findByText("客户资料已更新")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "查看" })).toHaveAttribute("href", "/crm?customer_id=100");
    expect(screen.getByRole("link", { name: "查看全部跟进记录 ›" })).toHaveAttribute("href", "/crm/follow-ups?customer_id=100");
    expect(screen.getByRole("link", { name: "拨打电话" })).toHaveAttribute("href", "tel:028-12345678");
    expect(screen.getByRole("link", { name: "发消息" })).toHaveAttribute("href", "mailto:hello@example.com?subject=%E8%B7%9F%E8%BF%9B%EF%BC%9A%E6%88%90%E9%83%BD%E5%90%AF%E6%98%8E%E6%98%9F%E6%95%99%E8%82%B2");
  });

  it("loads a CRM customer detail from query string", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({
        id: 100,
        user_id: 7,
        import_key: "enterprise_diagnosis_request:8",
        name: "30人销售团队需要AI获客陪跑",
        stage: "won",
        source: "enterprise",
        created_at: "2026-07-07T13:30:00Z",
        updated_at: "2026-07-07T13:30:00Z"
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        total: 1,
        new: 0,
        contacted: 0,
        qualified: 0,
        proposal: 0,
        won: 1,
        lost: 0,
        due_today: 0
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        activities: [{
          id: 1,
          user_id: 7,
          customer_id: 100,
          type: "follow_up_recorded",
          note: "企业交付客户复盘下一步",
          created_at: "2026-07-07T12:00:00Z"
        }]
      }), { status: 200 }));
    signIn();

    render(
      <MemoryRouter initialEntries={["/crm?customer_id=100"]}>
        <App />
      </MemoryRouter>
    );

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/crm/customers/100", expect.any(Object)));
    expect(await screen.findByRole("heading", { name: "30人销售团队需要AI获客陪跑" })).toBeInTheDocument();
    expect(screen.getAllByText("企业交付").length).toBeGreaterThan(0);
    expect(await screen.findByText("企业交付客户复盘下一步")).toBeInTheDocument();
  });

  it("filters CRM customers by enterprise delivery source", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ customers: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        total: 1,
        new: 0,
        contacted: 0,
        qualified: 0,
        proposal: 0,
        won: 1,
        lost: 0,
        due_today: 0
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        customers: [{
          id: 101,
          user_id: 7,
          import_key: "enterprise_diagnosis_request:8",
          name: "30人销售团队需要AI获客陪跑",
          stage: "won",
          source: "enterprise",
          created_at: "2026-07-07T13:30:00Z",
          updated_at: "2026-07-07T13:30:00Z"
        }]
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        total: 1,
        new: 0,
        contacted: 0,
        qualified: 0,
        proposal: 0,
        won: 1,
        lost: 0,
        due_today: 0
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ activities: [] }), { status: 200 }));
    renderCrmRoute();

    expect(await screen.findByText("暂无CRM客户")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "企业交付" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/crm/customers?source=enterprise&limit=20", expect.any(Object)));
    expect(await screen.findByRole("heading", { name: "30人销售团队需要AI获客陪跑" })).toBeInTheDocument();
    expect(screen.getAllByText("企业交付").length).toBeGreaterThan(0);
  });

  it("searches CRM customers from the backend API", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ customers: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        total: 1,
        new: 0,
        contacted: 0,
        qualified: 0,
        proposal: 0,
        won: 1,
        lost: 0,
        due_today: 0
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        customers: [{
          id: 102,
          user_id: 7,
          import_key: "enterprise_diagnosis_request:9",
          name: "企业AI陪跑复盘客户",
          stage: "won",
          source: "enterprise",
          created_at: "2026-07-08T09:30:00Z",
          updated_at: "2026-07-08T09:30:00Z"
        }]
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        total: 1,
        new: 0,
        contacted: 0,
        qualified: 0,
        proposal: 0,
        won: 1,
        lost: 0,
        due_today: 0
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ activities: [] }), { status: 200 }));
    renderCrmRoute();

    expect(await screen.findByText("暂无CRM客户")).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("搜索客户"), {
      target: { value: "陪跑" }
    });

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/crm/customers?q=%E9%99%AA%E8%B7%91&limit=20", expect.any(Object)));
    expect(await screen.findByRole("heading", { name: "企业AI陪跑复盘客户" })).toBeInTheDocument();
  });

  it("filters CRM customers by stage from the backend API", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ customers: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        total: 1,
        new: 0,
        contacted: 0,
        qualified: 0,
        proposal: 0,
        won: 1,
        lost: 0,
        due_today: 0
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        customers: [{
          id: 103,
          user_id: 7,
          import_key: "enterprise_diagnosis_request:10",
          name: "已成交企业交付客户",
          stage: "won",
          source: "enterprise",
          created_at: "2026-07-08T10:30:00Z",
          updated_at: "2026-07-08T10:30:00Z"
        }]
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        total: 1,
        new: 0,
        contacted: 0,
        qualified: 0,
        proposal: 0,
        won: 1,
        lost: 0,
        due_today: 0
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ activities: [] }), { status: 200 }));
    renderCrmRoute();

    expect(await screen.findByText("暂无CRM客户")).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("客户阶段筛选"), {
      target: { value: "won" }
    });

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/crm/customers?stage=won&limit=20", expect.any(Object)));
    expect(await screen.findByRole("heading", { name: "已成交企业交付客户" })).toBeInTheDocument();
    expect(screen.getAllByText("已成交").length).toBeGreaterThan(0);
  });

  it("records a follow-up from the selected CRM customer detail", async () => {
    const nextInputValue = "2026-06-26T10:00";
    const expectedNextAt = new Date(nextInputValue).toISOString();
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({
        customers: [{
          id: 100,
          user_id: 7,
          import_key: "lead_result:99",
          name: "成都启明星教育",
          phone: "028-12345678",
          stage: "contacted",
          source: "lead",
          created_at: "2026-06-24T12:00:00Z",
          updated_at: "2026-06-25T12:00:00Z"
        }]
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        total: 1,
        new: 0,
        contacted: 1,
        qualified: 0,
        proposal: 0,
        won: 0,
        lost: 0,
        due_today: 0
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ activities: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        id: 7,
        user_id: 7,
        customer_id: 100,
        note: "已发送企业AI运营方案",
        next_follow_up_at: expectedNextAt,
        created_at: "2026-06-25T12:00:00Z"
      }), { status: 200 }));
    renderCrmRoute();

    expect(await screen.findByRole("heading", { name: "成都启明星教育" })).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("跟进内容"), {
      target: { value: "已发送企业AI运营方案" }
    });
    fireEvent.change(screen.getByLabelText("下次跟进时间"), {
      target: { value: nextInputValue }
    });
    fireEvent.click(screen.getByRole("button", { name: "保存跟进" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/crm/customers/100/follow-ups", expect.objectContaining({
      method: "POST",
      body: JSON.stringify({
        note: "已发送企业AI运营方案",
        next_follow_up_at: expectedNextAt
      })
    })));
    expect(await screen.findByText("跟进已记录")).toBeInTheDocument();
  });

  it("generates follow-up copy from the selected CRM customer detail", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({
        customers: [{
          id: 100,
          user_id: 7,
          import_key: "lead_result:99",
          name: "成都启明星教育",
          phone: "028-12345678",
          stage: "contacted",
          source: "lead",
          created_at: "2026-06-24T12:00:00Z",
          updated_at: "2026-06-25T12:00:00Z"
        }]
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        total: 1,
        new: 0,
        contacted: 1,
        qualified: 0,
        proposal: 0,
        won: 0,
        lost: 0,
        due_today: 0
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ activities: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        subject: "邀约方案演示",
        body: "您好，我们想约您本周看一下企业AI运营方案。",
        channel: "wechat"
      }), { status: 200 }));
    renderCrmRoute();

    expect(await screen.findByRole("heading", { name: "成都启明星教育" })).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("跟进目标"), {
      target: { value: "邀约方案演示" }
    });
    fireEvent.click(screen.getByRole("button", { name: "生成话术" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/crm/customers/100/follow-up-copy", expect.objectContaining({
      method: "POST",
      body: JSON.stringify({ goal: "邀约方案演示" })
    })));
    expect(await screen.findByText("已生成wechat话术：邀约方案演示")).toBeInTheDocument();
    expect(screen.getByLabelText("跟进内容")).toHaveValue("您好，我们想约您本周看一下企业AI运营方案。");
  });

  it("updates the selected CRM customer stage from detail", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({
        customers: [{
          id: 100,
          user_id: 7,
          import_key: "lead_result:99",
          name: "成都启明星教育",
          stage: "contacted",
          source: "lead",
          created_at: "2026-06-24T12:00:00Z",
          updated_at: "2026-06-25T12:00:00Z"
        }]
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        total: 1,
        new: 0,
        contacted: 1,
        qualified: 0,
        proposal: 0,
        won: 0,
        lost: 0,
        due_today: 0
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ activities: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        id: 100,
        user_id: 7,
        import_key: "lead_result:99",
        name: "成都启明星教育",
        stage: "qualified",
        source: "lead",
        created_at: "2026-06-24T12:00:00Z",
        updated_at: "2026-06-25T13:00:00Z"
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        total: 1,
        new: 0,
        contacted: 0,
        qualified: 1,
        proposal: 0,
        won: 0,
        lost: 0,
        due_today: 0
      }), { status: 200 }));
    renderCrmRoute();

    expect(await screen.findByRole("heading", { name: "成都启明星教育" })).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("客户阶段"), {
      target: { value: "qualified" }
    });
    fireEvent.click(screen.getByRole("button", { name: "更新阶段" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/crm/customers/100/stage", expect.objectContaining({
      method: "POST",
      body: JSON.stringify({ stage: "qualified", note: "阶段更新为方案演示" })
    })));
    expect(await screen.findByText("阶段已更新")).toBeInTheDocument();
    expect(screen.getAllByText("方案演示").length).toBeGreaterThan(0);
  });

  it("updates the selected CRM customer profile from detail", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({
        customers: [{
          id: 100,
          user_id: 7,
          import_key: "lead_result:99",
          name: "成都启明星教育",
          stage: "contacted",
          source: "lead",
          created_at: "2026-06-24T12:00:00Z",
          updated_at: "2026-06-25T12:00:00Z"
        }]
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        total: 1,
        new: 0,
        contacted: 1,
        qualified: 0,
        proposal: 0,
        won: 0,
        lost: 0,
        due_today: 0
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ activities: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        id: 100,
        user_id: 7,
        import_key: "lead_result:99",
        name: "成都启明星教育",
        phone: "028-12345678",
        email: "hello@example.com",
        website: "https://example.com",
        stage: "contacted",
        source: "lead",
        created_at: "2026-06-24T12:00:00Z",
        updated_at: "2026-06-25T13:00:00Z"
      }), { status: 200 }));
    renderCrmRoute();

    expect(await screen.findByRole("heading", { name: "成都启明星教育" })).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("电话"), {
      target: { value: "028-12345678" }
    });
    fireEvent.change(screen.getByLabelText("邮箱"), {
      target: { value: "hello@example.com" }
    });
    fireEvent.change(screen.getByLabelText("网站"), {
      target: { value: "https://example.com" }
    });
    fireEvent.click(screen.getByRole("button", { name: "保存资料" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/crm/customers/100", expect.objectContaining({
      method: "PATCH",
      body: JSON.stringify({
        name: "成都启明星教育",
        phone: "028-12345678",
        email: "hello@example.com",
        website: "https://example.com"
      })
    })));
    expect(await screen.findByText("客户资料已更新")).toBeInTheDocument();
    expect(screen.getByText(/028-12345678 hello@example.com/)).toBeInTheDocument();
  });

  it("renders an empty follow-up list instead of static sample records", async () => {
    signIn();
    vi.mocked(membershipApi.featureAccess).mockResolvedValue({
      features: [{ key: "crm", label: "CRM客户管理", status: "available", allow_read_only: true, allow_workflow: true }]
    });
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ follow_ups: [] }), { status: 200 })
    );
    render(
      <MemoryRouter initialEntries={["/crm/follow-ups"]}>
        <App />
      </MemoryRouter>
    );

    expect(await screen.findByRole("heading", { name: "全部跟进" })).toBeInTheDocument();
    expect(screen.getByRole("table", { name: "全部跟进列表" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "跟进提醒" })).toBeInTheDocument();
    expect(await screen.findByText("暂无跟进记录")).toBeInTheDocument();
    expect(screen.getByText("暂无跟进提醒")).toBeInTheDocument();
    expect(screen.getByText("暂无最近更新")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("张女士 · 成都蓝鲸教育")).not.toBeInTheDocument());
  });

  it("loads follow-up records from API", async () => {
    signIn();
    vi.mocked(membershipApi.featureAccess).mockResolvedValue({
      features: [{ key: "crm", label: "CRM客户管理", status: "available", allow_read_only: true, allow_workflow: true }]
    });
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        follow_ups: [{
          id: 1,
          user_id: 7,
          customer_id: 100,
          note: "已发送企业AI运营方案，等待客户确认演示时间",
          next_follow_up_at: "2026-06-25T14:00:00Z",
          created_at: "2026-06-24T12:00:00Z"
        }]
      }), { status: 200 })
    );

    render(
      <MemoryRouter initialEntries={["/crm/follow-ups"]}>
        <App />
      </MemoryRouter>
    );

    const table = await screen.findByRole("table", { name: "全部跟进列表" });
    expect(await within(table).findByText("客户 #100")).toBeInTheDocument();
    expect(within(table).getByText("已发送企业AI运营方案，等待客户确认演示时间")).toBeInTheDocument();
    expect(within(table).getByRole("link", { name: "查看详情" })).toHaveAttribute("href", "/crm?customer_id=100");
  });

  it("searches follow-up records through the backend API", async () => {
    signIn();
    vi.mocked(membershipApi.featureAccess).mockResolvedValue({
      features: [{ key: "crm", label: "CRM客户管理", status: "available", allow_read_only: true, allow_workflow: true }]
    });
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((url) => {
      if (String(url).includes("/api/v1/crm/customers?")) {
        return Promise.resolve(new Response(JSON.stringify({ customers: [] }), { status: 200 }));
      }
      if (String(url).includes("/api/v1/crm/pipeline-stats")) {
        return Promise.resolve(new Response(JSON.stringify({
          total: 0,
          new: 0,
          contacted: 0,
          qualified: 0,
          proposal: 0,
          won: 0,
          lost: 0,
          due_today: 0
        }), { status: 200 }));
      }
      if (String(url).includes("q=%E6%BC%94%E7%A4%BA")) {
        return Promise.resolve(new Response(JSON.stringify({
          follow_ups: [{
            id: 2,
            user_id: 7,
            customer_id: 101,
            note: "预约下周演示",
            next_follow_up_at: "2026-06-26T10:00:00Z",
            created_at: "2026-06-24T13:00:00Z"
          }]
        }), { status: 200 }));
      }
      return Promise.resolve(new Response(JSON.stringify({
        follow_ups: [{
          id: 1,
          user_id: 7,
          customer_id: 100,
          note: "已发送企业AI运营方案",
          next_follow_up_at: "2026-06-25T14:00:00Z",
          created_at: "2026-06-24T12:00:00Z"
        }]
      }), { status: 200 }));
    });

    render(
      <MemoryRouter initialEntries={["/crm/follow-ups"]}>
        <App />
      </MemoryRouter>
    );

    expect((await screen.findAllByText("已发送企业AI运营方案")).length).toBeGreaterThan(0);
    fireEvent.change(screen.getByLabelText("搜索跟进记录"), {
      target: { value: "演示" }
    });

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/crm/follow-ups?q=%E6%BC%94%E7%A4%BA&limit=100", expect.any(Object)));
    expect((await screen.findAllByText("预约下周演示")).length).toBeGreaterThan(0);
    expect(screen.queryByText("已发送企业AI运营方案")).not.toBeInTheDocument();
  });

  it("filters follow-up records by due window through the backend API", async () => {
    signIn();
    vi.mocked(membershipApi.featureAccess).mockResolvedValue({
      features: [{ key: "crm", label: "CRM客户管理", status: "available", allow_read_only: true, allow_workflow: true }]
    });
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((url) => {
      if (String(url).includes("/api/v1/crm/customers?")) {
        return Promise.resolve(new Response(JSON.stringify({ customers: [] }), { status: 200 }));
      }
      if (String(url).includes("/api/v1/crm/pipeline-stats")) {
        return Promise.resolve(new Response(JSON.stringify({
          total: 0,
          new: 0,
          contacted: 0,
          qualified: 0,
          proposal: 0,
          won: 0,
          lost: 0,
          due_today: 0
        }), { status: 200 }));
      }
      if (String(url).includes("due=week")) {
        return Promise.resolve(new Response(JSON.stringify({
          follow_ups: [{
            id: 3,
            user_id: 7,
            customer_id: 102,
            note: "本周安排方案复盘",
            next_follow_up_at: "2026-06-27T10:00:00Z",
            created_at: "2026-06-24T13:00:00Z"
          }]
        }), { status: 200 }));
      }
      return Promise.resolve(new Response(JSON.stringify({ follow_ups: [] }), { status: 200 }));
    });

    render(
      <MemoryRouter initialEntries={["/crm/follow-ups"]}>
        <App />
      </MemoryRouter>
    );

    fireEvent.click(await screen.findByRole("button", { name: "本周待跟进" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/crm/follow-ups?due=week&limit=100", expect.any(Object)));
    expect((await screen.findAllByText("本周安排方案复盘")).length).toBeGreaterThan(0);
  });

  it("reschedules a follow-up record from the follow-up list", async () => {
    signIn();
    vi.mocked(membershipApi.featureAccess).mockResolvedValue({
      features: [{ key: "crm", label: "CRM客户管理", status: "available", allow_read_only: true, allow_workflow: true }]
    });
    const nextInputValue = "2026-06-27T15:00";
    const expectedNextAt = new Date(nextInputValue).toISOString();
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((url, init) => {
      if (String(url).includes("/api/v1/crm/customers?")) {
        return Promise.resolve(new Response(JSON.stringify({ customers: [] }), { status: 200 }));
      }
      if (String(url).includes("/api/v1/crm/pipeline-stats")) {
        return Promise.resolve(new Response(JSON.stringify({
          total: 0,
          new: 0,
          contacted: 0,
          qualified: 0,
          proposal: 0,
          won: 0,
          lost: 0,
          due_today: 0
        }), { status: 200 }));
      }
      if (String(url) === "/api/v1/crm/follow-ups/1" && init?.method === "PATCH") {
        return Promise.resolve(new Response(JSON.stringify({
          id: 1,
          user_id: 7,
          customer_id: 100,
          note: "已发送企业AI运营方案",
          next_follow_up_at: expectedNextAt,
          created_at: "2026-06-24T12:00:00Z"
        }), { status: 200 }));
      }
      return Promise.resolve(new Response(JSON.stringify({
        follow_ups: [{
          id: 1,
          user_id: 7,
          customer_id: 100,
          note: "已发送企业AI运营方案",
          next_follow_up_at: "2026-06-25T14:00:00Z",
          created_at: "2026-06-24T12:00:00Z"
        }]
      }), { status: 200 }));
    });

    render(
      <MemoryRouter initialEntries={["/crm/follow-ups"]}>
        <App />
      </MemoryRouter>
    );

    expect((await screen.findAllByText("客户 #100")).length).toBeGreaterThan(0);
    fireEvent.click(screen.getByRole("button", { name: "改期" }));
    fireEvent.change(screen.getByLabelText("改期时间 #1"), {
      target: { value: nextInputValue }
    });
    fireEvent.click(screen.getByRole("button", { name: "保存改期" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/crm/follow-ups/1", expect.objectContaining({
      method: "PATCH",
      body: JSON.stringify({ next_follow_up_at: expectedNextAt })
    })));
    expect(await screen.findByText("跟进时间已改期")).toBeInTheDocument();
  });

  it("records a new follow-up from the follow-up list", async () => {
    signIn();
    vi.mocked(membershipApi.featureAccess).mockResolvedValue({
      features: [{ key: "crm", label: "CRM客户管理", status: "available", allow_read_only: true, allow_workflow: true }]
    });
    const nextInputValue = "2026-06-28T10:00";
    const expectedNextAt = new Date(nextInputValue).toISOString();
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((url, init) => {
      if (String(url).includes("/api/v1/crm/customers?")) {
        return Promise.resolve(new Response(JSON.stringify({ customers: [] }), { status: 200 }));
      }
      if (String(url).includes("/api/v1/crm/pipeline-stats")) {
        return Promise.resolve(new Response(JSON.stringify({
          total: 0,
          new: 0,
          contacted: 0,
          qualified: 0,
          proposal: 0,
          won: 0,
          lost: 0,
          due_today: 0
        }), { status: 200 }));
      }
      if (String(url) === "/api/v1/crm/customers/100/follow-ups" && init?.method === "POST") {
        return Promise.resolve(new Response(JSON.stringify({
          id: 2,
          user_id: 7,
          customer_id: 100,
          note: "二次电话确认预算",
          next_follow_up_at: expectedNextAt,
          created_at: "2026-06-25T12:00:00Z"
        }), { status: 200 }));
      }
      return Promise.resolve(new Response(JSON.stringify({
        follow_ups: [{
          id: 1,
          user_id: 7,
          customer_id: 100,
          note: "已发送企业AI运营方案",
          next_follow_up_at: "2026-06-25T14:00:00Z",
          created_at: "2026-06-24T12:00:00Z"
        }]
      }), { status: 200 }));
    });

    render(
      <MemoryRouter initialEntries={["/crm/follow-ups"]}>
        <App />
      </MemoryRouter>
    );

    expect((await screen.findAllByText("已发送企业AI运营方案")).length).toBeGreaterThan(0);
    fireEvent.click(screen.getByRole("button", { name: "记录跟进" }));
    fireEvent.change(screen.getByLabelText("跟进内容 #1"), { target: { value: "二次电话确认预算" } });
    fireEvent.change(screen.getByLabelText("下次跟进时间 #1"), { target: { value: nextInputValue } });
    fireEvent.click(screen.getByRole("button", { name: "保存跟进" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/crm/customers/100/follow-ups", expect.objectContaining({
      method: "POST",
      body: JSON.stringify({
        note: "二次电话确认预算",
        next_follow_up_at: expectedNextAt
      })
    })));
    expect(await screen.findByText("跟进已记录")).toBeInTheDocument();
    expect(screen.getAllByText("二次电话确认预算").length).toBeGreaterThan(0);
  });

  it("filters, paginates, and exports follow-up records", async () => {
    signIn();
    vi.mocked(membershipApi.featureAccess).mockResolvedValue({
      features: [{ key: "crm", label: "CRM客户管理", status: "available", allow_read_only: true, allow_workflow: true }]
    });
    const createObjectURL = vi.fn(() => "blob:crm-followups");
    const revokeObjectURL = vi.fn();
    vi.spyOn(HTMLAnchorElement.prototype, "click").mockImplementation(() => {});
    Object.defineProperty(URL, "createObjectURL", { configurable: true, value: createObjectURL });
    Object.defineProperty(URL, "revokeObjectURL", { configurable: true, value: revokeObjectURL });
    const rows = Array.from({ length: 11 }, (_, index) => ({
      id: index + 1,
      user_id: 7,
      customer_id: 100 + index,
      note: `跟进记录 ${index + 1}`,
      next_follow_up_at: index === 0 ? "2026-06-25T14:00:00Z" : "2099-06-25T14:00:00Z",
      created_at: "2026-06-24T12:00:00Z"
    }));
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((url) => {
      if (String(url).includes("/api/v1/crm/customers?")) {
        return Promise.resolve(new Response(JSON.stringify({ customers: [] }), { status: 200 }));
      }
      if (String(url).includes("/api/v1/crm/pipeline-stats")) {
        return Promise.resolve(new Response(JSON.stringify({
          total: 0,
          new: 0,
          contacted: 0,
          qualified: 0,
          proposal: 0,
          won: 0,
          lost: 0,
          due_today: 0
        }), { status: 200 }));
      }
      return Promise.resolve(new Response(JSON.stringify({ follow_ups: rows }), { status: 200 }));
    });

    render(
      <MemoryRouter initialEntries={["/crm/follow-ups"]}>
        <App />
      </MemoryRouter>
    );

    const table = await screen.findByRole("table", { name: "全部跟进列表" });
    expect((await within(table).findAllByText("跟进记录 1")).length).toBeGreaterThan(0);
    fireEvent.change(screen.getByLabelText("跟进状态筛选"), { target: { value: "pending" } });
    await waitFor(() => expect(within(table).queryByText(/^跟进记录 1$/)).not.toBeInTheDocument());
    expect(within(table).getAllByText("跟进记录 2").length).toBeGreaterThan(0);
    fireEvent.change(screen.getByLabelText("跟进状态筛选"), { target: { value: "all" } });
    fireEvent.click(screen.getByRole("button", { name: "下一页" }));
    expect(await screen.findByText("跟进记录 11")).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("跟进时间范围"), { target: { value: "week" } });
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/crm/follow-ups?due=week&limit=100", expect.any(Object)));
    fireEvent.click(screen.getByRole("button", { name: "导出记录" }));
    expect(createObjectURL).toHaveBeenCalled();
    expect(await screen.findByText("已导出 11 条跟进记录")).toBeInTheDocument();
  });

  it("loads follow-up records for a specific customer from query string", async () => {
    signIn();
    vi.mocked(membershipApi.featureAccess).mockResolvedValue({
      features: [{ key: "crm", label: "CRM客户管理", status: "available", allow_read_only: true, allow_workflow: true }]
    });
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((url) => {
      if (String(url).includes("/api/v1/crm/customers?")) {
        return Promise.resolve(new Response(JSON.stringify({ customers: [] }), { status: 200 }));
      }
      if (String(url).includes("/api/v1/crm/pipeline-stats")) {
        return Promise.resolve(new Response(JSON.stringify({
          total: 0,
          new: 0,
          contacted: 0,
          qualified: 0,
          proposal: 0,
          won: 0,
          lost: 0,
          due_today: 0
        }), { status: 200 }));
      }
      return Promise.resolve(new Response(JSON.stringify({
        follow_ups: [{
          id: 2,
          user_id: 7,
          customer_id: 100,
          note: "企业交付客户复盘下一步",
          next_follow_up_at: "2026-07-08T10:00:00Z",
          created_at: "2026-07-07T12:00:00Z"
        }]
      }), { status: 200 }));
    });

    render(
      <MemoryRouter initialEntries={["/crm/follow-ups?customer_id=100"]}>
        <App />
      </MemoryRouter>
    );

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/crm/follow-ups?customer_id=100&limit=100", expect.any(Object)));
    await waitFor(() => expect(screen.getAllByText("企业交付客户复盘下一步").length).toBeGreaterThan(0));
  });

  it("shows the design preview for locked CRM overview without loading CRM APIs", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ customers: [] }), { status: 200 })
    );
    signIn();
    vi.mocked(membershipApi.featureAccess).mockResolvedValue({
      features: [{
        key: "crm",
        label: "CRM客户管理",
        status: "locked",
        required_plan: "pro",
        upgrade_url: "/membership",
        contact_url: "/enterprise",
        allow_read_only: true,
        allow_workflow: false
      }]
    });

    render(
      <MemoryRouter initialEntries={["/crm"]}>
        <App />
      </MemoryRouter>
    );

    expect(await screen.findByLabelText("CRM客户管理总览")).toBeInTheDocument();
    expect(screen.getByLabelText("CRM关键指标")).toBeInTheDocument();
    expect(screen.getAllByText("1,286").length).toBeGreaterThan(0);
    expect(screen.getByText("642")).toBeInTheDocument();
    expect(screen.getByText("328")).toBeInTheDocument();
    expect(screen.getByText("87")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "AI推荐跟进" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "销售阶段看板" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "客户来源分布" })).toBeInTheDocument();
    expect(screen.getByLabelText("CRM Copilot")).toBeInTheDocument();
    expect(screen.getAllByText("杭州智创科技有限公司").length).toBeGreaterThan(0);
    expect(screen.getAllByText("上海云联信息技术有限公司").length).toBeGreaterThan(0);
    expect(screen.getByText("42%")).toBeInTheDocument();
    expect(screen.getByText("28%")).toBeInTheDocument();
    expect(screen.queryByText("当前为示例预览，升级后接入真实客户数据")).not.toBeInTheDocument();
    expect(screen.queryByLabelText("CRM业务操作区")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "新建客户" })).not.toBeInTheDocument();
    expect(screen.getAllByRole("link", { name: /查看全部|生成今日跟进清单|识别高意向客户/ }).every((link) => (
      link.getAttribute("href") === "/membership/upgrade"
    ))).toBe(true);
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
