import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import App from "../App";
import { authSession } from "../lib/authSession";

describe("LeadDevelopmentPage", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  function signIn() {
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
  }

  function renderLeadRoute() {
    signIn();
    render(
      <MemoryRouter initialEntries={["/leads"]}>
        <App />
      </MemoryRouter>
    );
  }

  it("renders the AI lead development workbench instead of the placeholder", () => {
    renderLeadRoute();

    expect(screen.getByRole("heading", { name: "AI线索开发" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "新建线索任务" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "高意向线索" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "开发路径" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "星桥教育集团" })).toBeInTheDocument();
    expect(screen.queryByText("第一版正在实现")).not.toBeInTheDocument();
  });

  it("loads lead tasks from API", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        tasks: [
          {
            id: 99,
            user_id: 7,
            query: "华东 连锁教培 智能客服",
            status: "queued",
            idempotency_key: "lead-task-99",
            credit_cost: 1,
            created_at: "2026-06-25T12:00:00Z",
            updated_at: "2026-06-25T12:00:00Z"
          }
        ]
      }), { status: 200 })
    );
    renderLeadRoute();

    expect(await screen.findByRole("heading", { name: "华东 连锁教培 智能客服" })).toBeInTheDocument();
    expect(screen.getAllByText("排队中").length).toBeGreaterThan(0);
  });

  it("creates lead task from target profile", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ tasks: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        id: 100,
        user_id: 7,
        query: "成都 教培 私域转化",
        status: "queued",
        idempotency_key: "lead-task-100",
        credit_cost: 1,
        created_at: "2026-06-25T12:10:00Z",
        updated_at: "2026-06-25T12:10:00Z"
      }), { status: 200 }));
    renderLeadRoute();

    fireEvent.change(screen.getByLabelText("描述目标客户画像"), {
      target: { value: "成都 教培 私域转化" }
    });
    fireEvent.click(screen.getByRole("button", { name: "生成线索池" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/leads/tasks",
      expect.objectContaining({ method: "POST" })
    ));
    expect(await screen.findByRole("heading", { name: "成都 教培 私域转化" })).toBeInTheDocument();
  });
});
