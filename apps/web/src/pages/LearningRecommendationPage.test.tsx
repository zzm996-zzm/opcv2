import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import LearningRecommendationPage from "./LearningRecommendationPage";

describe("LearningRecommendationPage", () => {
  afterEach(() => { authSession.clear(); vi.restoreAllMocks(); });

  it("loads the persisted recommendation snapshot without fake courses", async () => {
    authSession.set({ access_token: "token", access_token_expires_at: "2026-07-14T00:00:00Z", is_new_user: false, user: { id: 7, nickname: "张婧", phone: "", status: "active" } });
    vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({
      diagnosis_id: 99, goal: "提升AI能力", project: "企业AI运营项目",
      focus: [{ name: "业务场景拆解", priority: "high", summary: "优先补齐业务场景拆解" }],
      recommendations: ["先完成一个项目拆解"], methods: [{ title: "用户计划投入", value: "每周 8 小时", detail: "来自本次提交" }],
      basis: "model_assessment", disclaimer: "模型评估说明", assumptions: ["基于用户自述"], evidence_sources: [], generated_at: "2026-07-13T08:00:00Z"
    }), { status: 200 }));
    render(<MemoryRouter><LearningRecommendationPage /></MemoryRouter>);
    expect(await screen.findByText("业务场景拆解")).toBeInTheDocument();
    expect(screen.getByText("每周 8 小时")).toBeInTheDocument();
    expect(screen.getByText("先完成一个项目拆解")).toBeInTheDocument();
    expect(screen.getByText(/自动匹配/)).toBeInTheDocument();
    expect(screen.queryByText("智能客服应用案例")).not.toBeInTheDocument();
  });
});
