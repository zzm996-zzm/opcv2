import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import LearningReportPage from "./LearningReportPage";

describe("LearningReportPage", () => {
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

  function renderReportPage() {
    signIn();
    render(
      <MemoryRouter>
        <LearningReportPage />
      </MemoryRouter>
    );
  }

  it("renders the diagnosis report overview, prioritized gaps and evidence", () => {
    renderReportPage();

    expect(screen.getByRole("heading", { name: "能力诊断" })).toBeInTheDocument();
    expect(screen.getByText("诊断概览")).toBeInTheDocument();
    expect(screen.getByText("智能客服与市场分析能力提升")).toBeInTheDocument();
    expect(screen.getByText("能力差距分布")).toBeInTheDocument();
    expect(screen.getByText("优先补齐能力")).toBeInTheDocument();
    expect(screen.getByText("诊断依据")).toBeInTheDocument();
    expect(screen.getByText("能力诊断报告.pdf")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /去补这些课程/ })).toHaveAttribute("href", "/learning/recommendation");
  });

  it("loads diagnosis dimensions into the report", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        id: 99,
        user_id: 7,
        goal: "提升企业AI落地能力",
        project: "企业AI运营项目",
        status: "completed",
        overall_score: 82,
        dimensions: [
          { name: "自动化运营能力", score: 88, gap: 6, summary: "自动化运营能力表现较好" },
          { name: "业务场景拆解", score: 58, gap: 24, summary: "需要补齐场景拆解方法" }
        ],
        recommendations: ["优先补齐业务场景拆解"],
        created_at: "2026-06-30T08:00:00Z",
        updated_at: "2026-06-30T08:00:00Z"
      }), { status: 200 })
    );

    renderReportPage();

    expect(await screen.findByText("82分")).toBeInTheDocument();
    expect(screen.getByText("自动化运营能力")).toBeInTheDocument();
    expect(screen.getByText("业务场景拆解")).toBeInTheDocument();
    expect(screen.getByText("需要补齐场景拆解方法")).toBeInTheDocument();
  });
});
