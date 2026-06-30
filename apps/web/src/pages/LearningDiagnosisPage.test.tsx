import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import App from "../App";
import { authSession } from "../lib/authSession";
import LearningDiagnosisPage from "./LearningDiagnosisPage";

describe("LearningDiagnosisPage", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  function signIn() {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-11T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张晨", phone: "13800138000", status: "active" }
    });
  }

  it("renders the diagnosis flow, connected data and start action", () => {
    signIn();

    render(
      <MemoryRouter>
        <LearningDiagnosisPage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "能力诊断" })).toBeInTheDocument();
    expect(screen.getByText("收集信息")).toBeInTheDocument();
    expect(screen.getByText("已接入分析的数据源")).toBeInTheDocument();
    expect(screen.getByText("本次诊断将分析什么")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /开始能力诊断/ })).toHaveAttribute("href", "/learning/assessment");
  });

  it("creates a diagnosis and routes to the assessment result", async () => {
    signIn();
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
      const url = String(input);
      if (url === "/api/v1/learning/diagnoses" && init?.method === "POST") {
        return Promise.resolve(new Response(JSON.stringify({
          id: 77,
          user_id: 7,
          goal: "提升专业能力",
          project: "智能客服与市场分析",
          status: "completed",
          overall_score: 84,
          dimensions: [
            { name: "智能客服方案设计", score: 84, gap: 10, summary: "方案设计能力稳定" }
          ],
          recommendations: ["优先补齐智能客服案例拆解"],
          created_at: "2026-06-30T08:00:00Z",
          updated_at: "2026-06-30T08:02:00Z"
        }), { status: 200 }));
      }
      if (url === "/api/v1/learning/diagnoses/latest" && init?.method === "GET") {
        return Promise.resolve(new Response(JSON.stringify({
          id: 77,
          user_id: 7,
          goal: "提升专业能力",
          project: "智能客服与市场分析",
          status: "completed",
          overall_score: 84,
          dimensions: [
            { name: "智能客服方案设计", score: 84, gap: 10, summary: "方案设计能力稳定" }
          ],
          recommendations: ["优先补齐智能客服案例拆解"],
          created_at: "2026-06-30T08:00:00Z",
          updated_at: "2026-06-30T08:02:00Z"
        }), { status: 200 }));
      }
      return Promise.reject(new Error(`unexpected request: ${url}`));
    });

    render(
      <MemoryRouter initialEntries={["/learning/diagnosis"]}>
        <App />
      </MemoryRouter>
    );
    fireEvent.click(screen.getByRole("link", { name: /开始能力诊断/ }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        "/api/v1/learning/diagnoses",
        expect.objectContaining({
          method: "POST",
          body: JSON.stringify({
            goal: "提升专业能力",
            project: "智能客服与市场分析"
          })
        })
      );
    });
    expect(await screen.findAllByText("84%")).not.toHaveLength(0);
    expect(screen.getByText("方案设计能力稳定")).toBeInTheDocument();
  });
});
