import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import App from "../App";
import { authSession } from "../lib/authSession";
import LearningDiagnosisPage from "./LearningDiagnosisPage";

function signIn() {
  authSession.set({ access_token: "token", access_token_expires_at: "2026-07-14T00:00:00Z", is_new_user: false, user: { id: 7, nickname: "张晨", phone: "", status: "active" } });
}

describe("LearningDiagnosisPage", () => {
  afterEach(() => { authSession.clear(); vi.restoreAllMocks(); });

  it("renders truthful diagnosis inputs and source boundaries", () => {
    signIn();
    render(<MemoryRouter><LearningDiagnosisPage /></MemoryRouter>);
    expect(screen.getByRole("heading", { name: "能力诊断" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "诊断依据说明" })).toBeInTheDocument();
    expect(screen.getByLabelText("目标项目或应用场景")).toBeInTheDocument();
    expect(screen.getByText(/不会声称读取未接入的数据/)).toBeInTheDocument();
  });

  it("submits an assessment and routes with the persisted result", async () => {
    signIn();
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      if (String(input) === "/api/v1/learning/assessments" && init?.method === "POST") {
        return Promise.resolve(new Response(JSON.stringify({
          id: 77, user_id: 7, goal: "拓展业务视野", project: "智能客服试点", status: "completed", overall_score: 84,
          dimensions: [{ name: "数据洞察能力", score: 84, gap: 10, summary: "模型评估摘要" }], recommendations: [],
          basis: "model_assessment", disclaimer: "模型评估，不是能力认证。", assumptions: ["基于用户自述"],
          evidence_sources: [{ type: "assessment_input", label: "用户本次提交", captured_at: "2026-07-13T08:00:00Z" }],
          answers: [], focus_abilities: ["数据洞察能力"], weekly_time: "5-8 小时", bottleneck: "缺少案例", created_at: "2026-07-13T08:00:00Z", updated_at: "2026-07-13T08:00:00Z"
        }), { status: 201 }));
      }
      return Promise.reject(new Error(`unexpected request: ${String(input)}`));
    });
    render(<MemoryRouter initialEntries={["/learning/diagnosis"]}><App /></MemoryRouter>);
    fireEvent.change(screen.getByLabelText("目标项目或应用场景"), { target: { value: "智能客服试点" } });
    fireEvent.click(screen.getByRole("button", { name: "拓展业务视野" }));
    fireEvent.change(screen.getByLabelText("希望提升的能力"), { target: { value: "数据洞察能力" } });
    fireEvent.click(screen.getByRole("button", { name: "5-8 小时" }));
    fireEvent.change(screen.getByLabelText("当前最大卡点"), { target: { value: "缺少案例" } });
    fireEvent.click(screen.getByRole("link", { name: /提交并生成模型评估/ }));
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/learning/assessments", expect.objectContaining({ method: "POST", body: expect.stringContaining('"project":"智能客服试点"') })));
    expect(await screen.findByText("84/100")).toBeInTheDocument();
    expect(screen.getByText("模型评估摘要")).toBeInTheDocument();
  });
});
