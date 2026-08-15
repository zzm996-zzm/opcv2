import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import App from "../App";
import { authSession } from "../lib/authSession";

describe("GrowthHistoryPage", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  function renderHistory(route = "/growth-calculator/history") {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-08-16T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "测试用户", phone: "", account: "tester", status: "active" }
    });
    render(
      <MemoryRouter initialEntries={[route]}>
        <App />
      </MemoryRouter>
    );
  }

  it("loads a filtered server page and links each model to its report", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        models: [{
          id: 99,
          user_id: 7,
          name: "SaaS 增长模型",
          assumptions: { monthly_visits: 24000, lead_rate: 0.068, deal_rate: 0.14, average_order: 820, acquisition_cost: 42, delivery_cost: 51000 },
          result: { monthly_revenue: 186960, leads: 1632, deals: 228, payback_days: 20, net_margin: 0.36 },
          created_at: "2026-08-15T08:30:00Z",
          updated_at: "2026-08-15T08:30:00Z"
        }],
        total: 21,
        limit: 10,
        offset: 10
      }), { status: 200 })
    );

    renderHistory("/growth-calculator/history?q=SaaS&page=2");

    expect(await screen.findByText("SaaS 增长模型")).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/growth/models?q=SaaS&limit=10&offset=10",
      expect.objectContaining({ method: "GET" })
    );
    expect(screen.getByRole("link", { name: "查看报告" })).toHaveAttribute("href", "/growth-calculator/report?model_id=99");
    expect(screen.getByRole("button", { name: "再次测算" })).toBeInTheDocument();
    expect(screen.getByText("2 / 3")).toBeInTheDocument();
  });

  it("submits search terms and resets pagination", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation(() => Promise.resolve(
      new Response(JSON.stringify({ models: [], total: 0, limit: 10, offset: 0 }), { status: 200 })
    ));
    renderHistory();

    fireEvent.change(screen.getByRole("textbox", { name: "搜索测算历史" }), { target: { value: "企业培训" } });
    fireEvent.click(screen.getByRole("button", { name: "查询" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/growth/models?q=%E4%BC%81%E4%B8%9A%E5%9F%B9%E8%AE%AD&limit=10&offset=0",
      expect.objectContaining({ method: "GET" })
    ));
    expect(await screen.findByText("没有匹配的测算记录")).toBeInTheDocument();
  });

  it("selects two models and opens the comparison page", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      if (String(input) === "/api/v1/growth/models/compare") {
        return Promise.resolve(new Response(JSON.stringify({
          models: [
            { id: 99, user_id: 7, name: "初版", assumptions: { monthly_visits: 10000, lead_rate: 0.08, deal_rate: 0.1, average_order: 1000, acquisition_cost: 50, delivery_cost: 20000 }, result: { monthly_revenue: 80000, leads: 800, deals: 80, payback_days: 10, net_margin: 0.5 }, created_at: "2026-08-15T08:00:00Z", updated_at: "2026-08-15T08:00:00Z" },
            { id: 100, user_id: 7, name: "优化版", assumptions: { monthly_visits: 12000, lead_rate: 0.09, deal_rate: 0.12, average_order: 1100, acquisition_cost: 45, delivery_cost: 22000 }, result: { monthly_revenue: 142560, leads: 1080, deals: 130, payback_days: 7, net_margin: 0.58 }, created_at: "2026-08-16T08:00:00Z", updated_at: "2026-08-16T08:00:00Z" }
          ],
          generated_at: "2026-08-16T08:30:00Z"
        }), { status: 200 }));
      }
      return Promise.resolve(new Response(JSON.stringify({
        models: [
          { id: 99, user_id: 7, name: "初版", assumptions: { monthly_visits: 10000, lead_rate: 0.08, deal_rate: 0.1, average_order: 1000, acquisition_cost: 50, delivery_cost: 20000 }, result: { monthly_revenue: 80000, leads: 800, deals: 80, payback_days: 10, net_margin: 0.5 }, created_at: "2026-08-15T08:00:00Z", updated_at: "2026-08-15T08:00:00Z" },
          { id: 100, user_id: 7, name: "优化版", assumptions: { monthly_visits: 12000, lead_rate: 0.09, deal_rate: 0.12, average_order: 1100, acquisition_cost: 45, delivery_cost: 22000 }, result: { monthly_revenue: 142560, leads: 1080, deals: 130, payback_days: 7, net_margin: 0.58 }, created_at: "2026-08-16T08:00:00Z", updated_at: "2026-08-16T08:00:00Z" }
        ], total: 2, limit: 10, offset: 0
      }), { status: 200 }));
    });

    renderHistory();
    expect(await screen.findByText("初版")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("checkbox", { name: "选择初版" }));
    fireEvent.click(screen.getByRole("checkbox", { name: "选择优化版" }));
    fireEvent.click(screen.getByRole("button", { name: "对比测算 (2/4)" }));

    expect(await screen.findByRole("heading", { name: "测算对比" })).toBeInTheDocument();
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/growth/models/compare",
      expect.objectContaining({ method: "POST", body: JSON.stringify({ model_ids: [99, 100] }) })
    ));
  });
});
