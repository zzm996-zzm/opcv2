import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import App from "../App";
import { authSession } from "../lib/authSession";

describe("EnterprisePage", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  it("renders public enterprise empty states without static claims", async () => {
    const fetchMock = mockEnterprisePublicLoad({
      overview: {
        headline: "",
        proof_points: [],
        stats: [],
        service_steps: []
      },
      cases: { cases: [] },
      contact: {}
    });

    renderEnterprisePage();

    expect(screen.getByRole("heading", { name: "企业定制化陪跑" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "预约企业咨询" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "服务路径" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "顾问联系方式" })).toBeInTheDocument();
    expect(await screen.findByText("暂无已发布企业服务介绍。")).toBeInTheDocument();
    expect(screen.getByText("暂无已发布服务路径")).toBeInTheDocument();
    expect(screen.getByText("暂无已发布顾问联系方式")).toBeInTheDocument();
    expect(screen.getByText("暂无已发布服务依据")).toBeInTheDocument();
    expect(screen.getByText("暂无已发布企业案例")).toBeInTheDocument();
    expect(screen.queryByText("服务企业")).not.toBeInTheDocument();
    expect(screen.queryByText("连锁教育集团")).not.toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/enterprise/public-overview", expect.any(Object));
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/enterprise/cases?limit=6", expect.any(Object));
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/enterprise/contact-config", expect.any(Object));
  });

  it("focuses the enterprise need input from the primary action", async () => {
    mockEnterprisePublicLoad({
      overview: { headline: "", proof_points: [], stats: [], service_steps: [] },
      cases: { cases: [] },
      contact: {}
    });

    renderEnterprisePage();

    fireEvent.click(screen.getByRole("button", { name: "预约企业咨询" }));

    expect(screen.getByLabelText("描述企业需求")).toHaveFocus();
  });

  it("renders published enterprise overview, cases, and contact config from APIs", async () => {
    mockEnterprisePublicLoad({
      overview: {
        headline: "企业AI落地陪跑",
        subheadline: "围绕获客、销售和运营搭建AI流程",
        description: "公开发布的企业服务说明",
        proof_points: [{ title: "来源可追溯", detail: "来自后台发布内容" }],
        stats: [{ key: "projects", label: "公开项目", value: "3", note: "后台统计" }],
        service_steps: [{ title: "业务诊断", detail: "拆解团队流程和目标" }],
        source_name: "企业服务后台",
        source_updated_at: "2026-07-13T08:00:00Z"
      },
      cases: {
        cases: [{
          id: 7,
          slug: "ai-sales",
          company: "启明星教育",
          title: "AI销售流程搭建",
          summary: "从线索收集到跟进复盘",
          result: "沉淀销售SOP",
          industry: "教育",
          services: ["CRM", "AI线索"],
          metrics: [],
          source_name: "案例后台"
        }]
      },
      contact: {
        consultant_name: "企业顾问",
        title: "AI增长顾问",
        description: "工作日联系",
        phone: "13800138000",
        email: "advisor@example.com",
        wechat: "ai-advisor",
        contact_url: "https://example.com/contact"
      }
    });

    renderEnterprisePage();

    expect(await screen.findByRole("heading", { name: "企业AI落地陪跑" })).toBeInTheDocument();
    expect(screen.getByText("公开发布的企业服务说明")).toBeInTheDocument();
    expect(screen.getByText("来源：企业服务后台 · 2026/7/13")).toBeInTheDocument();
    expect(screen.getByText("业务诊断")).toBeInTheDocument();
    expect(screen.getByText("来源可追溯")).toBeInTheDocument();
    expect(screen.getByText("AI销售流程搭建")).toBeInTheDocument();
    expect(screen.getByText("启明星教育 · 教育")).toBeInTheDocument();
    expect(screen.getByText("企业顾问")).toBeInTheDocument();
    expect(screen.getByText("ai-advisor")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "打开预约链接" })).toHaveAttribute("href", "https://example.com/contact");
  });

  it("submits an enterprise public inquiry", async () => {
    const fetchMock = mockEnterprisePublicLoad({
      overview: { headline: "", proof_points: [], stats: [], service_steps: [] },
      cases: { cases: [] },
      contact: {}
    });
    fetchMock.mockResolvedValueOnce(new Response(JSON.stringify({
      id: 11,
      company: "启明星教育",
      name: "张总",
      phone: "13800138000",
      need: "30人销售团队需要AI获客陪跑",
      status: "crm_synced",
      crm_customer_id: 300
    }), { status: 200 }));

    renderEnterprisePage();

    fireEvent.change(screen.getByLabelText("企业名称"), { target: { value: "启明星教育" } });
    fireEvent.change(screen.getByLabelText("联系人"), { target: { value: "张总" } });
    fireEvent.change(screen.getByLabelText("联系方式"), { target: { value: "13800138000" } });
    fireEvent.change(screen.getByLabelText("描述企业需求"), {
      target: { value: "30人销售团队需要AI获客陪跑" }
    });
    fireEvent.change(screen.getByLabelText("预算范围"), { target: { value: "5-10万" } });
    fireEvent.change(screen.getByLabelText("启动时间"), { target: { value: "本月" } });
    fireEvent.click(screen.getByRole("button", { name: "提交企业咨询" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/enterprise/inquiries", expect.objectContaining({
      method: "POST",
      body: JSON.stringify({
        company: "启明星教育",
        name: "张总",
        phone: "13800138000",
        need: "30人销售团队需要AI获客陪跑",
        budget: "5-10万",
        timeline: "本月",
        source_page: "/enterprise"
      })
    })));
    expect(await screen.findByText("咨询已提交，顾问将跟进联系")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "生成跟进任务" })).not.toBeInTheDocument();
  });
});

function renderEnterprisePage() {
  authSession.set({
    access_token: "access-token",
    access_token_expires_at: "2026-07-14T00:00:00Z",
    is_new_user: false,
    user: { id: 7, nickname: "张婧", phone: "", account: "zhangjing", status: "active" }
  });
  return render(
    <MemoryRouter initialEntries={["/enterprise"]}>
      <App />
    </MemoryRouter>
  );
}

function mockEnterprisePublicLoad(payload: { overview: unknown; cases: unknown; contact: unknown }) {
  return vi.spyOn(globalThis, "fetch")
    .mockResolvedValueOnce(new Response(JSON.stringify(payload.overview), { status: 200 }))
    .mockResolvedValueOnce(new Response(JSON.stringify(payload.cases), { status: 200 }))
    .mockResolvedValueOnce(new Response(JSON.stringify(payload.contact), { status: 200 }));
}
