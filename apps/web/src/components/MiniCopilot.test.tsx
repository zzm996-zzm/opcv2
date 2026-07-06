import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import MiniCopilot from "./MiniCopilot";

describe("MiniCopilot", () => {
  it("creates a thread and sends the user's question", async () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-07-06T12:00:00Z",
      is_new_user: false,
      user: {
        id: 7,
        nickname: "张晨",
        phone: "13800138000",
        status: "active"
      }
    });
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation(async (input, init) => {
      const url = String(input);
      if (url === "/api/v1/copilot/threads" && init?.method === "POST") {
        return new Response(JSON.stringify({
          id: 99,
          user_id: 7,
          title: "测试问题",
          mode: "chat",
          model: "gpt-5-5",
          created_at: "2026-07-06T12:00:00Z",
          updated_at: "2026-07-06T12:00:00Z"
        }), { status: 200, headers: { "Content-Type": "application/json" } });
      }
      if (url === "/api/v1/copilot/threads/99/messages" && init?.method === "POST") {
        return new Response(JSON.stringify({
          user_message: {
            id: 1,
            user_id: 7,
            thread_id: 99,
            role: "user",
            content: "测试问题",
            status: "completed",
            model: "gpt-5-5",
            metadata: {},
            created_at: "2026-07-06T12:00:01Z"
          },
          assistant_message: {
            id: 2,
            user_id: 7,
            thread_id: 99,
            role: "assistant",
            content: "测试成功",
            status: "completed",
            model: "gpt-5-5",
            metadata: {},
            created_at: "2026-07-06T12:00:02Z"
          }
        }), { status: 200, headers: { "Content-Type": "application/json" } });
      }
      return new Response(JSON.stringify({ error: "not_found" }), { status: 404 });
    });

    render(
      <MiniCopilot
        ariaLabel="智活 Copilot 测试助手"
        className="learning-copilot"
        inputClassName="learning-copilot-input"
        messagesClassName="learning-chat"
        placeholder="询问任何问题..."
        title="智活 Copilot"
      />
    );

    const assistant = screen.getByRole("complementary", { name: "智活 Copilot 测试助手" });
    fireEvent.change(within(assistant).getByLabelText("向 Copilot 提问"), { target: { value: "测试问题" } });
    fireEvent.click(within(assistant).getByRole("button", { name: "发送" }));

    await waitFor(() => expect(within(assistant).getByText("测试问题")).toBeInTheDocument());
    await waitFor(() => expect(within(assistant).getByText("测试成功")).toBeInTheDocument());
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/copilot/threads/99/messages",
      expect.objectContaining({
        method: "POST",
        headers: expect.objectContaining({ Authorization: "Bearer access-token" })
      })
    ));
  });
});
