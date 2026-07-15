import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";

import MembershipPaymentPage from "./MembershipPaymentPage";

describe("MembershipPaymentPage", () => {
  it("renders the checkout reference state", () => {
    render(<MemoryRouter><MembershipPaymentPage mode="checkout" /></MemoryRouter>);
    expect(screen.getByRole("heading", { name: "下单 / 支付页" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "订单摘要" })).toBeInTheDocument();
    expect(screen.getByLabelText("微信支付二维码")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /立即支付/ })).toHaveAttribute("href", "/membership/success");
  });

  it("renders the quota exhausted reference modal", () => {
    render(<MemoryRouter><MembershipPaymentPage mode="quota" /></MemoryRouter>);
    expect(screen.getByRole("dialog", { name: "本月额度已用尽" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: /当前功能已锁定/ })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "升级套餐" })).toHaveAttribute("href", "/membership/upgrade");
  });

  it("renders the payment success reference state", () => {
    render(<MemoryRouter><MembershipPaymentPage mode="success" /></MemoryRouter>);
    expect(screen.getByRole("heading", { name: "支付成功，会员已激活！" })).toBeInTheDocument();
    expect(screen.getByText("额度已刷新，可立即使用")).toBeInTheDocument();
    expect(screen.getByText("20250601101545987612")).toBeInTheDocument();
  });
});
