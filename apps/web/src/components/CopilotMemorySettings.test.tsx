import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import CopilotMemorySettings from "./CopilotMemorySettings";

describe("CopilotMemorySettings", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("loads, saves and deletes long-term memories", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation(async (input, init) => {
      const url = String(input);
      if (url === "/api/v1/copilot/memories?limit=4") {
        return jsonResponse({ memories: [memory(8, "行业", "企业服务")] });
      }
      if (url === "/api/v1/copilot/memories" && init?.method === "POST") {
        return jsonResponse(memory(9, "回答偏好", "先给结论"));
      }
      if (url === "/api/v1/copilot/memories/8" && init?.method === "DELETE") {
        return new Response(null, { status: 204 });
      }
      return jsonResponse({ error: "not_found" }, 404);
    });

    render(<MemoryRouter><CopilotMemorySettings /></MemoryRouter>);

    expect(await screen.findByText("企业服务")).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("记忆名称"), { target: { value: "回答偏好" } });
    fireEvent.change(screen.getByLabelText("记忆内容"), { target: { value: "先给结论" } });
    fireEvent.click(screen.getByRole("button", { name: "保存记忆" }));
    expect(await screen.findByText("先给结论")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "删除记忆 行业" }));
    await waitFor(() => expect(screen.queryByText("企业服务")).not.toBeInTheDocument());
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/copilot/memories/8",
      expect.objectContaining({ method: "DELETE" })
    );
  });
});

function memory(id: number, key: string, value: string) {
  return {
    id,
    user_id: 7,
    key,
    value,
    confidence: 1,
    source: "manual",
    created_at: "2026-08-20T08:00:00Z",
    updated_at: "2026-08-20T08:00:00Z"
  };
}

function jsonResponse(payload: unknown, status = 200) {
  return new Response(JSON.stringify(payload), { status, headers: { "Content-Type": "application/json" } });
}
