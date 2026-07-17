import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import InsightsPage from "./InsightsPage";

describe("InsightsPage", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  function renderPage(variant?: "list" | "detail" | "fileAnalysis", route = "/insights") {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-18T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "", account: "zhangjing", status: "active" }
    });

    render(
      <MemoryRouter initialEntries={[route]}>
        <InsightsPage variant={variant} />
      </MemoryRouter>
    );
  }

  it("renders the reference insights when the content API is empty", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ articles: [] }), { status: 200 })
    );
    renderPage();

    expect(screen.getByRole("heading", { name: "咨询通" })).toBeInTheDocument();
    expect(screen.getByLabelText("搜索资讯目录")).toBeInTheDocument();
    expect(screen.getByText("今日关注")).toBeInTheDocument();
    expect(await screen.findByText("企业智能客服落地实践：从成本中心到增长引擎")).toBeInTheDocument();
    expect(screen.getAllByRole("link", { name: "查看详情" })).toHaveLength(6);
    expect(screen.getByText("共 6 条")).toBeInTheDocument();
    expect(screen.getByText("第 1 / 1 页")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "下一页" })).not.toBeInTheDocument();
    expect(screen.getByText("嗨，张婧！", { exact: false })).toBeInTheDocument();
  });

  it("opens settings and collapses the Copilot panel", () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ articles: [] }), { status: 200 })
    );
    renderPage();

    const copilot = screen.getByRole("complementary", { name: "智活 Copilot 咨询助手" });
    const body = document.getElementById("insights-copilot-body");

    fireEvent.click(screen.getByRole("button", { name: "打开 Copilot 设置" }));

    expect(screen.getByRole("dialog", { name: "Copilot 设置" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /更多偏好设置/ })).toHaveAttribute("href", "/profile/preferences");

    fireEvent.click(screen.getByRole("checkbox", { name: "显示快捷建议" }));
    expect(screen.queryByRole("navigation", { name: "咨询助手快捷入口" })).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "收起 Copilot" }));

    expect(copilot).toHaveClass("is-collapsed");
    expect(body).toHaveAttribute("hidden");
    expect(screen.queryByRole("dialog", { name: "Copilot 设置" })).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "展开 Copilot" }));
    expect(copilot).not.toHaveClass("is-collapsed");
    expect(body).not.toHaveAttribute("hidden");
  });

  it("renders the reference detail when the content API is unavailable", async () => {
    vi.spyOn(globalThis, "fetch").mockRejectedValue(new Error("offline"));
    renderPage("detail");

    expect(await screen.findByRole("heading", { name: "企业智能客服进入规模化落地阶段：从效率工具走向增长引擎" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "市场背景" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /分析我能学到什么/ })).toHaveAttribute("href", "/insights/file-analysis?article=smart-customer-service-growth");
    expect(screen.getByText("您好，我是智活 Copilot。", { exact: false })).toBeInTheDocument();
  });

  it("asks a question against a published article and renders returned citations", async () => {
    const article = {
      id: 7, slug: "ai-growth-playbook", title: "AI获客增长手册", summary: "摘要", status: "published",
      tags: ["获客"], citations: [{ id: "source-1", label: "行业报告", source_name: "研究机构", source_url: "https://example.com/source" }],
      created_at: "2026-07-02T10:00:00Z", updated_at: "2026-07-02T10:00:00Z"
    };
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      if (String(input) === "/api/v1/content/insights/qa" && init?.method === "POST") {
        return Promise.resolve(new Response(JSON.stringify({
          answer: "市场仍在增长。", basis: "catalog_citations", assumptions: ["仅覆盖所选文章"], disclaimer: "仅基于已保存引用。",
          citations: [{ id: "ai-growth-playbook:source-1", article_slug: "ai-growth-playbook", label: "行业报告", source_name: "研究机构", source_url: "https://example.com/source", excerpt: "市场增长" }]
        }), { status: 200 }));
      }
      return Promise.resolve(new Response(JSON.stringify({ articles: [article] }), { status: 200 }));
    });
    renderPage("fileAnalysis");

    await screen.findByRole("option", { name: "AI获客增长手册" });
    fireEvent.change(screen.getByLabelText("资讯问答问题"), { target: { value: "市场趋势？" } });
    fireEvent.click(screen.getByRole("button", { name: "生成带引用回答" }));

    expect(await screen.findByText("市场仍在增长。")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "出处引用" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /研究机构/ })).toHaveAttribute("href", "https://example.com/source");
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/content/insights/qa", expect.objectContaining({ method: "POST" }));
  });

  it("loads article list from content API", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        articles: [
          {
            id: 7,
            slug: "ai-growth-playbook",
            title: "AI获客增长手册",
            summary: "整理低成本获客动作。",
            status: "published",
            published_at: "2026-07-02T10:00:00Z",
            created_at: "2026-07-02T10:00:00Z",
            updated_at: "2026-07-02T10:00:00Z"
          }
        ]
      }), { status: 200 })
    );

    renderPage();

    expect(await screen.findByText("AI获客增长手册")).toBeInTheDocument();
    expect(screen.getByText("整理低成本获客动作。")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "查看详情" })).toHaveAttribute("href", "/insights/detail?article=ai-growth-playbook");
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/content/articles?limit=20", expect.any(Object));
  });

  it("loads article detail and bookmarks it", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({
        id: 7,
        slug: "ai-growth-playbook",
        title: "AI获客增长手册",
        summary: "整理低成本获客动作。",
        body: "第一步，明确客户画像。",
        status: "published",
        source_name: "研究机构",
        source_url: "https://example.com/original",
        author: "研究团队",
        category: "获客案例",
        tags: ["AI获客"],
        citations: [{ id: "source-1", label: "行业报告", source_name: "研究机构", source_url: "https://example.com/source" }],
        published_at: "2026-07-02T10:00:00Z",
        created_at: "2026-07-02T10:00:00Z",
        updated_at: "2026-07-02T10:00:00Z"
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ slug: "ai-growth-playbook", bookmarked: true }), { status: 200 }));

    renderPage("detail", "/insights/detail?article=ai-growth-playbook");

    expect(await screen.findByRole("heading", { name: "AI获客增长手册" })).toBeInTheDocument();
    expect(screen.getByText("第一步，明确客户画像。")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "原始来源 ↗" })).toHaveAttribute("href", "https://example.com/original");
    expect(screen.getByRole("link", { name: /研究机构：行业报告/ })).toHaveAttribute("href", "https://example.com/source");
    fireEvent.click(screen.getByRole("button", { name: "收藏资讯" }));

    await waitFor(() => expect(fetchMock).toHaveBeenLastCalledWith(
      "/api/v1/content/articles/ai-growth-playbook/bookmark",
      expect.objectContaining({ method: "POST" })
    ));
    expect(screen.getByRole("button", { name: "取消收藏资讯" })).toBeInTheDocument();
  });
});
