import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import CommunityMembersPage from "./CommunityMembersPage";

describe("CommunityMembersPage", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  it("renders the free member community without static activity fallback data", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-17T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "13800138000", status: "active" }
    });

    render(
      <MemoryRouter>
        <CommunityMembersPage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "AI社群" })).toBeInTheDocument();
    expect(screen.getByText("创业成长互助社区")).toBeInTheDocument();
    expect(screen.getByText("企业决策者交流圈")).toBeInTheDocument();
    expect(screen.getByRole("dialog", { name: "加入会员社群" })).toBeInTheDocument();
    expect(screen.getByText("二维码未配置")).toBeInTheDocument();
    expect(screen.getAllByText("提交申请后由社群助手联系入群").length).toBeGreaterThan(0);
    expect(screen.getByRole("button", { name: "提交会员社群申请" })).toBeInTheDocument();
    expect(screen.getByText("社群价值")).toBeInTheDocument();
    expect(screen.getByText("社群动态")).toBeInTheDocument();
    expect(screen.getByText("本周活动预告")).toBeInTheDocument();
    expect(screen.getByText("暂无社群动态")).toBeInTheDocument();
    expect(screen.getByText("暂无活动数据")).toBeInTheDocument();
    expect(screen.getByText("暂无社群价值数据")).toBeInTheDocument();
    expect(screen.getByText("暂无社群助手对话")).toBeInTheDocument();
    expect(screen.getByText("暂无社群报告")).toBeInTheDocument();
    expect(screen.queryByText("实战经验：如何用AI提升私域运营效率")).not.toBeInTheDocument();
    expect(screen.queryByText("智能硬件市场分析报告.pdf")).not.toBeInTheDocument();
  });

  it("submits a members community join request", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({
        id: 1,
        headline: "加入智活会员社群",
        description: "一起交流真实增长问题",
        qr_variants: [{ key: "members", label: "会员社群二维码", description: "扫码添加社群助手", image_url: "/qr/members.png", join_url: "https://example.com/members", status: "published" }],
        created_at: "2026-07-02T10:00:00Z",
        updated_at: "2026-07-02T10:00:00Z"
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 9, community: "members", status: "submitted" }), { status: 200 }));
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-17T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "13800138000", status: "active" }
    });

    render(
      <MemoryRouter>
        <CommunityMembersPage />
      </MemoryRouter>
    );

    expect(await screen.findByText("加入智活会员社群")).toBeInTheDocument();
    expect(screen.getByRole("img", { name: "会员社群二维码" })).toHaveAttribute("src", "/qr/members.png");
    expect(screen.getByRole("link", { name: "打开入群链接" })).toHaveAttribute("href", "https://example.com/members");
    fireEvent.click(screen.getByRole("button", { name: "提交会员社群申请" }));

    await waitFor(() => expect(fetchMock).toHaveBeenLastCalledWith(
      "/api/v1/community/join-requests",
      expect.objectContaining({ method: "POST" })
    ));
    expect(await screen.findByText("申请已提交")).toBeInTheDocument();
  });
});
