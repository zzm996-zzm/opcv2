import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { authSession } from "../lib/authSession";
import CommunityMembersPage from "./CommunityMembersPage";

describe("CommunityMembersPage", () => {
  afterEach(() => {
    authSession.clear();
  });

  it("renders the free member community with join QR modal and community context", () => {
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
    expect(screen.getByText("与 1,200+ 创业者一起交流成长")).toBeInTheDocument();
    expect(screen.getByText("使用微信扫一扫，添加社群小助手，拉你入群")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "我知道了" })).toBeInTheDocument();
    expect(screen.getByText("社群价值")).toBeInTheDocument();
    expect(screen.getByText("社群动态")).toBeInTheDocument();
    expect(screen.getByText("本周活动预告")).toBeInTheDocument();
    expect(screen.getByText("智能硬件市场分析报告.pdf")).toBeInTheDocument();
  });
});
