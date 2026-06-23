import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { authSession } from "../lib/authSession";
import LearningCourseIntroPage from "./LearningCourseIntroPage";

describe("LearningCourseIntroPage", () => {
  afterEach(() => {
    authSession.clear();
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
});
