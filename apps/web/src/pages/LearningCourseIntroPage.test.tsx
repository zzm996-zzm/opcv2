import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import LearningCourseIntroPage from "./LearningCourseIntroPage";

const course = { id: 4, slug: "ai-market-analysis", title: "AI行业分析方法", description: "用 AI 洞察市场趋势与竞争格局", category: "工具", level: "实战", hours: 22, learners: 9100, price_label: "会员免费", tags: ["市场分析"], outline: ["行业地图", "竞品拆解"], created_at: "", updated_at: "" };

describe("LearningCourseIntroPage", () => {
  afterEach(() => { authSession.clear(); vi.restoreAllMocks(); });
  it("loads a published course without static PDF or related-course claims", async () => {
    authSession.set({ access_token: "token", access_token_expires_at: "2026-07-14T00:00:00Z", is_new_user: false, user: { id: 7, nickname: "张婧", phone: "", status: "active" } });
    vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify(course), { status: 200 }));
    render(<MemoryRouter><LearningCourseIntroPage /></MemoryRouter>);
    expect(await screen.findByRole("heading", { name: "AI行业分析方法" })).toBeInTheDocument();
    expect(screen.getByText("行业地图")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "开始学习" })).toBeInTheDocument();
    expect(screen.queryByText("能力诊断报告.pdf")).not.toBeInTheDocument();
  });

  it("starts the actual course and saves initial progress", async () => {
    authSession.set({ access_token: "token", access_token_expires_at: "2026-07-14T00:00:00Z", is_new_user: false, user: { id: 7, nickname: "张婧", phone: "", status: "active" } });
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      if (String(input) === "/api/v1/learning/courses/ai-market-analysis" && init?.method === "GET") return Promise.resolve(new Response(JSON.stringify(course), { status: 200 }));
      if (String(input) === "/api/v1/learning/progress/ai-market-analysis" && init?.method === "PUT") return Promise.resolve(new Response(JSON.stringify({ id: 7, course_slug: "ai-market-analysis", percent: 0, last_lesson: "行业地图" }), { status: 200 }));
      return Promise.reject(new Error(`unexpected request: ${String(input)}`));
    });
    render(<MemoryRouter><LearningCourseIntroPage /></MemoryRouter>);
    fireEvent.click(await screen.findByRole("button", { name: "开始学习" }));
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/learning/progress/ai-market-analysis", expect.objectContaining({ method: "PUT", body: expect.stringContaining('"last_lesson":"行业地图"') })));
  });
});
