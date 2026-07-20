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

  it("renders the curated reference catalog when the API is empty", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ tools: [] }), { status: 200 })
    );
    renderPage();

    expect(screen.getByRole("heading", { name: "工具箱" })).toBeInTheDocument();
    expect(screen.getByLabelText("搜索工具")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "本周热门工具" })).toBeInTheDocument();
    expect(await screen.findByText("Notion AI")).toBeInTheDocument();
    expect(screen.getByText("Midjourney")).toBeInTheDocument();
    expect(screen.getByLabelText("智活 Copilot 工具助手")).toBeInTheDocument();
    expect(screen.getByLabelText("产品侧边导航")).toBeInTheDocument();
  });

  it("renders the full four-column reference catalog without Copilot", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ tools: [] }), { status: 200 })
    );
    renderPage("all");

    expect(await screen.findByText("Perplexity")).toBeInTheDocument();
    expect(screen.getByText("Claude")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "平台" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "按热度排序" })).toBeInTheDocument();
    expect(screen.queryByLabelText("智活 Copilot 工具助手")).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: "打开智活 Copilot" })).toHaveAttribute("href", "/tools");
  });

  it("links the library Copilot collapse control to the complete catalog", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ tools: [] }), { status: 200 })
    );
    renderPage();

    expect(await screen.findByText("Notion AI")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "收起 Copilot 并查看完整工具箱" })).toHaveAttribute("href", "/tools/all");
    expect(screen.getByRole("link", { name: "打开智活 Copilot" })).toHaveAttribute("href", "/copilot");
  });

  it("refetches the real catalog from the category tabs", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ tools: [] }), { status: 200 })
    );
    renderPage();

    expect(await screen.findByText("Notion AI")).toBeInTheDocument();
    const drawingCategory = screen.getByRole("tab", { name: "绘图" });
    fireEvent.click(drawingCategory);

    expect(drawingCategory).toHaveClass("active");
    await waitFor(() => expect(fetchMock).toHaveBeenLastCalledWith(expect.stringContaining("category=%E7%BB%98%E5%9B%BE"), expect.any(Object)));
    expect(screen.getByText("Midjourney")).toBeInTheDocument();
    expect(screen.queryByText("Runway")).not.toBeInTheDocument();
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
    fireEvent.click(screen.getByRole("tab", { name: "营销" }));
    fireEvent.change(screen.getByLabelText("搜索工具"), { target: { value: "agent" } });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenLastCalledWith(
        expect.stringContaining("/api/v1/content/tools?"),
        expect.any(Object)
      );
    });
    const lastUrl = String(fetchMock.mock.calls.at(-1)?.[0]);
    expect(lastUrl).toContain("category=%E8%90%A5%E9%94%80");
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
      tools: [{ id: 9, slug: "deep-research-agent", name: "Deep Research Agent", description: "海报与市场研究", status: "published", category: "创业获客", tags: ["设计"], platforms: ["Web"], features: [], use_cases: ["社媒海报"], limitations: [], sort_weight: 10, created_at: "", updated_at: "" }]
    }), { status: 200 }));
    renderPage("recommend");

    expect(screen.getByRole("heading", { name: "工具推荐结果" })).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("工具匹配目标"), { target: { value: "获客" } });
    fireEvent.change(screen.getByLabelText("工具使用场景"), { target: { value: "社媒海报" } });
    fireEvent.click(screen.getByRole("button", { name: "换一换" }));

    expect(await screen.findByText("Deep Research Agent")).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/content/tools/recommendations", expect.objectContaining({ method: "POST" }));
  });

  it("renders the recommendation plan", () => {
    renderPage("plan");

    expect(screen.getByRole("heading", { name: /整套工具方案/ })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: /推荐执行流程/ })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "市场调研" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "执行建议" })).toBeInTheDocument();
    expect(screen.getByText("查看从市场调研到视频推广的工具方案")).toBeInTheDocument();
  });

  it("renders recommendation-specific Copilot prompts", () => {
    renderPage("recommend");

    expect(screen.getByText("如何用这三款工具做内容日历？")).toBeInTheDocument();
    expect(screen.getByText("帮我生成内容营销执行计划")).toBeInTheDocument();
    expect(screen.getByText("换一换")).toBeInTheDocument();
  });

  it("renders the Midjourney reference detail when no slug is supplied", async () => {
    vi.spyOn(globalThis, "fetch").mockRejectedValue(new Error("offline"));
    renderPage("detail");

    expect(await screen.findByRole("heading", { name: "Midjourney" })).toBeInTheDocument();
    expect(screen.getAllByText(/专业 AI 图像生成工具/).length).toBeGreaterThan(0);
    expect(screen.getByRole("heading", { name: "用它解决什么" })).toBeInTheDocument();
  });

  it("uses the Midjourney reference for a known API slug alias", async () => {
    vi.spyOn(globalThis, "fetch").mockRejectedValue(new Error("offline"));
    renderPage("detail", "/tools/detail?tool=midjourney-ai");

    expect(await screen.findByRole("heading", { name: "Midjourney" })).toBeInTheDocument();
    expect(screen.getAllByText(/专业 AI 图像生成工具/).length).toBeGreaterThan(0);
    expect(screen.queryByText(/暂时无法读取最新详情/)).not.toBeInTheDocument();
  });
});
