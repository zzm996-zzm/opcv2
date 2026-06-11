import { describe, expect, it, vi } from "vitest";

import { authSession } from "./authSession";

const loginResult = {
  access_token: "access-token",
  access_token_expires_at: "2026-06-11T12:00:00Z",
  is_new_user: true,
  user: {
    id: 7,
    nickname: "张晨",
    phone: "13800138000",
    status: "active"
  }
};

describe("authSession", () => {
  it("notifies subscribers when the session changes", () => {
    const listener = vi.fn();
    const unsubscribe = authSession.subscribe(listener);

    authSession.set(loginResult);
    expect(listener).toHaveBeenCalledTimes(1);
    expect(authSession.get().user?.nickname).toBe("张晨");

    authSession.clear();
    expect(listener).toHaveBeenCalledTimes(2);
    expect(authSession.get().user).toBeNull();

    unsubscribe();
  });
});
