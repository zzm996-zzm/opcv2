import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import App from "../App";
import { authSession } from "../lib/authSession";

describe("GrowthCalculatorPage", () => {
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

  function renderGrowthRoute() {
    signIn();
    render(
      <MemoryRouter initialEntries={["/growth-calculator"]}>
        <App />
      </MemoryRouter>
    );
  }

  it("renders the growth calculator workbench instead of the placeholder", () => {
    renderGrowthRoute();

    expect(screen.getByRole("heading", { name: "增长测算" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "保存测算模型" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "增长漏斗" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "情景对比" })).toBeInTheDocument();
    expect(screen.getByText("保守方案")).toBeInTheDocument();
    expect(screen.queryByText("第一版正在实现")).not.toBeInTheDocument();
  });

  it("loads the latest growth model from API", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        models: [
          {
            id: 9,
            user_id: 7,
            name: "商业沙盘标准模型",
            assumptions: {
              monthly_visits: 18000,
              lead_rate: 0.08,
              deal_rate: 0.16,
              average_order: 1200,
              acquisition_cost: 58,
              delivery_cost: 420
            },
            result: {
              monthly_revenue: 276480,
              leads: 1440,
              deals: 230,
              payback_days: 13,
              net_margin: 0.42
            },
            created_at: "2026-06-30T08:00:00Z",
            updated_at: "2026-06-30T08:30:00Z"
          }
        ]
      }), { status: 200 })
    );

    renderGrowthRoute();

    expect(await screen.findByText("商业沙盘标准模型")).toBeInTheDocument();
    expect(screen.getByText("¥276,480")).toBeInTheDocument();
    expect(screen.getByDisplayValue("18,000")).toBeInTheDocument();
    expect(screen.getByDisplayValue("8.0%")).toBeInTheDocument();
  });

  it("shows backend load errors while keeping fallback model values visible", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ error: "invalid_request" }), { status: 400 })
    );

    renderGrowthRoute();

    expect(await screen.findByText("请求参数有误，请检查后重试")).toBeInTheDocument();
    expect(screen.getByText("智能客服系统 · 标准方案")).toBeInTheDocument();
    expect(screen.getAllByText("¥18.6万").length).toBeGreaterThan(0);
  });

  it("saves a growth model and refreshes the displayed result", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/growth/models" && init?.method === "GET") {
        return Promise.resolve(new Response(JSON.stringify({ models: [] }), { status: 200 }));
      }
      if (url === "/api/v1/growth/models" && init?.method === "POST") {
        return Promise.resolve(new Response(JSON.stringify({
          id: 12,
          user_id: 7,
          name: "智能客服系统 · 标准方案",
          assumptions: {
            monthly_visits: 24000,
            lead_rate: 0.068,
            deal_rate: 0.14,
            average_order: 820,
            acquisition_cost: 42,
            delivery_cost: 260
          },
          result: {
            monthly_revenue: 187000,
            leads: 1632,
            deals: 228,
            payback_days: 15,
            net_margin: 0.41
          },
          created_at: "2026-06-30T08:00:00Z",
          updated_at: "2026-06-30T08:05:00Z"
        }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderGrowthRoute();
    fireEvent.click(screen.getByRole("button", { name: "保存测算模型" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        "/api/v1/growth/models",
        expect.objectContaining({
          method: "POST",
          body: JSON.stringify({
            name: "智能客服系统 · 标准方案",
            monthly_visits: 24000,
            lead_rate: 0.068,
            deal_rate: 0.14,
            average_order: 820,
            acquisition_cost: 42,
            delivery_cost: 260
          })
        })
      );
    });
    expect(await screen.findByText("¥187,000")).toBeInTheDocument();
    expect(screen.getByText("15天")).toBeInTheDocument();
  });
});
