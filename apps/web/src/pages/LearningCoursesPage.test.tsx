import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import LearningCoursesPage from "./LearningCoursesPage";

describe("LearningCoursesPage", () => {
  afterEach(() => { authSession.clear(); vi.restoreAllMocks(); });
  it("loads, filters, and links the published catalog", async () => {
    authSession.set({ access_token: "token", access_token_expires_at: "2026-07-14T00:00:00Z", is_new_user: false, user: { id: 7, nickname: "张婧", phone: "", status: "active" } });
    vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({ courses: [
      { id: 1, slug: "ai-basics", title: "AI基础入门", description: "核心概念", category: "入门", level: "入门", hours: 18, learners: 100, price_label: "免费", tags: ["基础"], outline: [], created_at: "", updated_at: "" },
      { id: 2, slug: "ai-market-analysis", title: "AI行业分析方法", description: "市场趋势", category: "工具", level: "实战", hours: 22, learners: 9100, price_label: "会员免费", tags: ["市场"], outline: [], created_at: "", updated_at: "" }
    ] }), { status: 200 }));
    render(<MemoryRouter><LearningCoursesPage /></MemoryRouter>);
    expect(await screen.findByRole("heading", { name: "AI基础入门" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /AI行业分析方法/ })).toHaveAttribute("href", "/learning/courses/ai-market-analysis");
    fireEvent.click(screen.getByRole("button", { name: "工具" }));
    expect(screen.queryByRole("heading", { name: "AI基础入门" })).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "收起课程助手" }));
    expect(screen.getByRole("button", { name: "展开课程助手" })).toBeInTheDocument();
    expect(screen.queryByText("本周热门课程")).not.toBeInTheDocument();
  });
});
