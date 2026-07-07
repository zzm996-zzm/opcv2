import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import LearningCoursesPage from "./LearningCoursesPage";

describe("LearningCoursesPage", () => {
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

  function renderCoursesPage() {
    signIn();
    render(
      <MemoryRouter>
        <LearningCoursesPage />
      </MemoryRouter>
    );
  }

  it("renders the all-courses catalog with filters, hot courses and recommendations", () => {
    renderCoursesPage();

    expect(screen.getByRole("heading", { name: "全部课程" })).toBeInTheDocument();
    expect(screen.getByPlaceholderText("搜索课程名称、关键词或讲师")).toBeInTheDocument();
    expect(screen.getByText("本周热门课程")).toBeInTheDocument();
    expect(screen.getByText("AI基础入门：从0到1了解AI")).toBeInTheDocument();
    expect(screen.getByText("提示词工程实战")).toBeInTheDocument();
    expect(screen.getAllByText("智能客服应用案例解析").length).toBeGreaterThan(0);
    expect(screen.getByText("为你推荐的课程")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /AI基础入门：从0到1了解AI/ })).toHaveAttribute("href", "/learning/courses/intro");
    expect(screen.getByRole("button", { name: "2" })).toBeInTheDocument();
    expect(screen.queryByLabelText("打开智活 Copilot")).not.toBeInTheDocument();
  });

  it("collapses the side course copilot without leaving the page", () => {
    renderCoursesPage();

    expect(screen.getByText("为你推荐的课程")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "收起课程助手" }));

    expect(screen.queryByText("为你推荐的课程")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "展开课程助手" })).toBeInTheDocument();
  });

  it("loads course catalog from API", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        courses: [
          {
            id: 31,
            slug: "enterprise-ai-playbook",
            title: "企业AI落地打法专题",
            description: "把课程、任务和业务场景串成落地路径",
            category: "工具",
            level: "实战",
            hours: 24,
            learners: 5500,
            price_label: "会员免费",
            tags: ["RAG"],
            outline: [],
            created_at: "2026-06-30T08:00:00Z",
            updated_at: "2026-06-30T08:00:00Z"
          }
        ]
      }), { status: 200 })
    );

    renderCoursesPage();

    expect(await screen.findByRole("heading", { name: "企业AI落地打法专题" })).toBeInTheDocument();
    expect(screen.getByText("把课程、任务和业务场景串成落地路径")).toBeInTheDocument();
    expect(screen.getAllByText("5.5k人学习").length).toBeGreaterThan(0);
    expect(screen.getByText("AI基础入门：从0到1了解AI")).toBeInTheDocument();
  });
});
