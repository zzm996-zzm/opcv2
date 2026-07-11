import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import LearningCourseDetailPage from "./LearningCourseDetailPage";

describe("LearningCourseDetailPage", () => {
  afterEach(() => {
    authSession.clear();
	vi.restoreAllMocks();
  });

  it("renders the active lesson player with outline, resources and copilot coaching", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-16T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "13800138000", status: "active" }
    });

    render(
      <MemoryRouter>
        <LearningCourseDetailPage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "2.2 行业生命周期与发展阶段判断" })).toBeInTheDocument();
    expect(screen.getByText("课程大纲")).toBeInTheDocument();
    expect(screen.getByText("课程总进度")).toBeInTheDocument();
    expect(screen.getAllByText("32%").length).toBeGreaterThan(0);
    expect(screen.getByRole("button", { name: "继续下一节" })).toBeInTheDocument();
    expect(screen.getByText("本节内容摘要")).toBeInTheDocument();
    expect(screen.getByText("学习资料")).toBeInTheDocument();
    expect(screen.getByText("行业生命周期分析框架.pptx")).toBeInTheDocument();
    expect(screen.getByText("总结本节关键知识点")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /去学习下一节/ })).toHaveAttribute("href", "/learning/courses/detail");
  });

  it("loads and advances persisted course progress", async () => {
	const fetchMock = vi.spyOn(globalThis, "fetch")
	  .mockResolvedValueOnce(new Response(JSON.stringify({
		id: 7,
		course_slug: "ai-market-analysis",
		course_title: "AI行业分析方法",
		percent: 32,
		last_lesson: "2.2 行业生命周期与发展阶段判断",
		recommended_action: "继续第2章",
		updated_at: "2026-07-11T08:00:00Z"
	  }), { status: 200 }))
	  .mockResolvedValueOnce(new Response(JSON.stringify({
		id: 7,
		course_slug: "ai-market-analysis",
		course_title: "AI行业分析方法",
		percent: 38,
		last_lesson: "2.3 行业规模与增长趋势分析",
		recommended_action: "继续完成第2章",
		updated_at: "2026-07-11T08:10:00Z"
	  }), { status: 200 }));

	render(
	  <MemoryRouter>
		<LearningCourseDetailPage />
	  </MemoryRouter>
	);
	expect((await screen.findAllByText("32%")).length).toBeGreaterThan(0);
	fireEvent.click(screen.getByRole("button", { name: "继续下一节" }));

	await waitFor(() => expect(fetchMock).toHaveBeenLastCalledWith(
	  "/api/v1/learning/progress/ai-market-analysis",
	  expect.objectContaining({ method: "PUT" })
	));
	expect((await screen.findAllByText("38%")).length).toBeGreaterThan(0);
  });
});
