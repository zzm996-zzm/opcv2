import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "./lib/authSession";
import App from "./App";

describe("App", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("renders the OPC product shell", () => {
    render(
      <MemoryRouter>
        <App />
      </MemoryRouter>
    );

    expect(
      screen.getByRole("heading", { name: "把商业想法，变成下一步行动" })
    ).toBeInTheDocument();
    expect(screen.getByRole("navigation", { name: "主导航" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "免费分析" })).toHaveAttribute(
      "href",
      "/analysis"
    );
  });

  it("redirects a signed-out user from a protected product route", async () => {
    authSession.finishRestore();

    render(
      <MemoryRouter initialEntries={["/leads"]}>
        <App />
      </MemoryRouter>
    );

    expect(
      await screen.findByRole("heading", { name: "登录或创建账号" })
    ).toBeInTheDocument();
  });

  it("renders a product route for a signed-in user", () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-11T12:00:00Z",
      is_new_user: false,
      user: {
        id: 7,
        nickname: "张晨",
        phone: "13800138000",
        status: "active"
      }
    });

    render(
      <MemoryRouter initialEntries={["/leads"]}>
        <App />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "VIP获客" })).toBeInTheDocument();
  });
});
