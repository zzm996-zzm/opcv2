import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { authSession } from "../lib/authSession";
import CommunityPage from "./CommunityPage";

describe("CommunityPage", () => {
  afterEach(() => {
    authSession.clear();
  });

  it("renders the AI community portal with free and VIP groups, activity and copilot help", () => {
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
    expect(screen.getByText("智能客服系统机会分析报告")).toBeInTheDocument();
  });
});
