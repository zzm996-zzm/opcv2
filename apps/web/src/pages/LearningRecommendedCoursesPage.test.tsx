import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import LearningRecommendedCoursesPage from "./LearningRecommendedCoursesPage";

describe("LearningRecommendedCoursesPage", () => {
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

  function renderRecommendedCoursesPage() {
    signIn();
    render(
      <MemoryRouter>
        <LearningRecommendedCoursesPage />
      </MemoryRouter>
    );
  }

  it("renders personalized recommended courses with context, sorting and copilot guidance", () => {
    renderRecommendedCoursesPage();

    expect(screen.getByRole("heading", { name: "推荐课程" })).toBeInTheDocument();
    expect(screen.getByText("提升智能客服与市场分析能力")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "匹配度" })).toBeInTheDocument();
    expect(screen.getByText("智能客服核心能力全景解析")).toBeInTheDocument();
    expect(screen.getByText("智能客服实战：从需求到落地")).toBeInTheDocument();
    expect(screen.getByText("推荐逻辑说明")).toBeInTheDocument();
    expect(screen.getByText("能力诊断报告.pdf")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /智能客服核心能力全景解析/ })).toHaveAttribute("href", "/learning/courses/intro");
    expect(screen.queryByLabelText("打开智活 Copilot")).not.toBeInTheDocument();
  });

  it("collapses the side recommended-courses copilot", () => {
    renderRecommendedCoursesPage();

    expect(screen.getByText("推荐逻辑说明")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "收起推荐课程助手" }));

    expect(screen.queryByText("推荐逻辑说明")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "展开推荐课程助手" })).toBeInTheDocument();
  });

  it("loads recommended courses from API", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        courses: [
          {
            id: 51,
            slug: "automation-ops",
            title: "AI自动化运营实战",
            description: "把诊断、任务和内容工作流串成自动化运营闭环",
            category: "工具",
            level: "实战",
            hours: 26,
            learners: 7800,
            price_label: "会员免费",
            tags: ["自动化"],
            outline: [],
            created_at: "2026-06-30T08:00:00Z",
            updated_at: "2026-06-30T08:00:00Z"
          }
        ]
      }), { status: 200 })
    );

    renderRecommendedCoursesPage();

    expect(await screen.findByRole("heading", { name: "AI自动化运营实战" })).toBeInTheDocument();
    expect(screen.getByText("把诊断、任务和内容工作流串成自动化运营闭环")).toBeInTheDocument();
    expect(screen.getByText("7.8k人学习")).toBeInTheDocument();
    expect(screen.getByText("智能客服核心能力全景解析")).toBeInTheDocument();
  });
});
