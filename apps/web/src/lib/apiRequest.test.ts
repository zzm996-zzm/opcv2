import { afterEach, describe, expect, it, vi } from "vitest";

import { ApiRequestError, apiRequest } from "./apiRequest";
import { authSession } from "./authSession";

describe("apiRequest", () => {
  afterEach(() => {
    authSession.clear();
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("uses relative api paths by default", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ ok: true }), { status: 200 })
    );

    await apiRequest<{ ok: boolean }>("/api/v1/tasks");

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/tasks",
      expect.objectContaining({ credentials: "include" })
    );
  });

  it("prefixes api paths with VITE_API_BASE_URL when configured", async () => {
    vi.stubEnv("VITE_API_BASE_URL", "https://api.example.com");
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ ok: true }), { status: 200 })
    );

    await apiRequest<{ ok: boolean }>("/api/v1/tasks");

    expect(fetchMock).toHaveBeenCalledWith(
      "https://api.example.com/api/v1/tasks",
      expect.any(Object)
    );
  });

  it("does not duplicate slashes when api base url ends with slash", async () => {
    vi.stubEnv("VITE_API_BASE_URL", "https://api.example.com/");
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ ok: true }), { status: 200 })
    );

    await apiRequest<{ ok: boolean }>("/api/v1/tasks");

    expect(fetchMock).toHaveBeenCalledWith(
      "https://api.example.com/api/v1/tasks",
      expect.any(Object)
    );
  });

  it("returns undefined for 204 responses", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(null, { status: 204 })
    );

    await expect(apiRequest<void>("/api/v1/auth/logout")).resolves.toBeUndefined();
  });

  it("maps backend error codes to user-facing messages and keeps the code", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ error: "invalid_request" }), { status: 400 })
    );

    await expect(apiRequest("/api/v1/tasks")).rejects.toMatchObject({
      code: "invalid_request",
      message: "请求参数有误，请检查后重试"
    });
    await expect(apiRequest("/api/v1/tasks")).rejects.toBeInstanceOf(ApiRequestError);
  });
});
