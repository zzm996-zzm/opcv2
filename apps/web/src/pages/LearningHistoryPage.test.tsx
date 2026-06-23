import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { authSession } from "../lib/authSession";
import LearningHistoryPage from "./LearningHistoryPage";

describe("LearningHistoryPage", () => {
  afterEach(() => {
    authSession.clear();
  });

  it("renders learning history stats, records, calendar and copilot suggestions", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-16T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "13800138000", status: "active" }
    });

    render(
      <MemoryRouter>
        <LearningHistoryPage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "学习历史进度" })).toBeInTheDocument();
    expect(screen.getByText("18 门")).toBeInTheDocument();
    expect(screen.getByText("42.6 小时")).toBeInTheDocument();
    expect(screen.getByText("最近学习记录")).toBeInTheDocument();
    expect(screen.getByText("AI基础入门")).toBeInTheDocument();
    expect(screen.getAllByText("提示词工程实战").length).toBeGreaterThan(0);
    expect(screen.getByText("学习打卡日历")).toBeInTheDocument();
    expect(screen.getByText("连续学习趋势")).toBeInTheDocument();
    expect(screen.getByText("今日学习建议")).toBeInTheDocument();
    expect(screen.getAllByRole("link", { name: /继续学习/ }).some((link) => link.getAttribute("href") === "/learning/courses/detail")).toBe(true);
  });
});
