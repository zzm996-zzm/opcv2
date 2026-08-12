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

  it("uploads project match files as multipart data without forcing a JSON content type", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ id: 501, name: "需求.txt", parse_status: "ready" }), { status: 201 })
    );
    const file = new File(["预算 3 万元"], "需求.txt", { type: "text/plain" });

    await projectsApi.uploadProjectMatchFile(file);

    expect(fetchMock).toHaveBeenCalledOnce();
    const [path, init] = fetchMock.mock.calls[0];
    expect(path).toBe("/api/v1/project-match-files");
    expect(init?.method).toBe("POST");
    expect(init?.body).toBeInstanceOf(FormData);
    expect((init?.body as FormData).get("file")).toEqual(file);
    expect(new Headers(init?.headers).has("Content-Type")).toBe(false);
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

  it("loads the V1.4 project home, catalog, and detail contracts", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ featured: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ items: [], page: 2, page_size: 8, total: 0 }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ slug: "ai-sales" }), { status: 200 }));

    await projectsApi.getHome();
    await projectsApi.listProjects({
      keyword: "AI销售",
      category: "service",
      track: "ai",
      budget: "1-3万",
      difficulty: "中等",
      resource: "销售经验",
      sort: "latest",
      isFeatured: true,
      page: 2,
      pageSize: 8
    });
    await projectsApi.getProject("ai-sales");

    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/projects/home", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/projects?keyword=AI%E9%94%80%E5%94%AE&category=service&track=ai&budget=1-3%E4%B8%87&difficulty=%E4%B8%AD%E7%AD%89&resource=%E9%94%80%E5%94%AE%E7%BB%8F%E9%AA%8C&sort=latest&is_featured=true&page=2&page_size=8", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(3, "/api/v1/projects/ai-sales", expect.objectContaining({ method: "GET" }));
  });

  it("loads public config and manages persisted project collections", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ feature_paywall_enabled: false }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ favorites: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ project_id: 42, slug: "ai-sales", title: "AI销售" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(null, { status: 204 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ items: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ project_id: 42, slug: "ai-sales", title: "AI销售" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(null, { status: 204 }));

    await projectsApi.getPublicConfig();
    await projectsApi.listProjectFavorites();
    await projectsApi.favoriteProject("ai-sales");
    await projectsApi.unfavoriteProject("ai-sales");
    await projectsApi.listProjectCompareItems();
    await projectsApi.addProjectCompareItem("ai-sales");
    await projectsApi.removeProjectCompareItem("ai-sales");

    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/config", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/projects/project-favorites", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(3, "/api/v1/projects/ai-sales/favorite", expect.objectContaining({ method: "POST" }));
    expect(fetchMock).toHaveBeenNthCalledWith(4, "/api/v1/projects/ai-sales/favorite", expect.objectContaining({ method: "DELETE" }));
    expect(fetchMock).toHaveBeenNthCalledWith(5, "/api/v1/projects/compare", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(6, "/api/v1/projects/compare-items", expect.objectContaining({ method: "POST", body: JSON.stringify({ project_id: "ai-sales" }) }));
    expect(fetchMock).toHaveBeenNthCalledWith(7, "/api/v1/projects/compare-items/ai-sales", expect.objectContaining({ method: "DELETE" }));
  });

  it("lists and gets V1.4 evidence cases", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ items: [], page: 2, page_size: 10, total: 0 }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 81, title: "AI销售试点" }), { status: 200 }));

    await projectsApi.listEvidenceCases({ caseType: "failure", industry: "AI", scale: "solo", page: 2, pageSize: 10 });
    await projectsApi.getEvidenceCase("81");

    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/project-cases?type=failure&industry=AI&scale=solo&page=2&page_size=10", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/project-cases/81", expect.objectContaining({ method: "GET" }));
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

  it("downloads a project export with bearer authentication", async () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-24T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张晨", phone: "", status: "active" }
    });
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ source_type: "match", source_id: 99 }), { status: 200, headers: { "Content-Type": "application/json" } })
    );

    const blob = await projectsApi.downloadExport(71);

    expect(await blob.text()).toContain('"source_id":99');
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/projects/exports/71/download",
      expect.objectContaining({
        method: "GET",
        headers: expect.objectContaining({ Authorization: "Bearer access-token" })
      })
    );
  });

  it("lists, favorites, and unfavorites project matches", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ favorites: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 7, user_id: 42, session_id: 99 }), { status: 200 }))
      .mockResolvedValueOnce(new Response(null, { status: 204 }));

    await projectsApi.listFavorites();
    await projectsApi.favoriteMatch(99);
    await projectsApi.unfavoriteMatch(99);

    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/projects/favorites", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(2,
      "/api/v1/projects/matches/99/favorite",
      expect.objectContaining({ method: "POST" })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(3, "/api/v1/projects/matches/99/favorite", expect.objectContaining({ method: "DELETE" }));
  });
});
