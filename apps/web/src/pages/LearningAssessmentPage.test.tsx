import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import LearningAssessmentPage from "./LearningAssessmentPage";

function signIn() {
  authSession.set({ access_token: "token", access_token_expires_at: "2026-07-14T00:00:00Z", is_new_user: false, user: { id: 7, nickname: "张婧", phone: "", status: "active" } });
}

describe("LearningAssessmentPage", () => {
  afterEach(() => { authSession.clear(); vi.restoreAllMocks(); });

  it("shows the reference assessment when no saved assessment exists", async () => {
    signIn();
    vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({ error: "diagnosis_not_found" }), { status: 404 }));
    render(<MemoryRouter><LearningAssessmentPage /></MemoryRouter>);
    expect(await screen.findByRole("heading", { name: "能力诊断" })).toBeInTheDocument();
    expect(screen.getByText("62/100")).toBeInTheDocument();
  });

  it("renders persisted model provenance and dimensions", async () => {
    signIn();
    vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({
      id: 99, user_id: 7, goal: "提升AI能力", project: "智能客服", status: "completed", overall_score: 82,
      dimensions: [{ name: "市场分析能力", score: 88, gap: 8, summary: "市场判断较强" }], recommendations: [],
      basis: "model_assessment", disclaimer: "模型评估，不是能力认证。", assumptions: ["基于用户自述"],
      evidence_sources: [{ type: "assessment_input", label: "用户本次提交", captured_at: "2026-07-13T08:00:00Z" }],
      answers: [], focus_abilities: [], weekly_time: "", bottleneck: "", created_at: "2026-07-13T08:00:00Z", updated_at: "2026-07-13T08:00:00Z"
    }), { status: 200 }));
    render(<MemoryRouter><LearningAssessmentPage /></MemoryRouter>);
    expect(await screen.findByRole("heading", { name: "能力诊断" })).toBeInTheDocument();
    expect(screen.getByText("82/100")).toBeInTheDocument();
    expect(screen.getByText("市场判断较强")).toBeInTheDocument();
    expect(screen.getByText("用户本次提交")).toBeInTheDocument();
    expect(screen.getByText("基于用户自述")).toBeInTheDocument();
  });
});
