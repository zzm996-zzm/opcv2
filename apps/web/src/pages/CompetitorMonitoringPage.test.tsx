import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import App from "../App";

describe("CompetitorMonitoringPage", () => {
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

  function renderMonitoringRoute() {
    signIn();
    render(
      <MemoryRouter initialEntries={["/competitor-monitoring"]}>
        <App />
      </MemoryRouter>
    );
  }

  it("renders the competitor monitoring workbench with empty backend state", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ watchlist: [], events: [] }), { status: 200 })
    );

    renderMonitoringRoute();

    expect(screen.getByRole("heading", { name: "竞品动态监测" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "新增监测对象" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "监测中竞品" })).toBeInTheDocument();
    expect(await screen.findByText("暂无监测对象")).toBeInTheDocument();
    expect(screen.queryByText("小鹅通")).not.toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "动态时间线" })).toBeInTheDocument();
    expect(screen.queryByText("第一版正在实现")).not.toBeInTheDocument();
  });

  it("loads competitor monitoring data from API", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        watchlist: [
          {
            id: 77,
            name: "增长雷达",
            category: "商业情报 / 自动化",
            status: "高频变化",
            threat: "强",
            last_seen_at: "2026-06-30T08:00:00Z",
            channels: ["产品页", "价格页"],
            signal: "上线自动化任务派发模块"
          }
        ],
        events: [
          {
            occurred_at: "2026-06-30T08:30:00Z",
            company: "增长雷达",
            title: "自动任务派发上线",
            detail: "竞品开始把监测事件直接转成执行清单。",
            level: "强"
          }
        ]
      }), { status: 200 })
    );

    renderMonitoringRoute();

    expect(await screen.findByRole("heading", { name: "增长雷达" })).toBeInTheDocument();
    expect(screen.getByText("上线自动化任务派发模块")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "自动任务派发上线" })).toBeInTheDocument();
  });

  it("shows backend load errors without rendering fallback monitoring data", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ error: "invalid_request" }), { status: 400 })
    );

    renderMonitoringRoute();

    expect(await screen.findByText("请求参数有误，请检查后重试")).toBeInTheDocument();
    expect(screen.getByText("暂无监测对象")).toBeInTheDocument();
    expect(screen.queryByText("小鹅通")).not.toBeInTheDocument();
  });

  it("creates monitoring watch items from the page", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/competitor/monitoring?limit=20" && init?.method === "GET") {
        return Promise.resolve(new Response(JSON.stringify({ watchlist: [], events: [] }), { status: 200 }));
      }
      if (url === "/api/v1/competitor/monitoring/watchlist" && init?.method === "POST") {
        return Promise.resolve(new Response(JSON.stringify({
          name: "增长雷达",
          id: 77,
          category: "商业情报",
          status: "监测中",
          threat: "中",
          last_seen_at: "2026-07-07T09:30:00Z",
          channels: ["价格页", "招聘动态"],
          signal: "已创建监测规则，等待首次巡检。"
        }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderMonitoringRoute();

    fireEvent.click(screen.getByRole("button", { name: "新增监测对象" }));
    fireEvent.change(screen.getByLabelText("监测对象名称"), { target: { value: "增长雷达" } });
    fireEvent.change(screen.getByLabelText("对象分类"), { target: { value: "商业情报" } });
    fireEvent.click(screen.getByRole("button", { name: "保存监测对象" }));

    expect(await screen.findByRole("heading", { name: "增长雷达" })).toBeInTheDocument();
    expect(screen.getByText("已创建监测规则，等待首次巡检。")).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/competitor/monitoring/watchlist",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ name: "增长雷达", category: "商业情报", channels: ["官网 / 价格页", "招聘动态"] })
      })
    );
  });

  it("removes monitoring watch items from the page", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/competitor/monitoring?limit=20" && init?.method === "GET") {
        return Promise.resolve(new Response(JSON.stringify({
          watchlist: [{
            id: 77,
            name: "增长雷达",
            category: "商业情报",
            status: "监测中",
            threat: "中",
            last_seen_at: "2026-07-07T09:30:00Z",
            channels: ["价格页"],
            signal: "已创建监测规则，等待首次巡检。"
          }],
          events: []
        }), { status: 200 }));
      }
      if (url === "/api/v1/competitor/monitoring/watchlist/77" && init?.method === "DELETE") {
        return Promise.resolve(new Response(JSON.stringify({ deleted: true }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderMonitoringRoute();

    expect(await screen.findByRole("heading", { name: "增长雷达" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "移除 增长雷达" }));

    expect(await screen.findByText("暂无监测对象")).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "增长雷达" })).not.toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/competitor/monitoring/watchlist/77",
      expect.objectContaining({ method: "DELETE" })
    );
  });
});
