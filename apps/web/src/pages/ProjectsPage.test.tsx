import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import App from "../App";
import { authSession } from "../lib/authSession";

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

function renderProjectRoute(path: string) {
  signIn();
  render(
    <MemoryRouter initialEntries={[path]}>
      <App />
    </MemoryRouter>
  );
}

describe("ProjectsPage", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  it("renders the project market home from the design reference", () => {
    renderProjectRoute("/projects");

    expect(screen.getByRole("heading", { name: "项目超市" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "发现下一个可落地机会" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "去匹配" })).toHaveAttribute("href", "/projects/match");
    expect(screen.getByRole("heading", { name: "精选机会" })).toBeInTheDocument();
    expect(screen.getByRole("complementary", { name: "智活 Copilot" })).toBeInTheDocument();
  });

  it("collapses and expands the project Copilot panel", () => {
    renderProjectRoute("/projects");

    const copilot = screen.getByRole("complementary", { name: "智活 Copilot" });
    const collapseButton = screen.getByRole("button", { name: "收起项目超市 Copilot" });
    const body = document.getElementById("project-copilot-body");

    expect(collapseButton).toHaveAttribute("aria-expanded", "true");
    expect(body).not.toHaveAttribute("hidden");

    fireEvent.click(collapseButton);

    expect(copilot).toHaveClass("is-collapsed");
    expect(screen.getByRole("button", { name: "展开项目超市 Copilot" })).toHaveAttribute("aria-expanded", "false");
    expect(body).toHaveAttribute("hidden");

    fireEvent.click(screen.getByRole("button", { name: "展开项目超市 Copilot" }));

    expect(copilot).not.toHaveClass("is-collapsed");
    expect(body).not.toHaveAttribute("hidden");
  });

  it("renders the AI matching request page", () => {
    renderProjectRoute("/projects/match");

    expect(screen.getByRole("heading", { name: "AI匹配" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "告诉我你的目标、资源与偏好" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "提交给 AI 分析" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "匹配历史" })).toHaveAttribute("href", "/projects/history");
  });

  it("submits match request and renders API result", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        session_id: 99,
        status: "completed",
        projects: [
          {
            rank: 1,
            title: "本地AI获客顾问",
            score: 91,
            tags: ["B端服务", "轻资产"],
            budget: "¥2,000 - ¥6,000",
            reasons: ["客户需求明确", "交付可标准化"],
            risk: "需要控制交付边界"
          }
        ]
      }), { status: 200 })
    );
    renderProjectRoute("/projects/match");

    fireEvent.change(screen.getByLabelText("项目匹配需求"), {
      target: { value: "我想做一个本地B端AI获客服务，预算6000元" }
    });
    fireEvent.click(screen.getByRole("button", { name: "提交给 AI 分析" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/projects/matches",
      expect.objectContaining({ method: "POST" })
    ));
    expect(await screen.findByRole("heading", { name: "本地AI获客顾问" })).toBeInTheDocument();
    expect(screen.getByText("91分")).toBeInTheDocument();
  });

  it("answers dynamic match questions and renders the persisted result", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ session_id:99, status:"needs_input", questions:[{ key:"background", text:"你擅长什么？", options:["销售经验","内容创作"] }] }), { status:200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ session_id:99, status:"completed", projects:[{ rank:1, title:"AI销售顾问", score:90, tags:["B端"], budget:"1万", reasons:["经验匹配"], risk:"需验证获客" }] }), { status:200 }));
    renderProjectRoute("/projects/match");
    fireEvent.change(screen.getByLabelText("项目匹配需求"), { target:{ value:"想找项目" } });
    fireEvent.click(screen.getByRole("button", { name:"提交给 AI 分析" }));
    fireEvent.click(await screen.findByRole("button", { name:"销售经验" }));
    fireEvent.click(screen.getByRole("button", { name:"生成匹配结果" }));
    await waitFor(() => expect(fetchMock).toHaveBeenLastCalledWith("/api/v1/projects/matches/99/answers", expect.objectContaining({ method:"POST" })));
    expect(await screen.findByRole("heading", { name:"AI销售顾问" })).toBeInTheDocument();
  });

  it("syncs a persisted match result to task center", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      if (String(input) === "/api/v1/projects/matches" ) return Promise.resolve(new Response(JSON.stringify({ session_id:99, status:"completed", projects:[{ rank:1, title:"AI销售顾问", score:90, tags:["B端"], budget:"1万", reasons:["经验匹配"], risk:"需验证" }] }), { status:200 }));
      if (String(input) === "/api/v1/tasks/generate" && init?.method === "POST") return Promise.resolve(new Response(JSON.stringify({ tasks:[{ id:1 }] }), { status:200 }));
      return Promise.reject(new Error("unexpected"));
    });
    renderProjectRoute("/projects/match");
    fireEvent.change(screen.getByLabelText("项目匹配需求"), { target:{ value:"完整的项目需求，预算1万，每周20小时，偏好线上服务型项目" } });
    fireEvent.click(screen.getByRole("button", { name:"提交给 AI 分析" }));
    fireEvent.click(await screen.findByRole("button", { name:"生成落地任务" }));
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/tasks/generate", expect.objectContaining({ method:"POST" })));
    expect(await screen.findByText("已创建 1 个项目任务")).toBeInTheDocument();
  });

  it("renders opportunity exploration from API", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({ opportunities: [{
      id: 42,
      slug: "ai-sales-consulting",
      title: "AI销售顾问",
      summary: "为中小企业搭建销售自动化流程",
      industry: "企业服务",
      tags: ["轻资产", "B端服务"],
      budget_band: "1-3万",
      difficulty: "中等",
      resource_requirements: ["销售经验"]
    }] }), { status: 200 }));
    renderProjectRoute("/projects/explore");

    expect(screen.getByRole("heading", { name: "机会探索" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "高潜力机会" })).toBeInTheDocument();
    expect(await screen.findByRole("heading", { name: "AI销售顾问" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "查看机会" })).toHaveAttribute("href", "/projects/opportunities/ai-sales-consulting");
    expect(screen.getByRole("heading", { name: "机会雷达" })).toBeInTheDocument();
  });

  it("renders real case library from API evidence", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({ cases: [{
      id: 81, slug: "ai-sales-pilot", title: "AI销售试点", summary: "从单一销售场景开始验证",
      case_type: "success", outcome: "完成首轮流程验证", key_actions: ["先限定客户范围"], pitfalls: ["不要承诺未验证收益"],
      source_title: "企业公开复盘", source_url: "https://example.com/case", captured_at: "2026-07-01T08:00:00Z"
    }] }), { status: 200 }));
    renderProjectRoute("/projects/cases");

    expect(screen.getByRole("heading", { name: "真实案例库" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "成功样板" })).toBeInTheDocument();
    expect(await screen.findByRole("heading", { name: "AI销售试点" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "企业公开复盘" })).toHaveAttribute("href", "https://example.com/case");
    expect(screen.getByRole("heading", { name: "案例共性" })).toBeInTheDocument();
  });

  it("answers the AI follow-up questions and enables result generation", () => {
    renderProjectRoute("/projects/questions");

    expect(screen.getByRole("heading", { name: "AI补充提问" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Copilot 还想确认以下问题" })).toBeInTheDocument();
    expect(screen.getByText("已完成 0/4")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "生成匹配结果" })).toBeDisabled();

    ["服务型", "1个月内", "可以", "纯线上"].forEach((answer) => {
      fireEvent.click(screen.getByRole("button", { name: answer }));
      expect(screen.getByRole("button", { name: answer })).toHaveAttribute("aria-pressed", "true");
    });

    expect(screen.getByText("已完成 4/4")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "生成匹配结果" })).toHaveAttribute("href", "/projects/results");
    expect(screen.getByRole("link", { name: "返回修改基础需求" })).toHaveAttribute("href", "/projects/match");
    expect(screen.getByRole("link", { name: "稍后继续" })).toHaveAttribute("href", "/projects");
  });

  it("renders the latest persisted matching result without a static fallback", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/projects/matches") return Promise.resolve(new Response(JSON.stringify({ matches: [{
        id: 99, user_id: 7, intent: "线上轻资产项目", status: "completed",
        result: { session_id: 99, status: "completed", projects: [{ rank: 1, title: "本地AI获客顾问", score: 91, tags: ["B端"], budget: "¥2,000", reasons: ["经验匹配"], risk: "需验证获客" }] },
        created_at: "2026-06-24T12:00:00Z", updated_at: "2026-06-24T12:00:00Z"
      }] }), { status: 200 }));
      if (url === "/api/v1/projects/favorites") return Promise.resolve(new Response(JSON.stringify({ favorites: [] }), { status: 200 }));
      if (url === "/api/v1/projects/matches/99/favorite" && init?.method === "POST") return Promise.resolve(new Response(JSON.stringify({ id: 7, user_id: 7, session_id: 99 }), { status: 200 }));
      if (url === "/api/v1/projects/matches/99/favorite" && init?.method === "DELETE") return Promise.resolve(new Response(null, { status: 204 }));
      return Promise.reject(new Error(`unexpected ${url}`));
    });
    renderProjectRoute("/projects/results");

    expect(screen.getByRole("heading", { name: "为你匹配到的项目机会" })).toBeInTheDocument();
    expect(await screen.findByRole("heading", { name: "本地AI获客顾问" })).toBeInTheDocument();
    expect(screen.getAllByRole("button", { name: /加入对比/ }).length).toBeGreaterThan(0);
    expect(screen.getByRole("heading", { name: "推荐依据说明" })).toBeInTheDocument();
    expect(screen.queryByText("AI短视频脚本工作室")).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "收藏结果" }));
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/projects/matches/99/favorite", expect.objectContaining({ method: "POST" })));
    fireEvent.click(screen.getAllByRole("button", { name: "取消收藏" })[0]);
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/projects/matches/99/favorite", expect.objectContaining({ method: "DELETE" })));
  });

  it("renders explicit empty project result state", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({ matches: [] }), { status: 200 }));
    renderProjectRoute("/projects/results");
    expect(await screen.findByText("暂无已完成的匹配结果，请先提交项目匹配需求。")).toBeInTheDocument();
    expect(screen.queryByText("AI短视频脚本工作室")).not.toBeInTheDocument();
  });

  it("renders history and saved matches", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const session = {
            id: 99,
            user_id: 7,
            intent: "线上轻资产项目",
            status: "completed",
            result: {
              session_id: 99,
              status: "completed",
              projects: [{ rank: 1, title: "本地AI获客顾问", score: 91, tags: [], budget: "¥2,000", reasons: [], risk: "获客验证" }]
            },
            created_at: "2026-06-24T12:00:00Z",
            updated_at: "2026-06-24T12:00:00Z"
      };
      if (String(input) === "/api/v1/projects/matches") return Promise.resolve(new Response(JSON.stringify({ matches: [session] }), { status: 200 }));
      if (String(input) === "/api/v1/projects/favorites") return Promise.resolve(new Response(JSON.stringify({ favorites: [{ id: 7, user_id: 7, session_id: 99, session }] }), { status: 200 }));
      if (String(input) === "/api/v1/projects/matches/99/favorite" && init?.method === "DELETE") return Promise.resolve(new Response(null, { status: 204 }));
      return Promise.reject(new Error(`unexpected ${String(input)}`));
    });
    renderProjectRoute("/projects/history");

    expect(screen.getByRole("heading", { name: "匹配历史与收藏" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "历史匹配" })).toBeInTheDocument();
    expect((await screen.findAllByText("本地AI获客顾问")).length).toBeGreaterThan(0);
    expect(screen.getAllByRole("link", { name: "查看结果" })[0]).toHaveAttribute("href", "/projects/matches/99");
    expect(screen.getByRole("heading", { name: "收藏项目" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "取消收藏" }));
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/projects/matches/99/favorite", expect.objectContaining({ method: "DELETE" })));
    expect(await screen.findByText("暂无收藏项目")).toBeInTheDocument();
  });

  it("renders match detail from API session", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        id: 99,
        user_id: 7,
        intent: "本地B端AI获客服务",
        status: "completed",
        result: {
          session_id: 99,
          status: "completed",
          projects: [{
            rank: 1,
            title: "本地AI获客顾问",
            score: 91,
            tags: ["B端服务", "轻资产"],
            budget: "¥2,000 - ¥6,000",
            reasons: ["客户需求明确", "交付可标准化"],
            risk: "需要控制交付边界"
          }]
        },
        created_at: "2026-06-24T12:00:00Z",
        updated_at: "2026-06-24T12:00:00Z"
      }), { status: 200 })
    );
    renderProjectRoute("/projects/matches/99");

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/projects/matches/99",
      expect.objectContaining({ method: "GET" })
    ));
    expect(await screen.findByRole("heading", { name: "本地AI获客顾问", level: 1 })).toBeInTheDocument();
    expect(screen.getByText("匹配度 91分")).toBeInTheDocument();
    expect(screen.getByText("预算 ¥2,000 - ¥6,000")).toBeInTheDocument();
    expect(screen.getByText(/未引用外部证据/)).toBeInTheDocument();
  });

  it("renders paid sample overlay state", () => {
    renderProjectRoute("/projects/results/paywall");

    expect(screen.getByRole("heading", { name: "解锁完整拆解" })).toBeInTheDocument();
    expect(screen.getByText("智活AI会员")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "立即解锁" })).toBeInTheDocument();
  });

  it("renders an explicit empty state for the legacy detail route", () => {
    renderProjectRoute("/projects/detail");

    expect(screen.getByRole("heading", { name: "项目详情" })).toBeInTheDocument();
    expect(screen.getByText(/当前地址没有关联项目记录/)).toBeInTheDocument();
    expect(screen.queryByText("AI短视频脚本工作室")).not.toBeInTheDocument();
  });

  it("renders opportunity sections with traceable case evidence", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      if (String(input) === "/api/v1/projects/opportunities/ai-sales") return Promise.resolve(new Response(JSON.stringify({
        id: 42, slug: "ai-sales", title: "AI销售顾问", summary: "销售流程试点", industry: "企业服务", tags: ["B端"], budget_band: "1万", difficulty: "中等", resource_requirements: [],
        sections: [{ title: "验证路径", body: "先验证单一销售环节", items: ["记录人工基线"] }]
      }), { status: 200 }));
      if (String(input) === "/api/v1/projects/cases") return Promise.resolve(new Response(JSON.stringify({ cases: [{
        id: 81, slug: "sales-pilot", title: "销售试点复盘", summary: "公开试点记录", case_type: "success", outcome: "完成验证", key_actions: [], lessons: [], pitfalls: [], source_title: "企业公开复盘", source_url: "https://example.com/case", captured_at: "2026-07-01T08:00:00Z"
      }] }), { status: 200 }));
      return Promise.reject(new Error(`unexpected ${String(input)}`));
    });
    renderProjectRoute("/projects/opportunities/ai-sales");

    expect(await screen.findByRole("heading", { name: "AI销售顾问" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "验证路径" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "来源与证据" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "企业公开复盘" })).toHaveAttribute("href", "https://example.com/case");
  });

  it("creates a persisted project comparison", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      const items = [
        { id: 1, slug: "ai-sales", title: "AI销售", summary: "销售流程", industry: "企业服务", tags: [], budget_band: "1万", difficulty: "中等", resource_requirements: [] },
        { id: 2, slug: "ai-content", title: "AI内容", summary: "内容生产", industry: "内容", tags: [], budget_band: "5000", difficulty: "低", resource_requirements: [] }
      ];
      if (url === "/api/v1/projects/opportunities") return Promise.resolve(new Response(JSON.stringify({ opportunities: items }), { status: 200 }));
      if (url === "/api/v1/projects/comparisons" && init?.method === "POST") return Promise.resolve(new Response(JSON.stringify({ id: 61, items }), { status: 200 }));
      return Promise.reject(new Error(`unexpected ${url}`));
    });
    renderProjectRoute("/projects/compare");

    expect(screen.getByRole("heading", { name: "项目对比" })).toBeInTheDocument();
    fireEvent.click(await screen.findByRole("checkbox", { name: "AI销售" }));
    fireEvent.click(screen.getByRole("checkbox", { name: "AI内容" }));
    fireEvent.click(screen.getByRole("button", { name: "开始对比" }));
    expect(await screen.findByRole("heading", { name: "AI销售" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "AI内容" })).toBeInTheDocument();
  });

  it("creates and exposes a downloadable match export", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/projects/opportunities") return Promise.resolve(new Response(JSON.stringify({ opportunities: [] }), { status: 200 }));
      if (url === "/api/v1/projects/matches") return Promise.resolve(new Response(JSON.stringify({ matches: [{ id: 99, status: "completed", intent: "AI项目", result: { status: "completed", projects: [] } }] }), { status: 200 }));
      if (url === "/api/v1/projects/exports" && init?.method === "POST") return Promise.resolve(new Response(JSON.stringify({ id: 71, status: "ready", download_url: "/api/v1/projects/exports/71/download" }), { status: 200 }));
      return Promise.reject(new Error(`unexpected ${url}`));
    });
    renderProjectRoute("/projects/export");

    expect(screen.getByRole("heading", { name: "导出匹配报告" })).toBeInTheDocument();
    const exportButton = await screen.findByRole("button", { name: "确认导出" });
    await waitFor(() => expect(exportButton).toBeEnabled());
    fireEvent.click(exportButton);
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/projects/exports", expect.objectContaining({ method: "POST" })));
    expect(await screen.findByRole("link", { name: "下载 JSON 报告" })).toHaveAttribute("href", "/api/v1/projects/exports/71/download");
  });
});
