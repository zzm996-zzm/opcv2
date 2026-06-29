import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import ToolsPage from "./ToolsPage";

describe("ToolsPage", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  function renderPage(variant?: "library" | "all" | "recommend" | "plan" | "detail") {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-21T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张晨", phone: "", account: "zhangchen", status: "active" }
    });

    render(
      <MemoryRouter>
        <ToolsPage variant={variant} />
      </MemoryRouter>
    );
  }

  it("renders the tool library with filters, hot scenarios and copilot", () => {
    renderPage();

    expect(screen.getByRole("heading", { name: "工具箱" })).toBeInTheDocument();
    expect(screen.getByLabelText("搜索工具")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: /本周热门工具/ })).toBeInTheDocument();
    expect(screen.getByText("Notion AI")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /推荐工具/ })).toHaveAttribute("href", "/tools/recommend");
  });

  it("renders the full tool library", () => {
    renderPage("all");

    expect(screen.getByText("Jasper")).toBeInTheDocument();
    expect(screen.getByText("Claude")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "平台⌄" })).toBeInTheDocument();
    expect(screen.queryByLabelText("智活 Copilot 工具助手")).not.toBeInTheDocument();
  });

  it("filters tools from the left category rail", () => {
    renderPage();

    fireEvent.click(screen.getByRole("button", { name: "视频剪辑" }));

    expect(screen.getByText("Runway")).toBeInTheDocument();
    expect(screen.queryByText("Notion AI")).not.toBeInTheDocument();
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

  it("renders tool recommendations", () => {
    renderPage("recommend");

    expect(screen.getByRole("heading", { name: "工具推荐结果" })).toBeInTheDocument();
    expect(screen.getByText("需求摘要")).toBeInTheDocument();
    expect(screen.getByText("ChatGPT")).toBeInTheDocument();
    expect(screen.getAllByRole("link", { name: /生成整套方案/ })[0]).toHaveAttribute("href", "/tools/recommendation-plan");
  });

  it("renders the recommendation plan", () => {
    renderPage("plan");

    expect(screen.getByRole("heading", { name: /整套工具方案/ })).toBeInTheDocument();
    expect(screen.getByText("推荐执行流程")).toBeInTheDocument();
    expect(screen.getByText("市场调研")).toBeInTheDocument();
    expect(screen.getByText("方案概览")).toBeInTheDocument();
  });

  it("renders tool detail information", () => {
    renderPage("detail");

    expect(screen.getByRole("heading", { name: "Midjourney" })).toBeInTheDocument();
    expect(screen.getByText("专业 AI 图像生成工具")).toBeInTheDocument();
    expect(screen.getByText("工具介绍")).toBeInTheDocument();
    expect(screen.getByText("用它解决什么")).toBeInTheDocument();
  });
});
