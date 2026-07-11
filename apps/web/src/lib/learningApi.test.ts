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
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: 99, status: "completed" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ diagnosis_id: 99, gaps: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ diagnosis_id: 99, focus: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ diagnosis_id: 99, stages: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ diagnosis_id: 99, priority_gaps: [] }), { status: 200 }));

    await learningApi.listCourses({ category: "实战", limit: 12 });
    await learningApi.getCourse("ai-basics");
    await learningApi.listProgress();
    await learningApi.createDiagnosis({
      goal: "提升AI能力",
      project: "智能客服",
      focus_abilities: ["数据洞察能力"],
      weekly_time: "5-8 小时",
      bottleneck: "缺少案例"
    });
    await learningApi.getLatestDiagnosis();
    await learningApi.getLatestGaps();
    await learningApi.getLatestRecommendations();
    await learningApi.getLatestPlan();
    await learningApi.getLatestReport();

    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/learning/courses?category=%E5%AE%9E%E6%88%98&limit=12", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/learning/courses/ai-basics", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(3, "/api/v1/learning/progress", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(
      4,
      "/api/v1/learning/diagnoses",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({
          goal: "提升AI能力",
          project: "智能客服",
          focus_abilities: ["数据洞察能力"],
          weekly_time: "5-8 小时",
          bottleneck: "缺少案例"
        })
      })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(5, "/api/v1/learning/diagnoses/latest", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(6, "/api/v1/learning/diagnoses/latest/gaps", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(7, "/api/v1/learning/diagnoses/latest/recommendations", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(8, "/api/v1/learning/diagnoses/latest/plan", expect.objectContaining({ method: "GET" }));
    expect(fetchMock).toHaveBeenNthCalledWith(9, "/api/v1/learning/diagnoses/latest/report", expect.objectContaining({ method: "GET" }));
  });
});
