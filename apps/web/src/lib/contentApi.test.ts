import { afterEach, describe, expect, it, vi } from "vitest";

import { contentApi } from "./contentApi";

describe("contentApi", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("lists public tools", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ tools: [] }), { status: 200 })
    );

    await contentApi.listTools();

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/content/tools",
      expect.objectContaining({ credentials: "include" })
    );
  });

  it("reads public articles, community config and brand content", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ articles: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 1, headline: "加入社群" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ metrics: [], cases: [] }), { status: 200 }));

    await contentApi.listArticles();
    await contentApi.getCommunityConfig();
    await contentApi.getBrand();

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "/api/v1/content/articles",
      expect.any(Object)
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "/api/v1/content/community",
      expect.any(Object)
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      3,
      "/api/v1/content/brand",
      expect.any(Object)
    );
  });
});
