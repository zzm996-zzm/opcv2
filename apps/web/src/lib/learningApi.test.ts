import { afterEach, describe, expect, it, vi } from "vitest";

import { authSession } from "./authSession";
import { learningApi } from "./learningApi";

describe("learningApi", () => {
  afterEach(() => {
    authSession.clear();
    vi.restoreAllMocks();
  });

  it("loads courses, progress, and creates diagnoses", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(JSON.stringify({ courses: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ slug: "ai-basics", title: "AI基础入门" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ progress: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 99, status: "completed" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 99, status: "completed" }), { status: 200 }));

    await learningApi.listCourses({ category: "实战", limit: 12 });
    await learningApi.getCourse("ai-basics");
    await learningApi.listProgress();
    await learningApi.createDiagnosis({ goal: "提升AI能力", project: "智能客服" });
    await learningApi.getLatestDiagnosis();

    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/learning/courses?category=%E5%AE%9E%E6%88%98&limit=12", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/learning/courses/ai-basics", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(3, "/api/v1/learning/progress", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(
      4,
      "/api/v1/learning/diagnoses",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ goal: "提升AI能力", project: "智能客服" })
      })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(5, "/api/v1/learning/diagnoses/latest", expect.objectContaining({ method: "GET" }));
  });
});
