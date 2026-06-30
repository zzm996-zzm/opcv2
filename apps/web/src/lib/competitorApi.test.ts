import { afterEach, describe, expect, it, vi } from "vitest";

import { competitorApi } from "./competitorApi";

describe("competitorApi", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("creates scans and gets monitoring snapshot", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 99, status: "completed" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ watchlist: [], events: [] }), { status: 200 }));

    await competitorApi.createScan({ targets: ["小鹅通"], focus: "价格变化" });
    await competitorApi.getMonitoring(10);

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "/api/v1/competitor/scans",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ targets: ["小鹅通"], focus: "价格变化" })
      })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/competitor/monitoring?limit=10", expect.objectContaining({ method: "GET" }));
  });
});
