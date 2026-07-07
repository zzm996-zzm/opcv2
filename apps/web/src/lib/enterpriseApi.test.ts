import { afterEach, describe, expect, it, vi } from "vitest";

import { enterpriseApi } from "./enterpriseApi";

describe("enterpriseApi", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("loads enterprise overview", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        stats: [],
        plans: [],
        delivery_board: [],
        milestones: [],
        cases: []
      }), { status: 200 })
    );

    const overview = await enterpriseApi.overview();

    expect(overview.plans).toEqual([]);
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/enterprise/overview", expect.any(Object));
  });

  it("creates an enterprise diagnosis request", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ id: 7, user_id: 42, need: "30人销售团队需要AI获客陪跑", status: "submitted" }), { status: 200 })
    );

    const request = await enterpriseApi.createDiagnosisRequest({ need: "30人销售团队需要AI获客陪跑" });

    expect(request.status).toBe("submitted");
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/enterprise/diagnosis-requests", expect.objectContaining({
      method: "POST",
      body: JSON.stringify({ need: "30人销售团队需要AI获客陪跑" })
    }));
  });

  it("lists enterprise diagnosis requests", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ requests: [{ id: 7, user_id: 42, need: "30人销售团队需要AI获客陪跑", status: "submitted" }] }), { status: 200 })
    );

    const response = await enterpriseApi.listDiagnosisRequests(5);

    expect(response.requests).toHaveLength(1);
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/enterprise/diagnosis-requests?limit=5", expect.any(Object));
  });

  it("updates an enterprise diagnosis request", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ id: 7, user_id: 42, need: "30人销售团队需要AI获客陪跑", status: "follow_up_created" }), { status: 200 })
    );

    const request = await enterpriseApi.updateDiagnosisRequest(7, { status: "follow_up_created" });

    expect(request.status).toBe("follow_up_created");
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/enterprise/diagnosis-requests/7", expect.objectContaining({
      method: "PATCH",
      body: JSON.stringify({ status: "follow_up_created" })
    }));
  });

  it("completes an enterprise diagnosis request", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ id: 7, user_id: 42, need: "30人销售团队需要AI获客陪跑", status: "completed" }), { status: 200 })
    );

    const request = await enterpriseApi.updateDiagnosisRequest(7, { status: "completed" });

    expect(request.status).toBe("completed");
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/enterprise/diagnosis-requests/7", expect.objectContaining({
      method: "PATCH",
      body: JSON.stringify({ status: "completed" })
    }));
  });
});
