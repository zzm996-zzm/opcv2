import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { authSession } from "../lib/authSession";
import ToolsPage from "./ToolsPage";

describe("ToolsPage", () => {
  afterEach(() => {
    authSession.clear();
  });

  it("renders the tool library and AI finder", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-11T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张晨", phone: "13800138000", status: "active" }
    });

    render(
      <MemoryRouter>
        <ToolsPage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "工具箱" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "生成工具方案" })).toBeInTheDocument();
    expect(screen.getByText("ChatGPT")).toBeInTheDocument();
  });
});
