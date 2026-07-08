import { describe, expect, it } from "vitest";

import { quotaSummary } from "./quotaUsage";

describe("quotaSummary", () => {
  it("computes remaining quota and depleted state", () => {
    expect(quotaSummary([{ key: "lead_tasks", label: "AI线索任务", used: 8, limit: 30, unit: "次/月" }], "lead_tasks", "线索任务")).toMatchObject({
      label: "AI线索任务",
      remaining: 22,
      blocked: false,
      value: "22/30"
    });

    expect(quotaSummary([{ key: "lead_tasks", label: "AI线索任务", used: 30, limit: 30, unit: "次/月" }], "lead_tasks", "线索任务")).toMatchObject({
      remaining: 0,
      blocked: true,
      value: "0/30"
    });
  });

  it("keeps unknown quota non-blocking while loading", () => {
    expect(quotaSummary([], "sandbox_runs", "商业沙盘")).toMatchObject({
      label: "商业沙盘",
      remaining: null,
      blocked: false,
      value: "读取中"
    });
  });
});
