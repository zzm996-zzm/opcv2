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

  it("loads public enterprise overview", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        headline: "企业AI落地陪跑",
        proof_points: [],
        stats: [],
        service_steps: []
      }), { status: 200 })
    );

    const overview = await enterpriseApi.publicOverview();

    expect(overview.headline).toBe("企业AI落地陪跑");
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/enterprise/public-overview", expect.any(Object));
  });

  it("loads public enterprise cases", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ cases: [{ id: 7, slug: "ai-sales", company: "启明星教育", title: "AI销售流程搭建", services: [], metrics: [] }] }), { status: 200 })
    );

    const response = await enterpriseApi.publicCases(6);

    expect(response.cases[0].slug).toBe("ai-sales");
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/enterprise/cases?limit=6", expect.any(Object));
  });

  it("loads a public enterprise case detail", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ id: 7, slug: "ai-sales", company: "启明星教育", title: "AI销售流程搭建", services: [], metrics: [] }), { status: 200 })
    );

    const item = await enterpriseApi.publicCase("ai sales");

    expect(item.slug).toBe("ai-sales");
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/enterprise/cases/ai%20sales", expect.any(Object));
  });

  it("loads enterprise contact config", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ consultant_name: "企业顾问", wechat: "ai-advisor" }), { status: 200 })
    );

    const config = await enterpriseApi.contactConfig();

    expect(config.wechat).toBe("ai-advisor");
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/enterprise/contact-config", expect.any(Object));
  });

  it("creates an enterprise public inquiry", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ id: 11, name: "张总", need: "AI销售陪跑", status: "crm_synced", crm_customer_id: 300 }), { status: 200 })
    );

    const inquiry = await enterpriseApi.createInquiry({ name: "张总", phone: "13800138000", need: "AI销售陪跑" });

    expect(inquiry.crm_customer_id).toBe(300);
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/enterprise/inquiries", expect.objectContaining({
      method: "POST",
      body: JSON.stringify({ name: "张总", phone: "13800138000", need: "AI销售陪跑" })
    }));
  });

  it("creates an enterprise diagnosis request", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ id: 7, user_id: 42, need: "30人销售团队需要AI获客陪跑", status: "submitted" }), { status: 200 })
    );

    const request = await enterpriseApi.createDiagnosisRequest({ need: "30人销售团队需要AI获客陪跑" });

    expect(request.status).toBe("submitted");
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/enterprise/diagnosis-requests", expect.objectContaining({
      method: "POST",
      body: JSON.stringify({ need: "30人销售团队需要AI获客陪跑" })
    }));
  });

  it("lists enterprise diagnosis requests", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ requests: [{ id: 7, user_id: 42, need: "30人销售团队需要AI获客陪跑", status: "submitted" }] }), { status: 200 })
    );

    const response = await enterpriseApi.listDiagnosisRequests(5);

    expect(response.requests).toHaveLength(1);
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/enterprise/diagnosis-requests?limit=5", expect.any(Object));
  });

  it("updates an enterprise diagnosis request", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ id: 7, user_id: 42, need: "30人销售团队需要AI获客陪跑", status: "follow_up_created" }), { status: 200 })
    );

    const request = await enterpriseApi.updateDiagnosisRequest(7, { status: "follow_up_created" });

    expect(request.status).toBe("follow_up_created");
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/enterprise/diagnosis-requests/7", expect.objectContaining({
      method: "PATCH",
      body: JSON.stringify({ status: "follow_up_created" })
    }));
  });

  it("completes an enterprise diagnosis request", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ id: 7, user_id: 42, need: "30人销售团队需要AI获客陪跑", status: "completed" }), { status: 200 })
    );

    const request = await enterpriseApi.updateDiagnosisRequest(7, { status: "completed" });

    expect(request.status).toBe("completed");
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/enterprise/diagnosis-requests/7", expect.objectContaining({
      method: "PATCH",
      body: JSON.stringify({ status: "completed" })
    }));
  });

  it("imports a completed enterprise diagnosis request into CRM", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        id: 100,
        user_id: 42,
        import_key: "enterprise_diagnosis_request:7",
        name: "30人销售团队需要AI获客陪跑",
        stage: "won",
        source: "enterprise",
        created_at: "2026-07-07T13:30:00Z",
        updated_at: "2026-07-07T13:30:00Z"
      }), { status: 200 })
    );

    const customer = await enterpriseApi.importDiagnosisRequestCustomer(7);

    expect(customer.source).toBe("enterprise");
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/enterprise/diagnosis-requests/7/crm-customer", expect.objectContaining({
      method: "POST"
    }));
  });
});
