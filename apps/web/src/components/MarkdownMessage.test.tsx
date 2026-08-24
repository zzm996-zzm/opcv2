import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import MarkdownMessage from "./MarkdownMessage";

describe("MarkdownMessage", () => {
  it("renders common AI Markdown as structured content", () => {
    render(
      <MarkdownMessage
        content={`## 项目建议\n\n- **先验证需求**\n- 记录结果\n\n| 项目 | 难度 |\n| --- | --- |\n| Photon Labs | 高 |`}
      />
    );

    expect(screen.getByRole("heading", { name: "项目建议" })).toBeInTheDocument();
    expect(screen.getByText("先验证需求")).toBeInTheDocument();
    expect(screen.getByRole("cell", { name: "Photon Labs" })).toBeInTheDocument();
  });

  it("does not render raw HTML", () => {
    const { container } = render(<MarkdownMessage content={'<script>alert("xss")</script>\n\n**安全文本**'} />);

    expect(container.querySelector("script")).toBeNull();
    expect(screen.getByText("安全文本")).toBeInTheDocument();
  });
});
