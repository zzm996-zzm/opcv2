import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import CommunityPage from "./CommunityPage";

describe("CommunityPage", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  it("renders the AI community portal without static activity fallback data", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-17T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "13800138000", status: "active" }
    });

    render(
      <MemoryRouter>
        <CommunityPage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "加入智活社群，与优秀创业者一起增长" })).toBeInTheDocument();
    expect(screen.getByText("创业成长互助社区")).toBeInTheDocument();
    expect(screen.getByText("企业家陪伴成长圈")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /免费加入社群/ })).toHaveAttribute("href", "/community/members");
    expect(screen.getByRole("link", { name: /升级VIP加入/ })).toHaveAttribute("href", "/community/enterprise");
    expect(screen.getByText("社群动态")).toBeInTheDocument();
    expect(screen.getByText("本周活动预告")).toBeInTheDocument();
    expect(screen.getByText("社群价值数据")).toBeInTheDocument();
    expect(screen.getByText("你的成长路径")).toBeInTheDocument();
    expect(screen.getByText("暂无社群动态")).toBeInTheDocument();
    expect(screen.getByText("暂无活动数据")).toBeInTheDocument();
    expect(screen.getByText("暂无社群价值数据")).toBeInTheDocument();
    expect(screen.queryByText("AI如何搭建私域的3个关键动作")).not.toBeInTheDocument();
    expect(screen.queryByText("企业私域增长的底层逻辑与实操打法")).not.toBeInTheDocument();
    expect(screen.queryByText("智能客服系统机会分析报告")).not.toBeInTheDocument();
  });

  it("loads community config from content API", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        id: 1,
        headline: "加入实战增长社区",
        description: "和创业者一起复盘获客案例",
        join_url: "https://example.com/community",
        qr_variants: [],
        created_at: "2026-07-02T10:00:00Z",
        updated_at: "2026-07-02T10:00:00Z"
      }), { status: 200 })
    );
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-17T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "13800138000", status: "active" }
    });

    render(
      <MemoryRouter>
        <CommunityPage />
      </MemoryRouter>
    );

    expect(await screen.findByRole("heading", { name: "加入实战增长社区" })).toBeInTheDocument();
    expect(screen.getByText("和创业者一起复盘获客案例")).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/content/community", expect.any(Object));
  });
});
