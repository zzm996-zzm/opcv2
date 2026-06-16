import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";

import RegisterDetailsPage from "./RegisterDetailsPage";

describe("RegisterDetailsPage", () => {
  it("renders the detailed registration profile form", () => {
    render(
      <MemoryRouter>
        <RegisterDetailsPage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: /完善企业资料/ })).toBeInTheDocument();
    expect(screen.getByText("基本身份")).toBeInTheDocument();
    expect(screen.getByText("我的业务 / 公司")).toBeInTheDocument();
    expect(screen.getByText("我的产品")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /保存并进入工作台/ })).toBeInTheDocument();
  });
});
