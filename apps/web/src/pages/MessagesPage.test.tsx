import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { authSession } from "../lib/authSession";
import MessagesPage from "./MessagesPage";

describe("MessagesPage", () => {
  afterEach(() => {
    authSession.clear();
  });

  it("renders the notification list from the V4 shell", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-11T12:00:00Z",
      is_new_user: false,
      user: {
        id: 7,
        nickname: "张晨",
        phone: "13800138000",
        status: "active"
      }
    });

    render(
      <MemoryRouter initialEntries={["/messages"]}>
        <MessagesPage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "消息中心" })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "全部 12" })).toHaveAttribute("aria-selected", "true");
    expect(screen.getByText("任务提醒：AI 智能硬件项目拆解完成")).toBeInTheDocument();
  });

  it("renders a message detail page", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-11T12:00:00Z",
      is_new_user: false,
      user: {
        id: 7,
        nickname: "张晨",
        phone: "13800138000",
        status: "active"
      }
    });

    render(
      <MemoryRouter initialEntries={["/messages/task-ai-hardware"]}>
        <Routes>
          <Route element={<MessagesPage />} path="/messages/:messageId" />
        </Routes>
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "任务提醒：AI 智能硬件项目拆解完成" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "查看拆解报告" })).toHaveAttribute("href", "/analysis");
  });
});
