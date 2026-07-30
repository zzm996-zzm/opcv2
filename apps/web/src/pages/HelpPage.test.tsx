import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import { contentApi } from "../lib/contentApi";
import { supportApi } from "../lib/supportApi";
import HelpPage from "./HelpPage";

vi.mock("../lib/contentApi", () => ({
  contentApi: {
    listHelpTopics: vi.fn()
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
    vi.mocked(supportApi.listTickets).mockResolvedValue({
      tickets: [
        {
          id: 22,
          topic: "套餐与额度",
          title: "额度没有更新",
          body: "充值后额度仍未更新",
          status: "processing",
          created_at: "2026-07-02T10:00:00Z",
          updated_at: "2026-07-02T10:20:00Z"
        },
        {
          id: 23,
          topic: "账号与安全",
          title: "无法修改登录方式",
          body: "设置页面没有保存成功",
          status: "replied",
          created_at: "2026-07-01T10:00:00Z"
        },
        {
          id: 24,
          topic: "账号与安全",
          title: "历史账号咨询",
          body: "历史咨询内容",
          status: "closed",
          created_at: "2026-06-28T10:00:00Z"
        }
      ]
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

  it("renders the ticket center and API-backed statistics", async () => {
    mockHelp();
    renderPage();

    expect(screen.getByRole("heading", { name: "工单中心" })).toBeInTheDocument();
    expect(await screen.findByText("额度没有更新")).toBeInTheDocument();
    expect(screen.getByText("无法修改登录方式")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "提交工单" })).toBeInTheDocument();
    expect(within(screen.getByLabelText("工单统计")).getByText("3")).toBeInTheDocument();
    expect(within(screen.getByLabelText("工单统计")).getByText("处理中")).toBeInTheDocument();
    expect(contentApi.listHelpTopics).toHaveBeenCalled();
    expect(supportApi.listTickets).toHaveBeenCalledWith(20);
  });

  it("shows reference tickets when APIs return empty collections", async () => {
    vi.mocked(contentApi.listHelpTopics).mockResolvedValue({ topics: [] });
    vi.mocked(supportApi.listTickets).mockResolvedValue({ tickets: [] });
    renderPage();

    expect(await screen.findByText("AI线索任务结果无法导入 CRM")).toBeInTheDocument();
    expect(screen.getByText("会员续费后额度未到账")).toBeInTheDocument();
    expect(screen.getByText("咨询导出报告是否支持自定义模板")).toBeInTheDocument();
  });

  it("filters, searches, and opens ticket details", async () => {
    mockHelp();
    renderPage();

    expect(await screen.findByText("额度没有更新")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("tab", { name: "已关闭" }));
    expect(screen.getByText("历史账号咨询")).toBeInTheDocument();
    expect(screen.queryByText("额度没有更新")).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("tab", { name: "全部" }));
    fireEvent.change(screen.getByLabelText("搜索工单"), { target: { value: "TK-00022" } });
    expect(screen.getByText("额度没有更新")).toBeInTheDocument();
    expect(screen.queryByText("无法修改登录方式")).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /额度没有更新/ }));
    const dialog = screen.getByRole("dialog", { name: "工单详情" });
    expect(within(dialog).getByRole("heading", { name: "额度没有更新" })).toBeInTheDocument();
    expect(within(dialog).getByText("充值后额度仍未更新")).toBeInTheDocument();
    fireEvent.click(within(dialog).getByRole("button", { name: "关闭" }));
    expect(screen.queryByRole("dialog", { name: "工单详情" })).not.toBeInTheDocument();
  });

  it("creates a support ticket from the composer dialog", async () => {
    mockHelp();
    vi.mocked(supportApi.createTicket).mockResolvedValue({
      id: 25,
      topic: "套餐与额度",
      title: "额度没有更新",
      body: "充值后额度仍未更新",
      status: "open",
      created_at: "2026-07-02T10:10:00Z"
    });
    renderPage();

    await screen.findByText("无法修改登录方式");
    fireEvent.click(screen.getByRole("button", { name: "提交工单" }));
    const dialog = screen.getByRole("dialog", { name: "提交工单" });
    fireEvent.change(within(dialog).getByLabelText("问题类型"), { target: { value: "套餐与额度" } });
    fireEvent.change(within(dialog).getByLabelText("问题标题"), { target: { value: "额度没有更新" } });
    fireEvent.change(within(dialog).getByLabelText("问题描述"), { target: { value: "充值后额度仍未更新" } });
    fireEvent.click(within(dialog).getByRole("button", { name: "提交工单" }));

    await waitFor(() => expect(supportApi.createTicket).toHaveBeenCalledWith({
      topic: "套餐与额度",
      title: "额度没有更新",
      body: "充值后额度仍未更新"
    }));
    expect(await screen.findByText("工单 TK-00025 已提交，我们会尽快处理。")).toBeInTheDocument();
    expect(screen.queryByRole("dialog", { name: "提交工单" })).not.toBeInTheDocument();
  });

  it("keeps submission errors inside the composer", async () => {
    mockHelp();
    vi.mocked(supportApi.createTicket).mockRejectedValue(new Error("offline"));
    renderPage();

    await screen.findByText("无法修改登录方式");
    fireEvent.click(screen.getByRole("button", { name: "提交工单" }));
    const dialog = screen.getByRole("dialog", { name: "提交工单" });
    fireEvent.change(within(dialog).getByLabelText("问题标题"), { target: { value: "额度没有更新" } });
    fireEvent.change(within(dialog).getByLabelText("问题描述"), { target: { value: "充值后额度仍未更新" } });
    fireEvent.click(within(dialog).getByRole("button", { name: "提交工单" }));

    expect(await within(dialog).findByRole("alert")).toHaveTextContent("工单提交失败，请稍后重试");
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
  });
});
