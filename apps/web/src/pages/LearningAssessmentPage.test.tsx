import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import LearningAssessmentPage from "./LearningAssessmentPage";

describe("LearningAssessmentPage", () => {
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

  function renderAssessmentPage() {
    signIn();
    render(
      <MemoryRouter>
        <LearningAssessmentPage />
      </MemoryRouter>
    );
  }

  it("renders the active assessment progress, metrics and assistant actions", () => {
    renderAssessmentPage();

    expect(screen.getByRole("heading", { name: "能力诊断" })).toBeInTheDocument();
    expect(screen.getByText("当前分析阶段：能力评估")).toBeInTheDocument();
    expect(screen.getByText("市场分析能力")).toBeInTheDocument();
    expect(screen.getByText("能力评估模型")).toBeInTheDocument();
    expect(screen.getByLabelText("能力评估雷达图")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /完善资料/ })).toHaveAttribute("href", "/learning/diagnosis");
    expect(screen.getByRole("link", { name: /咨询AI助手/ })).toHaveAttribute("href", "/copilot");
  });

  it("loads latest diagnosis from API", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        id: 99,
        user_id: 7,
        goal: "提升智能客服和市场分析能力",
        project: "智能客服系统",
        status: "completed",
        overall_score: 82,
        dimensions: [
          { name: "市场分析能力", score: 88, gap: 8, summary: "市场判断较强" },
          { name: "数据分析能力", score: 76, gap: 16, summary: "需要加强漏斗分析" }
        ],
        recommendations: ["优先学习 AI行业分析方法"],
        created_at: "2026-06-30T08:00:00Z",
        updated_at: "2026-06-30T08:00:00Z"
      }), { status: 200 })
    );

    renderAssessmentPage();

    expect(await screen.findByText("82%")).toBeInTheDocument();
    expect(screen.getByText("市场判断较强")).toBeInTheDocument();
    expect(screen.getByText("需要加强漏斗分析")).toBeInTheDocument();
  });
});
