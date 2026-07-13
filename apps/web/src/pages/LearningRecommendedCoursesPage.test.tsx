import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import LearningRecommendedCoursesPage from "./LearningRecommendedCoursesPage";

describe("LearningRecommendedCoursesPage", () => {
  afterEach(() => { authSession.clear(); vi.restoreAllMocks(); });
  it("labels the page as course selection instead of fake personalization", async () => {
    authSession.set({ access_token: "token", access_token_expires_at: "2026-07-14T00:00:00Z", is_new_user: false, user: { id: 7, nickname: "张婧", phone: "", status: "active" } });
    vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({ courses: [{ id: 51, slug: "automation-ops", title: "AI自动化运营实战", description: "自动化运营闭环", category: "工具", level: "实战", hours: 26, learners: 7800, price_label: "会员免费", tags: [], outline: [], created_at: "", updated_at: "" }] }), { status: 200 }));
    render(<MemoryRouter><LearningRecommendedCoursesPage /></MemoryRouter>);
    expect(await screen.findByRole("heading", { name: "AI自动化运营实战" })).toBeInTheDocument();
    expect(screen.getByText(/自动个性化课程匹配尚未实现/)).toBeInTheDocument();
    expect(screen.queryByText("能力诊断报告.pdf")).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "收起推荐课程助手" }));
    expect(screen.getByRole("button", { name: "展开推荐课程助手" })).toBeInTheDocument();
  });
});
