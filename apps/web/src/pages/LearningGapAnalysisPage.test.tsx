import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { authSession } from "../lib/authSession";
import LearningGapAnalysisPage from "./LearningGapAnalysisPage";

describe("LearningGapAnalysisPage", () => {
  afterEach(() => {
    authSession.clear();
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
});
