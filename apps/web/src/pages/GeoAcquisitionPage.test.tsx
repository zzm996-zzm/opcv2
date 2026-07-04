import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import App from "../App";
import { authSession } from "../lib/authSession";

describe("GeoAcquisitionPage", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  it("renders transparent empty states while GEO APIs are not connected", async () => {
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

    expect(screen.getByRole("heading", { name: "GEO获客" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "生成GEO方案" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "AI 搜索覆盖" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "内容阵地任务" })).toBeInTheDocument();
    expect(screen.getByText("GEO 后端接口未接入")).toBeInTheDocument();
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

  it("renders GEO overview records from the backend API", async () => {
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

    fireEvent.change(screen.getByLabelText("输入GEO获客目标"), {
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
});
