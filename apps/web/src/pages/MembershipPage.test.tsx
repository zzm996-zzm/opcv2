import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import MembershipPage from "./MembershipPage";

describe("MembershipPage", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("shows current plan and redeems a code", async () => {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-11T12:00:00Z",
      is_new_user: false,
      user: { id: 7, nickname: "张晨", phone: "13800138000", status: "active" }
    });
    const fetchMock = vi
      .spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            plan: {
              code: "free",
              name: "免费版",
              monthly_analysis_limit: 3,
              lead_export_limit: 0
            },
            credit_balance: 0
          }),
          { status: 200 }
        )
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            snapshot: {
              plan: {
                code: "pro",
                name: "专业版",
                monthly_analysis_limit: 100,
                lead_export_limit: 1000
              },
              credit_balance: 100
            },
            already_redeemed: false
          }),
          { status: 200 }
        )
      );

    render(
      <MemoryRouter>
        <MembershipPage />
      </MemoryRouter>
    );

    expect(await screen.findByText("免费版")).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("兑换码"), { target: { value: "PRO100" } });
    fireEvent.click(screen.getByRole("button", { name: "兑换" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
    expect(await screen.findByText("专业版")).toBeInTheDocument();
    expect(screen.getByText("100")).toBeInTheDocument();
  });
});
