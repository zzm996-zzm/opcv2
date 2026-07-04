import { afterEach, describe, expect, it, vi } from "vitest";

import { geoApi } from "./geoApi";

describe("geoApi", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("loads GEO overview", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        stats: [],
        engines: [],
        lead_signals: [],
        keywords: [],
        content_tasks: []
      }), { status: 200 })
    );

    const overview = await geoApi.overview();

    expect(overview.keywords).toEqual([]);
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/geo/overview", expect.any(Object));
  });

  it("creates a GEO analysis request", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        id: 7,
        user_id: 42,
        target: "面向制造业的 AI 质检工具",
        status: "queued",
        created_at: "2026-07-04T08:00:00Z",
        updated_at: "2026-07-04T08:00:00Z"
      }), { status: 200 })
    );

    const request = await geoApi.createAnalysisRequest({ target: "面向制造业的 AI 质检工具" });

    expect(request.status).toBe("queued");
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/geo/analysis-requests", expect.objectContaining({
      method: "POST",
      body: JSON.stringify({ target: "面向制造业的 AI 质检工具" })
    }));
  });

  it("lists GEO analysis requests", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        requests: [{
          id: 7,
          user_id: 42,
          target: "面向制造业的 AI 质检工具",
          status: "queued",
          created_at: "2026-07-04T08:00:00Z",
          updated_at: "2026-07-04T08:00:00Z"
        }]
      }), { status: 200 })
    );

    const payload = await geoApi.listAnalysisRequests(10);

    expect(payload.requests[0].target).toBe("面向制造业的 AI 质检工具");
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/geo/analysis-requests?limit=10", expect.any(Object));
  });

  it("loads a GEO analysis request detail", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        id: 7,
        user_id: 42,
        target: "面向制造业的 AI 质检工具",
        status: "queued",
        created_at: "2026-07-04T08:00:00Z",
        updated_at: "2026-07-04T08:00:00Z"
      }), { status: 200 })
    );

    const request = await geoApi.getAnalysisRequest(7);

    expect(request.id).toBe(7);
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/geo/analysis-requests/7", expect.any(Object));
  });
});
