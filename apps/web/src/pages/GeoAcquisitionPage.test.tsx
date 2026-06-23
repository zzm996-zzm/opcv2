import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import App from "../App";
import { authSession } from "../lib/authSession";

describe("GeoAcquisitionPage", () => {
  afterEach(() => {
    authSession.clear();
  });

  it("renders the GEO acquisition workbench instead of the placeholder", () => {
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
      <MemoryRouter initialEntries={["/geo"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "GEO获客" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "生成GEO方案" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "AI 搜索覆盖" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "内容阵地任务" })).toBeInTheDocument();
    expect(screen.getByText("智能客服系统怎么选")).toBeInTheDocument();
    expect(screen.queryByText("第一版正在实现")).not.toBeInTheDocument();
  });
});
