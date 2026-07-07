import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import CompetitorDataPage from "./CompetitorDataPage";

describe("CompetitorDataPage", () => {
  afterEach(() => {
    authSession.clear();
    vi.useRealTimers();
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

  function renderCompetitorDataPage() {
    signIn();
    render(
      <MemoryRouter>
        <CompetitorDataPage />
      </MemoryRouter>
    );
  }

  it("renders the competitor data cracking workbench with empty backend state", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ scans: [] }), { status: 200 })
    );

    renderCompetitorDataPage();

    expect(screen.getByRole("heading", { name: "竞品全盘数据破解" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "启动采集任务" })).toBeInTheDocument();
    expect(await screen.findByText("暂无竞品画像")).toBeInTheDocument();
    expect(screen.queryByText("小鹅通")).not.toBeInTheDocument();
    expect(screen.getByText("AI 破解结论")).toBeInTheDocument();
    expect(screen.queryByText("第一版正在实现")).not.toBeInTheDocument();
  });

  it("loads the latest competitor scan from API", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        scans: [
          {
            id: 11,
            user_id: 7,
            targets: ["商业沙盘竞品"],
            focus: "价格和产品变化",
            status: "completed",
            competitors: [
              {
                name: "增长雷达",
                category: "竞品监测 / 商业情报",
                score: 88,
                signal: "新增自动化竞品预警和任务派发能力",
                risk: "强",
                tags: ["产品更新", "自动化"]
              }
            ],
            conclusions: [
              { title: "自动化增强", detail: "竞品正在把监测结果直接转成销售动作。" }
            ],
            evidence_sources: [
              {
                source_type: "official_site",
                title: "增长雷达产品更新页",
                url: "https://example.com/release",
                summary: "新增自动化竞品预警和任务派发能力",
                captured_at: "2026-06-30T08:20:00Z"
              }
            ],
            created_at: "2026-06-30T08:00:00Z",
            updated_at: "2026-06-30T08:30:00Z"
          }
        ]
      }), { status: 200 })
    );

    renderCompetitorDataPage();

    expect(await screen.findByRole("heading", { name: "增长雷达" })).toBeInTheDocument();
    expect(screen.getAllByText("新增自动化竞品预警和任务派发能力")).toHaveLength(2);
    expect(screen.getByText("自动化增强")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "证据来源" })).toBeInTheDocument();
    expect(screen.getByText("增长雷达产品更新页")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "打开来源" })).toHaveAttribute("href", "https://example.com/release");
  });

  it("adds scan competitors to monitoring from the card", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/competitor/scans?limit=20" && init?.method === "GET") {
        return Promise.resolve(new Response(JSON.stringify({
          scans: [{
            id: 11,
            user_id: 7,
            targets: ["商业沙盘竞品"],
            focus: "价格和产品变化",
            status: "completed",
            competitors: [{
              name: "增长雷达",
              category: "竞品监测 / 商业情报",
              score: 88,
              signal: "新增自动化竞品预警和任务派发能力",
              risk: "强",
              tags: ["产品更新", "自动化"]
            }],
            conclusions: [],
            evidence_sources: [],
            created_at: "2026-06-30T08:00:00Z",
            updated_at: "2026-06-30T08:30:00Z"
          }]
        }), { status: 200 }));
      }
      if (url === "/api/v1/competitor/scans/11/watchlist" && init?.method === "POST") {
        return Promise.resolve(new Response(JSON.stringify({
          id: 77,
          name: "增长雷达",
          category: "竞品监测 / 商业情报",
          status: "监测中",
          threat: "强",
          last_seen_at: "2026-07-07T10:30:00Z",
          channels: ["产品页", "价格页", "内容矩阵"],
          signal: "新增自动化竞品预警和任务派发能力"
        }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderCompetitorDataPage();

    expect(await screen.findByRole("heading", { name: "增长雷达" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "加入监测 增长雷达" }));

    expect(await screen.findByText("已将增长雷达加入动态监测。")).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/competitor/scans/11/watchlist",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ competitor_name: "增长雷达" })
      })
    );
  });

  it("creates a counter task from the latest competitor conclusion", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/competitor/scans?limit=20" && init?.method === "GET") {
        return Promise.resolve(new Response(JSON.stringify({
          scans: [{
            id: 11,
            user_id: 7,
            targets: ["商业沙盘竞品"],
            focus: "价格和产品变化",
            status: "completed",
            competitors: [{
              name: "增长雷达",
              category: "竞品监测 / 商业情报",
              score: 88,
              signal: "新增自动化竞品预警和任务派发能力",
              risk: "强",
              tags: ["产品更新", "自动化"]
            }],
            conclusions: [
              { title: "销售自动化提速", detail: "竞品正在把AI能力嵌入销售跟进链路。" }
            ],
            evidence_sources: [],
            created_at: "2026-06-30T08:00:00Z",
            updated_at: "2026-06-30T08:30:00Z"
          }]
        }), { status: 200 }));
      }
      if (url === "/api/v1/tasks" && init?.method === "POST") {
        return Promise.resolve(new Response(JSON.stringify({
          id: 88,
          user_id: 7,
          title: "反击：销售自动化提速",
          project: "竞品动态监测",
          status: "todo",
          priority: "high",
          tools: ["竞品全盘数据破解", "任务中心"],
          learning: "竞品正在把AI能力嵌入销售跟进链路。",
          created_at: "2026-07-07T10:00:00Z",
          updated_at: "2026-07-07T10:00:00Z"
        }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderCompetitorDataPage();

    expect(await screen.findByText("销售自动化提速")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "生成反击任务" }));

    expect(await screen.findByText("已生成反击任务：反击：销售自动化提速")).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/tasks",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({
          title: "反击：销售自动化提速",
          project: "竞品动态监测",
          priority: "high",
          tools: ["竞品全盘数据破解", "任务中心"],
          learning: "竞品正在把AI能力嵌入销售跟进链路。"
        })
      })
    );
  });

  it("shows backend load errors without rendering fallback competitor data", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ error: "invalid_request" }), { status: 400 })
    );

    renderCompetitorDataPage();

    expect(await screen.findByText("请求参数有误，请检查后重试")).toBeInTheDocument();
    expect(screen.getByText("暂无竞品画像")).toBeInTheDocument();
    expect(screen.queryByText("小鹅通")).not.toBeInTheDocument();
  });

  it("creates a competitor scan and refreshes the displayed conclusion", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/competitor/scans?limit=20" && init?.method === "GET") {
        return Promise.resolve(new Response(JSON.stringify({ scans: [] }), { status: 200 }));
      }
      if (url === "/api/v1/competitor/scans" && init?.method === "POST") {
        return Promise.resolve(new Response(JSON.stringify({
          id: 12,
          user_id: 7,
          targets: ["小鹅通", "有赞教育", "企微管家"],
          focus: "价格、案例、招聘和 AI 功能",
          status: "completed",
          competitors: [
            {
              name: "私域增长助手",
              category: "私域运营 / AI 销售",
              score: 92,
              signal: "正在强化AI销售助手和企微自动跟进能力",
              risk: "强",
              tags: ["AI销售", "企微自动化"]
            }
          ],
          conclusions: [
            { title: "销售自动化提速", detail: "竞品正在把AI能力嵌入销售跟进链路。" }
          ],
          created_at: "2026-06-30T08:00:00Z",
          updated_at: "2026-06-30T08:05:00Z"
        }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderCompetitorDataPage();
    fireEvent.click(screen.getByRole("button", { name: "启动采集任务" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        "/api/v1/competitor/scans",
        expect.objectContaining({
          method: "POST",
          body: JSON.stringify({
            targets: ["小鹅通", "有赞教育", "企微管家"],
            focus: "价格、案例、招聘和 AI 功能"
          })
        })
      );
    });
    expect(await screen.findByRole("heading", { name: "私域增长助手" })).toBeInTheDocument();
    expect(screen.getByText("销售自动化提速")).toBeInTheDocument();
  });

  it("creates a competitor scan from the custom plan input", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/competitor/scans?limit=20" && init?.method === "GET") {
        return Promise.resolve(new Response(JSON.stringify({ scans: [] }), { status: 200 }));
      }
      if (url === "/api/v1/competitor/scans" && init?.method === "POST") {
        return Promise.resolve(new Response(JSON.stringify({
          id: 14,
          user_id: 7,
          targets: ["增长雷达", "线索火花"],
          focus: "获客链路、价格页和AI销售能力",
          status: "completed",
          competitors: [
            {
              name: "线索火花",
              category: "AI 销售 / 线索开发",
              score: 86,
              signal: "新增从竞品信号生成跟进任务的能力",
              risk: "强",
              tags: ["AI销售", "任务派发"]
            }
          ],
          conclusions: [
            { title: "获客动作前置", detail: "竞品正在把公开信号提前转成销售跟进动作。" }
          ],
          evidence_sources: [],
          created_at: "2026-06-30T08:00:00Z",
          updated_at: "2026-06-30T08:05:00Z"
        }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderCompetitorDataPage();

    fireEvent.change(screen.getByLabelText("输入竞品或关键词"), {
      target: { value: "增长雷达、线索火花；重点关注获客链路、价格页和AI销售能力" }
    });
    fireEvent.click(screen.getByRole("button", { name: "生成采集计划" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        "/api/v1/competitor/scans",
        expect.objectContaining({
          method: "POST",
          body: JSON.stringify({
            targets: ["增长雷达", "线索火花"],
            focus: "获客链路、价格页和AI销售能力"
          })
        })
      );
    });
    expect(await screen.findByRole("heading", { name: "线索火花" })).toBeInTheDocument();
    expect(screen.getByText("获客动作前置")).toBeInTheDocument();
  });

  it("shows queued scan progress after creating a competitor scan", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/competitor/scans?limit=20" && init?.method === "GET") {
        return Promise.resolve(new Response(JSON.stringify({ scans: [] }), { status: 200 }));
      }
      if (url === "/api/v1/competitor/scans" && init?.method === "POST") {
        return Promise.resolve(new Response(JSON.stringify({
          id: 13,
          user_id: 7,
          targets: ["小鹅通", "有赞教育", "企微管家"],
          focus: "价格、案例、招聘和 AI 功能",
          status: "queued",
          progress_percent: 0,
          current_step: "queued",
          competitors: [],
          conclusions: [],
          created_at: "2026-06-30T08:00:00Z",
          updated_at: "2026-06-30T08:00:00Z"
        }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderCompetitorDataPage();
    fireEvent.click(screen.getByRole("button", { name: "启动采集任务" }));

    expect(await screen.findByText("排队中")).toBeInTheDocument();
    expect(screen.getByText("脚本任务已排队，等待采集账号执行。")).toBeInTheDocument();
    expect(screen.getByText("0%")).toBeInTheDocument();
  });

  it("polls queued scans and displays worker progress updates", async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/competitor/scans?limit=20" && init?.method === "GET") {
        return Promise.resolve(new Response(JSON.stringify({
          scans: [{
            id: 13,
            user_id: 7,
            targets: ["小鹅通"],
            focus: "价格变化",
            status: "queued",
            progress_percent: 0,
            current_step: "queued",
            competitors: [],
            conclusions: [],
            created_at: "2026-06-30T08:00:00Z",
            updated_at: "2026-06-30T08:00:00Z"
          }]
        }), { status: 200 }));
      }
      if (url === "/api/v1/competitor/scans/13" && init?.method === "GET") {
        return Promise.resolve(new Response(JSON.stringify({
          id: 13,
          user_id: 7,
          targets: ["小鹅通"],
          focus: "价格变化",
          status: "running",
          progress_percent: 30,
          current_step: "collecting_sources",
          competitors: [],
          conclusions: [],
          created_at: "2026-06-30T08:00:00Z",
          updated_at: "2026-06-30T08:01:00Z"
        }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderCompetitorDataPage();

    expect(await screen.findByText("排队中")).toBeInTheDocument();
    await vi.advanceTimersByTimeAsync(3000);

    expect(await screen.findByText("脚本正在采集公开数据，完成后会生成 AI 破解结论。")).toBeInTheDocument();
    expect(screen.getByText("30%")).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/competitor/scans/13", expect.objectContaining({ method: "GET" }));
  });

  it("retries failed competitor scans from the status panel", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/competitor/scans?limit=20" && init?.method === "GET") {
        return Promise.resolve(new Response(JSON.stringify({
          scans: [{
            id: 13,
            user_id: 7,
            targets: ["小鹅通"],
            focus: "价格变化",
            status: "failed",
            progress_percent: 100,
            current_step: "failed",
            error_message: "scanner_not_configured",
            competitors: [],
            conclusions: [],
            evidence_sources: [],
            created_at: "2026-06-30T08:00:00Z",
            updated_at: "2026-06-30T08:01:00Z"
          }]
        }), { status: 200 }));
      }
      if (url === "/api/v1/competitor/scans/13/retry" && init?.method === "POST") {
        return Promise.resolve(new Response(JSON.stringify({
          id: 13,
          user_id: 7,
          targets: ["小鹅通"],
          focus: "价格变化",
          status: "queued",
          progress_percent: 0,
          current_step: "queued",
          competitors: [],
          conclusions: [],
          evidence_sources: [],
          created_at: "2026-06-30T08:00:00Z",
          updated_at: "2026-06-30T08:02:00Z"
        }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderCompetitorDataPage();

    expect(await screen.findByText("scanner_not_configured")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "重新采集" }));

    expect(await screen.findByText("排队中")).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/competitor/scans/13/retry", expect.objectContaining({ method: "POST" }));
  });
});
