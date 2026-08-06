import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import { contentApi } from "../lib/contentApi";
import { supportApi } from "../lib/supportApi";
import HelpPage from "./HelpPage";

vi.mock("../lib/contentApi", () => ({
  contentApi: {
    listHelpTopics: vi.fn(),
    listHelpArticles: vi.fn()
  }
}));

vi.mock("../lib/supportApi", () => ({
  supportApi: {
    createTicket: vi.fn(),
    listTickets: vi.fn()
  }
}));

describe("HelpPage", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  function mockHelp() {
    vi.mocked(contentApi.listHelpTopics).mockResolvedValue({
      topics: [{ key: "account", name: "账号与安全" }, { key: "quota", name: "套餐与额度" }]
    });
    vi.mocked(contentApi.listHelpArticles).mockResolvedValue({
      articles: [{ slug: "login-help", topic: "account", title: "如何登录账号", summary: "账号登录与安全验证" }]
    });
    vi.mocked(supportApi.listTickets).mockResolvedValue({
      tickets: [{
        id: 22,
        topic: "套餐与额度",
        title: "额度没有更新",
        body: "充值后额度仍未更新",
        status: "open",
        created_at: "2026-07-02T10:00:00Z"
      }]
    });
  }

  function renderPage() {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-08-11T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张晨", phone: "13800138000", status: "active" }
    });
    return render(
      <MemoryRouter>
        <HelpPage />
      </MemoryRouter>
    );
  }

  it("renders all four help and feedback workflows with API data", async () => {
    mockHelp();
    renderPage();

    expect(screen.getByRole("heading", { name: "帮助与反馈" })).toBeInTheDocument();
    expect(screen.getByText("1. 自助答疑 / 帮助文档")).toBeInTheDocument();
    expect(screen.getByText("2. 意见反馈 / 工单")).toBeInTheDocument();
    expect(screen.getByText("3. 反馈记录 / 工单状态")).toBeInTheDocument();
    expect(screen.getByText("4. 联系企业微信（人工支持 / 升级咨询）")).toBeInTheDocument();
    expect((await screen.findAllByText("账号与安全")).length).toBeGreaterThan(0);
    expect(screen.getByText("如何登录账号")).toBeInTheDocument();
    expect(screen.getByText("额度没有更新")).toBeInTheDocument();
    expect(contentApi.listHelpTopics).toHaveBeenCalled();
    expect(contentApi.listHelpArticles).toHaveBeenCalledWith({ limit: 10 });
    expect(supportApi.listTickets).toHaveBeenCalledWith(20);
  });

  it("shows reference help content when APIs return empty collections", async () => {
    vi.mocked(contentApi.listHelpTopics).mockResolvedValue({ topics: [] });
    vi.mocked(contentApi.listHelpArticles).mockResolvedValue({ articles: [] });
    vi.mocked(supportApi.listTickets).mockResolvedValue({ tickets: [] });
    renderPage();

    expect((await screen.findAllByText("账号与安全")).length).toBeGreaterThan(0);
    expect(screen.getByText("如何修改登录方式")).toBeInTheDocument();
    expect(screen.getByText("商业沙盘数据导出异常")).toBeInTheDocument();
    expect(screen.getByText("增长测算结果与预期不符")).toBeInTheDocument();
  });

  it("creates a support ticket from the inline feedback form", async () => {
    mockHelp();
    vi.mocked(supportApi.createTicket).mockResolvedValue({
      id: 23,
      topic: "套餐与额度",
      title: "额度没有更新",
      body: "充值后额度仍未更新",
      status: "open",
      created_at: "2026-07-02T10:10:00Z"
    });
    renderPage();

    await screen.findByRole("option", { name: "套餐与额度" });
    fireEvent.change(screen.getByLabelText("问题类型"), { target: { value: "套餐与额度" } });
    fireEvent.change(screen.getByLabelText("问题标题"), { target: { value: "额度没有更新" } });
    fireEvent.change(screen.getByLabelText("问题描述"), { target: { value: "充值后额度仍未更新" } });
    fireEvent.click(screen.getByRole("button", { name: "提交反馈" }));

    await waitFor(() => expect(supportApi.createTicket).toHaveBeenCalledWith({
      topic: "套餐与额度",
      title: "额度没有更新",
      body: "充值后额度仍未更新"
    }));
    expect(await screen.findByText("反馈已提交")).toBeInTheDocument();
    expect(screen.getAllByText("额度没有更新").length).toBeGreaterThan(0);
  });

  it("shows submission errors without removing the inline form", async () => {
    mockHelp();
    vi.mocked(supportApi.createTicket).mockRejectedValue(new Error("offline"));
    renderPage();

    await screen.findByRole("option", { name: "账号与安全" });
    fireEvent.change(screen.getByLabelText("问题标题"), { target: { value: "登录失败" } });
    fireEvent.change(screen.getByLabelText("问题描述"), { target: { value: "无法进入工作台" } });
    fireEvent.click(screen.getByRole("button", { name: "提交反馈" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("反馈提交失败，请稍后重试");
    expect(screen.getByLabelText("问题描述")).toHaveValue("无法进入工作台");
  });
});
