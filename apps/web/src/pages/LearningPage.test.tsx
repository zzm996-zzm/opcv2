import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { authSession } from "../lib/authSession";
import LearningPage from "./LearningPage";

describe("LearningPage", () => {
  afterEach(() => {
    authSession.clear();
  });

  it("renders the AI teaching home with course and Copilot sections", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-11T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张晨", phone: "13800138000", status: "active" }
    });

    render(
      <MemoryRouter>
        <LearningPage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "课程学习" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "推荐课程" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "AI基础入门" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /为我制定学习计划/ })).toHaveAttribute("href", "/learning/plan");
  });
});
