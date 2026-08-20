import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import LandingReferencePage, { type LandingModule, type LandingView } from "./LandingReferencePage";

function renderReference(module: LandingModule, view: LandingView) {
  return render(
    <MemoryRouter>
      <LandingReferencePage module={module} view={view} />
    </MemoryRouter>
  );
}

describe("LandingReferencePage", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("renders the task list reference state", () => {
    renderReference("tasks", "list");

    expect(screen.getByText("任务中心", { selector: ".landing-ref-h1" })).toBeInTheDocument();
    expect(screen.getByPlaceholderText("搜索任务、负责人、进度阶段、标签")).toBeInTheDocument();
  });

  it("hides and restores the complete Copilot panel", () => {
    const { container } = renderReference("tasks", "list");

    fireEvent.click(screen.getByRole("button", { name: "收起智活 Copilot" }));
    expect(container.querySelector(".landing-ref-layout")).toHaveClass("copilot-collapsed");
    expect(screen.queryByRole("complementary", { name: "智活 Copilot" })).not.toBeInTheDocument();
    expect(screen.queryByPlaceholderText("询问任何问题...")).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "打开智活 Copilot" }));
    expect(container.querySelector(".landing-ref-layout")).not.toHaveClass("copilot-collapsed");
    expect(screen.getByPlaceholderText("询问任何问题...")).toBeInTheDocument();
  });

  it("renders the task board reference state", () => {
    renderReference("tasks", "board");

    expect(screen.getByText("任务视图", { selector: ".landing-ref-h1" })).toBeInTheDocument();
    expect(screen.getByText("评审中", { selector: ".board-col > header" })).toBeInTheDocument();
  });

  it("renders the competitor query progress state", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({
      scans: [{
        id: 13,
        user_id: 7,
        targets: ["小鹅通"],
        focus: "产品能力",
        status: "running",
        progress_percent: 60,
        current_step: "analyzing",
        competitors: [],
        conclusions: [],
        evidence_sources: [],
        created_at: "2026-08-20T08:00:00Z",
        updated_at: "2026-08-20T08:01:00Z"
      }]
    }), { status: 200 }));

    renderReference("data", "progress");

    expect(screen.getByText("查询处理中", { selector: ".landing-ref-h1" })).toBeInTheDocument();
    expect(await screen.findByText("60%")).toBeInTheDocument();
    expect(screen.getAllByText("AI 分析中").length).toBeGreaterThan(0);
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
