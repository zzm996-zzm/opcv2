import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { authSession } from "../lib/authSession";
import TasksPage from "./TasksPage";

describe("TasksPage", () => {
  afterEach(() => {
    authSession.clear();
  });

  it("renders the task overview and AI generation entry", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-11T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张晨", phone: "13800138000", status: "active" }
    });

    render(
      <MemoryRouter>
        <TasksPage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "任务中心" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "生成任务表" })).toBeInTheDocument();
    expect(screen.getByText("完成智能客服系统项目商业画布")).toBeInTheDocument();
  });
});
