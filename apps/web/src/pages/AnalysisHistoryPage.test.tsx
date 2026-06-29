import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import App from "../App";
import { authSession } from "../lib/authSession";

function signIn() {
  authSession.set({
    access_token: "access-token",
    access_token_expires_at: "2026-06-27T12:00:00Z",
    is_new_user: false,
    user: { id: 7, nickname: "张晨", phone: "13800138000", status: "active" }
  });
}

describe("AnalysisHistoryPage", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  it("renders stored analysis reports from the history API", async () => {
    signIn();
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          sessions: [
            {
              id: 99,
              user_id: 7,
              mode: "direction",
              intent: "我有10年教培经验，5万本金，每周20小时",
              status: "completed",
              result: {
                session_id: 99,
                status: "completed",
                cards: [
                  {
                    name: "本地教培转化顾问",
                    score: 91,
                    reasons: ["经验匹配"],
                    market_evidence: "公开需求可验证",
                    difficulty: { level: "中", notes: ["首批客户"] },
                    benchmarks: ["案例A"],
                    actions: ["联系5家机构"],
                    upsell: "升级后生成线索"
                  }
                ]
              },
              created_at: "2026-06-27T08:00:00Z",
              updated_at: "2026-06-27T08:00:00Z"
            },
            {
              id: 100,
              user_id: 7,
              mode: "direction",
              intent: "想创业",
              status: "needs_input",
              questions: [{ key: "budget", text: "你大概有多少启动资金？", options: ["1-5万"] }],
              created_at: "2026-06-26T08:00:00Z",
              updated_at: "2026-06-26T08:00:00Z"
            }
          ]
        }),
        { status: 200 }
      )
    );

    render(
      <MemoryRouter initialEntries={["/analysis/history"]}>
        <App />
      </MemoryRouter>
    );

    expect(await screen.findByRole("heading", { name: "分析历史" })).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/analysis/sessions?limit=20",
      expect.objectContaining({ method: "GET" })
    );
    expect(screen.getByRole("heading", { name: "本地教培转化顾问" })).toBeInTheDocument();
    expect(screen.getByText("91分")).toBeInTheDocument();
    expect(screen.getByText("需要补充信息")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "新建分析" })).toHaveAttribute("href", "/analysis");
  });
});
