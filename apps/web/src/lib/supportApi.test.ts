import { afterEach, describe, expect, it, vi } from "vitest";

import { supportApi } from "./supportApi";

describe("supportApi", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("creates and lists tickets", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 22, status: "open" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ tickets: [] }), { status: 200 }));

    await supportApi.createTicket({ topic: "套餐与额度", title: "额度没有更新", body: "正文" });
    await supportApi.listTickets(10);

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "/api/v1/support/tickets",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ topic: "套餐与额度", title: "额度没有更新", body: "正文" })
      })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/support/tickets?limit=10", expect.any(Object));
  });
});
