import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { authSession } from "../lib/authSession";
import LearningRecommendedCoursesPage from "./LearningRecommendedCoursesPage";

describe("LearningRecommendedCoursesPage", () => {
  afterEach(() => {
    authSession.clear();
  });

  it("renders personalized recommended courses with context, sorting and copilot guidance", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-16T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "13800138000", status: "active" }
    });

    render(
      <MemoryRouter>
        <LearningRecommendedCoursesPage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "推荐课程" })).toBeInTheDocument();
    expect(screen.getByText("提升智能客服与市场分析能力")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "匹配度" })).toBeInTheDocument();
    expect(screen.getByText("智能客服核心能力全景解析")).toBeInTheDocument();
    expect(screen.getByText("智能客服实战：从需求到落地")).toBeInTheDocument();
    expect(screen.getByText("推荐逻辑说明")).toBeInTheDocument();
    expect(screen.getByText("能力诊断报告.pdf")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /智能客服核心能力全景解析/ })).toHaveAttribute("href", "/learning/courses/intro");
  });
});
