import { fireEvent, render, screen, waitFor } from "@testing-library/react";
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

describe("AnalysisReportPage", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  it("renders one stored analysis report with direction cards and plan", async () => {
    signIn();
    const fetchMock = vi
      .spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify(reportPayload()), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        items: [
          {
            id: 7,
            user_id: 7,
            session_id: 99,
            day_index: 1,
            title: "整理资源清单",
            detail: "把技能、预算、时间和人脉写成一页表格",
            completed: false,
            created_at: "2026-06-27T08:00:00Z",
            updated_at: "2026-06-27T08:00:00Z"
          }
        ]
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        id: 7,
        user_id: 7,
        session_id: 99,
        day_index: 1,
        title: "整理资源清单",
        detail: "把技能、预算、时间和人脉写成一页表格",
        completed: true,
        created_at: "2026-06-27T08:00:00Z",
        updated_at: "2026-06-27T08:05:00Z"
      }), { status: 200 }));

    render(
      <MemoryRouter initialEntries={["/analysis/sessions/99"]}>
        <App />
      </MemoryRouter>
    );

    expect(await screen.findByRole("heading", { name: "本地教培转化顾问" })).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/analysis/sessions/99",
      expect.objectContaining({ method: "GET" })
    );
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/analysis/sessions/99/action-items",
      expect.objectContaining({ method: "GET" })
    );
    expect(screen.getByRole("link", { name: "智活AI · OPC" })).toHaveAttribute("href", "/");
    expect(screen.getByText("91")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "为什么适合你" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "本报告下一步行动" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("checkbox", { name: "整理资源清单" }));
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/analysis/sessions/99/action-items/7",
      expect.objectContaining({
        method: "PATCH",
        body: JSON.stringify({ completed: true })
      })
    ));
    expect(screen.getByRole("link", { name: "← 返回报告列表" })).toHaveAttribute("href", "/analysis/history");
  });
});

function reportPayload() {
  return {
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
          reasons: ["教培经验可直接迁移", "本地客户更容易建立信任", "首批案例可从熟人圈验证"],
          market_evidence: "家长增长、续费和私域转化仍是高频经营问题。",
          difficulty: { level: "中", notes: ["需要先做出可展示样板"] },
          benchmarks: ["小班续费顾问", "私域招生陪跑"],
          actions: ["整理3个教培问题", "写1页诊断表", "约5个机构负责人访谈"],
          upsell: "升级后可按城市和机构类型生成线索池。"
        }
      ]
    },
    created_at: "2026-06-27T08:00:00Z",
    updated_at: "2026-06-27T08:00:00Z"
  };
}
