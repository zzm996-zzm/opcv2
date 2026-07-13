import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import LearningHistoryPage from "./LearningHistoryPage";

describe("LearningHistoryPage", () => {
  afterEach(() => { authSession.clear(); vi.restoreAllMocks(); });
  it("renders only persisted progress and derived counts", async () => {
    authSession.set({ access_token: "token", access_token_expires_at: "2026-07-14T00:00:00Z", is_new_user: false, user: { id: 7, nickname: "张婧", phone: "", status: "active" } });
    vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({ progress: [{ id: 41, user_id: 7, course_slug: "enterprise-ai-playbook", course_title: "企业AI落地打法专题", percent: 73, last_lesson: "第3章 业务场景拆解", recommended_action: "继续学习场景落地清单", updated_at: "2026-07-13T08:00:00Z" }] }), { status: 200 }));
    render(<MemoryRouter><LearningHistoryPage /></MemoryRouter>);
    expect(await screen.findByText("企业AI落地打法专题")).toBeInTheDocument();
    expect(screen.getByText("1 门")).toBeInTheDocument();
    expect(screen.getByText("73%")).toBeInTheDocument();
    expect(screen.queryByText("42.6 小时")).not.toBeInTheDocument();
    expect(screen.queryByText("学习打卡日历")).not.toBeInTheDocument();
  });
});
