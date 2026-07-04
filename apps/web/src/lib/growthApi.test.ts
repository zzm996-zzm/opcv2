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
});
