import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import LearningPage from "./LearningPage";

describe("LearningPage", () => {
  afterEach(() => { authSession.clear(); vi.restoreAllMocks(); });
  it("loads only published course API data", async () => {
    authSession.set({ access_token: "token", access_token_expires_at: "2026-07-14T00:00:00Z", is_new_user: false, user: { id: 7, nickname: "张晨", phone: "", status: "active" } });
    vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({ courses: [{ id: 21, slug: "ai-market-analysis", title: "AI行业分析方法", description: "洞察市场趋势", category: "工具", level: "实战", hours: 22, learners: 9100, price_label: "会员免费", tags: [], outline: [], created_at: "", updated_at: "" }] }), { status: 200 }));
    render(<MemoryRouter><LearningPage /></MemoryRouter>);
    expect(screen.getByRole("heading", { name: "课程学习" })).toBeInTheDocument();
    expect(await screen.findByRole("heading", { name: "AI行业分析方法" })).toBeInTheDocument();
    expect(screen.getByText("工具 · 22课时")).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "AI基础入门" })).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: /查看学习计划/ })).toHaveAttribute("href", "/learning/plan");
  });
});
