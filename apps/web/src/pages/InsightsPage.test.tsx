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
    expect(screen.getByLabelText("搜索资讯")).toBeInTheDocument();
    expect(screen.getByText("今日关注")).toBeInTheDocument();
    expect(await screen.findByText("暂无资讯数据")).toBeInTheDocument();
    expect(screen.getByText("接口当前没有返回资讯内容。你可以先使用右侧 AI 助手检索方向，或等内容入库后在这里查看文章列表。")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "让 AI 总结重点" })).toHaveAttribute("href", "/insights/file-analysis");
    expect(screen.getByRole("link", { name: "查看相关工具" })).toHaveAttribute("href", "/tools");
    expect(screen.queryByText("企业智能客服落地实践：从成本中心到增长引擎")).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: /查看详情/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "20" })).not.toBeInTheDocument();
    expect(screen.getByText("今日 AI 客服行业资讯摘要")).toBeInTheDocument();
  });

  it("renders empty article detail instead of static detail fallback", async () => {
    vi.spyOn(globalThis, "fetch").mockRejectedValue(new Error("offline"));
    renderPage("detail");

    expect(await screen.findByText("暂无资讯详情")).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: /企业智能客服进入规模化落地阶段/ })).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: "分析我能学到什么" })).toHaveAttribute("href", "/insights/file-analysis");
    expect(screen.getByText("您可以这样问（与资讯相关）")).toBeInTheDocument();
  });

  it("renders the file analysis view with references", () => {
    renderPage("fileAnalysis");

    expect(screen.getByText("为您分析如下：")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "出处引用（点击查看原文）" })).toBeInTheDocument();
    expect(screen.getByText("IDC")).toBeInTheDocument();
    expect(screen.getByLabelText("向咨询通 Copilot 提问")).toBeInTheDocument();
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
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/content/articles", expect.any(Object));
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
        published_at: "2026-07-02T10:00:00Z",
        created_at: "2026-07-02T10:00:00Z",
        updated_at: "2026-07-02T10:00:00Z"
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ slug: "ai-growth-playbook", bookmarked: true }), { status: 200 }));

    renderPage("detail", "/insights/detail?article=ai-growth-playbook");

    expect(await screen.findByRole("heading", { name: "AI获客增长手册" })).toBeInTheDocument();
    expect(screen.getByText("第一步，明确客户画像。")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "收藏资讯" }));

    await waitFor(() => expect(fetchMock).toHaveBeenLastCalledWith(
      "/api/v1/content/articles/ai-growth-playbook/bookmark",
      expect.objectContaining({ method: "POST" })
    ));
    expect(screen.getByRole("button", { name: "取消收藏资讯" })).toBeInTheDocument();
  });
});
