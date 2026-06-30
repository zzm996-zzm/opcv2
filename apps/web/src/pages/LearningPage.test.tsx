import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import LearningPage from "./LearningPage";

describe("LearningPage", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  function signIn() {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-11T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张晨", phone: "13800138000", status: "active" }
    });
  }

  function renderLearningPage() {
    signIn();
    render(
      <MemoryRouter>
        <LearningPage />
      </MemoryRouter>
    );
  }

  it("renders the AI teaching home with course and Copilot sections", () => {
    renderLearningPage();

    expect(screen.getByRole("heading", { name: "课程学习" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "推荐课程" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "AI基础入门" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /为我制定学习计划/ })).toHaveAttribute("href", "/learning/plan");
  });

  it("loads recommended courses from API", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        courses: [
          {
            id: 21,
            slug: "ai-market-analysis",
            title: "AI行业分析方法",
            description: "用 AI 洞察市场趋势与竞争格局",
            category: "工具",
            level: "实战",
            hours: 22,
            learners: 9100,
            price_label: "会员免费",
            tags: ["市场分析"],
            outline: [],
            created_at: "2026-06-30T08:00:00Z",
            updated_at: "2026-06-30T08:00:00Z"
          }
        ]
      }), { status: 200 })
    );

    renderLearningPage();

    expect(await screen.findByRole("heading", { name: "AI行业分析方法" })).toBeInTheDocument();
    expect(screen.getByText("工具 · 22课时")).toBeInTheDocument();
    expect(screen.getByText("9.1k人学习")).toBeInTheDocument();
  });
});
