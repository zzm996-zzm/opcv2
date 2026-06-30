import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import App from "../App";
import { authSession } from "../lib/authSession";

describe("DashboardPage", () => {
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

  function renderDashboardRoute() {
    signIn();
    render(
      <MemoryRouter initialEntries={["/dashboard"]}>
        <App />
      </MemoryRouter>
    );
  }

  it("renders the dashboard workbench instead of the placeholder", () => {
    renderDashboardRoute();

    expect(screen.getByRole("heading", { name: "仪表盘" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "生成经营周报" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "经营指标" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "增长趋势" })).toBeInTheDocument();
    expect(screen.getByText("智能客服系统")).toBeInTheDocument();
    expect(screen.queryByText("第一版正在实现")).not.toBeInTheDocument();
  });

  it("loads dashboard summary from API", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        metrics: [
          { label: "本月收入", value: "¥23.4万", change: "+18%" },
          { label: "新增线索", value: "512", change: "+27%" }
        ],
        projects: [
          { name: "商业沙盘联调", value: "¥12万", leads: "线索 38", stage: "验证中" }
        ],
        trend: [{ label: "周一", value: 40 }, { label: "周二", value: 76 }],
        pipeline: [{ stage: "成交", count: "19", percent: "18%" }],
        alerts: [{ title: "任务接口待联调", detail: "任务中心已经接入新后端接口" }],
        actions: [{ time: "今天 16:00", title: "检查仪表盘聚合接口" }]
      }), { status: 200 })
    );

    renderDashboardRoute();

    expect(await screen.findByText("¥23.4万")).toBeInTheDocument();
    expect(screen.getByText("商业沙盘联调")).toBeInTheDocument();
    expect(screen.getByText("任务接口待联调")).toBeInTheDocument();
    expect(screen.getByText("检查仪表盘聚合接口")).toBeInTheDocument();
  });

  it("shows backend load errors while keeping fallback dashboard data visible", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ error: "invalid_request" }), { status: 400 })
    );

    renderDashboardRoute();

    expect(await screen.findByText("请求参数有误，请检查后重试")).toBeInTheDocument();
    expect(screen.getByText("智能客服系统")).toBeInTheDocument();
  });
});
