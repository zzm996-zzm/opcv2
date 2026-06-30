import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import LearningHistoryPage from "./LearningHistoryPage";

describe("LearningHistoryPage", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  function signIn() {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-16T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "13800138000", status: "active" }
    });
  }

  function renderHistoryPage() {
    signIn();
    render(
      <MemoryRouter>
        <LearningHistoryPage />
      </MemoryRouter>
    );
  }

  it("renders learning history stats, records, calendar and copilot suggestions", () => {
    renderHistoryPage();

    expect(screen.getByRole("heading", { name: "学习历史进度" })).toBeInTheDocument();
    expect(screen.getByText("18 门")).toBeInTheDocument();
    expect(screen.getByText("42.6 小时")).toBeInTheDocument();
    expect(screen.getByText("最近学习记录")).toBeInTheDocument();
    expect(screen.getByText("AI基础入门")).toBeInTheDocument();
    expect(screen.getAllByText("提示词工程实战").length).toBeGreaterThan(0);
    expect(screen.getByText("学习打卡日历")).toBeInTheDocument();
    expect(screen.getByText("连续学习趋势")).toBeInTheDocument();
    expect(screen.getByText("今日学习建议")).toBeInTheDocument();
    expect(screen.getAllByRole("link", { name: /继续学习/ }).some((link) => link.getAttribute("href") === "/learning/courses/detail")).toBe(true);
  });

  it("loads learning progress from API", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        progress: [
          {
            id: 41,
            user_id: 7,
            course_slug: "enterprise-ai-playbook",
            course_title: "企业AI落地打法专题",
            percent: 73,
            last_lesson: "第3章 业务场景拆解",
            recommended_action: "继续学习场景落地清单",
            updated_at: "2026-06-30T08:00:00Z"
          }
        ]
      }), { status: 200 })
    );

    renderHistoryPage();

    expect(await screen.findByText("企业AI落地打法专题")).toBeInTheDocument();
    expect(screen.getByText("第3章 业务场景拆解")).toBeInTheDocument();
    expect(screen.getByText("73%")).toBeInTheDocument();
  });
});
