import { beforeEach, describe, expect, it, vi } from "vitest";

import { trackProjectEvent } from "./projectAnalytics";

describe("projectAnalytics", () => {
  beforeEach(() => {
    window.localStorage.clear();
    window.history.replaceState({}, "", "/projects/42?section=path");
    vi.restoreAllMocks();
  });

  it("persists an anonymous visitor and creates idempotent event IDs", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(null, { status: 202 }));
    trackProjectEvent("project_detail_view", { project_id: 42, tab: "path" });
    trackProjectEvent("project_tab_switch", { project_id: 42, from: "path", to: "data" });

    expect(fetchMock).toHaveBeenCalledTimes(2);
    const first = JSON.parse(String(fetchMock.mock.calls[0]?.[1]?.body));
    const second = JSON.parse(String(fetchMock.mock.calls[1]?.[1]?.body));
    expect(first.visitor_key).toMatch(/^visitor-/);
    expect(second.visitor_key).toBe(first.visitor_key);
    expect(second.event_id).not.toBe(first.event_id);
    expect(first.route).toBe("/projects/42?section=path");
    expect(fetchMock.mock.calls[0]?.[1]).toEqual(expect.objectContaining({ method: "POST", keepalive: true }));
    expect(first).not.toHaveProperty("user_id");
    expect(first).not.toHaveProperty("heat");
  });

  it("never rejects the product interaction when ingestion is unavailable", () => {
    vi.spyOn(globalThis, "fetch").mockRejectedValue(new Error("offline"));
    expect(() => trackProjectEvent("project_search", { keyword: "AI", result_count: 0 })).not.toThrow();
  });
});
