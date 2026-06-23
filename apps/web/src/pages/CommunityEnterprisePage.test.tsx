import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { authSession } from "../lib/authSession";
import CommunityEnterprisePage from "./CommunityEnterprisePage";

describe("CommunityEnterprisePage", () => {
  afterEach(() => {
    authSession.clear();
  });

  it("renders the enterprise community with gold join modal and resource context", () => {
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
    expect(screen.getByText("与 300+ 企业决策者一起链接资源")).toBeInTheDocument();
    expect(screen.getByText("使用微信扫一扫，添加社群顾问，拉你入群")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "我知道了" })).toBeInTheDocument();
    expect(screen.getByText("连接高价值脉络")).toBeInTheDocument();
    expect(screen.getByText("本周汇总")).toBeInTheDocument();
    expect(screen.getByText("智能硬件市场分析报告.pdf")).toBeInTheDocument();
  });
});
