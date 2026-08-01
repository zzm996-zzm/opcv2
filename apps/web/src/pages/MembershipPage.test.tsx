import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import { membershipApi } from "../lib/membershipApi";
import MembershipPage from "./MembershipPage";

vi.mock("../lib/membershipApi", () => ({
  membershipApi: {
    current: vi.fn(),
    listPlans: vi.fn(),
    usage: vi.fn(),
    listOrders: vi.fn(),
    checkout: vi.fn(),
    redeem: vi.fn()
  }
}));

describe("MembershipPage", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  function mockMembership() {
    vi.mocked(membershipApi.current).mockResolvedValue({
      plan: { code: "free", name: "免费版", monthly_analysis_limit: 3, lead_export_limit: 0 },
      credit_balance: 25
    });
    vi.mocked(membershipApi.listPlans).mockResolvedValue({
      plans: [
        {
          id: 1,
          code: "free",
          name: "免费版",
          price_cents: 0,
          billing_cycle: "month",
          features: ["基础分析"],
          quotas: [{ key: "analysis", label: "AI分析", limit: 3, unit: "次/月" }],
          recommended: false
        },
        {
          id: 2,
          code: "pro",
          name: "会员版",
          price_cents: 6900,
          billing_cycle: "month",
          features: ["线索数据实时更新", "去水印导出结果"],
          quotas: [{ key: "lead_tasks", label: "AI线索任务", limit: 30, unit: "次/月" }],
          recommended: true
        }
      ]
    });
    vi.mocked(membershipApi.usage).mockResolvedValue({
      usage: [{ key: "lead_tasks", label: "AI线索任务", used: 8, limit: 30, unit: "次/月" }]
    });
    vi.mocked(membershipApi.listOrders).mockResolvedValue({
      orders: [
        {
          id: 12,
          order_no: "ZS-20260702-0012",
          plan_code: "pro",
          amount_cents: 6900,
          status: "paid",
          created_at: "2026-07-02T10:00:00Z",
          paid_at: "2026-07-02T10:01:00Z"
        }
      ]
    });
  }

  it("loads membership plans, usage and orders from API", async () => {
    mockMembership();
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

    expect(await screen.findByRole("heading", { name: "会员版" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "会员与账单" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "免费版 会员生效" })).toBeInTheDocument();
    expect(screen.getByText("当前积分 25")).toBeInTheDocument();
    expect(screen.getByText("AI线索任务 8/30 次/月")).toBeInTheDocument();
    expect(screen.getByText("订单记录")).toBeInTheDocument();
    expect(screen.getByText("ZS-20260702-0012")).toBeInTheDocument();
    expect(membershipApi.current).toHaveBeenCalled();
    expect(membershipApi.listPlans).toHaveBeenCalled();
    expect(membershipApi.usage).toHaveBeenCalled();
    expect(membershipApi.listOrders).toHaveBeenCalledWith(20);
  });

  it("creates checkout for a non-current plan", async () => {
    mockMembership();
    vi.mocked(membershipApi.checkout).mockResolvedValue({
      order: {
        id: 13,
        order_no: "ZS-20260702-0013",
        plan_code: "pro",
        amount_cents: 6900,
        status: "pending",
        created_at: "2026-07-02T10:10:00Z"
      },
      payment: { mode: "manual", message: "客服会协助完成支付" }
    });

    render(
      <MemoryRouter>
        <MembershipPage />
      </MemoryRouter>
    );

    fireEvent.click(await screen.findByRole("button", { name: "开通会员版" }));

    await waitFor(() => expect(membershipApi.checkout).toHaveBeenCalledWith({
      plan_code: "pro",
      billing_cycle: "month"
    }));
    expect(await screen.findByText("已创建订单 ZS-20260702-0013，客服会协助完成支付")).toBeInTheDocument();
  });

  it("shows unconfigured quota state instead of a misleading percentage", async () => {
    mockMembership();
    vi.mocked(membershipApi.usage).mockResolvedValueOnce({
      usage: [{ key: "sandbox", label: "商业沙盘推演", used: 0, limit: 0, unit: "次/月" }]
    });

    render(
      <MemoryRouter>
        <MembershipPage />
      </MemoryRouter>
    );

    expect(await screen.findByText("未配置")).toBeInTheDocument();
    expect(screen.queryByText("剩余 100%")).not.toBeInTheDocument();
  });

  it("redeems membership code from the upgrade page", async () => {
    mockMembership();
    vi.mocked(membershipApi.redeem).mockResolvedValue({
      snapshot: {
        plan: { code: "pro", name: "会员版", monthly_analysis_limit: 30, lead_export_limit: 100 },
        credit_balance: 100
      },
      already_redeemed: false
    });

    render(
      <MemoryRouter>
        <MembershipPage showUpgrade />
      </MemoryRouter>
    );

    expect(screen.getByRole("dialog", { name: "升级套餐" })).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("兑换码"), { target: { value: "VIP20260702" } });
    fireEvent.click(screen.getByRole("button", { name: "立即兑换" }));

    await waitFor(() => expect(membershipApi.redeem).toHaveBeenCalledWith("VIP20260702"));
    expect(await screen.findByText("兑换成功，当前会员：会员版")).toBeInTheDocument();
    expect(screen.getByText("权益对比一览")).toBeInTheDocument();
  });
});
