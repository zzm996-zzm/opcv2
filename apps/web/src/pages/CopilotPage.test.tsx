import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { authSession } from "../lib/authSession";
import CopilotPage, { type CopilotVariant } from "./CopilotPage";

describe("CopilotPage", () => {
  afterEach(() => {
    authSession.clear();
  });

  function renderPage(variant: CopilotVariant = "home") {
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

    return render(
      <MemoryRouter>
        <CopilotPage variant={variant} />
      </MemoryRouter>
    );
  }

  it("renders the conversation workspace", () => {
    renderPage();

    expect(screen.getByRole("heading", { name: "智活 Copilot" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: /市场规模与增长趋势/ })).toBeInTheDocument();
    expect(screen.getByRole("complementary", { name: "会话记录" })).toBeInTheDocument();
  });

  it("renders a clean new conversation", () => {
    renderPage("new");

    expect(screen.getByRole("heading", { name: "开始一段新的对话" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "分析一个新项目机会" })).toBeInTheDocument();
  });

  it("renders the model picker", () => {
    renderPage("models");

    expect(screen.getByRole("dialog", { name: "模型选择" })).toBeInTheDocument();
    expect(screen.getByText("Claude opus4.8")).toBeInTheDocument();
    expect(screen.getByText("Grok4.3")).toBeInTheDocument();
  });

  it("renders the reference picker", () => {
    renderPage("files");

    expect(screen.getByRole("dialog", { name: "引用" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "插入引用" })).toBeInTheDocument();
  });

  it("renders the three-model comparison", () => {
    renderPage("compare");

    expect(screen.getByRole("heading", { name: "AI 对比分析" })).toBeInTheDocument();
    expect(screen.getAllByText("回答完成")).toHaveLength(3);
    expect(screen.getByRole("button", { name: "总结本次对比分析" })).toBeInTheDocument();
  });

  it("renders the rename dialog", () => {
    renderPage("rename");

    expect(screen.getByRole("dialog", { name: "重命名会话" })).toBeInTheDocument();
    expect(screen.getByDisplayValue("智能客服系统项目机会分析")).toBeInTheDocument();
  });

  it("renders the delete confirmation", () => {
    renderPage("delete");

    expect(screen.getByRole("dialog", { name: "删除会话" })).toBeInTheDocument();
    expect(screen.getByRole("checkbox", { name: "我已确认删除此会话" })).toBeInTheDocument();
  });
});
