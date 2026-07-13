import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import LearningCourseDetailPage from "./LearningCourseDetailPage";

const course = { id: 4, slug: "ai-market-analysis", title: "AI行业分析方法", description: "洞察市场趋势", category: "工具", level: "实战", hours: 22, learners: 9100, price_label: "会员免费", tags: [], outline: ["行业地图", "竞品拆解"], created_at: "", updated_at: "" };

describe("LearningCourseDetailPage", () => {
  afterEach(() => { authSession.clear(); vi.restoreAllMocks(); });

  it("loads real course, materials, and progress then saves user-selected progress", async () => {
    authSession.set({ access_token: "token", access_token_expires_at: "2026-07-14T00:00:00Z", is_new_user: false, user: { id: 7, nickname: "张婧", phone: "", status: "active" } });
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/learning/courses/ai-market-analysis") return Promise.resolve(new Response(JSON.stringify(course), { status: 200 }));
      if (url === "/api/v1/learning/courses/ai-market-analysis/materials") return Promise.resolve(new Response(JSON.stringify({ materials: [{ id: 9, course_slug: "ai-market-analysis", title: "行业分析讲义", material_type: "article", content_url: "/content/9", position: 1, downloadable: false, created_at: "", updated_at: "" }] }), { status: 200 }));
      if (url === "/api/v1/learning/progress/ai-market-analysis" && init?.method === "GET") return Promise.resolve(new Response(JSON.stringify({ id: 7, course_slug: "ai-market-analysis", course_title: "AI行业分析方法", percent: 32, last_lesson: "行业地图", recommended_action: "继续学习", updated_at: "" }), { status: 200 }));
      if (url === "/api/v1/learning/progress/ai-market-analysis" && init?.method === "PUT") return Promise.resolve(new Response(JSON.stringify({ id: 7, course_slug: "ai-market-analysis", course_title: "AI行业分析方法", percent: 60, last_lesson: "竞品拆解", recommended_action: "继续学习", updated_at: "" }), { status: 200 }));
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });
    render(<MemoryRouter><LearningCourseDetailPage /></MemoryRouter>);
    expect(await screen.findByRole("heading", { name: "AI行业分析方法" })).toBeInTheDocument();
    expect(screen.getByText("行业分析讲义")).toBeInTheDocument();
    expect(screen.getByDisplayValue("32")).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("学习进度百分比"), { target: { value: "60" } });
    fireEvent.change(screen.getByLabelText("最近学习课节"), { target: { value: "竞品拆解" } });
    fireEvent.click(screen.getByRole("button", { name: "保存学习进度" }));
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/learning/progress/ai-market-analysis", expect.objectContaining({ method: "PUT", body: JSON.stringify({ percent: 60, last_lesson: "竞品拆解", recommended_action: "继续选择下一课节学习" }) })));
    expect(await screen.findByRole("status")).toHaveTextContent("学习进度已保存");
  });
});
