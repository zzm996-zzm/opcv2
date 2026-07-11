import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import LearningCourseIntroPage from "./LearningCourseIntroPage";

describe("LearningCourseIntroPage", () => {
  afterEach(() => {
    authSession.clear();
	vi.restoreAllMocks();
  });

  it("renders the course intro with overview, syllabus and copilot guidance", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-16T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "13800138000", status: "active" }
    });

    render(
      <MemoryRouter>
        <LearningCourseIntroPage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "AI行业分析方法" })).toBeInTheDocument();
    expect(screen.getByText("用 AI 洞察市场趋势与竞争格局，驱动科学决策")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "立即学习" })).toBeInTheDocument();
    expect(screen.getByText("课程简介")).toBeInTheDocument();
    expect(screen.getByText("你将学到什么")).toBeInTheDocument();
    expect(screen.getByText("章节目录")).toBeInTheDocument();
    expect(screen.getByText("第1章")).toBeInTheDocument();
    expect(screen.getByText("与你的目标强相关")).toBeInTheDocument();
    expect(screen.getByText("能力诊断报告.pdf")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /提示词工程实战/ })).toHaveAttribute("href", "/learning/courses/detail");
  });

  it("starts the course and saves initial progress", async () => {
	const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({
	  id: 7,
	  course_slug: "ai-market-analysis",
	  percent: 1,
	  last_lesson: "第1章 行业分析概述与框架"
	}), { status: 200 }));

	render(
	  <MemoryRouter>
		<LearningCourseIntroPage />
	  </MemoryRouter>
	);
	fireEvent.click(screen.getByRole("button", { name: "立即学习" }));

	await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
	  "/api/v1/learning/progress/ai-market-analysis",
	  expect.objectContaining({ method: "PUT" })
	));
  });
});
