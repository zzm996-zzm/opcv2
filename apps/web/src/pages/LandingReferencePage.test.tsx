import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";

import LandingReferencePage, { type LandingModule, type LandingView } from "./LandingReferencePage";

function renderReference(module: LandingModule, view: LandingView) {
  return render(
    <MemoryRouter>
      <LandingReferencePage module={module} view={view} />
    </MemoryRouter>
  );
}

describe("LandingReferencePage", () => {
  it("renders the task list reference state", () => {
    renderReference("tasks", "list");

    expect(screen.getByText("任务中心", { selector: ".landing-ref-h1" })).toBeInTheDocument();
    expect(screen.getByPlaceholderText("搜索任务、负责人、进度阶段、标签")).toBeInTheDocument();
  });

  it("renders the task board reference state", () => {
    renderReference("tasks", "board");

    expect(screen.getByText("任务视图", { selector: ".landing-ref-h1" })).toBeInTheDocument();
    expect(screen.getByText("评审中", { selector: ".board-col > header" })).toBeInTheDocument();
  });

  it("renders the competitor query progress state", () => {
    renderReference("data", "progress");

    expect(screen.getByText("查询处理中", { selector: ".landing-ref-h1" })).toBeInTheDocument();
    expect(screen.getByText("38%")).toBeInTheDocument();
  });

  it("renders a competitor result section", () => {
    renderReference("data", "audience");

    expect(screen.getByText("竞品账号全盘数据结果", { exact: false, selector: ".landing-ref-h1" })).toBeInTheDocument();
    expect(screen.getByText("粉丝性别分布")).toBeInTheDocument();
  });

  it("renders the monitoring analysis state", () => {
    renderReference("monitoring", "analysis");

    expect(screen.getByText("完美日记官方旗舰店 动态监测结果", { exact: false, selector: ".landing-ref-h1" })).toBeInTheDocument();
    expect(screen.getByText("重点动态摘要")).toBeInTheDocument();
  });

  it("renders the growth report state", () => {
    renderReference("growth", "report");

    expect(screen.getByText("增长测算报告", { exact: false, selector: ".landing-ref-h1" })).toBeInTheDocument();
    expect(screen.getByText("90天行动计划")).toBeInTheDocument();
  });
});
