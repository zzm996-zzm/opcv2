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

  it("renders an empty lead workbench instead of static sample leads", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ tasks: [] }), { status: 200 })
    );
    renderLeadRoute();

    expect(screen.getByRole("heading", { name: "AI线索开发" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "新建线索任务" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "高意向线索" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "开发路径" })).toBeInTheDocument();
    expect(await screen.findByText("暂无高意向线索")).toBeInTheDocument();
    expect(screen.getByText("暂无CRM统计")).toBeInTheDocument();
    expect(screen.getByText("暂无待跟进客户")).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "星桥教育集团" })).not.toBeInTheDocument();
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

  it("loads completed lead task results from API", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/leads/tasks?limit=20") {
        return Promise.resolve(new Response(JSON.stringify({
          tasks: [
            {
              id: 99,
              user_id: 7,
              query: "成都 教培 私域转化",
              status: "succeeded",
              idempotency_key: "lead-task-99",
              credit_cost: 1,
              created_at: "2026-06-25T12:00:00Z",
              updated_at: "2026-06-25T12:05:00Z"
            }
          ]
        }), { status: 200 }));
      }
      if (url === "/api/v1/leads/tasks/99") {
        return Promise.resolve(new Response(JSON.stringify({
          task: {
            id: 99,
            user_id: 7,
            query: "成都 教培 私域转化",
            status: "succeeded",
            idempotency_key: "lead-task-99",
            credit_cost: 1,
            created_at: "2026-06-25T12:00:00Z",
            updated_at: "2026-06-25T12:05:00Z"
          },
          progress_percent: 100,
          message: "线索采集已完成，可以查看结果并导入 CRM。",
          results_count: 1
        }), { status: 200 }));
      }
      if (url === "/api/v1/leads/tasks/99/results?limit=20") {
        return Promise.resolve(new Response(JSON.stringify({
          results: [
            {
              id: 7,
              task_id: 99,
              name: "成都启明星教育",
              phone: "028-12345678",
              website: "https://example.com",
              evidence: [
                { type: "website", title: "官网出现暑期招生咨询入口", url: "https://example.com" },
                { type: "search", title: "公开页面提到私域转化", url: "https://example.com/case" }
              ],
              created_at: "2026-06-25T12:05:00Z"
            }
          ]
        }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    renderLeadRoute();

    expect(await screen.findByRole("heading", { name: "成都启明星教育" })).toBeInTheDocument();
    expect(screen.getByText("官网出现暑期招生咨询入口")).toBeInTheDocument();
    expect(screen.getByText("线索采集已完成，可以查看结果并导入 CRM。 已发现 1 条候选线索。")).toBeInTheDocument();
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
