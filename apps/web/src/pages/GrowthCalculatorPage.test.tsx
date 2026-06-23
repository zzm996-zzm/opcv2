import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import App from "../App";
import { authSession } from "../lib/authSession";

describe("GrowthCalculatorPage", () => {
  afterEach(() => {
    authSession.clear();
  });

  it("renders the growth calculator workbench instead of the placeholder", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-23T12:00:00Z",
      is_new_user: false,
      user: {
        id: 7,
        nickname: "张婧",
        phone: "",
        account: "zhangjing",
        status: "active"
      }
    });

    render(
      <MemoryRouter initialEntries={["/growth-calculator"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "增长测算" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "保存测算模型" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "增长漏斗" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "情景对比" })).toBeInTheDocument();
    expect(screen.getByText("保守方案")).toBeInTheDocument();
    expect(screen.queryByText("第一版正在实现")).not.toBeInTheDocument();
  });
});
