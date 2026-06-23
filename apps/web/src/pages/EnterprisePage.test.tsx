import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import App from "../App";
import { authSession } from "../lib/authSession";

describe("EnterprisePage", () => {
  afterEach(() => {
    authSession.clear();
  });

  it("renders the enterprise companion workbench instead of the placeholder", () => {
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
      <MemoryRouter initialEntries={["/enterprise"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "企业定制化陪跑" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "预约企业诊断" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "陪跑方案" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "交付看板" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "增长团队训练营" })).toBeInTheDocument();
    expect(screen.queryByText("第一版正在实现")).not.toBeInTheDocument();
  });
});
