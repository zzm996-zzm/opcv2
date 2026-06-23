import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { authSession } from "../lib/authSession";
import LearningAssessmentPage from "./LearningAssessmentPage";

describe("LearningAssessmentPage", () => {
  afterEach(() => {
    authSession.clear();
  });

  it("renders the active assessment progress, metrics and assistant actions", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-16T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "13800138000", status: "active" }
    });

    render(
      <MemoryRouter>
        <LearningAssessmentPage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "能力诊断" })).toBeInTheDocument();
    expect(screen.getByText("当前分析阶段：能力评估")).toBeInTheDocument();
    expect(screen.getByText("市场分析能力")).toBeInTheDocument();
    expect(screen.getByText("能力评估模型")).toBeInTheDocument();
    expect(screen.getByLabelText("能力评估雷达图")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /完善资料/ })).toHaveAttribute("href", "/learning/diagnosis");
    expect(screen.getByRole("link", { name: /咨询AI助手/ })).toHaveAttribute("href", "/copilot");
  });
});
