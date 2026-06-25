import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import App from "../App";
import { authSession } from "../lib/authSession";

describe("CrmPage", () => {
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

  function renderCrmRoute() {
    signIn();
    render(
      <MemoryRouter initialEntries={["/crm"]}>
        <App />
      </MemoryRouter>
    );
  }

  it("renders the CRM workbench instead of the placeholder", () => {
    renderCrmRoute();

    expect(screen.getByRole("heading", { name: "CRM客户管理" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "新建客户" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "客户列表" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "跟进看板" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "星桥教育集团" })).toBeInTheDocument();
    expect(screen.queryByText("第一版正在实现")).not.toBeInTheDocument();
  });

  it("loads due CRM customers from API", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        customers: [
          {
            id: 100,
            user_id: 7,
            import_key: "lead:99",
            name: "成都启明星教育",
            phone: "028-12345678",
            stage: "contacted",
            source: "lead",
            next_follow_up_at: "2026-06-25T14:00:00Z",
            created_at: "2026-06-24T12:00:00Z",
            updated_at: "2026-06-25T12:00:00Z"
          }
        ]
      }), { status: 200 })
    );
    renderCrmRoute();

    expect(await screen.findByRole("heading", { name: "成都启明星教育" })).toBeInTheDocument();
    expect(screen.getAllByText(/需求确认/).length).toBeGreaterThan(0);
  });
});
