import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { authSession } from "../lib/authSession";
import HelpPage from "./HelpPage";

describe("HelpPage", () => {
  afterEach(() => {
    authSession.clear();
  });

  it("renders help docs, feedback form, and ticket records", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-11T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张晨", phone: "13800138000", status: "active" }
    });

    render(
      <MemoryRouter>
        <HelpPage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "帮助与反馈" })).toBeInTheDocument();
    expect(screen.getByText("1. 自助答疑 / 帮助文档")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "提交反馈" })).toBeInTheDocument();
    expect(screen.getByText("关于项目超市筛选条件优化建议")).toBeInTheDocument();
  });
});
