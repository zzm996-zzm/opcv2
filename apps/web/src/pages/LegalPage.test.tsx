import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";

import LegalPage from "./LegalPage";

describe("LegalPage", () => {
  it("renders the privacy notice", () => {
    render(
      <MemoryRouter>
        <LegalPage kind="privacy" />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "隐私政策" })).toBeInTheDocument();
    expect(screen.getByText(/手机号码仅用于账号登录/)).toBeInTheDocument();
  });
});
