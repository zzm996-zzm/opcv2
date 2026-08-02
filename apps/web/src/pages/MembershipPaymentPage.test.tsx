import { fireEvent, render, screen } from "@testing-library/react";
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

  it("keeps checkout choices and coupon actions in real selectable states", () => {
    render(<MemoryRouter><MembershipPaymentPage mode="checkout" /></MemoryRouter>);

    const month = screen.getByRole("radio", { name: /月付/ });
    expect(month).toHaveAttribute("aria-checked", "false");
    fireEvent.click(month);
    expect(month).toHaveAttribute("aria-checked", "true");
    expect(screen.getByText("会员版（月付）")).toBeInTheDocument();
    expect(screen.getAllByText("¥98.00").length).toBeGreaterThan(0);

    const couponInput = screen.getByLabelText("优惠券或兑换码");
    const applyButton = screen.getByRole("button", { name: "应用" });
    expect(applyButton).toBeDisabled();
    fireEvent.change(couponInput, { target: { value: "SAVE10" } });
    expect(applyButton).toBeEnabled();

    const alipay = screen.getByRole("radio", { name: /支付宝/ });
    fireEvent.click(alipay);
    expect(alipay).toHaveAttribute("aria-checked", "true");
    expect(screen.getByRole("heading", { name: "支付宝" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "《会员服务协议》" })).toHaveAttribute("href", "/terms");
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
