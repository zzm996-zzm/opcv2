import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { authSession } from "../lib/authSession";
import MembershipPage from "./MembershipPage";

describe("MembershipPage", () => {
  afterEach(() => {
    authSession.clear();
  });

  it("renders the V4 membership and billing page", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-11T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张晨", phone: "13800138000", status: "active" }
    });

    render(
      <MemoryRouter>
        <MembershipPage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "会员与账单" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "企业版 会员生效" })).toBeInTheDocument();
    expect(screen.getByText("订单记录")).toBeInTheDocument();
    expect(screen.getByText("ZS-20250531-0012")).toBeInTheDocument();
  });

  it("renders the upgrade modal state", () => {
    render(
      <MemoryRouter>
        <MembershipPage showUpgrade />
      </MemoryRouter>
    );

    expect(screen.getByRole("dialog", { name: "升级套餐" })).toBeInTheDocument();
    expect(screen.getAllByText("会员版").length).toBeGreaterThan(0);
    expect(screen.getByText("权益对比一览")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "立即开通" })).toBeInTheDocument();
  });
});
