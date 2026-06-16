import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { authSession } from "../lib/authSession";
import ProjectsPage from "./ProjectsPage";

describe("ProjectsPage", () => {
  afterEach(() => {
    authSession.clear();
  });

  it("renders project shelves and matching entry", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-11T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张晨", phone: "13800138000", status: "active" }
    });

    render(
      <MemoryRouter>
        <ProjectsPage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "项目超市" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "开始 AI 匹配" })).toBeInTheDocument();
    expect(screen.getByText("智能客服系统")).toBeInTheDocument();
  });
});
