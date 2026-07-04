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

  it("lists tools with filters and reads tool detail", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ tools: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ slug: "canva-ai", name: "Canva AI" }), { status: 200 }));

    await contentApi.listTools({ category: "创业获客", q: "canva", sort: "hot", limit: 10 });
    await contentApi.getTool("canva-ai");

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "/api/v1/content/tools?category=%E5%88%9B%E4%B8%9A%E8%8E%B7%E5%AE%A2&q=canva&sort=hot&limit=10",
      expect.any(Object)
    );
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/content/tools/canva-ai", expect.any(Object));
  });

  it("favorites tools, joins community and reads help articles", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ slug: "canva-ai", favorited: true }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ slug: "canva-ai", favorited: false }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 9, status: "submitted" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ topics: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ articles: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ slug: "login-help", title: "如何登录账号" }), { status: 200 }));

    await contentApi.favoriteTool("canva-ai");
    await contentApi.unfavoriteTool("canva-ai");
    await contentApi.joinCommunity({ community: "members", contact: "13800138000" });
    await contentApi.listHelpTopics();
    await contentApi.listHelpArticles({ topic: "account", q: "登录", limit: 10 });
    await contentApi.getHelpArticle("login-help");

    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/content/tools/canva-ai/favorite", expect.objectContaining({ method: "POST" }));
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/content/tools/canva-ai/favorite", expect.objectContaining({ method: "DELETE" }));
    expect(fetchMock).toHaveBeenNthCalledWith(3, "/api/v1/community/join-requests", expect.objectContaining({ method: "POST" }));
    expect(fetchMock).toHaveBeenNthCalledWith(4, "/api/v1/help/topics", expect.any(Object));
    expect(fetchMock).toHaveBeenNthCalledWith(
      5,
      "/api/v1/help/articles?topic=account&q=%E7%99%BB%E5%BD%95&limit=10",
      expect.any(Object)
    );
    expect(fetchMock).toHaveBeenNthCalledWith(6, "/api/v1/help/articles/login-help", expect.any(Object));
  });

  it("reads public articles, community config and brand content", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ articles: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ slug: "ai-customer-service", bookmarked: true }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ slug: "ai-customer-service", bookmarked: false }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 1, headline: "加入社群" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ metrics: [], cases: [] }), { status: 200 }));

    await contentApi.listArticles();
    await contentApi.bookmarkArticle("ai-customer-service");
    await contentApi.unbookmarkArticle("ai-customer-service");
    await contentApi.getCommunityConfig();
    await contentApi.getBrand();

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "/api/v1/content/articles",
      expect.any(Object)
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "/api/v1/content/articles/ai-customer-service/bookmark",
      expect.objectContaining({ method: "POST" })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      3,
      "/api/v1/content/articles/ai-customer-service/bookmark",
      expect.objectContaining({ method: "DELETE" })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      4,
      "/api/v1/content/community",
      expect.any(Object)
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      5,
      "/api/v1/content/brand",
      expect.any(Object)
    );
  });
});
