import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import LearningRecommendationPage from "./LearningRecommendationPage";

describe("LearningRecommendationPage", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
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

  it("loads backend-derived recommendation focus and methods", async () => {
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
        focus: [{ name: "业务场景拆解", priority: "high", summary: "优先补齐业务场景拆解" }],
        recommendations: ["先完成一个企业AI运营项目拆解"],
        methods: [{ title: "建议每周学习节奏", value: "每周 8 小时", detail: "拆成 3 次学习" }],
        generated_at: "2026-06-30T08:00:00Z"
      }), { status: 200 })
    );

    render(
      <MemoryRouter>
        <LearningRecommendationPage />
      </MemoryRouter>
    );

    expect(await screen.findByText("业务场景拆解")).toBeInTheDocument();
    expect(screen.getByText("每周 8 小时")).toBeInTheDocument();
    expect(screen.getByText("先完成一个企业AI运营项目拆解")).toBeInTheDocument();
  });
});
