import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "./authSession";
import { crmApi } from "./crmApi";

describe("crmApi", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  it("lists customers with bearer token and filters", async () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-25T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张晨", phone: "", status: "active" }
    });
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ customers: [] }), { status: 200 })
    );

    await crmApi.listCustomers({ stage: "contacted", source: "enterprise", q: "启明星", limit: 10 });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/crm/customers?stage=contacted&source=enterprise&q=%E5%90%AF%E6%98%8E%E6%98%9F&limit=10",
      expect.objectContaining({
        method: "GET",
        headers: expect.objectContaining({ Authorization: "Bearer access-token" })
      })
    );
  });

  it("gets a CRM customer and due customers", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 100, name: "成都启明星教育" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 100, name: "成都启明星教育", phone: "028-12345678" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ activities: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ customers: [] }), { status: 200 }));

    await crmApi.getCustomer(100);
    await crmApi.updateCustomer(100, { name: "成都启明星教育", phone: "028-12345678" });
    await crmApi.listActivities(100, 10);
    await crmApi.listDueCustomers(10);

    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/crm/customers/100", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "/api/v1/crm/customers/100",
      expect.objectContaining({
        method: "PATCH",
        body: JSON.stringify({ name: "成都启明星教育", phone: "028-12345678" })
      })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(3, "/api/v1/crm/customers/100/activities?limit=10", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(4, "/api/v1/crm/customers/due?limit=10", expect.objectContaining({ method: "GET" }));
  });

  it("imports lead into CRM", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ id: 100, user_id: 7, name: "成都启明星教育" }), { status: 200 })
    );

    await crmApi.importLead({ leadResultId: 99, name: "成都启明星教育" });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/crm/customers/import-lead",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({
          lead_result_id: 99,
          name: "成都启明星教育"
        })
      })
    );
  });

  it("updates customer stage", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ id: 100, stage: "contacted" }), { status: 200 })
    );

    await crmApi.updateStage(100, { stage: "contacted", note: "电话已接通" });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/crm/customers/100/stage",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ stage: "contacted", note: "电话已接通" })
      })
    );
  });

  it("records follow-up and generates copy", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ follow_ups: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 1 }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ subject: "跟进方案", body: "您好", channel: "wechat" }), { status: 200 }));

    await crmApi.listFollowUps({ customerId: 100, q: "方案", due: "week", limit: 10 });
    await crmApi.recordFollowUp(100, { note: "已发资料", nextFollowUpAt: "2026-06-26T10:00:00Z" });
    await crmApi.generateFollowUpCopy(100, { goal: "推进方案会" });

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "/api/v1/crm/follow-ups?customer_id=100&q=%E6%96%B9%E6%A1%88&due=week&limit=10",
      expect.objectContaining({ method: "GET" })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "/api/v1/crm/customers/100/follow-ups",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ note: "已发资料", next_follow_up_at: "2026-06-26T10:00:00Z" })
      })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      3,
      "/api/v1/crm/customers/100/follow-up-copy",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ goal: "推进方案会" })
      })
    );
  });

  it("loads CRM pipeline stats", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ total: 3, new: 1, won: 1, due_today: 1 }), { status: 200 })
    );

    await crmApi.pipelineStats();

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/crm/pipeline-stats",
      expect.objectContaining({ method: "GET" })
    );
  });
});
