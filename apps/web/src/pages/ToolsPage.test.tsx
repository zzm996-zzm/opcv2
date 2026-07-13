import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import ToolsPage from "./ToolsPage";

describe("ToolsPage", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  function renderPage(variant?: "library" | "all" | "recommend" | "plan" | "detail", route = "/tools") {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-21T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张晨", phone: "", account: "zhangchen", status: "active" }
    });

    render(
      <MemoryRouter initialEntries={[route]}>
        <ToolsPage variant={variant} />
      </MemoryRouter>
    );
  }

  it("renders an explicit empty catalog without static tool defaults", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ tools: [] }), { status: 200 })
    );
    renderPage();

    expect(screen.getByRole("heading", { name: "工具箱" })).toBeInTheDocument();
    expect(screen.getByLabelText("搜索工具")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "常见工具场景" })).toBeInTheDocument();
    expect(await screen.findByText("没有匹配的工具，换个关键词或分类试试")).toBeInTheDocument();
    expect(screen.queryByText("Notion AI")).not.toBeInTheDocument();
    expect(screen.queryByText("Midjourney")).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: /开始匹配/ })).toHaveAttribute("href", "/tools/recommend");
  });

  it("renders the full catalog empty state when the API is empty", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ tools: [] }), { status: 200 })
    );
    renderPage("all");

    expect(await screen.findByText("没有匹配的工具，换个关键词或分类试试")).toBeInTheDocument();
    expect(screen.queryByText("Perplexity")).not.toBeInTheDocument();
    expect(screen.queryByText("Similarweb")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "平台⌄" })).toBeInTheDocument();
    expect(screen.queryByLabelText("智活 Copilot 工具助手")).not.toBeInTheDocument();
  });

  it("refetches the real catalog from the left category rail", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ tools: [] }), { status: 200 })
    );
    renderPage();

    expect(await screen.findByText("没有匹配的工具，换个关键词或分类试试")).toBeInTheDocument();
    const leadCategory = screen.getByRole("button", { name: "创业获客" });
    fireEvent.click(leadCategory);

    expect(leadCategory).toHaveClass("active");
    await waitFor(() => expect(fetchMock).toHaveBeenLastCalledWith(expect.stringContaining("category=%E5%88%9B%E4%B8%9A%E8%8E%B7%E5%AE%A2"), expect.any(Object)));
    expect(screen.queryByText("Canva AI")).not.toBeInTheDocument();
    expect(screen.queryByText("Apollo AI")).not.toBeInTheDocument();
    expect(screen.queryByText("Similarweb")).not.toBeInTheDocument();
  });

  it("loads public tools from content API", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        tools: [
          {
            id: 99,
            slug: "deep-research-agent",
            name: "Deep Research Agent",
            description: "自动整理行业资料和竞品线索。",
            url: "https://example.com",
            status: "published",
            created_at: "2026-06-25T12:00:00Z",
            updated_at: "2026-06-25T12:00:00Z"
          }
        ]
      }), { status: 200 })
    );
    renderPage("all");

    expect(await screen.findByText("Deep Research Agent")).toBeInTheDocument();
    expect(screen.getByText("自动整理行业资料和竞品线索。")).toBeInTheDocument();
  });

  it("passes tool filters to the content API", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ tools: [] }), { status: 200 })
    );
    renderPage("all");

    await waitFor(() => expect(fetchMock).toHaveBeenCalled());
    fireEvent.click(screen.getByRole("button", { name: "创业获客" }));
    fireEvent.click(screen.getByRole("tab", { name: "热门" }));
    fireEvent.change(screen.getByLabelText("搜索工具"), { target: { value: "agent" } });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenLastCalledWith(
        expect.stringContaining("/api/v1/content/tools?"),
        expect.any(Object)
      );
    });
    const lastUrl = String(fetchMock.mock.calls.at(-1)?.[0]);
    expect(lastUrl).toContain("category=%E5%88%9B%E4%B8%9A%E8%8E%B7%E5%AE%A2");
    expect(lastUrl).toContain("q=agent");
    expect(lastUrl).toContain("sort=hot");
  });

  it("favorites tools through the content API", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({
        tools: [
          {
            id: 99,
            slug: "deep-research-agent",
            name: "Deep Research Agent",
            description: "自动整理行业资料和竞品线索。",
            url: "https://example.com",
            status: "published",
            created_at: "2026-06-25T12:00:00Z",
            updated_at: "2026-06-25T12:00:00Z"
          }
        ]
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ slug: "deep-research-agent", favorited: true }), { status: 200 }));
    renderPage("all");

    expect(await screen.findByText("Deep Research Agent")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "收藏Deep Research Agent" }));

    await waitFor(() => expect(fetchMock).toHaveBeenLastCalledWith(
      "/api/v1/content/tools/deep-research-agent/favorite",
      expect.objectContaining({ method: "POST" })
    ));
    expect(screen.getByRole("button", { name: "取消收藏Deep Research Agent" })).toBeInTheDocument();
  });

  it("loads tool detail from content API", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        id: 99,
        slug: "deep-research-agent",
        name: "Deep Research Agent",
        description: "自动整理行业资料和竞品线索。",
        url: "https://example.com",
        status: "published",
        category: "数据分析",
        created_at: "2026-06-25T12:00:00Z",
        updated_at: "2026-06-25T12:00:00Z"
      }), { status: 200 })
    );
    renderPage("detail", "/tools/detail?tool=deep-research-agent");

    expect(await screen.findByRole("heading", { name: "Deep Research Agent" })).toBeInTheDocument();
    expect(screen.getAllByText("自动整理行业资料和竞品线索。").length).toBeGreaterThan(0);
    expect(screen.getByRole("link", { name: "访问官网" })).toHaveAttribute("href", "https://example.com");
  });

  it("submits catalog recommendation criteria and renders real matches", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({
      basis: "catalog_match",
      criteria: { goal: "获客", scenario: "社媒海报", limit: 6 },
      tools: [{ id: 9, slug: "canva-ai", name: "Canva AI", description: "海报设计", status: "published", category: "创业获客", tags: ["设计"], platforms: ["Web"], features: [], use_cases: ["社媒海报"], limitations: [], sort_weight: 10, created_at: "", updated_at: "" }]
    }), { status: 200 }));
    renderPage("recommend");

    expect(screen.getByRole("heading", { name: "工具推荐结果" })).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("工具匹配目标"), { target: { value: "获客" } });
    fireEvent.change(screen.getByLabelText("工具使用场景"), { target: { value: "社媒海报" } });
    fireEvent.click(screen.getByRole("button", { name: "匹配目录工具" }));

    expect(await screen.findByText("Canva AI")).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/content/tools/recommendations", expect.objectContaining({ method: "POST" }));
  });

  it("renders the recommendation plan", () => {
    renderPage("plan");

    expect(screen.getByRole("heading", { name: /整套工具方案/ })).toBeInTheDocument();
    expect(screen.getByText("暂无工具方案")).toBeInTheDocument();
    expect(screen.queryByText("推荐执行流程")).not.toBeInTheDocument();
    expect(screen.queryByText("市场调研")).not.toBeInTheDocument();
  });

  it("renders empty tool detail instead of static detail fallback", async () => {
    vi.spyOn(globalThis, "fetch").mockRejectedValue(new Error("offline"));
    renderPage("detail");

    expect(await screen.findByText("暂无工具详情")).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "Midjourney" })).not.toBeInTheDocument();
    expect(screen.queryByText("专业 AI 图像生成工具")).not.toBeInTheDocument();
  });
});
