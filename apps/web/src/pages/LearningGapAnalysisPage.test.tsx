import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import LearningGapAnalysisPage from "./LearningGapAnalysisPage";

describe("LearningGapAnalysisPage", () => {
  afterEach(() => { authSession.clear(); vi.restoreAllMocks(); });

  it("renders only backend-derived gaps and evidence", async () => {
    authSession.set({ access_token: "token", access_token_expires_at: "2026-07-14T00:00:00Z", is_new_user: false, user: { id: 7, nickname: "张婧", phone: "", status: "active" } });
    vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({
      diagnosis_id: 99, goal: "提升企业AI落地能力", project: "企业AI运营项目", overall_score: 82,
      gaps: [{ name: "业务场景拆解", current: 58, target: 82, gap: 24, priority: "high", summary: "需要补齐场景拆解方法", evidence: "用户输入", recommended: "完成一次项目拆解练习" }],
      evidence: [], basis: "model_assessment", disclaimer: "模型评估说明", assumptions: ["基于用户自述"],
      evidence_sources: [{ type: "assessment_input", label: "用户本次提交", captured_at: "2026-07-13T08:00:00Z" }], generated_at: "2026-07-13T08:00:00Z"
    }), { status: 200 }));
    render(<MemoryRouter><LearningGapAnalysisPage /></MemoryRouter>);
    expect(await screen.findByText("业务场景拆解")).toBeInTheDocument();
    expect(screen.getByText("用户本次提交")).toBeInTheDocument();
    expect(screen.getByText("基于用户自述")).toBeInTheDocument();
    expect(screen.queryByLabelText("差距分析雷达图")).not.toBeInTheDocument();
  });
});
