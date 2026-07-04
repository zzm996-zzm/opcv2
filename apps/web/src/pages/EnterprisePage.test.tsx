import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import App from "../App";
import { authSession } from "../lib/authSession";

describe("EnterprisePage", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  it("renders transparent empty states while enterprise APIs are not connected", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        stats: [],
        plans: [],
        delivery_board: [],
        milestones: [],
        cases: []
      }), { status: 200 })
    );
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-23T12:00:00Z",
      is_new_user: false,
      user: {
        id: 7,
        nickname: "张婧",
        phone: "",
        account: "zhangjing",
        status: "active"
      }
    });

    render(
      <MemoryRouter initialEntries={["/enterprise"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "企业定制化陪跑" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "预约企业诊断" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "陪跑方案" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "交付看板" })).toBeInTheDocument();
    expect(screen.getByText("企业陪跑后端接口未接入")).toBeInTheDocument();
    expect(screen.getByText("暂无陪跑方案")).toBeInTheDocument();
    expect(screen.getByText("暂无交付看板数据")).toBeInTheDocument();
    expect(screen.getByText("暂无陪跑里程碑")).toBeInTheDocument();
    expect(screen.getByText("暂无企业案例")).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "增长团队训练营" })).not.toBeInTheDocument();
    expect(screen.queryByText("连锁教育集团")).not.toBeInTheDocument();
    expect(screen.queryByText("服务企业")).not.toBeInTheDocument();
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/v1/enterprise/overview", expect.any(Object)));
    expect(screen.queryByText("第一版正在实现")).not.toBeInTheDocument();
  });

  it("renders enterprise overview records from the backend API", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        stats: [{ key: "companies", label: "服务企业数", value: "3" }],
        plans: [{
          id: 1,
          title: "后端陪跑方案",
          audience: "后端返回的适用对象",
          price_label: "待报价",
          focus: ["后端重点"],
          result: "后端返回的交付结果"
        }],
        delivery_board: [{ stage: "诊断中", count: 1, detail: "后端交付阶段" }],
        milestones: [{ time_label: "第1周", title: "后端里程碑", detail: "后端里程碑详情" }],
        cases: [{ id: 7, company: "后端企业案例", result: "后端案例结果" }]
      }), { status: 200 })
    );
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-23T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张婧", phone: "", account: "zhangjing", status: "active" }
    });

    render(
      <MemoryRouter initialEntries={["/enterprise"]}>
        <App />
      </MemoryRouter>
    );

    expect(await screen.findByRole("heading", { name: "后端陪跑方案" })).toBeInTheDocument();
    expect(screen.getByText("后端交付阶段")).toBeInTheDocument();
    expect(screen.getByText("后端里程碑")).toBeInTheDocument();
    expect(screen.getByText("后端企业案例")).toBeInTheDocument();
  });
});
