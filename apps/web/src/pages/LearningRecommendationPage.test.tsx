import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { authSession } from "../lib/authSession";
import LearningRecommendationPage from "./LearningRecommendationPage";

describe("LearningRecommendationPage", () => {
  afterEach(() => {
    authSession.clear();
  });

  it("renders prioritized directions, courses, learning style and next actions", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-16T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "13800138000", status: "active" }
    });

    render(
      <MemoryRouter>
        <LearningRecommendationPage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "推荐方案" })).toBeInTheDocument();
    expect(screen.getByText("优先补强方向")).toBeInTheDocument();
    expect(screen.getByText("推荐课程")).toBeInTheDocument();
    expect(screen.getByText("建议学习方式")).toBeInTheDocument();
    expect(screen.getByText("下一步可选动作")).toBeInTheDocument();
    expect(screen.getByText("智能客服应用案例")).toBeInTheDocument();
    const planLinks = screen.getAllByRole("link", { name: /生成系统学习路径/ });
    expect(planLinks.some((link) => link.getAttribute("href") === "/learning/plan")).toBe(true);
  });
});
