import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import LearningPlanPage from "./LearningPlanPage";

describe("LearningPlanPage", () => {
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

  function renderPlanPage() {
    signIn();
    render(
      <MemoryRouter>
        <LearningPlanPage />
      </MemoryRouter>
    );
  }

  it("renders a staged learning path, summary metrics and next actions", () => {
    renderPlanPage();

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

  it("loads diagnosis recommendations into the learning path", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        id: 99,
        user_id: 7,
        goal: "提升企业AI落地能力",
        project: "企业AI运营项目",
        status: "completed",
        overall_score: 82,
        dimensions: [],
        recommendations: [
          "先学习企业AI落地打法专题，建立业务场景拆解框架。",
          "补充AI自动化运营实战，把任务流串成闭环。"
        ],
        created_at: "2026-06-30T08:00:00Z",
        updated_at: "2026-06-30T08:00:00Z"
      }), { status: 200 })
    );

    renderPlanPage();

    expect(await screen.findByRole("heading", { name: "企业AI落地能力路径" })).toBeInTheDocument();
    expect(screen.getByText("先学习企业AI落地打法专题，建立业务场景拆解框架。")).toBeInTheDocument();
    expect(screen.getByText("补充AI自动化运营实战，把任务流串成闭环。")).toBeInTheDocument();
  });
});
