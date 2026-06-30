import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import App from "../App";

describe("SandboxPage", () => {
  afterEach(() => {
    cleanup();
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

  it("renders the business sandbox workbench", () => {
    signIn();
    render(<MemoryRouter initialEntries={["/sandbox"]}><App /></MemoryRouter>);

    expect(screen.getByRole("heading", { name: "商业沙盘" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /开始推演/ })).toHaveAttribute("href", "/sandbox/setup");
    expect(screen.getByText("多角色推演")).toBeInTheDocument();
    expect(screen.queryByText("第一版正在实现")).not.toBeInTheDocument();
  });

  it("renders sandbox setup and role selection states", () => {
    signIn();

    render(<MemoryRouter initialEntries={["/sandbox/roles"]}><App /></MemoryRouter>);

    expect(screen.getByRole("heading", { name: "你希望从谁的视角进行推演？" })).toBeInTheDocument();
    expect(screen.getByText("用户视角")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /确认角色/ })).toHaveAttribute("href", "/sandbox/start");
  });

  it("renders sandbox questions, report, history and quota states", () => {
    signIn();

    render(<MemoryRouter initialEntries={["/sandbox/questions"]}><App /></MemoryRouter>);
    expect(screen.getByRole("heading", { name: "你的目标用户更具体是哪些上班族？" })).toBeInTheDocument();
    cleanup();

    render(<MemoryRouter initialEntries={["/sandbox/report"]}><App /></MemoryRouter>);
    expect(screen.getByRole("heading", { name: "AI 驱动中小企业知识管理平台" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "核心结论" })).toBeInTheDocument();
    cleanup();

    render(<MemoryRouter initialEntries={["/sandbox/history"]}><App /></MemoryRouter>);
    expect(screen.getByRole("heading", { name: "历史推演" })).toBeInTheDocument();
    expect(screen.getByText("AI智能客服SaaS平台")).toBeInTheDocument();
    cleanup();

    render(<MemoryRouter initialEntries={["/sandbox/quota"]}><App /></MemoryRouter>);
    expect(screen.getByRole("dialog", { name: "本月沙盘次数已用尽" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "升级套餐" })).toHaveAttribute("href", "/membership");
  });

  it("loads sandbox sessions into history and report states", async () => {
    signIn();
    vi.spyOn(globalThis, "fetch").mockImplementation(() => Promise.resolve(
      new Response(JSON.stringify({
        sessions: [{
          id: 99,
          user_id: 7,
          goal: "验证企业AI运营平台",
          target_users: "连锁门店老板",
          product: "企业AI运营平台",
          roles: ["用户视角", "投资人视角"],
          status: "completed",
          report: {
            score: 91,
            summary: "AI运营平台具备清晰落地空间",
            metrics: [{ label: "综合可行性", value: "91" }],
            role_summaries: [{ role: "用户视角", view: "门店老板关注降本增效" }],
            risks: ["渠道教育成本偏高"],
            next_actions: ["先做3家门店试点"]
          },
          created_at: "2026-06-30T08:00:00Z",
          updated_at: "2026-06-30T08:10:00Z"
        }]
      }), { status: 200 })
    ));

    render(<MemoryRouter initialEntries={["/sandbox/history"]}><App /></MemoryRouter>);

    expect(await screen.findByText("企业AI运营平台")).toBeInTheDocument();
    expect(screen.getByText("验证企业AI运营平台")).toBeInTheDocument();
    cleanup();

    render(<MemoryRouter initialEntries={["/sandbox/report"]}><App /></MemoryRouter>);

    expect(await screen.findByRole("heading", { name: "企业AI运营平台" })).toBeInTheDocument();
    expect(screen.getByText("AI运营平台具备清晰落地空间")).toBeInTheDocument();
    expect(screen.getByText("门店老板关注降本增效")).toBeInTheDocument();
    expect(screen.getByText("先做3家门店试点")).toBeInTheDocument();
  });

  it("creates and runs a sandbox session from the start page", async () => {
    signIn();
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/sandbox/sessions?limit=10") {
        return Promise.resolve(new Response(JSON.stringify({ sessions: [] }), { status: 200 }));
      }
      if (url === "/api/v1/sandbox/sessions" && init?.method === "POST") {
        return Promise.resolve(new Response(JSON.stringify({
          id: 123,
          user_id: 7,
          goal: "验证 AI 低卡代餐奶昔",
          target_users: "上班族",
          product: "AI 低卡代餐奶昔",
          roles: ["用户视角", "投资人视角", "竞争对手视角", "运营视角"],
          status: "draft",
          created_at: "2026-06-30T08:00:00Z",
          updated_at: "2026-06-30T08:00:00Z"
        }), { status: 200 }));
      }
      if (url === "/api/v1/sandbox/sessions/123/run" && init?.method === "POST") {
        return Promise.resolve(new Response(JSON.stringify({
          id: 123,
          user_id: 7,
          goal: "验证 AI 低卡代餐奶昔",
          target_users: "上班族",
          product: "AI 低卡代餐奶昔",
          roles: ["用户视角", "投资人视角", "竞争对手视角", "运营视角"],
          status: "completed",
          report: {
            score: 86,
            summary: "代餐奶昔项目适合先做小范围验证",
            metrics: [{ label: "综合可行性", value: "86" }],
            role_summaries: [{ role: "用户视角", view: "用户需要口味和饱腹感双验证" }],
            risks: ["线下履约成本需要控制"],
            next_actions: ["先完成20位上班族访谈"]
          },
          created_at: "2026-06-30T08:00:00Z",
          updated_at: "2026-06-30T08:02:00Z"
        }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    render(<MemoryRouter initialEntries={["/sandbox/start"]}><App /></MemoryRouter>);
    fireEvent.click(screen.getByRole("link", { name: /开始推演/ }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith("/api/v1/sandbox/sessions", expect.objectContaining({ method: "POST" }));
    });
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/sandbox/sessions/123/run", expect.objectContaining({ method: "POST" }));
    expect(await screen.findByRole("heading", { name: "AI 低卡代餐奶昔" })).toBeInTheDocument();
    expect(screen.getByText("代餐奶昔项目适合先做小范围验证")).toBeInTheDocument();
  });

  it("shows backend error messages when starting a sandbox session fails", async () => {
    signIn();
    vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/sandbox/sessions?limit=10") {
        return Promise.resolve(new Response(JSON.stringify({ sessions: [] }), { status: 200 }));
      }
      if (url === "/api/v1/sandbox/sessions" && init?.method === "POST") {
        return Promise.resolve(new Response(JSON.stringify({ error: "invalid_request" }), { status: 400 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    render(<MemoryRouter initialEntries={["/sandbox/start"]}><App /></MemoryRouter>);
    fireEvent.click(screen.getByRole("link", { name: /开始推演/ }));

    expect(await screen.findByText("请求参数有误，请检查后重试")).toBeInTheDocument();
  });
});
