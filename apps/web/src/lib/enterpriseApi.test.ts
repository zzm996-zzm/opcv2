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
});
