import { afterEach, describe, expect, it, vi } from "vitest";

import { notificationsApi } from "./notificationsApi";

describe("notificationsApi", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("lists notifications with filters and reads detail", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ notifications: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 7, title: "任务提醒" }), { status: 200 }));

    await notificationsApi.list({ type: "task", status: "unread", limit: 10 });
    await notificationsApi.get(7);

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "/api/v1/notifications?type=task&status=unread&limit=10",
      expect.any(Object)
    );
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/notifications/7", expect.any(Object));
  });

  it("marks read, marks all read, and reads summary", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 7, read_at: "2026-07-02T10:00:00Z" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ updated: 12 }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ unread: 3, by_type: [], latest: [] }), { status: 200 }));

    await notificationsApi.markRead(7);
    await notificationsApi.markAllRead();
    await notificationsApi.summary();

    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/notifications/7/read", expect.objectContaining({ method: "PATCH" }));
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/notifications/read-all", expect.objectContaining({ method: "POST" }));
    expect(fetchMock).toHaveBeenNthCalledWith(3, "/api/v1/notifications/summary", expect.any(Object));
  });
});
