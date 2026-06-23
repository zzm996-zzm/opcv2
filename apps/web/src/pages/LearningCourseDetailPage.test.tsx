import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { authSession } from "../lib/authSession";
import LearningCourseDetailPage from "./LearningCourseDetailPage";

describe("LearningCourseDetailPage", () => {
  afterEach(() => {
    authSession.clear();
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
});
