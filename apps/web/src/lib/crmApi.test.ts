import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "./authSession";
import { crmApi } from "./crmApi";

describe("crmApi", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  it("lists due customers with bearer token", async () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-25T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张晨", phone: "", status: "active" }
    });
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ customers: [] }), { status: 200 })
    );

    await crmApi.listDueCustomers(10);

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/crm/customers/due?limit=10",
      expect.objectContaining({
        method: "GET",
        headers: expect.objectContaining({ Authorization: "Bearer access-token" })
      })
    );
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
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 1 }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ subject: "跟进方案", body: "您好", channel: "wechat" }), { status: 200 }));

    await crmApi.recordFollowUp(100, { note: "已发资料", nextFollowUpAt: "2026-06-26T10:00:00Z" });
    await crmApi.generateFollowUpCopy(100, { goal: "推进方案会" });

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "/api/v1/crm/customers/100/follow-ups",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ note: "已发资料", next_follow_up_at: "2026-06-26T10:00:00Z" })
      })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "/api/v1/crm/customers/100/follow-up-copy",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ goal: "推进方案会" })
      })
    );
  });
});
