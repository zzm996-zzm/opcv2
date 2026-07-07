import { afterEach, describe, expect, it, vi } from "vitest";

import { competitorApi } from "./competitorApi";

describe("competitorApi", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("creates scans and gets monitoring snapshot", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 99, status: "queued", progress_percent: 0, current_step: "queued" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ watchlist: [], events: [] }), { status: 200 }));

    const scan = await competitorApi.createScan({ targets: ["小鹅通"], focus: "价格变化" });
    await competitorApi.getMonitoring(10);

    expect(scan.status).toBe("queued");
    expect(scan.progress_percent).toBe(0);
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

  it("retries competitor scans", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 99, status: "queued", progress_percent: 0, current_step: "queued" }), { status: 200 }));

    const scan = await competitorApi.retryScan(99);

    expect(scan.status).toBe("queued");
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/competitor/scans/99/retry",
      expect.objectContaining({ method: "POST" })
    );
  });

  it("creates competitor monitoring watch items", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ name: "增长雷达", status: "监测中", channels: ["价格页"] }), { status: 200 }));

    const item = await competitorApi.createWatchItem({ name: "增长雷达", category: "商业情报", channels: ["价格页"] });

    expect(item.name).toBe("增长雷达");
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/competitor/monitoring/watchlist",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ name: "增长雷达", category: "商业情报", channels: ["价格页"] })
      })
    );
  });

  it("deletes competitor monitoring watch items", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ deleted: true }), { status: 200 }));

    await competitorApi.deleteWatchItem(77);

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/competitor/monitoring/watchlist/77",
      expect.objectContaining({ method: "DELETE" })
    );
  });
});
