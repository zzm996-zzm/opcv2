import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import LearningPlanPage from "./LearningPlanPage";

const planPayload = {
  diagnosis_id: 99, title: "企业AI落地能力路径", description: "基于诊断快照生成", recommendations: ["先完成业务场景拆解"],
  stages: [{ number: 1, title: "业务场景拆解", status: "not_started", courses: [], duration: "", goal: "建立拆解框架", milestone: "完成项目拆解" }],
  items: [{ id: 0, user_id: 7, diagnosis_id: 99, stage_number: 1, title: "业务场景拆解", completed: false, updated_at: "0001-01-01T00:00:00Z" }],
  estimated_hours: 0, weekly_suggestion: "每周 3 小时", basis: "model_assessment", disclaimer: "模型评估说明", assumptions: ["基于用户自述"], evidence_sources: [], generated_at: "2026-07-13T08:00:00Z"
};

function signIn() {
  authSession.set({ access_token: "token", access_token_expires_at: "2026-07-14T00:00:00Z", is_new_user: false, user: { id: 7, nickname: "张婧", phone: "", status: "active" } });
}

describe("LearningPlanPage", () => {
  afterEach(() => { authSession.clear(); vi.restoreAllMocks(); });

  it("loads a persisted learning path without static fallback", async () => {
    signIn();
    vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify(planPayload), { status: 200 }));
    render(<MemoryRouter><LearningPlanPage /></MemoryRouter>);
    expect(await screen.findByRole("heading", { name: "企业AI落地能力路径" })).toBeInTheDocument();
    expect(screen.getByText("业务场景拆解")).toBeInTheDocument();
    expect(screen.getByText("尚未匹配课程目录")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "标记为已完成" })).toBeInTheDocument();
  });

  it("persists stage completion and reloads the plan", async () => {
    signIn();
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/learning/diagnoses/latest/plan") return Promise.resolve(new Response(JSON.stringify(planPayload), { status: 200 }));
      if (url === "/api/v1/learning/diagnoses/99/plan/items/1" && init?.method === "PUT") return Promise.resolve(new Response(JSON.stringify({ ...planPayload.items[0], id: 77, completed: true }), { status: 200 }));
      if (url === "/api/v1/learning/diagnoses/99/plan") return Promise.resolve(new Response(JSON.stringify({ ...planPayload, stages: [{ ...planPayload.stages[0], status: "completed" }], items: [{ ...planPayload.items[0], id: 77, completed: true }] }), { status: 200 }));
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });
    render(<MemoryRouter><LearningPlanPage /></MemoryRouter>);
    fireEvent.click(await screen.findByRole("button", { name: "标记为已完成" }));
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/learning/diagnoses/99/plan/items/1", expect.objectContaining({ method: "PUT", body: JSON.stringify({ completed: true }) })));
    expect(await screen.findByRole("button", { name: "标记为未完成" })).toBeInTheDocument();
  });

  it("generates task-center tasks from the loaded plan", async () => {
    signIn();
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      if (String(input) === "/api/v1/learning/diagnoses/latest/plan") return Promise.resolve(new Response(JSON.stringify(planPayload), { status: 200 }));
      if (String(input) === "/api/v1/tasks/generate" && init?.method === "POST") return Promise.resolve(new Response(JSON.stringify({ tasks: [{ id: 501 }] }), { status: 200 }));
      return Promise.reject(new Error(`unexpected request: ${String(input)}`));
    });
    render(<MemoryRouter><LearningPlanPage /></MemoryRouter>);
    fireEvent.click(await screen.findByRole("button", { name: "同步到任务中心" }));
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/tasks/generate", expect.objectContaining({ method: "POST", body: expect.stringContaining('"source_type":"learning_diagnosis"') })));
    expect(await screen.findByRole("status")).toHaveTextContent("已创建 1 个学习任务");
  });
});
