import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
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

  it("renders the project market home from the V1.4 home contract", async () => {
    const featured = { id: 42, slug: "ai-sales", title: "AI销售顾问", summary: "销售流程试点", track: "企业服务", tags: ["B端"], budget_band: "1万", difficulty: "中等", resource_requirements: [] };
    const evidenceCase = { id: 81, title: "99dresses", result_summary: "虚拟货币增加了交易复杂度", type: "fail", primary_source_url: "https://www.loot-drop.io/database-view?id=1", source_count: 1 };
    vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      if (String(input) === "/api/v1/projects/home") return Promise.resolve(new Response(JSON.stringify({ hero: { title: "项目超市", subtitle: "发现机会", desc: "真实项目" }, quick_tags: [], entries: [], featured: [featured] }), { status: 200 }));
      if (String(input) === "/api/v1/project-cases?page_size=20") return Promise.resolve(new Response(JSON.stringify({ items: [evidenceCase], page: 1, page_size: 20, total: 1752 }), { status: 200 }));
      if (String(input) === "/api/v1/projects?page_size=20") return Promise.resolve(new Response(JSON.stringify({ items: [featured], page: 1, page_size: 20, total: 1 }), { status: 200 }));
      return Promise.reject(new Error(`unexpected ${String(input)}`));
    });
    renderProjectRoute("/projects");

    expect(screen.getByRole("heading", { name: "项目超市" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "发现下一个可落地机会" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "去匹配" })).toHaveAttribute("href", "/projects/match");
    expect(await screen.findByRole("heading", { name: "近期案例复盘" })).toBeInTheDocument();
    expect(screen.getByText("1,752")).toBeInTheDocument();
    expect(screen.queryByText("Loot Drop 数据源")).not.toBeInTheDocument();
    expect(within(screen.getByRole("region", { name: "真实商业案例" })).queryByRole("img")).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: "查看案例" })).toHaveAttribute("href", "/project-cases/81");
    expect(screen.getByRole("heading", { name: "精选机会" })).toBeInTheDocument();
    expect(await screen.findByRole("heading", { name: "AI销售顾问" })).toBeInTheDocument();
    expect(screen.getAllByRole("link", { name: "查看机会" }).find((link) => link.getAttribute("href") === "/projects/ai-sales")).toBeTruthy();
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

  it("uploads a parsed text file and supports a file-only match request", async () => {
    const uploaded = {
      id: 501,
      name: "项目需求.txt",
      mime_type: "text/plain",
      detected_mime: "text/plain",
      size_bytes: 24,
      sha256: "abc123",
      parse_status: "ready",
      expires_at: "2026-09-10T08:00:00Z",
      created_at: "2026-08-11T08:00:00Z",
      updated_at: "2026-08-11T08:00:00Z"
    };
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/project-match-files" && init?.method === "POST") {
        return Promise.resolve(new Response(JSON.stringify(uploaded), { status: 201 }));
      }
      if (url === "/api/v1/project-matches" && init?.method === "POST") {
        return Promise.resolve(new Response(JSON.stringify({ match_id: 99, status: "clarifying", completeness: 0.5, questions: [], revision: 1 }), { status: 200 }));
      }
      if (url === "/api/v1/project-matches/99") {
        return Promise.resolve(new Response(JSON.stringify({ match_id: 99, status: "clarifying", completeness: 0.5, questions: [], revision: 1, file_ids: [501] }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected ${url}`));
    });
    renderProjectRoute("/projects/match");

    const file = new File(["预算 3 万元，希望做线上服务"], "项目需求.txt", { type: "text/plain" });
    fireEvent.change(screen.getByLabelText("选择项目匹配资料"), { target: { files: [file] } });

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/project-match-files",
      expect.objectContaining({ method: "POST", body: expect.any(FormData) })
    ));
    expect(await screen.findByText("项目需求.txt")).toBeInTheDocument();
    expect(screen.getByText("解析完成")).toBeInTheDocument();
    const submit = screen.getByRole("button", { name: "提交给 AI 分析" });
    expect(submit).toBeEnabled();
    fireEvent.click(submit);

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/project-matches",
      expect.objectContaining({ method: "POST", body: JSON.stringify({ need: "", file_ids: [501] }), headers: expect.objectContaining({ "Idempotency-Key": expect.stringMatching(/^project-match-/) }) })
    ));
  });

  it("shows failed file state and supports retry and removal", async () => {
    const failed = {
      id: 502,
      name: "现场照片.png",
      mime_type: "image/png",
      detected_mime: "image/png",
      size_bytes: 12,
      sha256: "def456",
      parse_status: "failed",
      error_code: "ocr_unavailable",
      expires_at: "2026-09-10T08:00:00Z",
      created_at: "2026-08-11T08:00:00Z",
      updated_at: "2026-08-11T08:00:00Z"
    };
    const ready = { ...failed, parse_status: "ready", error_code: undefined, extracted_text: "线下门店照片说明" };
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/project-match-files" && init?.method === "POST") return Promise.resolve(new Response(JSON.stringify(failed), { status: 201 }));
      if (url === "/api/v1/project-match-files/502/retry" && init?.method === "POST") return Promise.resolve(new Response(JSON.stringify(ready), { status: 200 }));
      if (url === "/api/v1/project-match-files/502" && init?.method === "DELETE") return Promise.resolve(new Response(null, { status: 204 }));
      return Promise.reject(new Error(`unexpected ${url}`));
    });
    renderProjectRoute("/projects/match");

    const image = new File(["image bytes"], "现场照片.png", { type: "image/png" });
    fireEvent.change(screen.getByLabelText("选择项目匹配资料"), { target: { files: [image] } });

    expect(await screen.findByText("图片 OCR 暂不可用")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "提交给 AI 分析" })).toBeDisabled();
    fireEvent.click(screen.getByRole("button", { name: "重试" }));
    expect(await screen.findByText("解析完成")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "删除文件 现场照片.png" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/project-match-files/502", expect.objectContaining({ method: "DELETE" })));
    expect(screen.queryByText("现场照片.png")).not.toBeInTheDocument();
  });

  it("submits match request and renders API result", async () => {
    const project = { rank:1, opportunity_slug:"local-ai-sales-consulting", title:"本地AI获客顾问", score:91, tags:["B端服务","轻资产"], budget:"¥2,000 - ¥6,000", reasons:["客户需求明确","交付可标准化"], risk:"需要控制交付边界" };
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/project-matches" && init?.method === "POST") return Promise.resolve(new Response(JSON.stringify({ match_id:99, status:"ready", completeness:0.9, revision:1 }), { status:200 }));
      if (url === "/api/v1/project-matches/99") return Promise.resolve(new Response(JSON.stringify({ match_id:99, status:"ready", completeness:0.9, revision:1 }), { status:200 }));
      if (url === "/api/v1/project-matches/99/generate" && init?.method === "POST") return Promise.resolve(new Response(JSON.stringify({ match_id:99, status:"completed", attempt:1, progress_percent:100, current_step:"done", result:{ session_id:99, status:"completed", projects:[project], evidence:[], evidence_status:"sufficient" } }), { status:202 }));
      if (url === "/api/v1/projects/favorites") return Promise.resolve(new Response(JSON.stringify({ favorites:[] }), { status:200 }));
      return Promise.reject(new Error(`unexpected ${url}`));
    });
    renderProjectRoute("/projects/match");

    fireEvent.change(screen.getByLabelText("项目匹配需求"), {
      target: { value: "我想做一个本地B端AI获客服务，预算6000元" }
    });
    fireEvent.click(screen.getByRole("button", { name: "提交给 AI 分析" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/project-matches",
      expect.objectContaining({ method: "POST" })
    ));
    expect(await screen.findByRole("heading", { name: "本地AI获客顾问" })).toBeInTheDocument();
    expect(screen.getByText("91分")).toBeInTheDocument();
  });

  it("answers dynamic match questions and renders the persisted result", async () => {
    const project = { rank:1, opportunity_slug:"local-ai-sales-consulting", title:"AI销售顾问", score:90, tags:["B端"], budget:"1万", reasons:["经验匹配"], risk:"需验证获客" };
    const need = "预算2万元，每周20小时，一人公司，线上内容创作";
    let answered = false;
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/project-matches/99/answer" && init?.method === "POST") {
        answered = true;
        return Promise.resolve(new Response(JSON.stringify({ match_id:99, need, status:"ready", completeness:0.9, revision:2 }), { status:200 }));
      }
      if (url === "/api/v1/project-matches/99") {
        return Promise.resolve(new Response(JSON.stringify(answered
          ? { match_id:99, need, status:"completed", completeness:0.9, revision:2, generation:{ match_id:99, status:"completed", attempt:1, progress_percent:100, current_step:"done", result:{ session_id:99, status:"completed", projects:[project], evidence:[] } } }
          : { match_id:99, need, status:"clarifying", completeness:0.5, revision:1, questions:[{ id:"background", field:"background", type:"single", question:"你擅长什么？", options:["销售经验","内容创作","技术能力"], required:true }] }), { status:200 }));
      }
      if (url === "/api/v1/projects/favorites") return Promise.resolve(new Response(JSON.stringify({ favorites:[] }), { status:200 }));
      return Promise.reject(new Error(`unexpected ${url}`));
    });
    renderProjectRoute("/projects/matches/99/questions");
    expect(await screen.findByText("预算2万元")).toBeInTheDocument();
    expect(screen.getByText("一人公司")).toBeInTheDocument();
    expect(screen.getByText("线上优先")).toBeInTheDocument();
    fireEvent.click(await screen.findByRole("button", { name:"销售经验" }));
    fireEvent.click(screen.getByRole("button", { name:"生成匹配结果" }));
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/project-matches/99/answer", expect.objectContaining({ method:"POST", body:JSON.stringify({ revision:1, answers:[{ question_id:"background", field:"background", value:"销售经验" }] }) })));
    expect(await screen.findByRole("heading", { name:"AI销售顾问" })).toBeInTheDocument();
  });

  it("syncs a persisted match result to task center", async () => {
    const project = { rank:1, opportunity_slug:"local-ai-sales-consulting", title:"AI销售顾问", score:90, tags:["B端"], budget:"1万", reasons:["经验匹配"], risk:"需验证" };
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      if (String(input) === "/api/v1/project-matches/99") return Promise.resolve(new Response(JSON.stringify({ match_id:99, status:"completed", completeness:0.9, revision:2, generation:{ match_id:99, status:"completed", attempt:1, progress_percent:100, current_step:"done", result:{ session_id:99, status:"completed", projects:[project], evidence:[] } } }), { status:200 }));
      if (String(input) === "/api/v1/projects/favorites") return Promise.resolve(new Response(JSON.stringify({ favorites:[] }), { status:200 }));
      if (String(input) === "/api/v1/tasks/generate" && init?.method === "POST") return Promise.resolve(new Response(JSON.stringify({ tasks:[{ id:1 }] }), { status:200 }));
      return Promise.reject(new Error("unexpected"));
    });
    renderProjectRoute("/projects/matches/99/results");
    fireEvent.click(await screen.findByRole("button", { name:"生成落地任务" }));
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/tasks/generate", expect.objectContaining({ method:"POST" })));
    expect(await screen.findByText("已创建 1 个项目任务")).toBeInTheDocument();
  });

  it("restores running generation after refresh and supports cancellation", async () => {
    let canceled = false;
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/project-matches/99" && init?.method === "GET") {
        return Promise.resolve(new Response(JSON.stringify({
          match_id: 99,
          status: canceled ? "canceled" : "running",
          completeness: 0.9,
          revision: 2,
          generation: { match_id: 99, status: canceled ? "canceled" : "running", attempt: 1, progress_percent: canceled ? 100 : 70, current_step: canceled ? "canceled" : "merging" }
        }), { status: 200 }));
      }
      if (url === "/api/v1/project-matches/99/stream") return Promise.resolve(new Response("", { status: 200, headers: { "Content-Type": "text/event-stream" } }));
      if (url === "/api/v1/project-matches/99/cancel" && init?.method === "POST") {
        canceled = true;
        return Promise.resolve(new Response(JSON.stringify({ match_id:99, status:"canceled", attempt:1, progress_percent:100, current_step:"canceled" }), { status:200 }));
      }
      if (url === "/api/v1/projects/favorites") return Promise.resolve(new Response(JSON.stringify({ favorites:[] }), { status:200 }));
      return Promise.reject(new Error(`unexpected ${url}`));
    });

    renderProjectRoute("/projects/matches/99/results");

    expect(await screen.findByText("整理证据")).toBeInTheDocument();
    expect(screen.getByText("70%")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name:"取消生成" }));
    expect(await screen.findByText("本次生成已取消")).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/project-matches/99/cancel", expect.objectContaining({ method:"POST" }));
  });

  it("retries a failed durable generation", async () => {
    let retried = false;
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/project-matches/99" && init?.method === "GET") return Promise.resolve(new Response(JSON.stringify({
        match_id:99,
        status:retried ? "queued" : "failed",
        completeness:0.9,
        revision:2,
        generation:{ match_id:99, status:retried ? "queued" : "failed", attempt:retried ? 2 : 1, progress_percent:retried ? 0 : 100, current_step:retried ? "queued" : "error", error_code:retried ? "" : "insufficient_evidence" }
      }), { status:200 }));
      if (url === "/api/v1/project-matches/99/generate" && init?.method === "POST") {
        retried = true;
        return Promise.resolve(new Response(JSON.stringify({ match_id:99, status:"queued", attempt:2, progress_percent:0, current_step:"queued" }), { status:202 }));
      }
      if (url === "/api/v1/project-matches/99/stream") return Promise.resolve(new Response("", { status:200, headers:{ "Content-Type":"text/event-stream" } }));
      if (url === "/api/v1/projects/favorites") return Promise.resolve(new Response(JSON.stringify({ favorites:[] }), { status:200 }));
      return Promise.reject(new Error(`unexpected ${url}`));
    });

    renderProjectRoute("/projects/matches/99/results");

    expect(await screen.findByText("生成未完成")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name:"重试生成" }));
    expect(await screen.findByText("等待开始")).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/project-matches/99/generate", expect.objectContaining({ method:"POST" }));
  });

  it("renders opportunity exploration from API", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({ items: [{
      id: 42,
      slug: "ai-sales-consulting",
      title: "AI销售顾问",
      summary: "为中小企业搭建销售自动化流程",
      industry: "企业服务",
      tags: ["轻资产", "B端服务"],
      budget_band: "1-3万",
      difficulty: "中等",
      resource_requirements: ["销售经验"]
    }], page: 1, page_size: 8, total: 1 }), { status: 200 }));
    renderProjectRoute("/projects/explore");

    expect(screen.getByRole("heading", { name: "机会探索" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "高潜力机会" })).toBeInTheDocument();
    expect(await screen.findByRole("heading", { name: "AI销售顾问" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "查看机会" })).toHaveAttribute("href", "/projects/ai-sales-consulting");
    expect(screen.getByText("案例重建").closest("article")?.querySelector("img")).toBeNull();
    expect(screen.getByRole("navigation", { name: "项目机会分页" })).toBeInTheDocument();
  });

  it("keeps opportunity cards visible while animating to the next page", async () => {
    const firstPage = {
      id: 41, slug: "first-opportunity", title: "第一页机会", summary: "第一页项目摘要", track: "企业服务",
      tags: ["轻资产"], budget_band: "1万", difficulty: "中等", resource_requirements: []
    };
    const secondPage = {
      id: 42, slug: "second-opportunity", title: "第二页机会", summary: "第二页项目摘要", track: "企业服务",
      tags: ["可复制"], budget_band: "2万", difficulty: "中等", resource_requirements: []
    };
    let resolveSecondPage: ((response: Response) => void) | undefined;
    vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      const url = String(input);
      if (url.includes("/api/v1/projects?") && url.includes("page=2")) {
        return new Promise<Response>((resolve) => { resolveSecondPage = resolve; });
      }
      if (url.includes("/api/v1/projects?")) {
        return Promise.resolve(new Response(JSON.stringify({ items: [firstPage], page: 1, page_size: 8, total: 16 }), { status: 200 }));
      }
      if (url === "/api/v1/projects/project-favorites") return Promise.resolve(new Response(JSON.stringify({ favorites: [] }), { status: 200 }));
      if (url === "/api/v1/projects/compare") return Promise.resolve(new Response(JSON.stringify({ items: [] }), { status: 200 }));
      return Promise.reject(new Error(`unexpected ${url}`));
    });
    renderProjectRoute("/projects/explore");

    expect(await screen.findByRole("heading", { name: "第一页机会" })).toBeInTheDocument();
    const grid = screen.getByTestId("opportunity-grid");
    fireEvent.click(screen.getByRole("button", { name: "下一页" }));

    await waitFor(() => expect(grid).toHaveClass("is-switching", "is-entering-forward"));
    expect(screen.getByRole("heading", { name: "第一页机会" })).toBeInTheDocument();
    expect(screen.queryByText("正在搜索项目机会…")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "下一页" })).toBeDisabled();

    resolveSecondPage?.(new Response(JSON.stringify({ items: [secondPage], page: 2, page_size: 8, total: 16 }), { status: 200 }));

    expect(await screen.findByRole("heading", { name: "第二页机会" })).toBeInTheDocument();
    await waitFor(() => expect(grid).not.toHaveClass("is-switching"));
    expect(grid).toHaveClass("is-entering-forward");
  });

  it("persists a favorite directly from a catalog card", async () => {
    const project = {
      id: 42, slug: "ai-sales", title: "AI销售顾问", summary: "销售流程试点", track: "企业服务",
      tags: ["B端"], budget_band: "1万", difficulty: "中等", resource_requirements: []
    };
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url.startsWith("/api/v1/projects?") && init?.method === "GET") return Promise.resolve(new Response(JSON.stringify({ items: [project], page: 1, page_size: 8, total: 1 }), { status: 200 }));
      if (url === "/api/v1/projects/project-favorites") return Promise.resolve(new Response(JSON.stringify({ favorites: [] }), { status: 200 }));
      if (url === "/api/v1/projects/compare") return Promise.resolve(new Response(JSON.stringify({ items: [] }), { status: 200 }));
      if (url === "/api/v1/projects/ai-sales/favorite" && init?.method === "POST") return Promise.resolve(new Response(JSON.stringify({ project_id: 42, slug: "ai-sales", title: "AI销售顾问" }), { status: 200 }));
      return Promise.reject(new Error(`unexpected ${url}`));
    });

    renderProjectRoute("/projects/explore");

    fireEvent.click(await screen.findByRole("button", { name: "收藏 AI销售顾问" }));
    expect(screen.getByRole("button", { name: "取消收藏 AI销售顾问" })).toBeInTheDocument();
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/projects/ai-sales/favorite", expect.objectContaining({ method: "POST" })));
  });

  it("renders the persisted compare bar at the document root", async () => {
    const project = {
      id: 42, slug: "ai-sales", title: "AI销售顾问", summary: "销售流程试点", track: "企业服务",
      tags: ["B端"], budget_band: "1万", difficulty: "中等", resource_requirements: []
    };
    vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      const url = String(input);
      if (url.startsWith("/api/v1/projects?")) return Promise.resolve(new Response(JSON.stringify({ items: [project], page: 1, page_size: 8, total: 1 }), { status: 200 }));
      if (url === "/api/v1/projects/project-favorites") return Promise.resolve(new Response(JSON.stringify({ favorites: [] }), { status: 200 }));
      if (url === "/api/v1/projects/compare") return Promise.resolve(new Response(JSON.stringify({ items: [{ project_id: 42, slug: "ai-sales", title: "AI销售顾问" }] }), { status: 200 }));
      return Promise.reject(new Error(`unexpected ${url}`));
    });

    renderProjectRoute("/projects/explore");

    const compareBar = await screen.findByRole("complementary", { name: "项目对比栏" });
    expect(compareBar.parentElement).toBe(document.body);
    expect(compareBar).toHaveTextContent("已加入对比 1/5");
  });

  it("shows the real empty catalog without demo records or fake totals", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation(() => Promise.resolve(new Response(JSON.stringify({ items: [], page: 1, page_size: 8, total: 0 }), { status: 200 })));
    renderProjectRoute("/projects/explore");

    expect(await screen.findByText("暂无符合条件的已发布项目机会")).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "查看机会" })).not.toBeInTheDocument();
    expect(screen.queryByText(/共 120 条/)).not.toBeInTheDocument();
  });

  it("submits opportunity search on click instead of filtering while typing", async () => {
    const catalog = [
      { id: 1, slug: "excel", title: "Excel自动化报表定制", summary: "企业报表", track: "企业服务", tags: [], budget_band: "1万", difficulty: "中等", resource_requirements: [] },
      { id: 2, slug: "ai-resume", title: "AI智能简历优化服务", summary: "简历优化", track: "AI应用", tags: ["AI应用"], budget_band: "5000", difficulty: "较低", resource_requirements: [] },
      { id: 3, slug: "ai-art", title: "AI绘画定制服务", summary: "视觉服务", track: "AI应用", tags: ["AI应用"], budget_band: "5000", difficulty: "较低", resource_requirements: [] }
    ];
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      const url = String(input);
      const items = url.includes("keyword=") ? catalog.slice(1) : catalog;
      return Promise.resolve(new Response(JSON.stringify({ items, page: 1, page_size: 8, total: items.length }), { status: 200 }));
    });
    renderProjectRoute("/projects/explore");

    expect(await screen.findByRole("heading", { name: "Excel自动化报表定制" })).toBeInTheDocument();
    const input = screen.getByRole("textbox", { name: "搜索机会赛道" });
    fireEvent.change(input, { target: { value: "AI应用" } });

    expect(screen.getByRole("heading", { name: "Excel自动化报表定制" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "搜索机会" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/projects?keyword=AI%E5%BA%94%E7%94%A8&sort=heat&page=1&page_size=8",
      expect.objectContaining({ method: "GET" })
    ));
    expect(await screen.findByText("共 2 条")).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "Excel自动化报表定制" })).not.toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "AI智能简历优化服务" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "AI绘画定制服务" })).toBeInTheDocument();
    await waitFor(() => {
      const analytics = fetchMock.mock.calls
        .filter(([input]) => String(input) === "/api/v1/analytics/events")
        .map(([, init]) => JSON.parse(String(init?.body)))
        .find((payload) => payload.event_name === "project_search");
      expect(analytics?.properties).toEqual({ keyword: "AI应用", result_count: 2 });
    });
  });

  it("shows a query-specific empty state after search form submission", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      const empty = String(input).includes("keyword=");
      return Promise.resolve(new Response(JSON.stringify({
        items: empty ? [] : [{ id: 1, slug: "ai-resume", title: "AI智能简历优化服务", summary: "简历优化", track: "AI应用", tags: [], budget_band: "5000", difficulty: "较低", resource_requirements: [] }],
        page: 1, page_size: 8, total: empty ? 0 : 1
      }), { status: 200 }));
    });
    renderProjectRoute("/projects/explore");

    await screen.findByRole("heading", { name: "AI智能简历优化服务" });
    const input = screen.getByRole("textbox", { name: "搜索机会赛道" });
    fireEvent.change(input, { target: { value: "不存在的机会" } });
    fireEvent.submit(input.closest("form")!);

    expect(await screen.findByText("未找到“不存在的机会”相关的已发布项目机会")).toBeInTheDocument();
    expect(screen.queryByRole("navigation", { name: "项目机会分页" })).not.toBeInTheDocument();
  });

  it("restores catalog filters and pagination from the URL", async () => {
    const item = { id: 42, slug: "ai-sales", title: "AI销售顾问", summary: "销售流程", track: "企业服务", tags: ["B端"], budget_band: "1万", difficulty: "中等", resource_requirements: ["销售经验"] };
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({ items: [item], page: 2, page_size: 8, total: 24 }), { status: 200 }));

    renderProjectRoute("/projects/explore?q=AI&track=%E4%BC%81%E4%B8%9A%E6%9C%8D%E5%8A%A1&budget=1%E4%B8%87&difficulty=%E4%B8%AD%E7%AD%89&resource=%E9%94%80%E5%94%AE%E7%BB%8F%E9%AA%8C&sort=latest&page=2");

    expect(await screen.findByRole("heading", { name: "AI销售顾问" })).toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: "搜索机会赛道" })).toHaveValue("AI");
    expect(screen.getByRole("combobox", { name: "按行业筛选" })).toHaveValue("企业服务");
    expect(screen.getByText("共 24 条")).toBeInTheDocument();
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/projects?keyword=AI&track=%E4%BC%81%E4%B8%9A%E6%9C%8D%E5%8A%A1&budget=1%E4%B8%87&difficulty=%E4%B8%AD%E7%AD%89&resource=%E9%94%80%E5%94%AE%E7%BB%8F%E9%AA%8C&sort=latest&page=2&page_size=8",
      expect.objectContaining({ method: "GET" })
    ));
  });

  it("retries a failed catalog request without losing the current route", async () => {
    let catalogAttempts = 0;
    vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      const url = String(input);
      if (url.includes("page_size=20")) return Promise.resolve(new Response(JSON.stringify({ items: [], page: 1, page_size: 20, total: 0 }), { status: 200 }));
      catalogAttempts += 1;
      if (catalogAttempts === 1) return Promise.resolve(new Response(JSON.stringify({ error: "service_unavailable" }), { status: 503 }));
      return Promise.resolve(new Response(JSON.stringify({ items: [], page: 1, page_size: 8, total: 0 }), { status: 200 }));
    });

    renderProjectRoute("/projects/explore?q=AI");

    expect(await screen.findByRole("alert")).toHaveTextContent("请求失败，请稍后重试");
    fireEvent.click(screen.getByRole("button", { name: "重新加载" }));
    expect(await screen.findByText("未找到“AI”相关的已发布项目机会")).toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: "搜索机会赛道" })).toHaveValue("AI");
  });

  it("renders real case library from API evidence", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({ items: [{
      id: 81, title: "AI销售试点", result_summary: "完成首轮流程验证", industry: "企业服务", scale: "solo",
      type: "success", primary_source_url: "https://www.loot-drop.io/database-view?id=81", source_count: 2, published_at: "2026-07-01T08:00:00Z"
    }], page: 1, page_size: 12, total: 1 }), { status: 200 }));
    renderProjectRoute("/projects/cases");

    expect(screen.getByRole("heading", { name: "真实案例库" })).toBeInTheDocument();
    expect(within(screen.getByRole("complementary", { name: "产品侧边导航" })).getByRole("link", { name: "项目超市" })).toHaveClass("active");
    expect(screen.getByRole("complementary", { name: "智活 Copilot" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "成功案例" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "赛道样本" })).toHaveAttribute("href", "/projects/explore");
    expect(screen.getByRole("combobox", { name: "案例商业模式" })).toBeInTheDocument();
    expect(screen.getByRole("combobox", { name: "案例阶段" })).toBeInTheDocument();
    const caseHeading = await screen.findByRole("heading", { name: "AI销售试点" });
    expect(caseHeading).toBeInTheDocument();
    expect(caseHeading.closest("article")?.querySelector("img")).toBeNull();
    expect(screen.getByRole("link", { name: "原文来源" })).toHaveAttribute("href", "https://www.loot-drop.io/database-view?id=81");
    expect(screen.getByText(/来源：公开资料/)).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "可学要点" })).toBeInTheDocument();
    expect(screen.getByTestId("featured-case-grid").querySelectorAll("article")).toHaveLength(1);
  });

  it("animates between featured case batches without replacing the grid with a loader", async () => {
    let resolveSecondPage: ((response: Response) => void) | undefined;
    vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/project-cases?page=2&page_size=12") {
        return new Promise<Response>((resolve) => { resolveSecondPage = resolve; });
      }
      return Promise.resolve(new Response(JSON.stringify({ items: [{
        id: 81, title: "第一批案例", result_summary: "第一批案例摘要", type: "success",
        primary_source_url: "https://example.com/first", source_count: 1
      }], page: 1, page_size: 12, total: 24 }), { status: 200 }));
    });
    renderProjectRoute("/project-cases");

    expect(await screen.findByRole("heading", { name: "第一批案例" })).toBeInTheDocument();
    const grid = screen.getByTestId("featured-case-grid");
    fireEvent.click(screen.getByRole("button", { name: "下一批案例" }));

    await waitFor(() => expect(grid).toHaveClass("is-switching", "is-entering-forward"));
    expect(screen.getByRole("heading", { name: "第一批案例" })).toBeInTheDocument();
    expect(screen.queryByText("正在读取证据案例…")).not.toBeInTheDocument();

    resolveSecondPage?.(new Response(JSON.stringify({ items: [{
      id: 93, title: "第二批案例", result_summary: "第二批案例摘要", type: "fail",
      primary_source_url: "https://example.com/second", source_count: 1
    }], page: 2, page_size: 12, total: 24 }), { status: 200 }));

    expect(await screen.findByRole("heading", { name: "第二批案例" })).toBeInTheDocument();
    await waitFor(() => expect(grid).not.toHaveClass("is-switching"));
    expect(grid).toHaveClass("is-entering-forward");
  });

  it("renders the case library at its public entry route", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({
      items: [], page: 1, page_size: 12, total: 0
    }), { status: 200 }));

    renderProjectRoute("/project-cases");

    expect(screen.getByRole("heading", { name: "真实案例库" })).toBeInTheDocument();
    expect(await screen.findByText("暂无符合条件的已发布案例")).toBeInTheDocument();
  });

  it("renders a directly addressable evidence case detail", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({
      id: 81, title: "AI销售试点", result_summary: "完成首轮流程验证", type: "success",
      primary_source_url: "https://example.com/case", source_count: 1, content_md: "# 复盘",
      facts: [{ field: "验证周期", value: "两周", source_refs: [101] }],
      analyses: [{ point: "先限定场景", detail: "降低验证成本", is_model_generated: true, source_refs: [101] }],
      sources: [{ id: 101, title: "企业公开复盘", url: "https://example.com/case", fetched_at: "2026-07-01T08:00:00Z", kind: "primary", is_primary: true, claim_fields: ["验证周期"] }]
    }), { status: 200 }));

    renderProjectRoute("/project-cases/81");

    expect(await screen.findByRole("heading", { name: "AI销售试点" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "案例概览" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "关键复盘" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "关键事实" })).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "公开来源" })).not.toBeInTheDocument();
    expect(screen.getByText("两周")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "返回案例库" })).toHaveAttribute("href", "/project-cases");
    expect(screen.queryByRole("link", { name: /企业公开复盘/ })).not.toBeInTheDocument();
    expect(screen.queryByText("1 条公开来源")).not.toBeInTheDocument();
    expect(screen.queryByText(/来源编号/)).not.toBeInTheDocument();
    expect(document.querySelector(".pm-catalog-hero-art")).toBeNull();
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

  it("restores and removes persisted project favorites from history", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/projects/matches") return Promise.resolve(new Response(JSON.stringify({ matches: [] }), { status: 200 }));
      if (url === "/api/v1/projects/favorites") return Promise.resolve(new Response(JSON.stringify({ favorites: [] }), { status: 200 }));
      if (url === "/api/v1/projects/project-favorites") return Promise.resolve(new Response(JSON.stringify({ favorites: [{ project_id: 42, slug: "ai-sales", title: "AI销售顾问", created_at: "2026-08-11T08:00:00Z" }] }), { status: 200 }));
      if (url === "/api/v1/projects/ai-sales/favorite" && init?.method === "DELETE") return Promise.resolve(new Response(null, { status: 204 }));
      return Promise.reject(new Error(`unexpected ${url}`));
    });

    renderProjectRoute("/projects/history");

    expect(await screen.findByRole("heading", { name: "收藏项目" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "查看项目" })).toHaveAttribute("href", "/projects/ai-sales");
    fireEvent.click(screen.getByRole("button", { name: "取消收藏 AI销售顾问" }));
    expect(await screen.findByText("暂无收藏项目")).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/projects/ai-sales/favorite", expect.objectContaining({ method: "DELETE" }));
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
      "/api/v1/project-matches/99",
      expect.objectContaining({ method: "GET" })
    ));
    expect(await screen.findByRole("heading", { name: "本地AI获客顾问", level: 1 })).toBeInTheDocument();
    expect(screen.getByText("匹配度 91分")).toBeInTheDocument();
    expect(screen.getByText("预算 ¥2,000 - ¥6,000")).toBeInTheDocument();
    expect(screen.getByText(/未引用外部证据/)).toBeInTheDocument();
  });

  it("does not show an unlock shell when the paywall is disabled", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/config") return Promise.resolve(new Response(JSON.stringify({ feature_paywall_enabled: false }), { status: 200 }));
      if (url === "/api/v1/projects/ai-sales") return Promise.resolve(new Response(JSON.stringify({
        id: 42,
        slug: "ai-sales",
        title: "AI销售顾问",
        summary: "销售流程试点",
        industry: "企业服务",
        tags: ["B端"],
        budget_band: "1万",
        difficulty: "中等",
        resource_requirements: [],
        sections: [{ key: "data", title: "当前数据", body: "当前数据正文", items: [] }]
      }), { status: 200 }));
      if (url === "/api/v1/project-cases?page_size=100") return Promise.resolve(new Response(JSON.stringify({ items: [], page: 1, page_size: 100, total: 0 }), { status: 200 }));
      if (url === "/api/v1/projects/project-favorites") return Promise.resolve(new Response(JSON.stringify({ favorites: [] }), { status: 200 }));
      if (url === "/api/v1/projects/compare-items") return Promise.resolve(new Response(JSON.stringify({ items: [] }), { status: 200 }));
      return Promise.reject(new Error(`unexpected ${url}`));
    });
    renderProjectRoute("/projects/ai-sales?section=data");

    await screen.findByRole("heading", { name: "当前数据" });
    expect(screen.queryByRole("link", { name: "解锁完整拆解" })).not.toBeInTheDocument();
    expect(screen.queryByRole("dialog", { name: "解锁完整拆解" })).not.toBeInTheDocument();
    expect(fetchMock.mock.calls.some(([input]) => String(input) === "/api/v1/membership/plans")).toBe(false);
  });

  it("redirects the legacy detail route to opportunity exploration", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({ items: [], page: 1, page_size: 8, total: 0 }), { status: 200 }));
    renderProjectRoute("/projects/detail");

    expect(await screen.findByRole("heading", { name: "机会探索" })).toBeInTheDocument();
    expect(screen.queryByText(/当前地址没有关联项目记录/)).not.toBeInTheDocument();
  });

  it("renders opportunity sections with traceable case evidence", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      if (String(input) === "/api/v1/projects/ai-sales") return Promise.resolve(new Response(JSON.stringify({
        id: 42, slug: "ai-sales", title: "AI销售顾问", summary: "销售流程试点", industry: "企业服务", tags: ["B端"], budget_band: "1万", difficulty: "中等", resource_requirements: [],
        sections: [{ title: "验证路径", body: "先验证单一销售环节", items: ["记录人工基线"] }]
      }), { status: 200 }));
      if (String(input) === "/api/v1/project-cases?page_size=100") return Promise.resolve(new Response(JSON.stringify({ items: [{
        id: 81, project_id: 42, title: "销售试点复盘", result_summary: "完成验证", type: "success", primary_source_url: "https://example.com/case", source_count: 1, verified_at: "2026-07-01T08:00:00Z"
      }], page: 1, page_size: 100, total: 1 }), { status: 200 }));
      return Promise.reject(new Error(`unexpected ${String(input)}`));
    });
    renderProjectRoute("/projects/ai-sales");

    expect(await screen.findByRole("heading", { name: "AI销售顾问" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "验证路径" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "来源与证据" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "查看首要来源" })).toHaveAttribute("href", "https://example.com/case");
  });

  it("keeps opportunity detail visible when the cases API fails", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/projects/ai-sales") return Promise.resolve(new Response(JSON.stringify({
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
      if (url === "/api/v1/project-cases?page_size=100") return Promise.resolve(new Response(JSON.stringify({ message: "cases unavailable" }), { status: 503 }));
      return Promise.reject(new Error(`unexpected ${url}`));
    });

    renderProjectRoute("/projects/ai-sales");

    expect(await screen.findByRole("heading", { name: "AI销售顾问" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "验证路径" })).toBeInTheDocument();
    expect(screen.getByText("先验证单一销售环节")).toBeInTheDocument();
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/project-cases?page_size=100",
      expect.objectContaining({ method: "GET" })
    ));
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("restores a project favorite after refresh and rolls back a failed removal", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/projects/ai-sales") return Promise.resolve(new Response(JSON.stringify({
        id: 42, slug: "ai-sales", title: "AI销售顾问", summary: "销售流程试点", track: "企业服务", tags: ["B端"], budget_band: "1万", difficulty: "中等", resource_requirements: [], sections: []
      }), { status: 200 }));
      if (url === "/api/v1/project-cases?page_size=100") return Promise.resolve(new Response(JSON.stringify({ items: [], page: 1, page_size: 100, total: 0 }), { status: 200 }));
      if (url === "/api/v1/projects/project-favorites") return Promise.resolve(new Response(JSON.stringify({ favorites: [{ project_id: 42, slug: "ai-sales", title: "AI销售顾问" }] }), { status: 200 }));
      if (url === "/api/v1/projects/compare") return Promise.resolve(new Response(JSON.stringify({ items: [] }), { status: 200 }));
      if (url === "/api/v1/projects/ai-sales/favorite" && init?.method === "DELETE") return Promise.resolve(new Response(JSON.stringify({ error: "internal_error" }), { status: 500 }));
      return Promise.reject(new Error(`unexpected ${url}`));
    });

    renderProjectRoute("/projects/ai-sales");

    const favoriteButton = await screen.findByRole("button", { name: "取消收藏项目" });
    fireEvent.click(favoriteButton);
    expect(screen.getByRole("button", { name: "收藏项目" })).toBeInTheDocument();
    expect(await screen.findByRole("alert")).toHaveTextContent("取消收藏失败");
    expect(screen.getByRole("button", { name: "取消收藏项目" })).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/projects/project-favorites", expect.objectContaining({ method: "GET" }));
  });

  it("restores the persisted five-item compare list and rolls back a sixth item", async () => {
    const items = Array.from({ length: 6 }, (_, index) => ({
      id: index + 1,
      slug: `project-${index + 1}`,
      title: `项目${index + 1}`,
      summary: `项目${index + 1}摘要`,
      track: "企业服务",
      tags: [],
      budget_band: "1万",
      difficulty: "中等",
      resource_requirements: []
    }));
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/projects?page_size=100") return Promise.resolve(new Response(JSON.stringify({ items, page: 1, page_size: 100, total: 6 }), { status: 200 }));
      if (url === "/api/v1/projects/compare") return Promise.resolve(new Response(JSON.stringify({ items: items.slice(0, 5).map((item) => ({ project_id: item.id, slug: item.slug, title: item.title })) }), { status: 200 }));
      if (url === "/api/v1/projects/compare-items" && init?.method === "POST") return Promise.resolve(new Response(JSON.stringify({ error: "compare_limit_reached", max: 5 }), { status: 409 }));
      return Promise.reject(new Error(`unexpected ${url}`));
    });

    renderProjectRoute("/projects/compare");

    expect(await screen.findByText("已保存 5/5 个项目")).toBeInTheDocument();
    const sixth = screen.getByRole("checkbox", { name: "项目6" });
    fireEvent.click(sixth);
    expect(sixth).toBeChecked();
    expect(await screen.findByRole("alert")).toHaveTextContent("最多只能同时对比 5 个项目");
    expect(sixth).not.toBeChecked();
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/projects/compare-items", expect.objectContaining({ method: "POST", body: JSON.stringify({ project_id: "project-6" }) }));
  });

  it("opens a directly addressable detail tab backed by API sections", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      if (String(input) === "/api/v1/projects/ai-short-video-studio") return Promise.resolve(new Response(JSON.stringify({
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
      if (String(input) === "/api/v1/project-cases?page_size=100") return Promise.resolve(new Response(JSON.stringify({ items: [], page: 1, page_size: 100, total: 0 }), { status: 200 }));
      return Promise.reject(new Error(`unexpected ${String(input)}`));
    });

    renderProjectRoute("/projects/ai-short-video-studio?section=swot");

    expect(await screen.findByRole("heading", { name: "AI短视频脚本工作室" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "优劣势" })).toHaveClass("active");
    expect(screen.getByText("启动成本低，但需要建立差异化。")).toBeInTheDocument();
    expect(screen.queryByText("先完成首个付费验证")).not.toBeInTheDocument();
    await waitFor(() => {
      const analytics = fetchMock.mock.calls
        .filter(([input]) => String(input) === "/api/v1/analytics/events")
        .map(([, init]) => JSON.parse(String(init?.body)))
        .find((payload) => payload.event_name === "project_detail_view");
      expect(analytics?.properties).toEqual({ project_id: 42, tab: "swot" });
    });
  });

  it("renders the project diagnosis as a closeable route state", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      if (String(input) === "/api/v1/projects/ai-short-video-studio") return Promise.resolve(new Response(JSON.stringify({
        id: 42, slug: "ai-short-video-studio", title: "AI短视频脚本工作室", summary: "短视频脚本服务",
        industry: "内容服务", tags: [], budget_band: "0.8-3万元", difficulty: "中等", resource_requirements: [], sections: []
      }), { status: 200 }));
      if (String(input) === "/api/v1/project-cases?page_size=100") return Promise.resolve(new Response(JSON.stringify({ items: [], page: 1, page_size: 100, total: 0 }), { status: 200 }));
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
      if (url === "/api/v1/projects?page_size=100" || url === "/api/v1/projects?page_size=20") return Promise.resolve(new Response(JSON.stringify({ items, page: 1, page_size: 100, total: items.length }), { status: 200 }));
      if (url === "/api/v1/projects/compare" && init?.method === "GET") return Promise.resolve(new Response(JSON.stringify({ items: [] }), { status: 200 }));
      if (url === "/api/v1/projects/compare-items" && init?.method === "POST") {
        const slug = JSON.parse(String(init.body)).project_id;
        const item = items.find((candidate) => candidate.slug === slug);
        return Promise.resolve(new Response(JSON.stringify({ project_id: item?.id, slug: item?.slug, title: item?.title }), { status: 200 }));
      }
      if (url === "/api/v1/projects/comparisons" && init?.method === "POST") return Promise.resolve(new Response(JSON.stringify({ id: 61, items }), { status: 200 }));
      if (url === "/api/v1/projects/exports" && init?.method === "POST") return Promise.resolve(new Response(JSON.stringify({ id: 72, status: "ready", format: "pdf", download_url: "/api/v1/projects/exports/72/download" }), { status: 202 }));
      if (url === "/api/v1/projects/exports/72/download" && init?.method === "GET") return Promise.resolve(new Response("%PDF-1.7\ncomparison", { status: 200, headers: { "Content-Type": "application/pdf" } }));
      return Promise.reject(new Error(`unexpected ${url}`));
    });
    renderProjectRoute("/projects/compare");

    expect(screen.getByRole("heading", { name: "项目对比" })).toBeInTheDocument();
    fireEvent.click(await screen.findByRole("checkbox", { name: "AI销售" }));
    fireEvent.click(screen.getByRole("checkbox", { name: "AI内容" }));
    const compareButton = screen.getByRole("button", { name: "开始对比" });
    await waitFor(() => expect(compareButton).toBeEnabled());
    fireEvent.click(compareButton);
    expect(await screen.findByRole("heading", { name: "AI销售" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "AI内容" })).toBeInTheDocument();
    await waitFor(() => expect(fetchMock.mock.calls.filter(([input, init]) =>
      String(input) === "/api/v1/projects/comparisons" && init?.method === "POST"
    )).toHaveLength(1));
    fireEvent.click(screen.getByRole("button", { name: "导出对比" }));
    fireEvent.click(await screen.findByRole("button", { name: "下载 PDF 对比报告" }));
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
      if (url === "/api/v1/project-matches/99") return Promise.resolve(new Response(JSON.stringify({
        match_id: 99,
        need: "AI项目",
        status: "completed",
        completeness: 0.9,
        revision: 2,
        generation: { match_id: 99, status: "completed", attempt: 1, progress_percent: 100, current_step: "done", result: { session_id: 99, status: "completed", projects: [], evidence: [] } }
      }), { status: 200 }));
      if (url === "/api/v1/projects/exports" && init?.method === "POST") return Promise.resolve(new Response(JSON.stringify({ id: 71, status: "ready", format: "pdf", download_url: "/api/v1/projects/exports/71/download" }), { status: 202 }));
      if (url === "/api/v1/projects/exports/71/download" && init?.method === "GET") return Promise.resolve(new Response("%PDF-1.7\nmatch", { status: 200, headers: { "Content-Type": "application/pdf" } }));
      return Promise.reject(new Error(`unexpected ${url}`));
    });
    renderProjectRoute("/projects/matches/99/export");

    expect(screen.getByRole("heading", { name: "导出匹配报告" })).toBeInTheDocument();
    const exportButton = await screen.findByRole("button", { name: "确认导出" });
    await waitFor(() => expect(exportButton).toBeEnabled());
    fireEvent.click(exportButton);
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/projects/exports", expect.objectContaining({ method: "POST" })));
    fireEvent.click(await screen.findByRole("button", { name: "下载 PDF 报告" }));
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
      match_id: 77,
      need: "等待补充的项目需求",
      status: "pending",
      completeness: 0.4,
      revision: 1,
      generation: { match_id: 77, status: "pending", attempt: 0, progress_percent: 0 }
    };
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/project-matches/77") return Promise.resolve(new Response(JSON.stringify(session), { status: 200 }));
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
