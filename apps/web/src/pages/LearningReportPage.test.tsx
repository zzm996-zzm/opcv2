import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { authSession } from "../lib/authSession";
import LearningReportPage from "./LearningReportPage";

describe("LearningReportPage", () => {
  afterEach(() => {
    authSession.clear();
  });

  it("renders the diagnosis report overview, prioritized gaps and evidence", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-16T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "13800138000", status: "active" }
    });

    render(
      <MemoryRouter>
        <LearningReportPage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "能力诊断" })).toBeInTheDocument();
    expect(screen.getByText("诊断概览")).toBeInTheDocument();
    expect(screen.getByText("智能客服与市场分析能力提升")).toBeInTheDocument();
    expect(screen.getByText("能力差距分布")).toBeInTheDocument();
    expect(screen.getByText("优先补齐能力")).toBeInTheDocument();
    expect(screen.getByText("诊断依据")).toBeInTheDocument();
    expect(screen.getByText("能力诊断报告.pdf")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /去补这些课程/ })).toHaveAttribute("href", "/learning/recommendation");
  });
});
