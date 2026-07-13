import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import CommunityEnterprisePage from "./CommunityEnterprisePage";

describe("CommunityEnterprisePage", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  it("renders the enterprise community without static activity fallback data", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-17T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "13800138000", status: "active" }
    });

    render(
      <MemoryRouter>
        <CommunityEnterprisePage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "AI社群" })).toBeInTheDocument();
    expect(screen.getByText("创业成长互助社区")).toBeInTheDocument();
    expect(screen.getByText("企业决策者交流圈")).toBeInTheDocument();
    expect(screen.getByRole("dialog", { name: "加入企业社群" })).toBeInTheDocument();
    expect(screen.getByText("二维码未配置")).toBeInTheDocument();
    expect(screen.getAllByText("提交申请后由社群顾问联系入群").length).toBeGreaterThan(0);
    expect(screen.getByRole("button", { name: "提交企业社群申请" })).toBeInTheDocument();
    expect(screen.getByText("社群价值")).toBeInTheDocument();
    expect(screen.getByText("暂无社群动态")).toBeInTheDocument();
    expect(screen.getByText("暂无活动数据")).toBeInTheDocument();
    expect(screen.getByText("暂无社群价值数据")).toBeInTheDocument();
    expect(screen.getByText("暂无社群助手对话")).toBeInTheDocument();
    expect(screen.getByText("暂无社群报告")).toBeInTheDocument();
    expect(screen.queryByText("分享了智能AI眼镜的底盘？")).not.toBeInTheDocument();
    expect(screen.queryByText("智能AI眼镜赛道的增长与实战复盘")).not.toBeInTheDocument();
    expect(screen.queryByText("智能硬件市场分析报告.pdf")).not.toBeInTheDocument();
  });

  it("submits an enterprise community join request", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({
        id: 1,
        headline: "加入智活企业社群",
        description: "链接企业决策者与增长资源",
        qr_variants: [{ key: "enterprise", label: "企业社群二维码", description: "扫码添加企业顾问", image_url: "/qr/enterprise.png", join_url: "https://example.com/enterprise", status: "published" }],
        created_at: "2026-07-02T10:00:00Z",
        updated_at: "2026-07-02T10:00:00Z"
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 10, community: "enterprise", status: "submitted" }), { status: 200 }));
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-17T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "13800138000", status: "active" }
    });

    render(
      <MemoryRouter>
        <CommunityEnterprisePage />
      </MemoryRouter>
    );

    expect(await screen.findByText("加入智活企业社群")).toBeInTheDocument();
    expect(screen.getByRole("img", { name: "企业社群二维码" })).toHaveAttribute("src", "/qr/enterprise.png");
    expect(screen.getByRole("link", { name: "打开入群链接" })).toHaveAttribute("href", "https://example.com/enterprise");
    fireEvent.click(screen.getByRole("button", { name: "提交企业社群申请" }));

    await waitFor(() => expect(fetchMock).toHaveBeenLastCalledWith(
      "/api/v1/community/join-requests",
      expect.objectContaining({ method: "POST" })
    ));
    expect(await screen.findByText("申请已提交")).toBeInTheDocument();
  });
});
