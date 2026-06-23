import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { authSession } from "../lib/authSession";
import SandboxPage from "./SandboxPage";

describe("SandboxPage", () => {
  afterEach(() => {
    authSession.clear();
  });

  it("renders the business sandbox workbench", () => {
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
      <MemoryRouter>
        <SandboxPage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "商业沙盘" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "开始推演" })).toBeInTheDocument();
    expect(screen.getByText("智能客服系统沙盘")).toBeInTheDocument();
    expect(screen.queryByText("第一版正在实现")).not.toBeInTheDocument();
  });
});
