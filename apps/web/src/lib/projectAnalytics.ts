export type ProjectEventName =
  | "project_home_view"
  | "project_search"
  | "project_filter_apply"
  | "project_card_click"
  | "project_detail_view"
  | "project_tab_switch"
  | "project_lock_view"
  | "project_unlock_click"
  | "project_diagnose_submit"
  | "project_diagnose_result"
  | "project_banner_submit"
  | "match_file_upload"
  | "match_file_parse_result"
  | "match_start"
  | "match_answer"
  | "match_research"
  | "match_result"
  | "project_explore_view"
  | "project_explore_source_click"
  | "project_case_view"
  | "project_case_source_click"
  | "match_export_click"
  | "project_compare_add"
  | "project_compare_view";

type ProjectEventProperty = string | number | boolean | null | ProjectEventProperty[] | { [key: string]: ProjectEventProperty };

const visitorStorageKey = "opcv2:project-analytics-visitor";

function randomID(prefix: string) {
  const value = typeof crypto !== "undefined" && typeof crypto.randomUUID === "function"
    ? crypto.randomUUID()
    : `${Date.now()}-${Math.random().toString(36).slice(2)}`;
  return `${prefix}-${value}`;
}

function projectVisitorKey() {
  try {
    const stored = window.localStorage.getItem(visitorStorageKey);
    if (stored) return stored;
    const visitor = randomID("visitor");
    window.localStorage.setItem(visitorStorageKey, visitor);
    return visitor;
  } catch {
    return randomID("visitor");
  }
}

export function trackProjectEvent(
  eventName: ProjectEventName,
  properties: Record<string, ProjectEventProperty> = {},
  refModule = "project_market"
) {
  if (typeof window === "undefined" || typeof fetch !== "function") return;
  const payload = {
    event_id: randomID("event"),
    event_name: eventName,
    visitor_key: projectVisitorKey(),
    route: `${window.location.pathname}${window.location.search}`,
    ref_module: refModule,
    properties
  };
  void fetch("/api/v1/analytics/events", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
    keepalive: true
  }).catch(() => undefined);
}
