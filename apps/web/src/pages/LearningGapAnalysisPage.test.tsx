import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import LearningGapAnalysisPage from "./LearningGapAnalysisPage";

describe("LearningGapAnalysisPage", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  it("renders the gap analysis comparison, causes and next actions", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-16T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "13800138000", status: "active" }
    });

    render(
      <MemoryRouter>
        <LearningGapAnalysisPage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "能力诊断" })).toBeInTheDocument();
    expect(screen.getByText("目标要求 vs 当前水平")).toBeInTheDocument();
    expect(screen.getByLabelText("差距分析雷达图")).toBeInTheDocument();
    expect(screen.getByText("关键差距与原因")).toBeInTheDocument();
    expect(screen.getByText("判断依据")).toBeInTheDocument();
    expect(screen.getByText("优先补齐顺序")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /继续生成诊断报告/ })).toHaveAttribute("href", "/learning/report");
  });

  it("loads backend-derived gaps", async () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-16T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "13800138000", status: "active" }
    });
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        diagnosis_id: 99,
        goal: "提升企业AI落地能力",
        project: "企业AI运营项目",
        overall_score: 82,
        gaps: [{
          name: "业务场景拆解",
          current: 58,
          target: 82,
          gap: 24,
          priority: "high",
          summary: "需要补齐场景拆解方法",
          evidence: "诊断显示业务场景拆解差距最大",
          recommended: "完成一次项目拆解练习"
        }],
        evidence: ["项目方向：企业AI运营项目"],
        generated_at: "2026-06-30T08:00:00Z"
      }), { status: 200 })
    );

    render(
      <MemoryRouter>
        <LearningGapAnalysisPage />
      </MemoryRouter>
    );

    expect((await screen.findAllByText("业务场景拆解")).length).toBeGreaterThan(0);
    expect(screen.getByText("项目方向：企业AI运营项目")).toBeInTheDocument();
  });
});
