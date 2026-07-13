import type { MembershipUsageItem } from "./membershipApi";

export const quotaKeys = {
  leadTasks: "lead_tasks",
  sandboxRuns: "sandbox_runs",
  competitorScans: "competitor_scans",
  copilotMessages: "copilot_messages",
  copilotCompareCalls: "copilot_compare_calls"
} as const;

export function quotaSummary(usage: MembershipUsageItem[], key: string, fallbackLabel: string) {
  const item = usage.find((entry) => entry.key === key);
  const remaining = item ? Math.max(item.limit - item.used, 0) : null;
  return {
    item,
    label: item?.label ?? fallbackLabel,
    remaining,
    blocked: item ? item.limit <= 0 || item.used >= item.limit : false,
    value: item ? `${remaining}/${item.limit}` : "读取中",
    unit: item?.unit ?? "次/月"
  };
}
