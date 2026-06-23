import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import App from "../App";
import { authSession } from "../lib/authSession";

describe("DashboardPage", () => {
  afterEach(() => {
    authSession.clear();
  });

  it("renders the dashboard workbench instead of the placeholder", () => {
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
      <MemoryRouter initialEntries={["/dashboard"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "仪表盘" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "生成经营周报" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "经营指标" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "增长趋势" })).toBeInTheDocument();
    expect(screen.getByText("智能客服系统")).toBeInTheDocument();
    expect(screen.queryByText("第一版正在实现")).not.toBeInTheDocument();
  });
});
