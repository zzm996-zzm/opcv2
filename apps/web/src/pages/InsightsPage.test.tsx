import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { authSession } from "../lib/authSession";
import InsightsPage from "./InsightsPage";

describe("InsightsPage", () => {
  afterEach(() => {
    authSession.clear();
  });

  function renderPage(variant?: "list" | "detail" | "fileAnalysis") {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-18T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "", account: "zhangjing", status: "active" }
    });

    render(
      <MemoryRouter>
        <InsightsPage variant={variant} />
      </MemoryRouter>
    );
  }

  it("renders the insights list with focus topics and copilot actions", () => {
    renderPage();

    expect(screen.getByRole("heading", { name: "咨询通" })).toBeInTheDocument();
    expect(screen.getByLabelText("搜索资讯")).toBeInTheDocument();
    expect(screen.getByText("今日关注")).toBeInTheDocument();
    expect(screen.getByText("企业智能客服落地实践：从成本中心到增长引擎")).toBeInTheDocument();
    expect(screen.getAllByRole("link", { name: /查看详情/ })[0]).toHaveAttribute("href", "/insights/detail");
    expect(screen.getByText("今日 AI 客服行业资讯摘要")).toBeInTheDocument();
  });

  it("renders the article detail view with summary and recommendations", () => {
    renderPage("detail");

    expect(screen.getByRole("heading", { name: /企业智能客服进入规模化落地阶段/ })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "分析我能学到什么" })).toHaveAttribute("href", "/insights/file-analysis");
    expect(screen.getByText("资讯摘要")).toBeInTheDocument();
    expect(screen.getByText("相关推荐")).toBeInTheDocument();
    expect(screen.getByText("您可以这样问（与资讯相关）")).toBeInTheDocument();
  });

  it("renders the file analysis view with references", () => {
    renderPage("fileAnalysis");

    expect(screen.getByText("为您分析如下：")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "出处引用（点击查看原文）" })).toBeInTheDocument();
    expect(screen.getByText("IDC")).toBeInTheDocument();
    expect(screen.getByLabelText("向咨询通 Copilot 提问")).toBeInTheDocument();
  });
});
