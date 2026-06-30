import { afterEach, describe, expect, it, vi } from "vitest";

import { dashboardApi } from "./dashboardApi";

describe("dashboardApi", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("gets dashboard summary", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ metrics: [], projects: [], trend: [], pipeline: [], alerts: [], actions: [] }), { status: 200 })
    );

    await dashboardApi.getSummary();

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/dashboard/summary",
      expect.objectContaining({ method: "GET" })
    );
  });
});
