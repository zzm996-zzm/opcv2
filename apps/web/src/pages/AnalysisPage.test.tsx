import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import AnalysisPage from "./AnalysisPage";

describe("AnalysisPage", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("shows follow-up questions for thin input", async () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-11T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张晨", phone: "13800138000", status: "active" }
    });
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          session_id: 1,
          status: "needs_input",
          questions: [{ key: "budget", text: "你大概有多少启动资金？", options: ["1-5万"] }]
        }),
        { status: 200 }
      )
    );

    render(
      <MemoryRouter>
        <AnalysisPage />
      </MemoryRouter>
    );

    fireEvent.change(screen.getByLabelText("描述你的资源和目标"), {
      target: { value: "想创业" }
    });
    fireEvent.click(screen.getByRole("button", { name: "开始分析" }));

    expect(await screen.findByText("你大概有多少启动资金？")).toBeInTheDocument();
  });

  it("renders three direction cards", async () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-11T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张晨", phone: "13800138000", status: "active" }
    });
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          session_id: 1,
          status: "completed",
          cards: [
            card("本地教培小班陪跑"),
            card("教师副业内容账号"),
            card("中小机构招生顾问")
          ]
        }),
        { status: 200 }
      )
    );

    render(
      <MemoryRouter>
        <AnalysisPage />
      </MemoryRouter>
    );

    fireEvent.change(screen.getByLabelText("描述你的资源和目标"), {
      target: { value: "我有10年教培经验，5万本金，每周20小时，想在成都创业" }
    });
    fireEvent.click(screen.getByRole("button", { name: "开始分析" }));

    await waitFor(() => expect(screen.getAllByRole("article")).toHaveLength(3));
    expect(screen.getByText("本地教培小班陪跑")).toBeInTheDocument();
  });
});

function card(name: string) {
  return {
    name,
    score: 90,
    reasons: ["经验匹配", "预算够", "本地可验证"],
    market_evidence: "公开来源可验证需求。",
    difficulty: { level: "中", notes: ["首批客户"] },
    benchmarks: ["案例A", "案例B"],
    actions: ["今天写画像", "本周发内容", "联系10人"],
    upsell: "升级后可获取线索"
  };
}
