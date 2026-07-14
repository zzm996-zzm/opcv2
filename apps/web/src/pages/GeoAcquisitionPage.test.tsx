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

describe("GeoAcquisitionPage", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  function mockGeoAvailable() {
    vi.mocked(membershipApi.featureAccess).mockResolvedValue({
      features: [{ key: "geo_acquisition", label: "GEO获客", status: "available", allow_read_only: true, allow_workflow: true }]
    });
  }

  it("renders backend-connected GEO empty states", async () => {
    mockGeoAvailable();
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({
        stats: [],
        engines: [],
        lead_signals: [],
        keywords: [],
        content_tasks: []
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
      <MemoryRouter initialEntries={["/geo"]}>
        <App />
      </MemoryRouter>
    );

    expect(await screen.findByRole("button", { name: "生成GEO方案" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "GEO获客" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "AI 搜索覆盖" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "内容阵地任务" })).toBeInTheDocument();
    expect(screen.getByText("暂无GEO概览数据，提交一次分析请求后将逐步沉淀覆盖、关键词和内容任务。")).toBeInTheDocument();
    expect(screen.queryByText("GEO 后端接口未接入")).not.toBeInTheDocument();
    expect(screen.getByText("暂无 AI 搜索覆盖数据")).toBeInTheDocument();
    expect(screen.getByText("暂无线索机会")).toBeInTheDocument();
    expect(screen.getByText("暂无关键词机会")).toBeInTheDocument();
    expect(screen.getByText("暂无内容阵地任务")).toBeInTheDocument();
    expect(await screen.findByText("暂无 GEO 分析请求")).toBeInTheDocument();
    expect(screen.queryByText("智能客服系统怎么选")).not.toBeInTheDocument();
    expect(screen.queryByText("ChatGPT")).not.toBeInTheDocument();
    expect(screen.queryByText("潜在线索")).not.toBeInTheDocument();
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/geo/overview", expect.any(Object)));
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/geo/analysis-requests?limit=5", expect.any(Object));
    expect(screen.queryByText("第一版正在实现")).not.toBeInTheDocument();
  });

  it("focuses the GEO target input from the primary action", async () => {
    mockGeoAvailable();
    vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({
        stats: [],
        engines: [],
        lead_signals: [],
        keywords: [],
        content_tasks: []
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ requests: [] }), { status: 200 }));
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-23T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "", account: "zhangjing", status: "active" }
    });

    render(
      <MemoryRouter initialEntries={["/geo"]}>
        <App />
      </MemoryRouter>
    );

    fireEvent.click(await screen.findByRole("button", { name: "生成GEO方案" }));

    expect(screen.getByLabelText("输入GEO获客目标")).toHaveFocus();
    expect(await screen.findByText("暂无 GEO 分析请求")).toBeInTheDocument();
  });

  it("renders GEO overview records from the backend API", async () => {
    mockGeoAvailable();
    vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({
        stats: [{ key: "coverage", label: "AI引用覆盖", value: "12%" }],
        engines: [{ name: "Perplexity", coverage_percent: 22, status: "待优化" }],
        lead_signals: [{ title: "高意向问题", detail: "来自后端的线索信号" }],
        keywords: [{
          id: 1,
          query: "AI客服选型后端关键词",
          intent: "选型",
          coverage: "待覆盖",
          score: 71,
          action: "补充对比页"
        }],
        content_tasks: [{
          id: 2,
          type: "对比页",
          title: "后端返回的内容任务",
          priority: "高",
          due_at: "2026-07-08"
        }]
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        requests: [{
          id: 7,
          user_id: 42,
          target: "面向制造业的 AI 质检工具",
          status: "queued",
          created_at: "2026-07-04T08:00:00Z",
          updated_at: "2026-07-04T08:00:00Z"
        }]
      }), { status: 200 }));
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-23T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "", account: "zhangjing", status: "active" }
    });

    render(
      <MemoryRouter initialEntries={["/geo"]}>
        <App />
      </MemoryRouter>
    );

    expect(await screen.findByText("AI客服选型后端关键词")).toBeInTheDocument();
    expect(screen.getByText("Perplexity")).toBeInTheDocument();
    expect(screen.getByText("来自后端的线索信号")).toBeInTheDocument();
    expect(screen.getByText("后端返回的内容任务")).toBeInTheDocument();
    expect(screen.getByText("面向制造业的 AI 质检工具")).toBeInTheDocument();
    expect(screen.getByText("queued")).toBeInTheDocument();
  });

  it("submits a GEO analysis request without rendering generated mock results", async () => {
    mockGeoAvailable();
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({
        stats: [],
        engines: [],
        lead_signals: [],
        keywords: [],
        content_tasks: []
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ requests: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        id: 7,
        user_id: 42,
        target: "面向制造业的 AI 质检工具",
        status: "queued",
        created_at: "2026-07-04T08:00:00Z",
        updated_at: "2026-07-04T08:00:00Z"
      }), { status: 200 }));
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-23T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "", account: "zhangjing", status: "active" }
    });

    render(
      <MemoryRouter initialEntries={["/geo"]}>
        <App />
      </MemoryRouter>
    );

    fireEvent.change(await screen.findByLabelText("输入GEO获客目标"), {
      target: { value: "面向制造业的 AI 质检工具" }
    });
    fireEvent.click(screen.getByRole("button", { name: "分析 AI 搜索机会" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/geo/analysis-requests", expect.objectContaining({
      method: "POST",
      body: JSON.stringify({ target: "面向制造业的 AI 质检工具" })
    })));
    expect(await screen.findByText("GEO 分析请求已提交")).toBeInTheDocument();
    expect(screen.getByText("面向制造业的 AI 质检工具")).toBeInTheDocument();
    expect(screen.queryByText("智能客服系统怎么选")).not.toBeInTheDocument();
  });

  it("shows locked state without loading GEO business APIs", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ stats: [] }), { status: 200 })
    );
    vi.mocked(membershipApi.featureAccess).mockResolvedValue({
      features: [{
        key: "geo_acquisition",
        label: "GEO获客",
        status: "locked",
        required_plan: "pro",
        upgrade_url: "/membership",
        contact_url: "/enterprise",
        allow_read_only: false,
        allow_workflow: false
      }]
    });
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-23T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "", account: "zhangjing", status: "active" }
    });

    render(
      <MemoryRouter initialEntries={["/geo"]}>
        <App />
      </MemoryRouter>
    );

    expect(await screen.findByText("当前不会读取业务数据，也不会创建任务、客户或分析请求。")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "分析 AI 搜索机会" })).not.toBeInTheDocument();
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
