import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { authSession } from "../lib/authSession";
import LearningPlanPage from "./LearningPlanPage";

describe("LearningPlanPage", () => {
  afterEach(() => {
    authSession.clear();
  });

  it("renders a staged learning path, summary metrics and next actions", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-16T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "13800138000", status: "active" }
    });

    render(
      <MemoryRouter>
        <LearningPlanPage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "系统学习路径" })).toBeInTheDocument();
    expect(screen.getByText("AI基础认知")).toBeInTheDocument();
    expect(screen.getByText("提示词与工具实操")).toBeInTheDocument();
    expect(screen.getByText("行业分析方法")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "智能客服应用案例" })).toBeInTheDocument();
    expect(screen.getByText("预计总学习时长")).toBeInTheDocument();
    expect(screen.getByText("每周学习节奏建议")).toBeInTheDocument();
    expect(screen.getByText("学习目标产出")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /开始第一阶段/ })).toHaveAttribute("href", "/learning/courses/intro");
  });
});
