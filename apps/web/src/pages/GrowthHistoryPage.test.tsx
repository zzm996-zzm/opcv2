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
});
