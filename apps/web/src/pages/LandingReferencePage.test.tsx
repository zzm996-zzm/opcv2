import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import LandingReferencePage, { type LandingModule, type LandingView } from "./LandingReferencePage";

function renderReference(module: LandingModule, view: LandingView, route = "/") {
  return render(
    <MemoryRouter initialEntries={[route]}>
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

  it("sends the active competitor scan context to Copilot", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation(async (input, init) => {
      const url = String(input);
      if (url === "/api/v1/competitor/scans/13") {
        return jsonResponse({
          id: 13, user_id: 7, targets: ["小鹅通"], focus: "产品能力", status: "running",
          progress_percent: 60, current_step: "analyzing", competitors: [], conclusions: [], evidence_sources: [],
          created_at: "2026-08-20T08:00:00Z", updated_at: "2026-08-20T08:01:00Z"
        });
      }
      if (url === "/api/v1/copilot/threads" && init?.method === "POST") {
        return jsonResponse({ id: 99, user_id: 7, title: "竞品分析", mode: "chat", model: "deepseek-chat", created_at: "2026-08-20T08:00:00Z", updated_at: "2026-08-20T08:00:00Z" });
      }
      if (url === "/api/v1/copilot/threads/99/messages" && init?.method === "POST") {
        return jsonResponse({
          user_message: { id: 1, user_id: 7, thread_id: 99, role: "user", content: "分析当前进度", status: "completed", created_at: "2026-08-20T08:00:01Z" },
          assistant_message: { id: 2, user_id: 7, thread_id: 99, role: "assistant", content: "当前正在分析产品能力。", status: "completed", created_at: "2026-08-20T08:00:02Z" }
        });
      }
      return jsonResponse({ error: "not_found" }, 404);
    });

    renderReference("data", "progress", "/competitor-data/progress?scanId=13");
    fireEvent.change(screen.getByLabelText("询问落地 Copilot"), { target: { value: "分析当前进度" } });
    fireEvent.click(screen.getByRole("button", { name: "发送" }));

    expect(await screen.findByText("当前正在分析产品能力。")).toBeInTheDocument();
    await waitFor(() => {
      const sendCall = fetchMock.mock.calls.find(([input]) => String(input) === "/api/v1/copilot/threads/99/messages");
      expect(JSON.parse(String(sendCall?.[1]?.body))).toEqual(expect.objectContaining({
        current_view: "/competitor-data/progress?scanId=13",
        active_filters: { module: "data", view: "progress", scan_id: "13" }
      }));
    });
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

function jsonResponse(payload: unknown, status = 200) {
  return new Response(JSON.stringify(payload), { status, headers: { "Content-Type": "application/json" } });
}
