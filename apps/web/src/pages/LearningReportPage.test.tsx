import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import LearningReportPage from "./LearningReportPage";

describe("LearningReportPage", () => {
  afterEach(() => { authSession.clear(); vi.restoreAllMocks(); });

  it("renders the persisted report and removes fake PDF export", async () => {
    authSession.set({ access_token: "token", access_token_expires_at: "2026-07-14T00:00:00Z", is_new_user: false, user: { id: 7, nickname: "张婧", phone: "", status: "active" } });
    vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({
      diagnosis_id: 99, goal: "提升AI能力", project: "企业AI运营项目", overall_score: 82,
      dimensions: [{ name: "自动化运营能力", score: 88, gap: 6, summary: "表现较好" }],
      priority_gaps: [{ name: "业务场景拆解", current: 58, target: 82, gap: 24, priority: "high", summary: "需补齐", evidence: "用户输入", recommended: "完成一次项目拆解" }],
      recommendations: ["优先完成项目拆解"], evidence: [], basis: "model_assessment", disclaimer: "模型评估说明",
      assumptions: ["基于用户自述"], evidence_sources: [{ type: "assessment_input", label: "用户本次提交", captured_at: "2026-07-13T08:00:00Z" }], generated_at: "2026-07-13T08:00:00Z"
    }), { status: 200 }));
    render(<MemoryRouter><LearningReportPage /></MemoryRouter>);
    expect(await screen.findByRole("heading", { name: "能力诊断" })).toBeInTheDocument();
    expect(screen.getByText("82/100")).toBeInTheDocument();
    expect(screen.getByText("自动化运营能力")).toBeInTheDocument();
    expect(screen.getByText("用户本次提交")).toBeInTheDocument();
    expect(screen.queryByText("能力诊断报告.pdf")).not.toBeInTheDocument();
  });
});
