import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import App from "../App";
import { authSession } from "../lib/authSession";

describe("GrowthComparisonPage", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  it("loads selected models and renders comparable metrics", async () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-08-16T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "测试用户", phone: "", account: "tester", status: "active" }
    });
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({
      models: [
        {
          id: 99, user_id: 7, name: "初版模型",
          assumptions: { monthly_visits: 12000, lead_rate: 0.08, deal_rate: 0.15, average_order: 6000, acquisition_cost: 80, delivery_cost: 120000 },
          result: { monthly_revenue: 864000, leads: 960, deals: 144, payback_days: 7, net_margin: 0.77 },
          created_at: "2026-08-15T08:00:00Z", updated_at: "2026-08-15T08:00:00Z"
        },
        {
          id: 100, user_id: 7, name: "优化模型",
          assumptions: { monthly_visits: 14000, lead_rate: 0.09, deal_rate: 0.17, average_order: 6200, acquisition_cost: 72, delivery_cost: 125000 },
          result: { monthly_revenue: 1325340, leads: 1260, deals: 214, payback_days: 5, net_margin: 0.81 },
          created_at: "2026-08-16T08:00:00Z", updated_at: "2026-08-16T08:00:00Z"
        }
      ],
      generated_at: "2026-08-16T08:30:00Z"
    }), { status: 200 }));

    render(
      <MemoryRouter initialEntries={["/growth-calculator/compare?model_ids=99,100"]}>
        <App />
      </MemoryRouter>
    );

    expect(await screen.findByRole("heading", { name: "测算对比" })).toBeInTheDocument();
    expect(screen.getByText("初版模型")).toBeInTheDocument();
    expect(screen.getByText("优化模型")).toBeInTheDocument();
    expect(screen.getByText("¥864,000")).toBeInTheDocument();
    expect(screen.getByText("¥1,325,340")).toBeInTheDocument();
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/growth/models/compare",
      expect.objectContaining({ method: "POST", body: JSON.stringify({ model_ids: [99, 100] }) })
    ));
  });

  it("rejects a comparison URL with fewer than two models", async () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-08-16T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "测试用户", phone: "", account: "tester", status: "active" }
    });
    const fetchMock = vi.spyOn(globalThis, "fetch");

    render(
      <MemoryRouter initialEntries={["/growth-calculator/compare?model_ids=99"]}>
        <App />
      </MemoryRouter>
    );

    expect(await screen.findByRole("alert")).toHaveTextContent("请选择 2 到 4 个测算模型进行对比");
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
