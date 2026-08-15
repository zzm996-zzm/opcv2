import { afterEach, describe, expect, it, vi } from "vitest";

import { growthApi } from "./growthApi";

describe("growthApi", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("creates and lists growth models", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 99, name: "标准方案" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ models: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ scenarios: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ months: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ action_items: [] }), { status: 200 }));

    await growthApi.createModel({
      name: "标准方案",
      monthlyVisits: 24000,
      leadRate: 0.068,
      dealRate: 0.14,
      averageOrder: 820,
      acquisitionCost: 42,
      deliveryCost: 51000
    });
    await growthApi.listModels();
    await growthApi.modelScenarios(99);
    await growthApi.modelForecast(99);
    await growthApi.modelRecommendations(99);

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "/api/v1/growth/models",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({
          name: "标准方案",
          monthly_visits: 24000,
          lead_rate: 0.068,
          deal_rate: 0.14,
          average_order: 820,
          acquisition_cost: 42,
          delivery_cost: 51000
        })
      })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/growth/models", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(3, "/api/v1/growth/models/99/scenarios", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(4, "/api/v1/growth/models/99/forecast", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(5, "/api/v1/growth/models/99/recommendations", expect.objectContaining({ method: "GET" }));
  });

  it("creates, answers and calculates a growth draft", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 71, status: "needs_input" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 71, status: "needs_input" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 71, status: "ready" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ draft: { id: 71 }, model: { id: 99 } }), { status: 200 }));

    await growthApi.createDraft("企业培训服务，需要测算收入和成本");
    await growthApi.getDraft(71);
    await growthApi.answerDraft(71, { acquisition_cost: 80, delivery_cost: 120000 });
    await growthApi.calculateDraft(71);

    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/growth/drafts", expect.objectContaining({
      method: "POST",
      body: JSON.stringify({ input: "企业培训服务，需要测算收入和成本" })
    }));
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/growth/drafts/71", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(3, "/api/v1/growth/drafts/71/answers", expect.objectContaining({
      method: "POST",
      body: JSON.stringify({ answers: { acquisition_cost: 80, delivery_cost: 120000 } })
    }));
    expect(fetchMock).toHaveBeenNthCalledWith(4, "/api/v1/growth/drafts/71/calculate", expect.objectContaining({
      method: "POST"
    }));
  });

  it("lists persisted model snapshots", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ snapshots: [{ id: 501, model_id: 99 }] }), { status: 200 })
    );

    await growthApi.listSnapshots(99);

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/growth/models/99/snapshots",
      expect.objectContaining({ method: "GET" })
    );
  });

  it("recalculates a model through the versioned endpoint", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ model: { id: 100 }, snapshot: { id: 501 } }), { status: 200 })
    );

    await growthApi.recalculateModel(99);

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/growth/models/99/recalculate",
      expect.objectContaining({ method: "POST" })
    );
  });

  it("lists growth history with server-side filters", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ models: [], total: 0, limit: 10, offset: 20 }), { status: 200 })
    );

    await growthApi.listModels({ q: "SaaS", limit: 10, offset: 20 });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/growth/models?q=SaaS&limit=10&offset=20",
      expect.objectContaining({ method: "GET" })
    );
  });
});
