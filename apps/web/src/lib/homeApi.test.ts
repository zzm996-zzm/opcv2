import { afterEach, describe, expect, it, vi } from "vitest";

import { homeApi } from "./homeApi";

describe("homeApi", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("loads home summary", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ metrics: [], hero_cards: [], recommendations: [], recent_tasks: [], notification_summary: { unread: 0, latest: [] }, account_summary: { plan_name: "会员版", quota_warnings: [] } }), { status: 200 })
    );

    const summary = await homeApi.summary();

    expect(summary.account_summary.plan_name).toBe("会员版");
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/home/summary", expect.any(Object));
  });
});
