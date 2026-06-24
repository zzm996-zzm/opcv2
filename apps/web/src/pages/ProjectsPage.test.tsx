import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import App from "../App";
import { authSession } from "../lib/authSession";

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

function renderProjectRoute(path: string) {
  signIn();
  render(
    <MemoryRouter initialEntries={[path]}>
      <App />
    </MemoryRouter>
  );
}

describe("ProjectsPage", () => {
  afterEach(() => {
    authSession.clear();
  });

  it("renders the project market home from the design reference", () => {
    renderProjectRoute("/projects");

    expect(screen.getByRole("heading", { name: "项目超市" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "发现下一个可落地机会" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "去匹配" })).toHaveAttribute("href", "/projects/match");
    expect(screen.getByRole("heading", { name: "精选机会" })).toBeInTheDocument();
    expect(screen.getByRole("complementary", { name: "智活 Copilot" })).toBeInTheDocument();
  });

  it("renders the AI matching request page", () => {
    renderProjectRoute("/projects/match");

    expect(screen.getByRole("heading", { name: "AI匹配" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "告诉我你的目标、资源与偏好" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "提交给 AI 分析" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "匹配历史" })).toHaveAttribute("href", "/projects/history");
  });

  it("renders opportunity exploration page", () => {
    renderProjectRoute("/projects/explore");

    expect(screen.getByRole("heading", { name: "机会探索" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "高潜力机会" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "AI视频矩阵" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "机会雷达" })).toBeInTheDocument();
  });

  it("renders real case library page", () => {
    renderProjectRoute("/projects/cases");

    expect(screen.getByRole("heading", { name: "真实案例库" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "成功样板" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Excel自动化顾问" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "案例共性" })).toBeInTheDocument();
  });

  it("renders the AI follow-up questions page", () => {
    renderProjectRoute("/projects/questions");

    expect(screen.getByRole("heading", { name: "AI补充提问" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Copilot 还想确认以下问题" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "生成匹配结果" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "服务型" })).toBeInTheDocument();
  });

  it("renders matching results with ranked opportunities", () => {
    renderProjectRoute("/projects/results");

    expect(screen.getByRole("heading", { name: "为你匹配到的项目机会" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "AI短视频脚本工作室" })).toBeInTheDocument();
    expect(screen.getAllByRole("button", { name: /加入对比/ }).length).toBeGreaterThan(0);
    expect(screen.getByRole("heading", { name: "为什么推荐这些项目" })).toBeInTheDocument();
  });

  it("renders history and saved matches", () => {
    renderProjectRoute("/projects/history");

    expect(screen.getByRole("heading", { name: "匹配历史与收藏" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "历史匹配" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "收藏项目" })).toBeInTheDocument();
  });

  it("renders paid sample overlay state", () => {
    renderProjectRoute("/projects/results/paywall");

    expect(screen.getByRole("heading", { name: "解锁完整拆解" })).toBeInTheDocument();
    expect(screen.getByText("智活AI会员")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "立即解锁" })).toBeInTheDocument();
  });

  it("renders project detail tabs and diagnostics", () => {
    renderProjectRoute("/projects/detail");

    expect(screen.getByRole("heading", { name: "AI短视频脚本工作室" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "诊断是否能做" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "成功路径" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "当前数据" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "真实案例库" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "优劣势" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "可学经验" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "要避免行为" })).toBeInTheDocument();
    expect(screen.getByText("综合可做度 81分")).toBeInTheDocument();
  });

  it("renders project comparison and export modal states", () => {
    renderProjectRoute("/projects/compare");

    expect(screen.getByRole("heading", { name: "项目对比" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "AI短视频脚本工作室" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "导出对比报告" })).toBeInTheDocument();

    authSession.clear();
    renderProjectRoute("/projects/export");

    expect(screen.getByRole("heading", { name: "导出匹配报告" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "确认导出" })).toBeInTheDocument();
  });
});
