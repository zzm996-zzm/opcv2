import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "./authSession";
import { sandboxApi } from "./sandboxApi";

describe("sandboxApi", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  it("creates and runs a sandbox session with bearer token", async () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-30T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张晨", phone: "", status: "active" }
    });
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 99, status: "draft" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 99, status: "completed", report: { score: 83 } }), { status: 200 }));

    await sandboxApi.createSession({ goal: "验证项目", targetUsers: "教培机构", product: "AI客服", roles: ["用户"] });
    await sandboxApi.runSession(99);

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "/api/v1/sandbox/sessions",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ goal: "验证项目", target_users: "教培机构", product: "AI客服", roles: ["用户"] }),
        headers: expect.objectContaining({ Authorization: "Bearer access-token" })
      })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "/api/v1/sandbox/sessions/99/run",
      expect.objectContaining({ method: "POST" })
    );
  });

  it("gets a sandbox session by id", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ id: 99, status: "completed" }), { status: 200 })
    );

    await sandboxApi.getSession(99);

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/sandbox/sessions/99",
      expect.objectContaining({ method: "GET" })
    );
  });

  it("gets status, retries, and cancels a sandbox session", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockImplementation(() => Promise.resolve(
        new Response(JSON.stringify({ id: 99, status: "queued" }), { status: 200 })
      ));

    await sandboxApi.getStatus(99);
    await sandboxApi.retrySession(99);
    await sandboxApi.cancelSession(99);

    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/sandbox/sessions/99/status", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/sandbox/sessions/99/retry", expect.objectContaining({ method: "POST" }));
    expect(fetchMock).toHaveBeenNthCalledWith(3, "/api/v1/sandbox/sessions/99/cancel", expect.objectContaining({ method: "POST" }));
  });

  it("loads roles and updates a sandbox draft", async () => {
    const settings = {
      depth: "deep" as const,
      output_style: "structured_report" as const,
      generate_outline: true,
      variables: { pricing: "39 元/月" }
    };
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ roles: [{ key: "user", label: "用户视角" }] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 99, roles: ["用户视角"] }), { status: 200 }));

    await sandboxApi.listRoles();
    await sandboxApi.updateDraft(99, { roles: ["用户视角"], settings });

    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/sandbox/roles", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/sandbox/sessions/99/draft", expect.objectContaining({
      method: "PATCH",
      body: JSON.stringify({ roles: ["用户视角"], settings })
    }));
  });

  it("loads sandbox options and completes the intake flow", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation(() => Promise.resolve(
      new Response(JSON.stringify({ id: 99, status: "draft" }), { status: 200 })
    ));

    await sandboxApi.getOptions();
    await sandboxApi.createIntake("我想做一款面向上班族的低卡代餐奶昔");
    await sandboxApi.answerIntakeQuestion(99, "target/user", { answer: "25-35 岁白领", skipped: false });
    await sandboxApi.completeIntake(99);

    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/sandbox/options", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/sandbox/sessions/intake", expect.objectContaining({
      method: "POST",
      body: JSON.stringify({ initial_idea: "我想做一款面向上班族的低卡代餐奶昔" })
    }));
    expect(fetchMock).toHaveBeenNthCalledWith(3, "/api/v1/sandbox/sessions/99/intake/questions/target%2Fuser", expect.objectContaining({
      method: "PUT",
      body: JSON.stringify({ answer: "25-35 岁白领", skipped: false })
    }));
    expect(fetchMock).toHaveBeenNthCalledWith(4, "/api/v1/sandbox/sessions/99/intake/complete", expect.objectContaining({ method: "POST" }));
  });

  it("lists and creates role follow-up messages", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ messages: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 1, role: "投资人视角", answer: "关注留存" }), { status: 200 }));

    await sandboxApi.listMessages(99);
    await sandboxApi.askRole(99, { role: "投资人视角", question: "最关注什么？" });

    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/sandbox/sessions/99/messages", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/sandbox/sessions/99/messages", expect.objectContaining({
      method: "POST",
      body: JSON.stringify({ role: "投资人视角", question: "最关注什么？" })
    }));
  });
});
