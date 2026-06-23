import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { authSession } from "../lib/authSession";
import LearningCoursesPage from "./LearningCoursesPage";

describe("LearningCoursesPage", () => {
  afterEach(() => {
    authSession.clear();
  });

  it("renders the all-courses catalog with filters, hot courses and recommendations", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-16T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "13800138000", status: "active" }
    });

    render(
      <MemoryRouter>
        <LearningCoursesPage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "全部课程" })).toBeInTheDocument();
    expect(screen.getByPlaceholderText("搜索课程名称、关键词或讲师")).toBeInTheDocument();
    expect(screen.getByText("本周热门课程")).toBeInTheDocument();
    expect(screen.getByText("AI基础入门：从0到1了解AI")).toBeInTheDocument();
    expect(screen.getByText("提示词工程实战")).toBeInTheDocument();
    expect(screen.getAllByText("智能客服应用案例解析").length).toBeGreaterThan(0);
    expect(screen.getByText("为你推荐的课程")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /AI基础入门：从0到1了解AI/ })).toHaveAttribute("href", "/learning/courses/intro");
    expect(screen.getByRole("button", { name: "2" })).toBeInTheDocument();
  });
});
