import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { authSession } from "../lib/authSession";
import CompetitorDataPage from "./CompetitorDataPage";

describe("CompetitorDataPage", () => {
  afterEach(() => {
    authSession.clear();
  });

  it("renders the competitor data cracking workbench", () => {
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
      <MemoryRouter>
        <CompetitorDataPage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "竞品全盘数据破解" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "启动采集任务" })).toBeInTheDocument();
    expect(screen.getByText("小鹅通")).toBeInTheDocument();
    expect(screen.getByText("AI 破解结论")).toBeInTheDocument();
    expect(screen.queryByText("第一版正在实现")).not.toBeInTheDocument();
  });
});
