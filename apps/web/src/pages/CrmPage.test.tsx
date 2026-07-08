import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import App from "../App";
import { authSession } from "../lib/authSession";

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

    expect(screen.getByRole("heading", { name: "CRM客户管理" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "新建客户" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "客户列表" })).toBeInTheDocument();
    expect(await screen.findByText("暂无CRM客户")).toBeInTheDocument();
    expect(screen.getByText("暂无客户详情")).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "星桥教育集团" })).not.toBeInTheDocument();
    expect(screen.queryByText("第一版正在实现")).not.toBeInTheDocument();
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
    expect(within(stats).getByText("总客户")).toBeInTheDocument();
    expect(within(stats).getByText("3")).toBeInTheDocument();
    expect(await screen.findByText("客户资料已更新")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "查看全部跟进记录 ›" })).toHaveAttribute("href", "/crm/follow-ups?customer_id=100");
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

  it("renders an empty follow-up list instead of static sample records", async () => {
    signIn();
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ follow_ups: [] }), { status: 200 })
    );
    render(
      <MemoryRouter initialEntries={["/crm/follow-ups"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "全部跟进" })).toBeInTheDocument();
    expect(screen.getByRole("table", { name: "全部跟进列表" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "跟进提醒" })).toBeInTheDocument();
    expect(await screen.findByText("暂无跟进记录")).toBeInTheDocument();
    expect(screen.getByText("暂无跟进提醒")).toBeInTheDocument();
    expect(screen.getByText("暂无最近更新")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("张女士 · 成都蓝鲸教育")).not.toBeInTheDocument());
  });

  it("loads follow-up records from API", async () => {
    signIn();
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

    const table = screen.getByRole("table", { name: "全部跟进列表" });
    expect(await within(table).findByText("客户 #100")).toBeInTheDocument();
    expect(within(table).getByText("已发送企业AI运营方案，等待客户确认演示时间")).toBeInTheDocument();
    expect(within(table).getByRole("link", { name: "查看详情" })).toHaveAttribute("href", "/crm?customer_id=100");
  });

  it("loads follow-up records for a specific customer from query string", async () => {
    signIn();
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        follow_ups: [{
          id: 2,
          user_id: 7,
          customer_id: 100,
          note: "企业交付客户复盘下一步",
          next_follow_up_at: "2026-07-08T10:00:00Z",
          created_at: "2026-07-07T12:00:00Z"
        }]
      }), { status: 200 })
    );

    render(
      <MemoryRouter initialEntries={["/crm/follow-ups?customer_id=100"]}>
        <App />
      </MemoryRouter>
    );

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/crm/follow-ups?customer_id=100&limit=20", expect.any(Object)));
    await waitFor(() => expect(screen.getAllByText("企业交付客户复盘下一步").length).toBeGreaterThan(0));
  });
});
