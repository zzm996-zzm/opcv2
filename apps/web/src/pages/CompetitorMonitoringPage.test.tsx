import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { authSession } from "../lib/authSession";
import App from "../App";

describe("CompetitorMonitoringPage", () => {
  afterEach(() => {
    authSession.clear();
  });

  it("renders the competitor monitoring workbench instead of the placeholder", () => {
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
      <MemoryRouter initialEntries={["/competitor-monitoring"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "竞品动态监测" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "新增监测对象" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "监测中竞品" })).toBeInTheDocument();
    expect(screen.getAllByText("小鹅通").length).toBeGreaterThan(0);
    expect(screen.getByRole("heading", { name: "动态时间线" })).toBeInTheDocument();
    expect(screen.queryByText("第一版正在实现")).not.toBeInTheDocument();
  });
});
