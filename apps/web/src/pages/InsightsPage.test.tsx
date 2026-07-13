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

  it("renders an empty insights list instead of static sample articles", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ articles: [] }), { status: 200 })
    );
    renderPage();

    expect(screen.getByRole("heading", { name: "咨询通" })).toBeInTheDocument();
    expect(screen.getByLabelText("搜索资讯目录")).toBeInTheDocument();
    expect(screen.getByText("最新发布")).toBeInTheDocument();
    expect(await screen.findByText("暂无资讯数据")).toBeInTheDocument();
    expect(screen.getByText("接口当前没有返回已发布资讯。请等待内容管理员完成入库和发布。")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "基于已发布资讯提问" })).toHaveAttribute("href", "/insights/file-analysis");
    expect(screen.getByRole("link", { name: "查看相关工具" })).toHaveAttribute("href", "/tools");
    expect(screen.queryByText("企业智能客服落地实践：从成本中心到增长引擎")).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: /查看详情/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "20" })).not.toBeInTheDocument();
    expect(screen.getByText("打开资讯详情可核对原始来源；进入“基于资讯提问”可获得带引用的回答。")).toBeInTheDocument();
  });

  it("renders empty article detail instead of static detail fallback", async () => {
    vi.spyOn(globalThis, "fetch").mockRejectedValue(new Error("offline"));
    renderPage("detail");

    expect(await screen.findByText("暂无资讯详情")).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: /企业智能客服进入规模化落地阶段/ })).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: "选择资讯后提问" })).toHaveAttribute("href", "/insights/file-analysis");
    expect(screen.getByText("打开资讯详情可核对原始来源；进入“基于资讯提问”可获得带引用的回答。")).toBeInTheDocument();
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

    expect(await screen.findAllByText("AI获客增长手册")).toHaveLength(2);
    expect(screen.getAllByText("整理低成本获客动作。")).toHaveLength(2);
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
