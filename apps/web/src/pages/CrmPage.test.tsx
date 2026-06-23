import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import App from "../App";
import { authSession } from "../lib/authSession";

describe("CrmPage", () => {
  afterEach(() => {
    authSession.clear();
  });

  it("renders the CRM workbench instead of the placeholder", () => {
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

    render(
      <MemoryRouter initialEntries={["/crm"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "CRM客户管理" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "新建客户" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "客户列表" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "跟进看板" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "星桥教育集团" })).toBeInTheDocument();
    expect(screen.queryByText("第一版正在实现")).not.toBeInTheDocument();
  });
});
