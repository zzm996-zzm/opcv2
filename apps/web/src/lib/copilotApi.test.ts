import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "./authSession";
import { copilotApi } from "./copilotApi";

describe("copilotApi", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  it("creates a copilot thread with bearer token", async () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-30T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张晨", phone: "", status: "active" }
    });
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ id: 99, title: "新会话", mode: "chat" }), { status: 200 })
    );

    await copilotApi.createThread({ title: "新会话" });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/copilot/threads",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ title: "新会话" }),
        headers: expect.objectContaining({ Authorization: "Bearer access-token" })
      })
    );
  });

  it("sends a message to a thread", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ user_message: {}, assistant_message: {} }), { status: 200 })
    );

    await copilotApi.sendMessage(99, { content: "帮我分析机会", model: "gpt-test", reference_ids: [17] });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/copilot/threads/99/messages",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ content: "帮我分析机会", model: "gpt-test", reference_ids: [17] })
      })
    );
  });

  it("compares a message across selected models", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ user_message: {}, answers: [] }), { status: 200 })
    );

    await copilotApi.compareMessages(99, { content: "分析机会", models: ["deepseek", "gpt-main"] });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/copilot/threads/99/compare",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ content: "分析机会", models: ["deepseek", "gpt-main"] })
      })
    );
  });

  it("summarizes comparison answers", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ summary_message: {} }), { status: 200 })
    );

    await copilotApi.summarizeComparison(99, {
      content: "分析机会",
      model: "deepseek",
      answers: [{
        model: "deepseek",
        assistant_message: {
          id: 11,
          user_id: 7,
          thread_id: 99,
          role: "assistant",
          content: "先做验证。",
          status: "completed",
          model: "deepseek",
          created_at: "2026-07-01T09:51:00Z"
        }
      }]
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/copilot/threads/99/compare/summary",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({
          content: "分析机会",
          model: "deepseek",
          answers: [{
            model: "deepseek",
            assistant_message: {
              id: 11,
              user_id: 7,
              thread_id: 99,
              role: "assistant",
              content: "先做验证。",
              status: "completed",
              model: "deepseek",
              created_at: "2026-07-01T09:51:00Z"
            }
          }]
        })
      })
    );
  });

  it("lists configured copilot models", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ models: [{ name: "DeepSeek", value: "deepseek", is_default: true }] }), { status: 200 })
    );

    await copilotApi.listModels();

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/copilot/models",
      expect.objectContaining({ method: "GET" })
    );
  });

  it("smoke tests a configured copilot model", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ ok: true, model: "deepseek", reply: "pong" }), { status: 200 })
    );

    await copilotApi.smokeModel({ model: "deepseek", prompt: "ping" });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/copilot/models/smoke",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ model: "deepseek", prompt: "ping" })
      })
    );
  });

  it("lists recent copilot AI runs", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ runs: [{ id: 11, feature: "copilot.model_smoke", status: "failed" }] }), { status: 200 })
    );

    await copilotApi.listAIRuns(20);

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/copilot/ai-runs?limit=20",
      expect.objectContaining({ method: "GET" })
    );
  });

  it("lists and saves structured memories", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ memories: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 7, key: "industry", value: "教培" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(null, { status: 204 }));

    await copilotApi.listMemories();
    await copilotApi.saveMemory({ key: "industry", value: "教培", confidence: 0.9, source: "manual" });
    await copilotApi.deleteMemory(7);

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "/api/v1/copilot/memories?limit=50",
      expect.objectContaining({ method: "GET" })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "/api/v1/copilot/memories",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ key: "industry", value: "教培", confidence: 0.9, source: "manual" })
      })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      3,
      "/api/v1/copilot/memories/7",
      expect.objectContaining({ method: "DELETE" })
    );
  });

  it("lists and uploads copilot reference files", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ files: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 17, name: "竞品对比.txt" }), { status: 200 }));

    await copilotApi.listFiles();
    await copilotApi.saveFile({ name: "竞品对比.txt", mime_type: "text/plain", content: "小鹅通：私域工具强。" });

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "/api/v1/copilot/files?limit=50",
      expect.objectContaining({ method: "GET" })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "/api/v1/copilot/files",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ name: "竞品对比.txt", mime_type: "text/plain", content: "小鹅通：私域工具强。" })
      })
    );
  });
});
