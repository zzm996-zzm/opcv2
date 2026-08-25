import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";

import ReferenceShell from "./ReferenceShell";
import V4PageShell from "./V4PageShell";

describe("V4PageShell", () => {
  it.each(["/tasks", "/crm", "/copilot"])("uses the canonical navigation on %s", (path) => {
    render(
      <MemoryRouter initialEntries={[path]}>
        <V4PageShell><p>页面内容</p></V4PageShell>
      </MemoryRouter>
    );

    expect(screen.getByRole("button", { name: "项目确定及拆解" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "项目超市" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "我的项目" })).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: "竞品全盘数据破解" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "增长洞察" })).toHaveAttribute("href", "/growth-calculator");
    expect(screen.getByRole("link", { name: "CRM客户管理" })).toBeInTheDocument();
    expect(screen.queryByText("项目管理及系统")).not.toBeInTheDocument();
    expect(screen.queryByText("增长引擎")).not.toBeInTheDocument();
  });

  it("highlights My Projects separately from the project marketplace", () => {
    render(
      <MemoryRouter initialEntries={["/projects/mine"]}>
        <V4PageShell><p>我的项目页面</p></V4PageShell>
      </MemoryRouter>
    );

    expect(screen.queryByRole("link", { name: "我的项目" })).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: "项目超市" })).not.toHaveClass("active");
  });

  it("uses V4PageShell as the ReferenceShell chrome", () => {
    const { container } = render(
      <MemoryRouter>
        <ReferenceShell accountSlot={<span>自定义账号区</span>} className="reference-adapter" mainClassName="reference-main">
          <p>参考页内容</p>
        </ReferenceShell>
      </MemoryRouter>
    );

    expect(container.querySelector(".v4-shell.reference-adapter")).toBeInTheDocument();
    expect(container.querySelector(".v4-page-main.reference-main")).toBeInTheDocument();
    expect(screen.getByText("自定义账号区")).toBeInTheDocument();
    expect(screen.getByRole("navigation", { name: "顶部全局功能区" })).toBeInTheDocument();
  });
});
