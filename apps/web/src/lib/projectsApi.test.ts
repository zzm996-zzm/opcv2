import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "./authSession";
import { projectsApi } from "./projectsApi";

describe("projectsApi", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  it("creates project match with bearer token", async () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-24T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张晨", phone: "", status: "active" }
    });
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ session_id: 99, status: "completed", projects: [] }), { status: 200 })
    );

    await projectsApi.createMatch({ intent: "我想做一人公司项目" });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/projects/matches",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ intent: "我想做一人公司项目" }),
        headers: expect.objectContaining({ Authorization: "Bearer access-token" })
      })
    );
  });

  it("lists project match history", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ matches: [] }), { status: 200 })
    );

    await projectsApi.listMatches();

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/projects/matches",
      expect.objectContaining({ method: "GET" })
    );
  });

  it("lists and gets published opportunities", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ opportunities: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ slug: "ai-sales" }), { status: 200 }));

    await projectsApi.listOpportunities({ query: "AI销售" });
    await projectsApi.getOpportunity("ai-sales");

    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/projects/opportunities?q=AI%E9%94%80%E5%94%AE", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/projects/opportunities/ai-sales", expect.objectContaining({ method: "GET" }));
  });

  it("lists published project cases", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({ cases: [] }), { status: 200 }));
    await projectsApi.listCases({ caseType: "success" });
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/projects/cases?type=success", expect.objectContaining({ method: "GET" }));
  });

  it("gets project match detail", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ id: 99, user_id: 42, intent: "线上项目", status: "completed" }), { status: 200 })
    );

    await projectsApi.getMatch(99);

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/projects/matches/99",
      expect.objectContaining({ method: "GET" })
    );
  });

  it("answers project match follow-up questions", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({ session_id:99, status:"completed" }), { status:200 }));
    await projectsApi.answerMatch(99, [{ key:"background", value:"销售经验" }]);
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/projects/matches/99/answers", expect.objectContaining({ method:"POST", body:JSON.stringify({ answers:[{ key:"background", value:"销售经验" }] }) }));
  });

  it("creates and gets a project comparison", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 61, items: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 61, items: [] }), { status: 200 }));
    await projectsApi.createComparison(["ai-sales", "ai-content"]);
    await projectsApi.getComparison(61);
    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/projects/comparisons", expect.objectContaining({ method: "POST", body: JSON.stringify({ opportunity_slugs: ["ai-sales", "ai-content"] }) }));
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/projects/comparisons/61", expect.objectContaining({ method: "GET" }));
  });

  it("creates a project export", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({ id:71, status:"ready" }), { status:200 }));
    await projectsApi.createExport("match", 99);
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/projects/exports", expect.objectContaining({ method:"POST", body:JSON.stringify({ source_type:"match", source_id:99 }) }));
  });

  it("favorites a project match", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ id: 7, user_id: 42, session_id: 99 }), { status: 200 })
    );

    await projectsApi.favoriteMatch(99);

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/projects/matches/99/favorite",
      expect.objectContaining({ method: "POST" })
    );
  });
});
