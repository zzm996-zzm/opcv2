import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import App from "../App";
import { authSession } from "../lib/authSession";

describe("LeadDevelopmentPage", () => {
  afterEach(() => {
    authSession.clear();
  });

  it("renders the AI lead development workbench instead of the placeholder", () => {
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
      <MemoryRouter initialEntries={["/leads"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "AI线索开发" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "新建线索任务" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "高意向线索" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "开发路径" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "星桥教育集团" })).toBeInTheDocument();
    expect(screen.queryByText("第一版正在实现")).not.toBeInTheDocument();
  });
});
