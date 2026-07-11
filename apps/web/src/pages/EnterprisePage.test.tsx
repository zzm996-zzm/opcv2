import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import App from "../App";
import { authSession } from "../lib/authSession";

describe("EnterprisePage", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  it("renders backend-connected enterprise empty states", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({
        stats: [],
        plans: [],
        delivery_board: [],
        milestones: [],
        cases: []
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ requests: [] }), { status: 200 }));
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
      <MemoryRouter initialEntries={["/enterprise"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "企业定制化陪跑" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "预约企业诊断" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "陪跑方案" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "交付看板" })).toBeInTheDocument();
    expect(screen.getByText("暂无企业陪跑概览数据，提交一次诊断需求后将逐步沉淀方案、交付和案例数据。")).toBeInTheDocument();
    expect(screen.queryByText("企业陪跑后端接口未接入")).not.toBeInTheDocument();
    expect(screen.getByText("暂无陪跑方案")).toBeInTheDocument();
    expect(screen.getByText("暂无交付看板数据")).toBeInTheDocument();
    expect(screen.getByText("暂无陪跑里程碑")).toBeInTheDocument();
    expect(screen.getByText("暂无企业案例")).toBeInTheDocument();
    expect(await screen.findByText("暂无企业诊断预约")).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "增长团队训练营" })).not.toBeInTheDocument();
    expect(screen.queryByText("连锁教育集团")).not.toBeInTheDocument();
    expect(screen.queryByText("服务企业")).not.toBeInTheDocument();
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/enterprise/overview", expect.any(Object)));
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/enterprise/diagnosis-requests?limit=5", expect.any(Object));
    expect(screen.queryByText("第一版正在实现")).not.toBeInTheDocument();
  });

  it("focuses the enterprise need input from the primary action", async () => {
    vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({
        stats: [],
        plans: [],
        delivery_board: [],
        milestones: [],
        cases: []
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ requests: [] }), { status: 200 }));
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-23T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "", account: "zhangjing", status: "active" }
    });

    render(
      <MemoryRouter initialEntries={["/enterprise"]}>
        <App />
      </MemoryRouter>
    );

    fireEvent.click(screen.getByRole("button", { name: "预约企业诊断" }));

    expect(screen.getByLabelText("描述企业需求")).toHaveFocus();
  });

  it("renders enterprise overview records from the backend API", async () => {
    vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({
        stats: [{ key: "companies", label: "服务企业数", value: "3" }],
        plans: [{
          id: 1,
          title: "后端陪跑方案",
          audience: "后端返回的适用对象",
          price_label: "待报价",
          focus: ["后端重点"],
          result: "后端返回的交付结果"
        }],
        delivery_board: [
          { stage: "诊断中", count: 1, detail: "后端交付阶段" },
          { stage: "已生成跟进", count: 2, detail: "已生成任务，等待进入交付" }
        ],
        milestones: [
          { time_label: "第1周", title: "后端里程碑", detail: "后端里程碑详情" },
          { time_label: "07-07", title: "交付启动", detail: "30人销售团队需要AI获客陪跑" }
        ],
        cases: [{ id: 7, company: "后端企业案例", result: "后端案例结果" }]
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        requests: [{
          id: 8,
          user_id: 42,
          need: "后端返回的诊断预约",
          status: "submitted",
          created_at: "2026-07-07T10:30:00Z"
        }]
      }), { status: 200 }));
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-23T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "", account: "zhangjing", status: "active" }
    });

    render(
      <MemoryRouter initialEntries={["/enterprise"]}>
        <App />
      </MemoryRouter>
    );

    expect(await screen.findByRole("heading", { name: "后端陪跑方案" })).toBeInTheDocument();
    expect(screen.getByText("后端交付阶段")).toBeInTheDocument();
    expect(screen.getByText("已生成任务，等待进入交付")).toBeInTheDocument();
    expect(screen.getByText("后端里程碑")).toBeInTheDocument();
    expect(screen.getByText("交付启动")).toBeInTheDocument();
    expect(screen.getByText("后端企业案例")).toBeInTheDocument();
    expect(screen.getByText("后端返回的诊断预约")).toBeInTheDocument();
  });

  it("submits an enterprise diagnosis request", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({
        stats: [],
        plans: [],
        delivery_board: [],
        milestones: [],
        cases: []
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ requests: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        id: 7,
        user_id: 42,
        need: "30人销售团队需要AI获客陪跑",
        status: "submitted",
        created_at: "2026-07-07T10:30:00Z"
      }), { status: 200 }));
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-23T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "", account: "zhangjing", status: "active" }
    });

    render(
      <MemoryRouter initialEntries={["/enterprise"]}>
        <App />
      </MemoryRouter>
    );

    fireEvent.change(screen.getByLabelText("描述企业需求"), {
      target: { value: "30人销售团队需要AI获客陪跑" }
    });
    fireEvent.click(screen.getByRole("button", { name: "提交诊断预约" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/enterprise/diagnosis-requests", expect.objectContaining({
      method: "POST",
      body: JSON.stringify({ need: "30人销售团队需要AI获客陪跑" })
    })));
    expect(await screen.findByText("企业诊断预约已提交")).toBeInTheDocument();
    expect(screen.getByText("30人销售团队需要AI获客陪跑")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "生成诊断提纲" })).not.toBeInTheDocument();
  });

  it("creates a follow-up task from an enterprise diagnosis request", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({
        stats: [],
        plans: [],
        delivery_board: [],
        milestones: [],
        cases: []
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        requests: [{
          id: 8,
          user_id: 42,
          need: "30人销售团队需要AI获客陪跑",
          status: "submitted",
          created_at: "2026-07-07T10:30:00Z"
        }]
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        id: 99,
        title: "跟进企业诊断：30人销售团队需要AI获客陪跑",
        status: "todo"
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        id: 8,
        user_id: 42,
        need: "30人销售团队需要AI获客陪跑",
        status: "follow_up_created",
        created_at: "2026-07-07T10:30:00Z",
        updated_at: "2026-07-07T11:30:00Z"
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        stats: [],
        plans: [],
        delivery_board: [{ stage: "已生成跟进", count: 1, detail: "已生成任务，等待进入交付" }],
        milestones: [],
        cases: []
      }), { status: 200 }));
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-23T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "", account: "zhangjing", status: "active" }
    });

    render(
      <MemoryRouter initialEntries={["/enterprise"]}>
        <App />
      </MemoryRouter>
    );

    expect(await screen.findByText("30人销售团队需要AI获客陪跑")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "生成跟进任务" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/tasks", expect.objectContaining({
      method: "POST",
      body: JSON.stringify({
        title: "跟进企业诊断：30人销售团队需要AI获客陪跑",
        project: "企业定制化陪跑",
        priority: "high",
        tools: ["企业诊断", "CRM"],
        learning: "围绕企业需求制定陪跑方案：30人销售团队需要AI获客陪跑",
        source_type: "enterprise_diagnosis",
        source_id: 8,
        source_title: "企业诊断：30人销售团队需要AI获客陪跑",
        source_url: "/enterprise"
      })
    })));
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/enterprise/diagnosis-requests/8", expect.objectContaining({
      method: "PATCH",
      body: JSON.stringify({ status: "follow_up_created" })
    })));
    expect(await screen.findByText("跟进任务已生成")).toBeInTheDocument();
    expect(screen.getByText(/follow_up_created/)).toBeInTheDocument();
    expect(screen.getByText("已生成任务，等待进入交付")).toBeInTheDocument();
  });

  it("moves a followed-up enterprise diagnosis request into delivery", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({
        stats: [],
        plans: [],
        delivery_board: [{ stage: "已生成跟进", count: 1, detail: "已生成任务，等待进入交付" }],
        milestones: [],
        cases: []
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        requests: [{
          id: 8,
          user_id: 42,
          need: "30人销售团队需要AI获客陪跑",
          status: "follow_up_created",
          created_at: "2026-07-07T10:30:00Z",
          updated_at: "2026-07-07T11:30:00Z"
        }]
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        id: 8,
        user_id: 42,
        need: "30人销售团队需要AI获客陪跑",
        status: "in_delivery",
        created_at: "2026-07-07T10:30:00Z",
        updated_at: "2026-07-07T12:30:00Z"
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        stats: [],
        plans: [],
        delivery_board: [{ stage: "交付中预约", count: 1, detail: "已进入企业陪跑交付" }],
        milestones: [],
        cases: []
      }), { status: 200 }));
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-23T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "", account: "zhangjing", status: "active" }
    });

    render(
      <MemoryRouter initialEntries={["/enterprise"]}>
        <App />
      </MemoryRouter>
    );

    expect(await screen.findByText("30人销售团队需要AI获客陪跑")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "进入交付" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/enterprise/diagnosis-requests/8", expect.objectContaining({
      method: "PATCH",
      body: JSON.stringify({ status: "in_delivery" })
    })));
    expect(await screen.findByText("已进入交付")).toBeInTheDocument();
    expect(screen.getByText(/in_delivery/)).toBeInTheDocument();
    expect(screen.getByText("已进入企业陪跑交付")).toBeInTheDocument();
  });

  it("completes an in-delivery diagnosis request and refreshes enterprise cases", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({
        stats: [],
        plans: [],
        delivery_board: [{ stage: "交付中预约", count: 1, detail: "已进入企业陪跑交付" }],
        milestones: [{ time_label: "07-07", title: "交付启动", detail: "30人销售团队需要AI获客陪跑" }],
        cases: []
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        requests: [{
          id: 8,
          user_id: 42,
          need: "30人销售团队需要AI获客陪跑",
          status: "in_delivery",
          created_at: "2026-07-07T10:30:00Z",
          updated_at: "2026-07-07T12:30:00Z"
        }]
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        id: 8,
        user_id: 42,
        need: "30人销售团队需要AI获客陪跑",
        status: "completed",
        created_at: "2026-07-07T10:30:00Z",
        updated_at: "2026-07-07T13:30:00Z"
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        stats: [],
        plans: [],
        delivery_board: [{ stage: "已完成交付", count: 1, detail: "已完成交付并沉淀案例" }],
        milestones: [],
        cases: [{ id: -8, company: "企业诊断交付", result: "30人销售团队需要AI获客陪跑" }]
      }), { status: 200 }));
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-23T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "", account: "zhangjing", status: "active" }
    });

    render(
      <MemoryRouter initialEntries={["/enterprise"]}>
        <App />
      </MemoryRouter>
    );

    const completeButton = await screen.findByRole("button", { name: "完成交付" });
    fireEvent.click(completeButton);

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/enterprise/diagnosis-requests/8", expect.objectContaining({
      method: "PATCH",
      body: JSON.stringify({ status: "completed" })
    })));
    await waitFor(() => expect(screen.getAllByText("已完成交付").length).toBeGreaterThan(0));
    expect(screen.getByText(/completed/)).toBeInTheDocument();
    expect(screen.getByText("已完成交付并沉淀案例")).toBeInTheDocument();
    expect(screen.getByText("企业诊断交付")).toBeInTheDocument();
  });

  it("imports a completed diagnosis request into CRM", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({
        stats: [],
        plans: [],
        delivery_board: [{ stage: "已完成交付", count: 1, detail: "已完成交付并沉淀案例" }],
        milestones: [],
        cases: [{ id: -8, company: "企业诊断交付", result: "30人销售团队需要AI获客陪跑" }]
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        requests: [{
          id: 8,
          user_id: 42,
          need: "30人销售团队需要AI获客陪跑",
          status: "completed",
          created_at: "2026-07-07T10:30:00Z",
          updated_at: "2026-07-07T13:30:00Z"
        }]
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        id: 100,
        user_id: 42,
        import_key: "enterprise_diagnosis_request:8",
        name: "30人销售团队需要AI获客陪跑",
        stage: "won",
        source: "enterprise",
        created_at: "2026-07-07T13:30:00Z",
        updated_at: "2026-07-07T13:30:00Z"
      }), { status: 200 }));
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-23T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "", account: "zhangjing", status: "active" }
    });

    render(
      <MemoryRouter initialEntries={["/enterprise"]}>
        <App />
      </MemoryRouter>
    );

    const crmButton = await screen.findByRole("button", { name: "同步CRM" });
    fireEvent.click(crmButton);

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/enterprise/diagnosis-requests/8/crm-customer", expect.objectContaining({
      method: "POST"
    })));
    expect(await screen.findByText("已同步CRM客户：30人销售团队需要AI获客陪跑")).toBeInTheDocument();
  });
});
