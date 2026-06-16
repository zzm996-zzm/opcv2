import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { authSession } from "../lib/authSession";
import LearningDiagnosisPage from "./LearningDiagnosisPage";

describe("LearningDiagnosisPage", () => {
  afterEach(() => {
    authSession.clear();
  });

  it("renders the diagnosis flow, connected data and start action", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-11T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张晨", phone: "13800138000", status: "active" }
    });

    render(
      <MemoryRouter>
        <LearningDiagnosisPage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "能力诊断" })).toBeInTheDocument();
    expect(screen.getByText("收集信息")).toBeInTheDocument();
    expect(screen.getByText("已接入分析的数据源")).toBeInTheDocument();
    expect(screen.getByText("本次诊断将分析什么")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /开始能力诊断/ })).toHaveAttribute("href", "/learning/assessment");
  });
});
