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
    expect(screen.queryByRole("heading", { name: "全部页面" })).not.toBeInTheDocument();
  });

  it("hides and restores the complete project Copilot panel", () => {
    renderProjectRoute("/projects");

    const collapseButton = screen.getByRole("button", { name: "收起项目超市 Copilot" });

    fireEvent.click(collapseButton);

    expect(screen.queryByRole("complementary", { name: "智活 Copilot" })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "打开智活 Copilot" })).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "打开智活 Copilot" }));

    expect(screen.getByRole("complementary", { name: "智活 Copilot" })).toBeInTheDocument();
    expect(document.getElementById("project-copilot-body")).toBeInTheDocument();
  });

  it("renders the AI matching request page", () => {
    renderProjectRoute("/projects/match");

    expect(screen.getByRole("heading", { name: "AI匹配" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "告诉我你的目标、资源与偏好" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "提交给 AI 分析" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "匹配历史" })).toHaveAttribute("href", "/projects/history");
  });

  it("submits match request and renders API result", async () => {
    const project = { rank:1, opportunity_slug:"local-ai-sales-consulting", title:"本地AI获客顾问", score:91, tags:["B端服务","轻资产"], budget:"¥2,000 - ¥6,000", reasons:["客户需求明确","交付可标准化"], risk:"需要控制交付边界" };
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/projects/matches" && init?.method === "POST") return Promise.resolve(new Response(JSON.stringify({ session_id:99, status:"completed", projects:[project] }), { status:200 }));
      if (url === "/api/v1/projects/matches/99") return Promise.resolve(new Response(JSON.stringify({ id:99, user_id:7, intent:"本地AI获客服务", status:"completed", result:{ session_id:99, status:"completed", projects:[project] }, created_at:"2026-06-24T12:00:00Z", updated_at:"2026-06-24T12:00:00Z" }), { status:200 }));
      if (url === "/api/v1/projects/favorites") return Promise.resolve(new Response(JSON.stringify({ favorites:[] }), { status:200 }));
      return Promise.reject(new Error(`unexpected ${url}`));
    });
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
    const project = { rank:1, opportunity_slug:"local-ai-sales-consulting", title:"AI销售顾问", score:90, tags:["B端"], budget:"1万", reasons:["经验匹配"], risk:"需验证获客" };
    let answered = false;
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/projects/matches/99/answers" && init?.method === "POST") {
        answered = true;
        return Promise.resolve(new Response(JSON.stringify({ session_id:99, status:"completed", projects:[project] }), { status:200 }));
      }
      if (url === "/api/v1/projects/matches/99") {
        return Promise.resolve(new Response(JSON.stringify(answered
          ? { id:99, user_id:7, intent:"想找项目", answers:[{ key:"background", value:"销售经验" }], status:"completed", result:{ session_id:99, status:"completed", projects:[project] }, created_at:"2026-06-24T12:00:00Z", updated_at:"2026-06-24T12:00:00Z" }
          : { id:99, user_id:7, intent:"想找项目", status:"needs_input", questions:[{ key:"background", text:"你擅长什么？", options:["销售经验","内容创作","技术能力"] }], created_at:"2026-06-24T12:00:00Z", updated_at:"2026-06-24T12:00:00Z" }), { status:200 }));
      }
      if (url === "/api/v1/projects/favorites") return Promise.resolve(new Response(JSON.stringify({ favorites:[] }), { status:200 }));
      return Promise.reject(new Error(`unexpected ${url}`));
    });
    renderProjectRoute("/projects/matches/99/questions");
    fireEvent.click(await screen.findByRole("button", { name:"销售经验" }));
    fireEvent.click(screen.getByRole("button", { name:"生成匹配结果" }));
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/projects/matches/99/answers", expect.objectContaining({ method:"POST" })));
    expect(await screen.findByRole("heading", { name:"AI销售顾问" })).toBeInTheDocument();
  });

  it("syncs a persisted match result to task center", async () => {
    const project = { rank:1, opportunity_slug:"local-ai-sales-consulting", title:"AI销售顾问", score:90, tags:["B端"], budget:"1万", reasons:["经验匹配"], risk:"需验证" };
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      if (String(input) === "/api/v1/projects/matches/99") return Promise.resolve(new Response(JSON.stringify({ id:99, user_id:7, intent:"线上服务项目", status:"completed", result:{ session_id:99, status:"completed", projects:[project] }, created_at:"2026-06-24T12:00:00Z", updated_at:"2026-06-24T12:00:00Z" }), { status:200 }));
      if (String(input) === "/api/v1/projects/favorites") return Promise.resolve(new Response(JSON.stringify({ favorites:[] }), { status:200 }));
      if (String(input) === "/api/v1/tasks/generate" && init?.method === "POST") return Promise.resolve(new Response(JSON.stringify({ tasks:[{ id:1 }] }), { status:200 }));
      return Promise.reject(new Error("unexpected"));
    });
    renderProjectRoute("/projects/matches/99/results");
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
    expect(screen.getByRole("navigation", { name: "项目机会分页" })).toBeInTheDocument();
  });

  it("keeps the reference opportunity grid when the catalog is empty", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation(() => Promise.resolve(new Response(JSON.stringify({ opportunities: [] }), { status: 200 })));
    renderProjectRoute("/projects/explore");

    expect(await screen.findByRole("heading", { name: "AI智能简历优化服务" })).toBeInTheDocument();
    expect(screen.getAllByRole("link", { name: "查看机会" })).toHaveLength(8);
    expect(screen.getByText("共 120 条")).toBeInTheDocument();
    expect(screen.queryByText("暂无符合条件的已发布项目机会")).not.toBeInTheDocument();
  });

  it("renders real case library from API evidence", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({ cases: [{
      id: 81, slug: "ai-sales-pilot", title: "AI销售试点", summary: "从单一销售场景开始验证",
      case_type: "success", outcome: "完成首轮流程验证", key_actions: ["先限定客户范围"], pitfalls: ["不要承诺未验证收益"],
      source_title: "企业公开复盘", source_url: "https://example.com/case", captured_at: "2026-07-01T08:00:00Z"
    }] }), { status: 200 }));
    renderProjectRoute("/projects/cases");

    expect(screen.getByRole("heading", { name: "真实案例库" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "成功案例" })).toBeInTheDocument();
    expect(await screen.findByRole("heading", { name: "AI销售试点" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "企业公开复盘" })).toHaveAttribute("href", "https://example.com/case");
    expect(screen.getByRole("heading", { name: "案例共性" })).toBeInTheDocument();
  });

  it("answers persisted AI follow-up questions and enables result generation", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({
      id:99, user_id:7, intent:"想找线上项目", status:"needs_input",
      questions:[
        { key:"type", text:"你更偏好服务型还是产品型？", options:["服务型","产品型","都可以"] },
        { key:"income", text:"你希望多久看到第一笔收入？", options:["1个月内","1-3个月","3个月以上"] },
        { key:"camera", text:"你是否接受出镜或打造个人IP？", options:["可以","尽量不出镜","无所谓"] },
        { key:"channel", text:"更偏好线上项目还是本地项目？", options:["纯线上","本地服务","都可以"] }
      ], created_at:"2026-06-24T12:00:00Z", updated_at:"2026-06-24T12:00:00Z"
    }), { status:200 }));
    renderProjectRoute("/projects/matches/99/questions");

    expect(await screen.findByRole("heading", { name: "AI补充提问" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Copilot 还想确认以下问题" })).toBeInTheDocument();
    expect(screen.getByText("已完成 0/4")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "生成匹配结果" })).toBeDisabled();

    ["服务型", "1个月内", "可以", "纯线上"].forEach((answer) => {
      fireEvent.click(screen.getByRole("button", { name: answer }));
      expect(screen.getByRole("button", { name: answer })).toHaveAttribute("aria-pressed", "true");
    });

    expect(screen.getByText("已完成 4/4")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "生成匹配结果" })).toBeEnabled();
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
    expect(screen.getAllByRole("link", { name: /加入对比/ }).length).toBeGreaterThan(0);
    expect(screen.getByRole("heading", { name: "推荐依据说明" })).toBeInTheDocument();
    expect(screen.queryByText("AI短视频脚本工作室")).not.toBeInTheDocument();
    fireEvent.click(screen.getAllByRole("button", { name: "收藏结果" })[0]);
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
    expect(screen.getAllByRole("link", { name: "查看结果" })[0]).toHaveAttribute("href", "/projects/matches/99/results");
    expect(screen.getByRole("heading", { name: "收藏结果" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "取消收藏" }));
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/projects/matches/99/favorite", expect.objectContaining({ method: "DELETE" })));
    expect(await screen.findByText("暂无收藏结果")).toBeInTheDocument();
  });

  it("keeps match history visible when favorites fail to load", async () => {
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
    vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/projects/matches") return Promise.resolve(new Response(JSON.stringify({ matches: [session] }), { status: 200 }));
      if (url === "/api/v1/projects/favorites") return Promise.resolve(new Response(JSON.stringify({ message: "favorites unavailable" }), { status: 503 }));
      return Promise.reject(new Error(`unexpected ${url}`));
    });

    renderProjectRoute("/projects/history");

    expect((await screen.findAllByText("本地AI获客顾问")).length).toBeGreaterThan(0);
    expect(screen.getAllByRole("link", { name: "查看结果" })[0]).toHaveAttribute("href", "/projects/matches/99/results");
    expect(await screen.findByText("暂无收藏结果")).toBeInTheDocument();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
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

  it("renders paid sample overlay state from the membership API", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/membership/plans") return Promise.resolve(new Response(JSON.stringify({ plans: [{ id: 2, code: "pro", name: "会员版", price_cents: 6900, billing_cycle: "month", features: ["完整项目拆解"], quotas: [], recommended: true }] }), { status: 200 }));
      if (url === "/api/v1/projects/matches") return Promise.resolve(new Response(JSON.stringify({ matches: [] }), { status: 200 }));
      if (url === "/api/v1/projects/favorites") return Promise.resolve(new Response(JSON.stringify({ favorites: [] }), { status: 200 }));
      return Promise.reject(new Error(`unexpected ${url}`));
    });
    renderProjectRoute("/projects/results/paywall");

    expect(screen.getByRole("heading", { name: "解锁完整拆解" })).toBeInTheDocument();
    expect(await screen.findByText("会员版")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "立即解锁" })).toHaveAttribute("href", "/membership/upgrade");
  });

  it("shows a disabled paywall action when membership plans fail to load", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/membership/plans") return Promise.resolve(new Response(JSON.stringify({ message: "plans unavailable" }), { status: 503 }));
      if (url === "/api/v1/projects/matches") return Promise.resolve(new Response(JSON.stringify({ matches: [] }), { status: 200 }));
      return Promise.reject(new Error(`unexpected ${url}`));
    });

    renderProjectRoute("/projects/results/paywall");

    expect(await screen.findByText("会员方案暂不可用，请稍后重试")).toHaveAttribute("role", "alert");
    expect(screen.getByRole("button", { name: "方案暂不可用" })).toBeDisabled();
    expect(screen.queryByRole("link", { name: "立即解锁" })).not.toBeInTheDocument();
  });

  it("redirects the legacy detail route to opportunity exploration", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({ opportunities: [] }), { status: 200 }));
    renderProjectRoute("/projects/detail");

    expect(await screen.findByRole("heading", { name: "机会探索" })).toBeInTheDocument();
    expect(screen.queryByText(/当前地址没有关联项目记录/)).not.toBeInTheDocument();
  });

  it("renders opportunity sections with traceable case evidence", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      if (String(input) === "/api/v1/projects/opportunities/ai-sales") return Promise.resolve(new Response(JSON.stringify({
        id: 42, slug: "ai-sales", title: "AI销售顾问", summary: "销售流程试点", industry: "企业服务", tags: ["B端"], budget_band: "1万", difficulty: "中等", resource_requirements: [],
        sections: [{ title: "验证路径", body: "先验证单一销售环节", items: ["记录人工基线"] }]
      }), { status: 200 }));
      if (String(input) === "/api/v1/projects/cases?opportunity_slug=ai-sales") return Promise.resolve(new Response(JSON.stringify({ cases: [{
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

  it("keeps opportunity detail visible when the cases API fails", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/projects/opportunities/ai-sales") return Promise.resolve(new Response(JSON.stringify({
        id: 42,
        slug: "ai-sales",
        title: "AI销售顾问",
        summary: "销售流程试点",
        industry: "企业服务",
        tags: ["B端"],
        budget_band: "1万",
        difficulty: "中等",
        resource_requirements: [],
        sections: [{ title: "验证路径", body: "先验证单一销售环节", items: ["记录人工基线"] }]
      }), { status: 200 }));
      if (url === "/api/v1/projects/cases?opportunity_slug=ai-sales") return Promise.resolve(new Response(JSON.stringify({ message: "cases unavailable" }), { status: 503 }));
      return Promise.reject(new Error(`unexpected ${url}`));
    });

    renderProjectRoute("/projects/opportunities/ai-sales");

    expect(await screen.findByRole("heading", { name: "AI销售顾问" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "验证路径" })).toBeInTheDocument();
    expect(screen.getByText("先验证单一销售环节")).toBeInTheDocument();
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/projects/cases?opportunity_slug=ai-sales",
      expect.objectContaining({ method: "GET" })
    ));
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("opens a directly addressable detail tab backed by API sections", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      if (String(input) === "/api/v1/projects/opportunities/ai-short-video-studio") return Promise.resolve(new Response(JSON.stringify({
        id: 42,
        slug: "ai-short-video-studio",
        title: "AI短视频脚本工作室",
        summary: "短视频脚本服务",
        industry: "内容服务",
        tags: ["轻资产"],
        budget_band: "0.8-3万元",
        difficulty: "中等",
        resource_requirements: [],
        sections: [
          { title: "成功路径", body: "先完成首个付费验证", items: ["选择一个细分行业"] },
          { title: "优劣势", body: "启动成本低，但需要建立差异化。", items: ["优势：交付快", "短板：同质化竞争"] }
        ]
      }), { status: 200 }));
      if (String(input) === "/api/v1/projects/cases?opportunity_slug=ai-short-video-studio") return Promise.resolve(new Response(JSON.stringify({ cases: [] }), { status: 200 }));
      return Promise.reject(new Error(`unexpected ${String(input)}`));
    });

    renderProjectRoute("/projects/opportunities/ai-short-video-studio?section=swot");

    expect(await screen.findByRole("heading", { name: "AI短视频脚本工作室" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "优劣势" })).toHaveClass("active");
    expect(screen.getByText("启动成本低，但需要建立差异化。")).toBeInTheDocument();
    expect(screen.queryByText("先完成首个付费验证")).not.toBeInTheDocument();
  });

  it("renders the project diagnosis as a closeable route state", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      if (String(input) === "/api/v1/projects/opportunities/ai-short-video-studio") return Promise.resolve(new Response(JSON.stringify({
        id: 42, slug: "ai-short-video-studio", title: "AI短视频脚本工作室", summary: "短视频脚本服务",
        industry: "内容服务", tags: [], budget_band: "0.8-3万元", difficulty: "中等", resource_requirements: [], sections: []
      }), { status: 200 }));
      if (String(input) === "/api/v1/projects/cases?opportunity_slug=ai-short-video-studio") return Promise.resolve(new Response(JSON.stringify({ cases: [] }), { status: 200 }));
      return Promise.reject(new Error(`unexpected ${String(input)}`));
    });

    renderProjectRoute("/projects/opportunities/ai-short-video-studio/diagnosis");

    expect(screen.getByRole("dialog", { name: "诊断我能否做这个项目？" })).toBeInTheDocument();
    expect(screen.getByText("按实际填写进度")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "关闭项目诊断" })).toHaveAttribute("href", "/projects/opportunities/ai-short-video-studio");
    expect(screen.getByRole("button", { name: "开始诊断" })).toBeInTheDocument();
  });

  it("creates a persisted project comparison", async () => {
    const createObjectURL = vi.fn(() => "blob:project-comparison");
    const revokeObjectURL = vi.fn();
    const clickSpy = vi.spyOn(HTMLAnchorElement.prototype, "click").mockImplementation(() => {});
    Object.defineProperty(URL, "createObjectURL", { configurable: true, value: createObjectURL });
    Object.defineProperty(URL, "revokeObjectURL", { configurable: true, value: revokeObjectURL });
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      const items = [
        { id: 1, slug: "ai-sales", title: "AI销售", summary: "销售流程", industry: "企业服务", tags: [], budget_band: "1万", difficulty: "中等", resource_requirements: [] },
        { id: 2, slug: "ai-content", title: "AI内容", summary: "内容生产", industry: "内容", tags: [], budget_band: "5000", difficulty: "低", resource_requirements: [] }
      ];
      if (url === "/api/v1/projects/opportunities") return Promise.resolve(new Response(JSON.stringify({ opportunities: items }), { status: 200 }));
      if (url === "/api/v1/projects/comparisons" && init?.method === "POST") return Promise.resolve(new Response(JSON.stringify({ id: 61, items }), { status: 200 }));
      if (url === "/api/v1/projects/exports" && init?.method === "POST") return Promise.resolve(new Response(JSON.stringify({ id: 72, status: "ready", download_url: "/api/v1/projects/exports/72/download" }), { status: 200 }));
      if (url === "/api/v1/projects/exports/72/download" && init?.method === "GET") return Promise.resolve(new Response(JSON.stringify({ source_type: "comparison", source_id: 61 }), { status: 200 }));
      return Promise.reject(new Error(`unexpected ${url}`));
    });
    renderProjectRoute("/projects/compare");

    expect(screen.getByRole("heading", { name: "项目对比" })).toBeInTheDocument();
    fireEvent.click(await screen.findByRole("checkbox", { name: "AI销售" }));
    fireEvent.click(screen.getByRole("checkbox", { name: "AI内容" }));
    fireEvent.click(screen.getByRole("button", { name: "开始对比" }));
    expect(await screen.findByRole("heading", { name: "AI销售" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "AI内容" })).toBeInTheDocument();
    await waitFor(() => expect(fetchMock.mock.calls.filter(([input, init]) =>
      String(input) === "/api/v1/projects/comparisons" && init?.method === "POST"
    )).toHaveLength(1));
    fireEvent.click(screen.getByRole("button", { name: "导出对比" }));
    fireEvent.click(await screen.findByRole("button", { name: "下载对比报告" }));
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/projects/exports/72/download",
      expect.objectContaining({ method: "GET", headers: expect.objectContaining({ Authorization: "Bearer access-token" }) })
    ));
    expect(createObjectURL).toHaveBeenCalled();
    expect(clickSpy).toHaveBeenCalled();
    expect(revokeObjectURL).toHaveBeenCalledWith("blob:project-comparison");
  });

  it("creates and downloads an authenticated match export", async () => {
    const createObjectURL = vi.fn(() => "blob:project-match");
    const revokeObjectURL = vi.fn();
    const clickSpy = vi.spyOn(HTMLAnchorElement.prototype, "click").mockImplementation(() => {});
    Object.defineProperty(URL, "createObjectURL", { configurable: true, value: createObjectURL });
    Object.defineProperty(URL, "revokeObjectURL", { configurable: true, value: revokeObjectURL });
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/projects/opportunities") return Promise.resolve(new Response(JSON.stringify({ opportunities: [] }), { status: 200 }));
      if (url === "/api/v1/projects/matches") return Promise.resolve(new Response(JSON.stringify({ matches: [{ id: 99, status: "completed", intent: "AI项目", result: { status: "completed", projects: [] } }] }), { status: 200 }));
      if (url === "/api/v1/projects/exports" && init?.method === "POST") return Promise.resolve(new Response(JSON.stringify({ id: 71, status: "ready", download_url: "/api/v1/projects/exports/71/download" }), { status: 200 }));
      if (url === "/api/v1/projects/exports/71/download" && init?.method === "GET") return Promise.resolve(new Response(JSON.stringify({ source_type: "match", source_id: 99 }), { status: 200 }));
      return Promise.reject(new Error(`unexpected ${url}`));
    });
    renderProjectRoute("/projects/export");

    expect(screen.getByRole("heading", { name: "导出匹配报告" })).toBeInTheDocument();
    const exportButton = await screen.findByRole("button", { name: "确认导出" });
    await waitFor(() => expect(exportButton).toBeEnabled());
    fireEvent.click(exportButton);
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/projects/exports", expect.objectContaining({ method: "POST" })));
    fireEvent.click(await screen.findByRole("button", { name: "下载已生成的 JSON 报告" }));
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/projects/exports/71/download",
      expect.objectContaining({ method: "GET", headers: expect.objectContaining({ Authorization: "Bearer access-token" }) })
    ));
    expect(createObjectURL).toHaveBeenCalled();
    expect(clickSpy).toHaveBeenCalled();
    expect(revokeObjectURL).toHaveBeenCalledWith("blob:project-match");
  });

  it("rejects an invalid match id before creating an export", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/projects/matches") return Promise.resolve(new Response(JSON.stringify({ matches: [] }), { status: 200 }));
      return Promise.reject(new Error(`unexpected ${url}`));
    });

    renderProjectRoute("/projects/matches/not-a-number/export");

    expect(await screen.findByText("匹配记录不存在")).toHaveAttribute("role", "alert");
    const exportButton = screen.getByRole("button", { name: "确认导出" });
    expect(exportButton).toBeDisabled();
    fireEvent.click(exportButton);
    expect(fetchMock.mock.calls.filter(([input, init]) =>
      String(input) === "/api/v1/projects/exports" && init?.method === "POST"
    )).toHaveLength(0);
  });

  it("rejects an incomplete match before creating an export", async () => {
    const session = {
      id: 77,
      user_id: 7,
      intent: "等待补充的项目需求",
      status: "pending",
      created_at: "2026-07-17T08:00:00Z",
      updated_at: "2026-07-17T08:00:00Z"
    };
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/projects/matches/77") return Promise.resolve(new Response(JSON.stringify(session), { status: 200 }));
      if (url === "/api/v1/projects/favorites") return Promise.resolve(new Response(JSON.stringify({ favorites: [] }), { status: 200 }));
      return Promise.reject(new Error(`unexpected ${url}`));
    });

    renderProjectRoute("/projects/matches/77/export");

    expect(await screen.findByText("该匹配记录尚未完成，暂时不能导出")).toHaveAttribute("role", "alert");
    const exportButton = screen.getByRole("button", { name: "确认导出" });
    expect(exportButton).toBeDisabled();
    fireEvent.click(exportButton);
    expect(fetchMock.mock.calls.filter(([input, init]) =>
      String(input) === "/api/v1/projects/exports" && init?.method === "POST"
    )).toHaveLength(0);
  });
});
